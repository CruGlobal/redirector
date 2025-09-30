//nolint:recvcheck // Caddy mixes receiver ptr/non-ptr on interfaces
package permission

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"go.uber.org/zap"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddytls"
)

var (
	_ caddy.Module                = (*DynamoDBPermission)(nil)
	_ caddy.Provisioner           = (*DynamoDBPermission)(nil)
	_ caddyfile.Unmarshaler       = (*DynamoDBPermission)(nil)
	_ caddytls.OnDemandPermission = (*DynamoDBPermission)(nil)
	_ caddy.Provisioner           = (*DynamoDBPermission)(nil)
)

const (
	DefaultDynamoDBRegion  = "us-east-1"
	DefaultDynamoDBTable   = "RedirectorAppProd"
	DefaultDynamoDBHashKey = "Hostname"
)

type DynamoDBPermission struct {
	Client *dynamodb.Client
	logger *zap.Logger

	Region  string `json:"region"`
	Table   string `json:"table"`
	HashKey string `json:"hash_key"`
}

func (m DynamoDBPermission) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID: "tls.permission.dynamodb",
		New: func() caddy.Module {
			return new(DynamoDBPermission)
		},
	}
}

func (m *DynamoDBPermission) Provision(ctx caddy.Context) error {
	if m.logger == nil {
		m.logger = ctx.Logger(m)
	}

	m.logger.Info("Creating new DynamoDB client")

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(m.Region))
	if err != nil {
		return err
	}

	m.Client = dynamodb.NewFromConfig(cfg)
	return nil
}

func (m *DynamoDBPermission) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		key := d.Val()
		var value string

		if !d.Args(&value) {
			continue
		}

		switch key {
		case "region":
			if value != "" {
				m.Region = value
			} else {
				m.Region = DefaultDynamoDBRegion
			}
		case "table":
			if value != "" {
				m.Table = value
			} else {
				m.Table = DefaultDynamoDBTable
			}
		case "hash_key":
			if value != "" {
				m.HashKey = value
			} else {
				m.HashKey = DefaultDynamoDBHashKey
			}
		}
	}
	return nil
}

func (m *DynamoDBPermission) CertificateAllowed(ctx context.Context, name string) error {
	item, err := m.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(m.Table),
		Key: map[string]types.AttributeValue{
			m.HashKey: &types.AttributeValueMemberS{Value: name},
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
