package redirector_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/CruGlobal/redirector/internal/redirector"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedirect_ServeHTTP(t *testing.T) {
	type response struct {
		status  int
		headers http.Header
	}
	tests := []struct {
		name      string
		redirect  redirector.Redirect
		response  response
		expectErr bool
	}{
		{
			name: "missing location",
			redirect: redirector.Redirect{
				Location: "",
			},
			expectErr: true,
		},
		{
			name: "defaults",
			redirect: redirector.Redirect{
				Location: "https://example.com",
			},
			response: response{
				status:  302,
				headers: map[string][]string{"Location": {"https://example.com"}},
			},
		},
		{
			name: "permanent redirect",
			redirect: redirector.Redirect{
				Location: "https://example.com",
				Status:   redirector.StatusPermanent,
			},
			response: response{
				status:  301,
				headers: map[string][]string{"Location": {"https://example.com"}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			err := tt.redirect.ServeHTTP(w, nil)
			if tt.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.response.status, w.Code)
			assert.Equal(t, tt.response.headers, w.Result().Header)
		})
	}
}
