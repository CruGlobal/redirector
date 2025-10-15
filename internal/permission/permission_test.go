package permission_test

import (
	"testing"

	"github.com/CruGlobal/redirector/internal/permission"
	"github.com/CruGlobal/redirector/redirtest"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestPermission_NewPermission(t *testing.T) {
	perm := permission.NewPermission()
	assert.NotNil(t, perm)
	assert.IsType(t, &permission.Permission{}, perm)
	assert.Equal(t, permission.DefaultTable, perm.Table)
	assert.Equal(t, permission.DefaultKey, perm.Key)
}

func TestPermission_CaddyModule(t *testing.T) {
	module := permission.Permission{}.CaddyModule()
	assert.IsType(t, caddy.ModuleInfo{}, module)
	assert.Equal(t, caddy.ModuleID("tls.permission.dynamodb"), module.ID)
	assert.IsType(t, &permission.Permission{}, module.New())
}

func TestPermission_Provision(t *testing.T) {
	ctx := redirtest.NewRedirectorCaddyContext(t)

	perm := permission.NewPermission()
	err := perm.Provision(ctx)
	require.NoError(t, err)
}

func TestPermission_UnmarshalCaddyfile(t *testing.T) {
	testcases := []struct {
		name      string
		caddyfile string
		expected  *permission.Permission
		expectErr bool
	}{
		{
			name: "valid1",
			caddyfile: `dynamodb {
				table TestTableName
				key TestKey
			}`,
			expected: &permission.Permission{
				Table: "TestTableName",
				Key:   "TestKey",
			},
			expectErr: false,
		},
		{
			name: "valid2",
			caddyfile: `dynamodb {
				table TestTableName
			}`,
			expected: &permission.Permission{
				Table: "TestTableName",
				Key:   permission.DefaultKey,
			},
			expectErr: false,
		},
		{
			name: "valid3",
			caddyfile: `dynamodb {
				key TestKey
			}`,
			expected: &permission.Permission{
				Table: permission.DefaultTable,
				Key:   "TestKey",
			},
			expectErr: false,
		},
		{
			name:      "valid4",
			caddyfile: `dynamodb`,
			expected: &permission.Permission{
				Table: permission.DefaultTable,
				Key:   permission.DefaultKey,
			},
			expectErr: false,
		},
		{
			name: "valid5",
			caddyfile: `dynamodb {
			}`,
			expected: &permission.Permission{
				Table: permission.DefaultTable,
				Key:   permission.DefaultKey,
			},
			expectErr: false,
		},
		{
			name: "invalid",
			caddyfile: `dynamodb name {
				key TestKey
			}`,
			expected:  nil,
			expectErr: true,
		},
		{
			name: "invalid2",
			caddyfile: `dynamodb {
				key TestKey
				extra value
			}`,
			expected:  nil,
			expectErr: true,
		},
		{
			name:      "invalid3",
			caddyfile: `dynamodb {}`,
			expected:  nil,
			expectErr: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			perm := permission.NewPermission()
			err := perm.UnmarshalCaddyfile(caddyfile.NewTestDispenser(tc.caddyfile))
			if tc.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expected, perm)
		})
	}
}

type PermissionTestSuite struct {
	suite.Suite

	permission *permission.Permission
}

const (
	testTable = "TestingTableName"
	testKey   = "TestKey"
)

func (ts *PermissionTestSuite) SetupSuite() {
	client := redirtest.NewDynamoDBClient(ts.T())
	redirtest.CreateDynamoDBTable(ts.T(), client, testTable, testKey)

	perm := permission.Permission{
		Table:  testTable,
		Key:    testKey,
		Client: client,
	}

	ts.permission = &perm
}

func (ts *PermissionTestSuite) TearDownSuite() {
	redirtest.DeleteDynamoDBTable(ts.T(), ts.permission.Client, testTable)
}

func (ts *PermissionTestSuite) TestPermission_CertificateAllowed() {
	ctx := ts.T().Context()
	validKeys := []string{"example.com", "www.example.com", "starkindustries.com"}
	invalidKeys := []string{"www.starkindustries.com", "ftp.example.com"}

	for _, key := range validKeys {
		_, err := ts.permission.Client.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: aws.String(testTable),
			Item: map[string]types.AttributeValue{
				testKey: &types.AttributeValueMemberS{Value: key},
			},
		})
		ts.Require().NoError(err)
	}

	for _, valid := range validKeys {
		err := ts.permission.CertificateAllowed(ctx, valid)
		ts.Require().NoError(err)
	}

	for _, invalid := range invalidKeys {
		err := ts.permission.CertificateAllowed(ctx, invalid)
		ts.Require().Error(err)
	}
}

func TestPermissionTestSuite(t *testing.T) {
	suite.Run(t, new(PermissionTestSuite))
}
