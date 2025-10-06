package permission

import (
	"context"
	"errors"
	"fmt"

	"github.com/CruGlobal/redirector/internal/redirector/app"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"go.uber.org/zap"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddytls"
)

var (
	_ caddy.Module                = (*Permission)(nil)
	_ caddyfile.Unmarshaler       = (*Permission)(nil)
	_ caddy.Provisioner           = (*Permission)(nil)
	_ caddytls.OnDemandPermission = (*Permission)(nil)
)

const (
	DefaultDynamoDBTable = "RedirectorAppProd"
	DefaultDynamoDBKey   = "Hostname"
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
		Table: DefaultDynamoDBTable,
		Key:   DefaultDynamoDBKey,
	}
}

func (perm Permission) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID: "tls.permission.dynamodb",
		New: func() caddy.Module {
			return NewPermission()
		},
	}
}

func (perm *Permission) Provision(ctx caddy.Context) error {
	perm.logger = ctx.Logger(perm)

	module, err := ctx.App("redirector")
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
		return errors.New("DynamoDB client has been initialized")
	}

	perm.Client = redir.Client

	return nil
}

func (perm *Permission) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
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
				perm.Table = configVal
			case "key":
				perm.Key = configVal
			default:
				return d.Errf("unknown parameter '%s' for 'dynamodb'", configKey)
			}
		}
	}
	return nil
}

func (perm *Permission) CertificateAllowed(ctx context.Context, name string) error {
	item, err := perm.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(perm.Table),
		Key: map[string]types.AttributeValue{
			perm.Key: &types.AttributeValueMemberS{Value: name},
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
