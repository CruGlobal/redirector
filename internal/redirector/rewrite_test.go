package redirector_test

import (
	"regexp"
	"testing"

	"github.com/CruGlobal/redirector/internal/redirector"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRewrite_Marshal(t *testing.T) {
	tests := []struct {
		name     string
		input    redirector.Rewrite
		expected *types.AttributeValueMemberM
	}{
		{
			name: "valid",
			input: redirector.Rewrite{
				RegExp:  redirector.RewriteRegexp{Regexp: regexp.MustCompile("^(.*)$")},
				Replace: "$1",
				Final:   true,
			},
			expected: &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
				"RegExp":  &types.AttributeValueMemberS{Value: "^(.*)$"},
				"Replace": &types.AttributeValueMemberS{Value: "$1"},
				"Final":   &types.AttributeValueMemberBOOL{Value: true},
			}},
		},
		{
			name: "empty",
			input: redirector.Rewrite{
				RegExp:  redirector.RewriteRegexp{Regexp: regexp.MustCompile("")},
				Replace: "",
				Final:   false,
			},
			expected: &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
				"RegExp": &types.AttributeValueMemberS{Value: ""},
				"Final":  &types.AttributeValueMemberBOOL{Value: false},
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := attributevalue.Marshal(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRewrite_Unmarshal(t *testing.T) {
	tests := []struct {
		name     string
		input    *types.AttributeValueMemberM
		expected redirector.Rewrite
	}{
		{
			name: "valid",
			input: &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
				"RegExp":  &types.AttributeValueMemberS{Value: "^(.*)$"},
				"Replace": &types.AttributeValueMemberS{Value: "$1"},
				"Final":   &types.AttributeValueMemberBOOL{Value: true},
			}},
			expected: redirector.Rewrite{
				RegExp:  redirector.RewriteRegexp{Regexp: regexp.MustCompile("^(.*)$")},
				Replace: "$1",
				Final:   true,
			},
		},
		{
			name: "invalid",
			input: &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
				"RegExp": &types.AttributeValueMemberS{Value: "[abc"},
				"Final":  &types.AttributeValueMemberBOOL{Value: false},
			}},
			expected: redirector.Rewrite{
				RegExp:  redirector.RewriteRegexp{Regexp: nil},
				Replace: "$1",
				Final:   false,
			},
		},
		{
			name: "missing",
			input: &types.AttributeValueMemberM{Value: map[string]types.AttributeValue{
				"RegExp":  &types.AttributeValueMemberN{Value: "1"},
				"Replace": &types.AttributeValueMemberS{Value: "foo"},
			}},
			expected: redirector.Rewrite{
				RegExp:  redirector.RewriteRegexp{Regexp: nil},
				Replace: "foo",
				Final:   true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result redirector.Rewrite
			err := attributevalue.Unmarshal(tt.input, &result)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
