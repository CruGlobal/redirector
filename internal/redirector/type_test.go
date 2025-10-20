package redirector_test

import (
	"testing"

	"github.com/CruGlobal/redirector/internal/redirector"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestType_String(t *testing.T) {
	tests := []struct {
		name     string
		input    redirector.Type
		expected string
	}{
		{
			name:     "REDIRECT",
			input:    redirector.TypeRedirect,
			expected: "REDIRECT",
		},
		{
			name:     "PROXY",
			input:    redirector.TypeProxy,
			expected: "PROXY",
		},
		{
			name:     "unknown type",
			input:    redirector.Type(42),
			expected: "REDIRECT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestType_MarshalDynamoDBAttributeValue(t *testing.T) {
	tests := []struct {
		name     string
		input    redirector.Type
		expected *types.AttributeValueMemberS
	}{
		{
			name:     "valid REDIRECT type",
			input:    redirector.TypeRedirect,
			expected: &types.AttributeValueMemberS{Value: "REDIRECT"},
		},
		{
			name:     "valid PROXY type",
			input:    redirector.TypeProxy,
			expected: &types.AttributeValueMemberS{Value: "PROXY"},
		},
		{
			name:     "invalid type",
			input:    redirector.Type(42),
			expected: &types.AttributeValueMemberS{Value: "REDIRECT"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.input.MarshalDynamoDBAttributeValue()
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestType_UnmarshalDynamoDBAttributeValue(t *testing.T) {
	tests := []struct {
		name     string
		input    types.AttributeValue
		expected redirector.Type
	}{
		{
			name:     "REDIRECT",
			input:    &types.AttributeValueMemberS{Value: "REDIRECT"},
			expected: redirector.TypeRedirect,
		},
		{
			name:     "PROXY",
			input:    &types.AttributeValueMemberS{Value: "PROXY"},
			expected: redirector.TypeProxy,
		},
		{
			name:     "unknown type",
			input:    &types.AttributeValueMemberS{Value: "FOO"},
			expected: redirector.DefaultType,
		},
		{
			name:     "incorrect type",
			input:    &types.AttributeValueMemberN{Value: "0"},
			expected: redirector.DefaultType,
		},
		{
			name:     "nil",
			input:    nil,
			expected: redirector.DefaultType,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p redirector.Type
			err := p.UnmarshalDynamoDBAttributeValue(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, p)
		})
	}
}
