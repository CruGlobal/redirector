package redirector

import "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

type Type int

const (
	TypeRedirect Type = iota
	TypeProxy
)

const (
	DefaultType = TypeRedirect
)

func typeNames() map[Type]string {
	return map[Type]string{
		TypeRedirect: "REDIRECT",
		TypeProxy:    "PROXY",
	}
}

func (t Type) String() string {
	value, ok := typeNames()[t]
	if !ok {
		return typeNames()[DefaultType]
	}
	return value
}

func (t Type) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	return &types.AttributeValueMemberS{Value: t.String()}, nil
}

func (t *Type) UnmarshalDynamoDBAttributeValue(value types.AttributeValue) error {
	if value == nil {
		*t = DefaultType
		return nil
	}
	val, ok := value.(*types.AttributeValueMemberS)
	if !ok {
		*t = DefaultType
		return nil
	}
	switch val.Value {
	case typeNames()[TypeRedirect]:
		*t = TypeRedirect
	case typeNames()[TypeProxy]:
		*t = TypeProxy
	default:
		*t = DefaultType
	}
	return nil
}
