package redirector

import (
	"net/http"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type Status int

const (
	StatusTemporary Status = iota
	StatusPermanent
)

const (
	DefaultStatus = StatusTemporary
)

func statusNames() map[Status]string {
	return map[Status]string{
		StatusTemporary: "TEMPORARY",
		StatusPermanent: "PERMANENT",
	}
}

func (s Status) String() string {
	value, ok := statusNames()[s]
	if !ok {
		return statusNames()[DefaultStatus]
	}
	return value
}

func (s Status) StatusCode() int {
	switch s {
	case StatusTemporary:
		return http.StatusFound // Maybe http.StatusTemporaryRedirect?
	case StatusPermanent:
		return http.StatusMovedPermanently // Maybe http.StatusPermanentRedirect?
	}
	return http.StatusFound
}

func (s Status) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	return &types.AttributeValueMemberS{Value: s.String()}, nil
}

func (s *Status) UnmarshalDynamoDBAttributeValue(value types.AttributeValue) error {
	if value == nil {
		*s = DefaultStatus
		return nil
	}
	val, ok := value.(*types.AttributeValueMemberS)
	if !ok {
		*s = DefaultStatus
		return nil
	}
	switch val.Value {
	case statusNames()[StatusTemporary]:
		*s = StatusTemporary
	case statusNames()[StatusPermanent]:
		*s = StatusPermanent
	default:
		*s = DefaultStatus
	}
	return nil
}
