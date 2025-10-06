package app

import (
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/caddyserver/caddy/v2"
	"go.uber.org/zap"
)

const (
	appName = "redirector"
)

var (
	// Interface guards.
	_ caddy.Provisioner = (*App)(nil)
	_ caddy.Module      = (*App)(nil)
	_ caddy.App         = (*App)(nil)
)

func init() {
	caddy.RegisterModule(App{})
}

// App implements redirector.
type App struct {
	Name   string           `json:"-"`
	Client *dynamodb.Client `json:"-"`
	logger *zap.Logger

	Region     string `json:"region,omitempty"`
	Endpoint   string `json:"endpoint,omitempty"`
	DisableSSL bool   `json:"disable_ssl,omitempty"`
}

func NewApp() *App {
	r := App{
		Region:     "us-east-1",
		DisableSSL: false,
	}
	return &r
}

func (app App) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "redirector",
		New: func() caddy.Module { return NewApp() },
	}
}

func (app *App) Provision(ctx caddy.Context) error {
	app.Name = appName
	app.logger = ctx.Logger(app)

	app.logger.Info(
		"provisioning app instance",
		zap.String("app", app.Name),
	)

	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(app.Region),
		config.WithBaseEndpoint(app.Endpoint),
	)
	if err != nil {
		return err
	}
	app.Client = dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		o.EndpointOptions.DisableHTTPS = app.DisableSSL
	})

	return nil
}

func (app App) Start() error {
	app.logger.Debug(
		"started app instance",
		zap.String("app", app.Name),
	)
	return nil
}

func (app App) Stop() error {
	app.logger.Debug(
		"stopped app instance",
		zap.String("app", app.Name),
	)
	return nil
}
