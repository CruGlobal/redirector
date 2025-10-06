package app

import (
	"strconv"

	"github.com/caddyserver/caddy/v2/caddyconfig"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
)

func init() {
	httpcaddyfile.RegisterGlobalOption("redirector", ParseRedirector)
}

// ParseRedirector sets up the App from Caddyfile tokens. Syntax:
//
//		{
//		  redirector {
//		    region us-east-1
//		    endpoint 127.0.0.1:8000
//	        disable_ssl true
//		  }
//		}
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
			case "disable_ssl":
				disableSSL, err := strconv.ParseBool(configVal)
				if err != nil {
					return nil, d.Errf("invalid boolean value for 'disable_ssl': %s", configVal)
				}
				app.DisableSSL = disableSSL
			default:
				return nil, d.Errf("unknown parameter '%s' for 'redirector'", configKey)
			}
		}
	}

	return httpcaddyfile.App{
		Name:  appName,
		Value: caddyconfig.JSON(app, nil),
	}, nil
}
