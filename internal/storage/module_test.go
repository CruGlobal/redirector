package storage_test

import (
	"testing"

	"cirello.io/dynamolock/v2"
	storage2 "github.com/CruGlobal/redirector/internal/storage"
	"github.com/CruGlobal/redirector/redirtest"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStorage_NewDynamoDBStorage(t *testing.T) {
	s := storage2.NewDynamoDBStorage()
	assert.NotNil(t, s)
	assert.IsType(t, &storage2.DynamoDBStorage{}, s)
	assert.Equal(t, storage2.DefaultTable, s.Table)
}

func TestStorage_CaddyModule(t *testing.T) {
	module := storage2.DynamoDBStorage{}.CaddyModule()
	assert.IsType(t, caddy.ModuleInfo{}, module)
	assert.Equal(t, caddy.ModuleID("caddy.storage.dynamodb"), module.ID)
	assert.IsType(t, &storage2.DynamoDBStorage{}, module.New())
}

func TestStorage_Provision(t *testing.T) {
	ctx := redirtest.NewRedirectorCaddyContext(t)

	s := storage2.NewDynamoDBStorage()
	err := s.Provision(ctx)
	require.NoError(t, err)
}

func TestStorage_UnmarshalCaddyfile(t *testing.T) {
	testcases := []struct {
		name      string
		caddyfile string
		expected  string
		expectErr bool
	}{
		{
			name: "valid1",
			caddyfile: `dynamodb {
				table TestTableName
			}`,
			expected:  "TestTableName",
			expectErr: false,
		},
		{
			name:      "valid2",
			caddyfile: `dynamodb`,
			expected:  storage2.DefaultTable,
			expectErr: false,
		},
		{
			name: "valid3",
			caddyfile: `dynamodb {
			}`,
			expected:  storage2.DefaultTable,
			expectErr: false,
		},
		{
			name: "invalid",
			caddyfile: `dynamodb name {
				key TestKey
			}`,
			expectErr: true,
		},
		{
			name: "invalid2",
			caddyfile: `dynamodb {
				key TestKey
			}`,
			expectErr: true,
		},
		{
			name:      "invalid3",
			caddyfile: `dynamodb {}`,
			expectErr: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			s := storage2.NewDynamoDBStorage()
			err := s.UnmarshalCaddyfile(caddyfile.NewTestDispenser(tc.caddyfile))
			if tc.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.IsType(t, &dynamodb.Client{}, s.Client)
			require.IsType(t, &dynamolock.Client{}, s.Locker)
			require.Equal(t, tc.expected, s.Table)
		})
	}
}
