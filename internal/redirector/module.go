package redirector

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/CruGlobal/redirector/internal/app"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"go.uber.org/zap"
)

var (
	// Interface guards.
	_ caddy.Provisioner           = (*Redirector)(nil)
	_ caddy.Module                = (*Redirector)(nil)
	_ caddyhttp.MiddlewareHandler = (*Redirector)(nil)
)

func init() {
	caddy.RegisterModule(Redirector{})
	httpcaddyfile.RegisterHandlerDirective("redirector", parseCaddyfile)
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

	m, ok := module.(*app.App)
	if !ok {
		return fmt.Errorf("unexpected module type: %T", module)
	}
	if m == nil {
		return errors.New("redirector has not been initialized")
	}

	if m.Client == nil {
		return errors.New("DynamoDB client has not been initialized")
	}

	r.Client = m.Client
	r.Table = m.Table
	r.Client = m.Client

	return nil
}

func parseCaddyfile(_ httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	// Redirector has no configuration, so we just return a new instance
	return NewRedirector(), nil
}

func (r Redirector) ServeHTTP(writer http.ResponseWriter, request *http.Request, handler caddyhttp.Handler) error {
	// Split host and port
	hostname, _, err := net.SplitHostPort(request.Host)
	if err != nil {
		hostname = request.Host // Probably OK, host just didn't have a port
	}
	useCache := !request.URL.Query().Has("skip_cache")

	writer.Header().Set("Server", "redirector")

	redirect, err := r.GetRedirect(request.Context(), hostname, useCache)
	if err != nil {
		return handler.ServeHTTP(writer, request)
	}
	return redirect.ServeHTTP(writer, request)
}

func (r *Redirector) GetRedirect(ctx context.Context, hostname string, _ bool) (*Redirect, error) {
	// TODO: Implement caching
	item, err := r.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.Table),
		Key: map[string]types.AttributeValue{
			r.Key: &types.AttributeValueMemberS{Value: hostname},
		},
	})
	if err != nil {
		return nil, err
	}
	if item.Item == nil {
		return nil, errors.New("no redirect found")
	}

	var redirect Redirect
	err = attributevalue.UnmarshalMap(item.Item, &redirect)
	if err != nil {
		return nil, err
	}
	return &redirect, nil
}
