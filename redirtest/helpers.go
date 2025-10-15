package redirtest

import (
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/stretchr/testify/require"
)

func NewRedirectorCaddyContext(t *testing.T) caddy.Context {
	t.Helper()

	caddyfileInput := fmt.Sprintf(`{
	redirector {
		region %s
		endpoint %s
	}
	log {
		level ERROR
	}
}
`, os.Getenv("DYNAMODB_TESTING_REGION"), os.Getenv("DYNAMODB_TESTING_ENDPOINT"))
	adapter := caddyfile.Adapter{ServerType: &httpcaddyfile.ServerType{}}
	adaptedJSON, warnings, err := adapter.Adapt([]byte(caddyfileInput), nil)
	require.NoError(t, err)
	require.Empty(t, warnings)

	cfg := &caddy.Config{}
	err = caddy.StrictUnmarshalJSON(adaptedJSON, cfg)
	require.NoError(t, err)

	ctx, err := caddy.ProvisionContext(cfg)
	require.NoError(t, err)

	return ctx
}

func NewDynamoDBClient(t *testing.T) *dynamodb.Client {
	cfg, err := config.LoadDefaultConfig(
		t.Context(),
		config.WithRegion(os.Getenv("DYNAMODB_TESTING_REGION")),
		config.WithBaseEndpoint(os.Getenv("DYNAMODB_TESTING_ENDPOINT")),
	)
	require.NoError(t, err)
	return dynamodb.NewFromConfig(cfg)
}

func CreateDynamoDBTable(t *testing.T, client *dynamodb.Client, table string, key string) {
	_, err := client.CreateTable(t.Context(), &dynamodb.CreateTableInput{
		TableName:   aws.String(table),
		BillingMode: types.BillingModePayPerRequest,
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String(key),
				KeyType:       types.KeyTypeHash,
			},
		},
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String(key),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
	})
	require.NoError(t, err)
}

func DeleteDynamoDBTable(t *testing.T, client *dynamodb.Client, table string) {
	_, err := client.DeleteTable(t.Context(), &dynamodb.DeleteTableInput{
		TableName: aws.String(table),
	})
	require.NoError(t, err)
}
