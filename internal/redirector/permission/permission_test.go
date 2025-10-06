package permission_test

import (
	"context"
	"testing"

	"github.com/CruGlobal/redirector/internal/redirector/permission"
	"github.com/CruGlobal/redirector/redirtest"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestPermission_NewPermission(t *testing.T) {
	perm := permission.NewPermission()
	assert.NotNil(t, perm)
	assert.IsType(t, &permission.Permission{}, perm)
	assert.Equal(t, permission.DefaultDynamoDBTable, perm.Table)
	assert.Equal(t, permission.DefaultDynamoDBKey, perm.Key)
}

func TestPermission_CaddyModule(t *testing.T) {
	module := permission.Permission{}.CaddyModule()
	assert.IsType(t, caddy.ModuleInfo{}, module)
	assert.Equal(t, caddy.ModuleID("tls.permission.dynamodb"), module.ID)
	assert.IsType(t, &permission.Permission{}, module.New())
}

func TestPermission_Provision(t *testing.T) {
	ctx := redirtest.NewRedirectorCaddyContext(t)

	perm := permission.NewPermission()
	err := perm.Provision(ctx)
	require.NoError(t, err)
}

func TestPermission_UnmarshalCaddyfile(t *testing.T) {
	testcases := []struct {
		name      string
		caddyfile string
		expected  *permission.Permission
		expectErr bool
	}{
		{
			name: "valid1",
			caddyfile: `dynamodb {
				table TestTableName
				key TestKey
			}`,
			expected: &permission.Permission{
				Table: "TestTableName",
				Key:   "TestKey",
			},
			expectErr: false,
		},
		{
			name: "valid2",
			caddyfile: `dynamodb {
				table TestTableName
			}`,
			expected: &permission.Permission{
				Table: "TestTableName",
				Key:   permission.DefaultDynamoDBKey,
			},
			expectErr: false,
		},
		{
			name: "valid3",
			caddyfile: `dynamodb {
				key TestKey
			}`,
			expected: &permission.Permission{
				Table: permission.DefaultDynamoDBTable,
				Key:   "TestKey",
			},
			expectErr: false,
		},
		{
			name:      "valid4",
			caddyfile: `dynamodb`,
			expected: &permission.Permission{
				Table: permission.DefaultDynamoDBTable,
				Key:   permission.DefaultDynamoDBKey,
			},
			expectErr: false,
		},
		{
			name: "valid5",
			caddyfile: `dynamodb {
			}`,
			expected: &permission.Permission{
				Table: permission.DefaultDynamoDBTable,
				Key:   permission.DefaultDynamoDBKey,
			},
			expectErr: false,
		},
		{
			name: "invalid",
			caddyfile: `dynamodb name {
				key TestKey
			}`,
			expected:  nil,
			expectErr: true,
		},
		{
			name: "invalid2",
			caddyfile: `dynamodb {
				key TestKey
				extra value
			}`,
			expected:  nil,
			expectErr: true,
		},
		{
			name:      "invalid3",
			caddyfile: `dynamodb {}`,
			expected:  nil,
			expectErr: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			perm := permission.NewPermission()
			err := perm.UnmarshalCaddyfile(caddyfile.NewTestDispenser(tc.caddyfile))
			if tc.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expected, perm)
		})
	}
}

type PermissionTestSuite struct {
	suite.Suite

	permission *permission.Permission
}

const (
	testTable = "TestingTableName"
	testKey   = "TestKey"
)

func (s *PermissionTestSuite) SetupSuite() {
	ctx := context.Background()

	cfg, _ := config.LoadDefaultConfig(
		ctx,
		config.WithRegion("local"),
		config.WithBaseEndpoint("http://localhost:8000"),
	)
	client := dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		o.EndpointOptions.DisableHTTPS = true
	})

	_, err := client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName:   aws.String(testTable),
		BillingMode: types.BillingModePayPerRequest,
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String(testKey),
				KeyType:       types.KeyTypeHash,
			},
		},
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String(testKey),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
	})
	if err != nil {
		panic(err)
	}

	perm := permission.Permission{
		Table:  testTable,
		Key:    testKey,
		Client: client,
	}

	s.permission = &perm
}

func (s *PermissionTestSuite) TearDownSuite() {
	_, err := s.permission.Client.DeleteTable(s.T().Context(), &dynamodb.DeleteTableInput{
		TableName: aws.String(testTable),
	})
	s.Require().NoError(err)
}

func (s *PermissionTestSuite) TestPermission_CertificateAllowed() {
	ctx := s.T().Context()
	validKeys := []string{"example.com", "www.example.com", "starkindustries.com"}
	invalidKeys := []string{"www.starkindustries.com", "ftp.example.com"}

	for _, key := range validKeys {
		_, err := s.permission.Client.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: aws.String(testTable),
			Item: map[string]types.AttributeValue{
				testKey: &types.AttributeValueMemberS{Value: key},
			},
		})
		s.Require().NoError(err)
	}

	for _, valid := range validKeys {
		err := s.permission.CertificateAllowed(ctx, valid)
		s.Require().NoError(err)
	}

	for _, invalid := range invalidKeys {
		err := s.permission.CertificateAllowed(ctx, invalid)
		s.Require().Error(err)
	}
}

func TestPermissionTestSuite(t *testing.T) {
	suite.Run(t, new(PermissionTestSuite))
}
