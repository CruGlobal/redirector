package redirector

import (
	"errors"
	"fmt"

	"github.com/CruGlobal/redirector/internal/app"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"go.uber.org/zap"
)

var (
	// Interface guards.
	_ caddy.Provisioner     = (*Redirector)(nil)
	_ caddy.Module          = (*Redirector)(nil)
	_ caddyfile.Unmarshaler = (*Redirector)(nil)
)

func init() {
	caddy.RegisterModule(Redirector{})
}

type Redirector struct {
	Table  string           `json:"table,omitempty"`
	Key    string           `json:"key,omitempty"`
	Client *dynamodb.Client `json:"-"`

	logger *zap.Logger
}

func NewRedirector() *Redirector {
	return &Redirector{
		Table: app.DefaultTable,
		Key:   app.DefaultKey,
	}
}

func (r Redirector) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID: "http.handlers.redirector",
		New: func() caddy.Module {
			return NewRedirector()
		},
	}
}

func (r *Redirector) Provision(ctx caddy.Context) error {
	r.logger = ctx.Logger(r)

	module, err := ctx.App(app.AppName)
	if err != nil {
		return err
	}

	redir, ok := module.(*app.App)
	if !ok {
		return fmt.Errorf("unexpected module type: %T", module)
	}
	if redir == nil {
		return errors.New("redirector has not been initialized")
	}

	if redir.Client == nil {
		return errors.New("DynamoDB client has not been initialized")
	}

	r.Client = redir.Client
	r.Table = redir.Table
	r.Client = redir.Client

	return nil
}

func (r *Redirector) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		if d.NextArg() {
			return d.ArgErr()
		}

		for nesting := d.Nesting(); d.NextBlock(nesting); {
			configKey := d.Val()
			var configVal string

			if !d.Args(&configVal) {
				return d.ArgErr()
			}

			switch configKey {
			case "table":
				r.Table = configVal
			case "key":
				r.Key = configVal
			default:
				return d.Errf("unknown parameter '%s' for 'dynamodb'", configKey)
			}
		}
	}
	return nil
}
