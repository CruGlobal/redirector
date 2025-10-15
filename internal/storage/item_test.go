package storage_test

import (
	"testing"
	"time"

	"github.com/CruGlobal/redirector/internal/storage"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestItem_Item(t *testing.T) {
	testTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	testcases := []struct {
		name string
		item *storage.Item
		want map[string]types.AttributeValue
	}{
		{
			name: "empty item",
			item: &storage.Item{Key: "test"},
			want: map[string]types.AttributeValue{
				"Key": &types.AttributeValueMemberS{Value: "test"},
			},
		},
		{
			name: "full item",
			item: &storage.Item{
				Key:      "test",
				Contents: []byte("content"),
				Modified: &testTime,
				Size:     7,
			},
			want: map[string]types.AttributeValue{
				"Key":      &types.AttributeValueMemberS{Value: "test"},
				"Contents": &types.AttributeValueMemberB{Value: []byte("content")},
				"Modified": &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
				"Size":     &types.AttributeValueMemberN{Value: "7"},
			},
		},
		{
			name: "only contents",
			item: &storage.Item{
				Key:      "test",
				Contents: []byte("content"),
			},
			want: map[string]types.AttributeValue{
				"Key":      &types.AttributeValueMemberS{Value: "test"},
				"Contents": &types.AttributeValueMemberB{Value: []byte("content")},
			},
		},
		{
			name: "only modified",
			item: &storage.Item{
				Key:      "test",
				Modified: &testTime,
			},
			want: map[string]types.AttributeValue{
				"Key":      &types.AttributeValueMemberS{Value: "test"},
				"Modified": &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
			},
		},
		{
			name: "only size",
			item: &storage.Item{
				Key:  "test",
				Size: 7,
			},
			want: map[string]types.AttributeValue{
				"Key":  &types.AttributeValueMemberS{Value: "test"},
				"Size": &types.AttributeValueMemberN{Value: "7"},
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.item.Item()
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestItem_Load(t *testing.T) {
	testTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	testcases := []struct {
		name    string
		input   map[string]types.AttributeValue
		want    *storage.Item
		wantErr bool
	}{
		{
			name: "empty item",
			input: map[string]types.AttributeValue{
				"Key": &types.AttributeValueMemberS{Value: "test"},
			},
			want: &storage.Item{Key: "test"},
		},
		{
			name: "full item",
			input: map[string]types.AttributeValue{
				"Key":      &types.AttributeValueMemberS{Value: "test"},
				"Contents": &types.AttributeValueMemberB{Value: []byte("content")},
				"Modified": &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
				"Size":     &types.AttributeValueMemberN{Value: "7"},
			},
			want: &storage.Item{
				Key:      "test",
				Contents: []byte("content"),
				Modified: &testTime,
				Size:     7,
			},
		},
		{
			name: "invalid modified time",
			input: map[string]types.AttributeValue{
				"Key":      &types.AttributeValueMemberS{Value: "test"},
				"Modified": &types.AttributeValueMemberS{Value: "invalid"},
			},
			wantErr: true,
		},
		{
			name: "invalid size",
			input: map[string]types.AttributeValue{
				"Key":  &types.AttributeValueMemberS{Value: "test"},
				"Size": &types.AttributeValueMemberS{Value: "invalid"},
			},
			wantErr: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			got := &storage.Item{}
			err := got.Load(tc.input)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}
