package redirector_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"testing"

	"github.com/CruGlobal/redirector/internal/app"
	"github.com/CruGlobal/redirector/internal/redirector"
	"github.com/CruGlobal/redirector/redirtest"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/caddyserver/caddy/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestRedirector_NewRedirector(t *testing.T) {
	r := redirector.NewRedirector()
	assert.NotNil(t, r)
	assert.IsType(t, &redirector.Redirector{}, r)
	assert.Equal(t, app.DefaultTable, r.Table)
	assert.Equal(t, app.DefaultKey, r.Key)
}

func TestRedirector_CaddyModule(t *testing.T) {
	module := redirector.Redirector{}.CaddyModule()
	assert.IsType(t, caddy.ModuleInfo{}, module)
	assert.Equal(t, caddy.ModuleID("http.handlers.redirector"), module.ID)
	assert.IsType(t, &redirector.Redirector{}, module.New())
}

func TestRedirector_Provision(t *testing.T) {
	ctx := redirtest.NewRedirectorCaddyContext(t)

	r := redirector.NewRedirector()
	err := r.Provision(ctx)
	require.NoError(t, err)

	assert.NotNil(t, r.Client)
	assert.Equal(t, os.Getenv("DYNAMODB_TESTING_TABLE"), r.Table)
	assert.Equal(t, os.Getenv("DYNAMODB_TESTING_KEY"), r.Key)
}

type RedirectorTestSuite struct {
	suite.Suite

	redirector *redirector.Redirector
}

func (ts *RedirectorTestSuite) SetupSuite() {
	ctx := redirtest.NewRedirectorCaddyContext(ts.T())

	r := redirector.NewRedirector()
	err := r.Provision(ctx)
	ts.Require().NoError(err)

	ts.redirector = r
}

//nolint:gochecknoglobals // redirects are global to the test suite
var redirects = []redirector.Redirect{
	{
		Hostname: "www.example.com",
		Location: "example.com",
	},
	{
		Hostname: "example.org",
		Type:     redirector.TypeRedirect,
		Status:   redirector.StatusPermanent,
		Location: "www.example.org",
	},
	{
		Hostname: "www.example.info",
		Type:     redirector.TypeRedirect,
		Status:   redirector.StatusTemporary,
		Location: "example.info",
		Rewrites: []redirector.Rewrite{
			{
				RegExp:  redirector.RewriteRegexp{Regexp: regexp.MustCompile(`^(.*)$`)},
				Replace: "$1",
				Final:   true,
			},
		},
	},
}

// SetupTest creates the table before each test.
func (ts *RedirectorTestSuite) SetupTest() {
	redirtest.CreateDynamoDBTable(ts.T(), ts.redirector.Client, ts.redirector.Table, ts.redirector.Key)

	for _, r := range redirects {
		item, err := attributevalue.MarshalMap(r)
		ts.Require().NoError(err)
		_, err = ts.redirector.Client.PutItem(ts.T().Context(), &dynamodb.PutItemInput{
			TableName: aws.String(ts.redirector.Table),
			Item:      item,
		})
		ts.Require().NoError(err)
	}
}

// TearDownTest deletes the table after each test.
func (ts *RedirectorTestSuite) TearDownTest() {
	redirtest.DeleteDynamoDBTable(ts.T(), ts.redirector.Client, ts.redirector.Table)
}

// TestRedirectorTestSuite runs the test suite.
func TestRedirectorTestSuite(t *testing.T) {
	suite.Run(t, new(RedirectorTestSuite))
}

func (ts *RedirectorTestSuite) TestRedirector_GetRedirect() {
	tests := []struct {
		name      string
		hostname  string
		expect    redirector.Redirect
		expectErr bool
	}{
		{
			name:     "temporary redirect",
			hostname: "www.example.com",
			expect:   redirects[0],
		},
		{
			name:     "permanent redirect",
			hostname: "example.org",
			expect:   redirects[1],
		},
		{
			name:      "missing redirect",
			hostname:  "example.edu",
			expectErr: true,
		},
		{
			name:     "redirect with rewrites",
			hostname: "www.example.info",
			expect:   redirects[2],
		},
	}
	for _, tc := range tests {
		ts.Run(tc.name, func() {
			r, err := ts.redirector.GetRedirect(ts.T().Context(), tc.hostname, true)
			if tc.expectErr {
				ts.Require().Error(err)
				return
			}
			ts.Require().NoError(err)
			ts.Equal(tc.expect, *r)
		})
	}
}

type MockCaddyHandler struct {
	mock.Mock
}

func (m *MockCaddyHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) error {
	args := m.Called(writer, request)
	return args.Error(0)
}

func (ts *RedirectorTestSuite) TestRedirector_ServeHTTP() {
	type response struct {
		status  int
		headers http.Header
	}
	tests := []struct {
		name      string
		method    string
		url       string
		response  response
		expectErr bool
	}{
		{
			name:   "temporary redirect",
			method: "GET",
			url:    "https://www.example.com",
			response: response{
				status:  302,
				headers: map[string][]string{"Location": {"https://example.com"}, "Server": {"redirector"}},
			},
		},
		{
			name:   "permanent redirect",
			method: "GET",
			url:    "https://example.org",
			response: response{
				status:  301,
				headers: map[string][]string{"Location": {"https://www.example.org"}, "Server": {"redirector"}},
			},
		},
		{
			name:      "unknown hostname",
			method:    "GET",
			url:       "https://example.edu",
			expectErr: true,
		},
		{
			name:   "redirect with rewrites",
			method: "GET",
			url:    "https://www.example.info/foo/bar/baz",
			response: response{
				status: 302,
				headers: map[string][]string{
					"Location": {"https://example.info/foo/bar/baz"},
					"Server":   {"redirector"},
				},
			},
		},
	}
	for _, tc := range tests {
		ts.Run(tc.name, func() {
			w := httptest.NewRecorder()
			r, err := http.NewRequest(tc.method, tc.url, nil)
			ts.Require().NoError(err)
			mockHandler := new(MockCaddyHandler)
			mockHandler.On("ServeHTTP", mock.Anything, mock.Anything).Return(nil)

			err = ts.redirector.ServeHTTP(w, r, mockHandler)
			ts.Require().NoError(err)
			if tc.expectErr {
				mockHandler.AssertNumberOfCalls(ts.T(), "ServeHTTP", 1)
			} else {
				mockHandler.AssertNotCalled(ts.T(), "ServeHTTP")
				ts.Equal(tc.response.status, w.Code)
				ts.Equal(tc.response.headers, w.Result().Header)
			}
		})
	}
}
