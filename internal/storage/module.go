package storage

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"cirello.io/dynamolock/v2"
	"github.com/CruGlobal/redirector/internal/app"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/certmagic"
)

var (
	// Interface guards.
	_ caddy.Module           = (*DynamoDBStorage)(nil)
	_ caddy.Provisioner      = (*DynamoDBStorage)(nil)
	_ caddy.StorageConverter = (*DynamoDBStorage)(nil)
	_ caddy.CleanerUpper     = (*DynamoDBStorage)(nil)
	_ caddyfile.Unmarshaler  = (*DynamoDBStorage)(nil)
)

const (
	DefaultTable  = "RedirectorCertsProd"
	LeaseDuration = 15 * time.Second
)

func init() {
	caddy.RegisterModule(DynamoDBStorage{})
}

// NewDynamoDBStorage creates a new DynamoDBStorage instance with default settings.
func NewDynamoDBStorage() *DynamoDBStorage {
	return &DynamoDBStorage{
		Table: DefaultTable,
		locks: make(map[string]*dynamolock.Lock),
		mutex: &sync.RWMutex{},
	}
}

// CaddyModule returns the Caddy module information.
func (dbs DynamoDBStorage) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID: "caddy.storage.dynamodb",
		New: func() caddy.Module {
			return NewDynamoDBStorage()
		},
	}
}

// Provision sets up the storage.
func (dbs *DynamoDBStorage) Provision(ctx caddy.Context) error {
	dbs.logger = ctx.Logger(dbs)

	module, err := ctx.App(app.AppName)
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
		return errors.New("DynamoDB client has not been initialized")
	}
	dbs.Client = redir.Client

	dbs.Locker, err = dynamolock.New(dbs.Client, dbs.Table,
		dynamolock.WithPartitionKeyName("Key"),
		dynamolock.WithLeaseDuration(LeaseDuration),
	)
	if err != nil {
		return fmt.Errorf("failed to create DynamoDB lock client: %w", err)
	}

	return nil
}

func (dbs DynamoDBStorage) Cleanup() error {
	if dbs.Locker != nil {
		return dbs.Locker.Close()
	}
	return nil
}

// UnmarshalCaddyfile sets up the storage from Caddyfile tokens. Syntax:
//
//	storage dynamodb {
//	    table <table_name>
//	}
func (dbs *DynamoDBStorage) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
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
				dbs.Table = configVal
			default:
				return d.Errf("unknown parameter '%s' for storage 'dynamodb'", configKey)
			}
		}
	}
	return nil
}

func (dbs DynamoDBStorage) CertMagicStorage() (certmagic.Storage, error) {
	return dbs, nil
}
