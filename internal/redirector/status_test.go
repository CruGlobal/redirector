package redirector_test

import (
	"testing"

	"github.com/CruGlobal/redirector/internal/redirector"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatus_String(t *testing.T) {
	tests := []struct {
		name     string
		input    redirector.Status
		expected string
	}{
		{
			name:     "temporary",
			input:    redirector.StatusTemporary,
			expected: "TEMPORARY",
		},
		{
			name:     "permanent",
			input:    redirector.StatusPermanent,
			expected: "PERMANENT",
		},
		{
			name:     "unknown status",
			input:    redirector.Status(42),
			expected: "TEMPORARY",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStatus_StatusCode(t *testing.T) {
	tests := []struct {
		name     string
		input    redirector.Status
		expected int
	}{
		{
			name:     "temporary",
			input:    redirector.StatusTemporary,
			expected: 302,
		},
		{
			name:     "permanent",
			input:    redirector.StatusPermanent,
			expected: 301,
		},
		{
			name:     "unknown status",
			input:    redirector.Status(42),
			expected: 302,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.StatusCode()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStatus_MarshalDynamoDBAttributeValue(t *testing.T) {
	tests := []struct {
		name     string
		input    redirector.Status
		expected *types.AttributeValueMemberS
	}{
		{
			name:     "temporary",
			input:    redirector.StatusTemporary,
			expected: &types.AttributeValueMemberS{Value: "TEMPORARY"},
		},
		{
			name:     "permanent",
			input:    redirector.StatusPermanent,
			expected: &types.AttributeValueMemberS{Value: "PERMANENT"},
		},
		{
			name:     "invalid status",
			input:    redirector.Status(42),
			expected: &types.AttributeValueMemberS{Value: "TEMPORARY"},
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

func TestStatus_UnmarshalDynamoDBAttributeValue(t *testing.T) {
	tests := []struct {
		name     string
		input    types.AttributeValue
		expected redirector.Status
	}{
		{
			name:     "temporary",
			input:    &types.AttributeValueMemberS{Value: "TEMPORARY"},
			expected: redirector.StatusTemporary,
		},
		{
			name:     "permanent",
			input:    &types.AttributeValueMemberS{Value: "PERMANENT"},
			expected: redirector.StatusPermanent,
		},
		{
			name:     "unknown status",
			input:    &types.AttributeValueMemberS{Value: "FOO"},
			expected: redirector.DefaultStatus,
		},
		{
			name:     "incorrect status",
			input:    &types.AttributeValueMemberN{Value: "0"},
			expected: redirector.DefaultStatus,
		},
		{
			name:     "nil",
			input:    nil,
			expected: redirector.DefaultStatus,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s redirector.Status
			err := s.UnmarshalDynamoDBAttributeValue(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, s)
		})
	}
}
