package app_test

import (
	"testing"

	"github.com/CruGlobal/redirector/internal/redirector/app"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRedirector(t *testing.T) {
	testcases := []struct {
		name      string
		d         *caddyfile.Dispenser
		want      string
		shouldErr bool
		err       string
	}{
		{
			name: "missing",
			d: caddyfile.NewTestDispenser(`{
            }`),
			want: "{}",
		},
		{
			name: "blank",
			d: caddyfile.NewTestDispenser(`{
                redirector
            }`),
			want: "{}",
		},
		{
			name: "empty",
			d: caddyfile.NewTestDispenser(`{
                redirector {
                }
            }`),
			want: "{}",
		},
		{
			name: "unexpected",
			d: caddyfile.NewTestDispenser(`{
                redirector something {}
            }`),
			shouldErr: true,
			err:       "wrong argument count or unexpected line ending after 'something'",
			want:      "{}",
		},
		{
			name: "valid1",
			d: caddyfile.NewTestDispenser(`{
                redirector {
                  region local
                  endpoint example.com
                  disable_ssl true
                }
            }`),
			want: `{"disable_ssl":true,"region":"local","endpoint":"example.com"}`,
		},
		{
			name: "valid2",
			d: caddyfile.NewTestDispenser(`{
                redirector {
                  region us-west-2
                }
            }`),
			want: `{"region":"us-west-2"}`,
		},
		{
			name: "invalid1",
			d: caddyfile.NewTestDispenser(`{
                redirector {
                  region
                }
            }`),
			shouldErr: true,
			err:       "wrong argument count or unexpected line ending after 'region'",
		},
		{
			name: "invalid2",
			d: caddyfile.NewTestDispenser(`{
                redirector {
                  region local
                  table name
                  endpoint example.com
                  disable_ssl true
                }
            }`),
			shouldErr: true,
			err:       "unknown parameter 'table'",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			app, err := app.ParseRedirector(tc.d, nil)
			if err != nil {
				if !tc.shouldErr {
					t.Fatalf("unexpected error: %v", err)
				}
				require.ErrorContains(t, err, tc.err)
				return
			}
			if tc.shouldErr {
				t.Fatalf("unexpected success: %v", err)
			}

			json := string(app.(httpcaddyfile.App).Value)
			assert.JSONEq(t, tc.want, json)
		})
	}
}
