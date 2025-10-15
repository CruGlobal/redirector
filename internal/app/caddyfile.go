package app

import (
	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
)

func init() {
	httpcaddyfile.RegisterGlobalOption("redirector", ParseRedirector)
}

// ParseRedirector sets up the App from Caddyfile tokens. Syntax:
//
//	{
//	    redirector {
//	        region <region>
//	        endpoint <endpoint>
//	        table <table_name>
//	        key <key_name>
//	    }
//	}
func ParseRedirector(d *caddyfile.Dispenser, _ any) (any, error) {
	app := new(App)

	for d.Next() {
		if d.NextArg() {
			return nil, d.ArgErr()
		}

		for nesting := d.Nesting(); d.NextBlock(nesting); {
			configKey := d.Val()
			var configVal string

			if !d.Args(&configVal) {
				return nil, d.ArgErr()
			}

			switch configKey {
			case "region":
				app.Region = configVal
			case "endpoint":
				app.Endpoint = configVal
			case "table":
				app.Table = configVal
			case "key":
				app.Key = configVal
			default:
				return nil, d.Errf("unknown parameter '%s' for 'redirector'", configKey)
			}
		}
	}

	return httpcaddyfile.App{
		Name:  AppName,
		Value: caddyconfig.JSON(app, nil),
	}, nil
}
