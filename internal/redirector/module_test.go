package redirector_test

import (
	"os"
	"testing"

	"github.com/CruGlobal/redirector/internal/app"
	"github.com/CruGlobal/redirector/internal/redirector"
	"github.com/CruGlobal/redirector/redirtest"
	"github.com/caddyserver/caddy/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
