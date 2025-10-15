package storage

import (
	"context"
	"errors"
	"io/fs"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/caddyserver/certmagic"
)

var (
	// Interface guards.
	_ certmagic.Storage = (*DynamoDBStorage)(nil)
)

// Store stores the value at key.
func (dbs DynamoDBStorage) Store(ctx context.Context, key string, value []byte) error {
	if key == "" {
		return errors.New("key cannot be empty")
	}
	item := &Item{
		Key:      key,
		Contents: value,
		Modified: aws.Time(time.Now()),
		Size:     int64(len(value)),
	}
	_, err := dbs.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(dbs.Table),
		Item:      item.Item(),
	})
	return err
}

// Load retrieves the value at key.
func (dbs DynamoDBStorage) Load(ctx context.Context, key string) ([]byte, error) {
	output, err := dbs.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName:      &dbs.Table,
		ConsistentRead: aws.Bool(true),
		Key: map[string]types.AttributeValue{
			"Key": &types.AttributeValueMemberS{Value: key},
		},
	})
	if err != nil {
		return nil, err
	}
	if len(output.Item) == 0 {
		return nil, fs.ErrNotExist
	}
	var item Item
	if err = item.Load(output.Item); err != nil {
		return nil, err
	}
	return item.Contents, nil
}

// Delete deletes the named key.
func (dbs DynamoDBStorage) Delete(ctx context.Context, key string) error {
	_, err := dbs.Client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: &dbs.Table,
		Key: map[string]types.AttributeValue{
			"Key": &types.AttributeValueMemberS{Value: key},
		},
		ReturnValues: types.ReturnValueNone,
	})
	if err != nil {
		return err
	}
	return nil
}

// Exists returns true if the key exists.
func (dbs DynamoDBStorage) Exists(ctx context.Context, key string) bool {
	output, err := dbs.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName:      &dbs.Table,
		ConsistentRead: aws.Bool(true),
		Key: map[string]types.AttributeValue{
			"Key": &types.AttributeValueMemberS{Value: key},
		},
	})
	return err == nil && len(output.Item) > 0
}

// List returns all keys in the given path.
func (dbs DynamoDBStorage) List(ctx context.Context, path string, recursive bool) ([]string, error) {
	input := &dynamodb.ScanInput{
		TableName:                &dbs.Table,
		ConsistentRead:           aws.Bool(true),
		ExpressionAttributeNames: map[string]string{"#key": "Key"},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":key": &types.AttributeValueMemberS{Value: path + "/"},
		},
		FilterExpression: aws.String("begins_with(#key, :key)"),
	}

	var matchingKeys []string

	paginator := dynamodb.NewScanPaginator(dbs.Client, input)
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, i := range output.Items {
			var item Item
			if err = item.Load(i); err != nil {
				return nil, err
			}
			if !recursive {
				// these two paths go through foo:
				// foo/cert/key
				// foo/cert/chain
				//
				// So List(prefix: "foo", recursive: false) would return:
				// foo/cert
				name := strings.TrimPrefix(item.Key, path+"/")
				if strings.Contains(name, "/") {
					continue
				}
			}
			matchingKeys = append(matchingKeys, item.Key)
		}
	}

	if len(matchingKeys) == 0 {
		return nil, fs.ErrNotExist
	}
	return matchingKeys, nil
}

// Stat returns information about key.
func (dbs DynamoDBStorage) Stat(ctx context.Context, key string) (certmagic.KeyInfo, error) {
	output, err := dbs.Client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName:      &dbs.Table,
		ConsistentRead: aws.Bool(true),
		Key: map[string]types.AttributeValue{
			"Key": &types.AttributeValueMemberS{Value: key},
		},
	})
	if err != nil {
		return certmagic.KeyInfo{}, err
	}
	if len(output.Item) == 0 {
		return certmagic.KeyInfo{}, fs.ErrNotExist
	}
	var item Item
	if err = item.Load(output.Item); err != nil {
		return certmagic.KeyInfo{}, err
	}
	return certmagic.KeyInfo{
		Key:        item.Key,
		Modified:   *item.Modified,
		Size:       item.Size,
		IsTerminal: true,
	}, nil
}
