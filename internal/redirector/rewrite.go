package redirector

import (
	"regexp"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type Rewrite struct {
	RegExp  RewriteRegexp `dynamodbav:"RegExp,omitempty"`
	Replace string        `dynamodbav:"Replace,omitempty"`
	Final   bool          `dynamodbav:"Final"`
}

type RewriteRegexp struct {
	*regexp.Regexp
}

func (rr RewriteRegexp) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	return &types.AttributeValueMemberS{Value: rr.String()}, nil
}

func (rr *RewriteRegexp) UnmarshalDynamoDBAttributeValue(value types.AttributeValue) error {
	if value == nil {
		rr.Regexp = nil
		return nil
	}
	val, ok := value.(*types.AttributeValueMemberS)
	if !ok {
		rr.Regexp = nil
		return nil
	}
	regex, err := regexp.Compile(val.Value)
	if err != nil {
		rr.Regexp = nil
		return nil //nolint:nilerr // Swallow err and set RegExp to nil
	}
	rr.Regexp = regex
	return nil
}

func (r *Rewrite) UnmarshalDynamoDBAttributeValue(value types.AttributeValue) error {
	type Alias Rewrite
	// Defaults for missing fields
	rewrite := &Alias{
		Replace: "$1",
		Final:   true,
	}

	err := attributevalue.Unmarshal(value, rewrite)
	if err != nil {
		return err
	}

	r.RegExp = rewrite.RegExp
	r.Replace = rewrite.Replace
	r.Final = rewrite.Final
	return nil
}
