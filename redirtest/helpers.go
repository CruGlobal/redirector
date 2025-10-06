package redirtest

import (
	"testing"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/stretchr/testify/require"
)

func NewRedirectorCaddyContext(t *testing.T) caddy.Context {
	t.Helper()

	caddyfileInput := `{
	redirector {
		region local
		endpoint http://localhost:8000
		disable_ssl true
	}
	log {
		level ERROR
	}
}
`
	adapter := caddyfile.Adapter{ServerType: &httpcaddyfile.ServerType{}}
	adaptedJSON, warnings, err := adapter.Adapt([]byte(caddyfileInput), nil)
	require.NoError(t, err)
	require.Empty(t, warnings)

	cfg := &caddy.Config{}
	err = caddy.StrictUnmarshalJSON(adaptedJSON, cfg)
	require.NoError(t, err)

	ctx, err := caddy.ProvisionContext(cfg)
	require.NoError(t, err)

	return ctx
}
