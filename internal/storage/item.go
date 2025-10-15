package storage

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type Item struct {
	Key      string     `dynamodbav:"Key"`
	Contents []byte     `dynamodbav:"Contents,omitempty"`
	Modified *time.Time `dynamodbav:"Modified,omitempty"`
	Size     int64      `dynamodbav:"Size,omitempty"`
}

func (i *Item) Item() map[string]types.AttributeValue {
	item, _ := attributevalue.MarshalMap(i)
	return item
}

func (i *Item) Load(item map[string]types.AttributeValue) error {
	return attributevalue.UnmarshalMap(item, i)
}
