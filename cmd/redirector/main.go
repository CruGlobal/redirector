package main

import (
	"github.com/CruGlobal/redirector/internal/permission"
	"github.com/caddyserver/caddy/v2"
	caddycmd "github.com/caddyserver/caddy/v2/cmd"

	_ "github.com/caddyserver/caddy/v2/modules/standard"
)

func main() {
	caddy.RegisterModule(permission.DynamoDBPermission{})
	caddycmd.Main()
}
