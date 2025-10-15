package storage

import (
	"sync"

	"cirello.io/dynamolock/v2"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"go.uber.org/zap"
)

type DynamoDBStorage struct {
	logger *zap.Logger
	locks  map[string]*dynamolock.Lock
	mutex  *sync.RWMutex

	Client *dynamodb.Client   `json:"-"`
	Locker *dynamolock.Client `json:"-"`
	Table  string             `json:"table,omitempty"`
}
