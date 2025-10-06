package main

import (
	caddycmd "github.com/caddyserver/caddy/v2/cmd"

	_ "github.com/CruGlobal/redirector/internal/redirector/app"
	_ "github.com/CruGlobal/redirector/internal/redirector/permission"
	_ "github.com/caddyserver/caddy/v2/modules/standard"
)

func main() {
	caddycmd.Main()
}
