package main

import (
	caddycmd "github.com/caddyserver/caddy/v2/cmd"

	_ "github.com/CruGlobal/redirector/internal/app"
	_ "github.com/CruGlobal/redirector/internal/permission"
	_ "github.com/CruGlobal/redirector/internal/storage"
	_ "github.com/caddyserver/caddy/v2/modules/standard"
)

func main() {
	caddycmd.Main()
}
