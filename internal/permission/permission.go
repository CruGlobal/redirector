package permission

import (
	"context"
	"errors"
	"fmt"

	"github.com/CruGlobal/redirector/internal/app"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"go.uber.org/zap"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddytls"
)

var (
	// Interface guards.
	_ caddy.Module                = (*Permission)(nil)
	_ caddy.Provisioner           = (*Permission)(nil)
	_ caddytls.OnDemandPermission = (*Permission)(nil)
)

type Permission struct {
	Table  string           `json:"table,omitempty"`
	Key    string           `json:"key,omitempty"`
	Client *dynamodb.Client `json:"-"`

	logger *zap.Logger
}

func init() {
	caddy.RegisterModule(Permission{})
}

func NewPermission() *Permission {
	return &Permission{
		Table: app.DefaultTable,
		Key:   app.DefaultKey,
	}
}

func (p Permission) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID: "tls.permission.dynamodb",
		New: func() caddy.Module {
			return NewPermission()
		},
	}
}

func (p *Permission) Provision(ctx caddy.Context) error {
	p.logger = ctx.Logger(p)

	module, err := ctx.App("redirector")
	if err != nil {
		return err
	}

	redirectorApp, ok := module.(*app.App)
	if !ok {
		return fmt.Errorf("unexpected module type: %T", module)
	}
	if redirectorApp == nil {
		return errors.New("redirector has not been initialized")
	}

	if redirectorApp.Client == nil {
		return errors.New("DynamoDB client has not been initialized")
	}

	p.Client = redirectorApp.Client
	p.Table = redirectorApp.Table
	p.Key = redirectorApp.Key

	return nil
}

func (p *Permission) CertificateAllowed(ctx context.Context, name string) error {
	item, err := p.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(p.Table),
		Key: map[string]types.AttributeValue{
			p.Key: &types.AttributeValueMemberS{Value: name},
		},
	})
	if err != nil {
		return fmt.Errorf("%s: %w (error looking up %w)", name, caddytls.ErrPermissionDenied, err)
	}
	if item.Item != nil {
		return nil
	}
	return fmt.Errorf("%s: %w", name, caddytls.ErrPermissionDenied)
}
