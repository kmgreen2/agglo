package common_test

import (
	"fmt"
	"github.com/kmgreen2/agglo/internal/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

var reservedKeys = []common.InternalKey{common.PartitionIDKey, common.ResourceNameKey, common.AccumulatorKey,
	common.CheckpointIndexKey, common.CheckpointDataKey, common.TeeMetadataKey}

var reservedPrefixes = []common.InternalKeyPrefix{common.AggregationDataPrefix, common.CompletionStatusPrefix}

func TestInternalKeySetGet(t *testing.T) {
	in := make(map[string]interface{})

	for i, k := range reservedKeys {
		err := common.SetUsingInternalKey(k, fmt.Sprintf("%d", i), in, false)
		assert.Nil(t, err)
		val, ok := common.GetFromInternalKey(k, in)
		assert.True(t, ok)
		assert.Equal(t, fmt.Sprintf("%d", i), val)
		assert.True(t, common.ContainsReservedKey(in))
	}

	for i, k := range reservedKeys {
		err := common.SetUsingInternalKey(k, fmt.Sprintf("%d", i), in, false)
		assert.Error(t, err)
	}

	for _, k := range reservedKeys {
		err := common.SetUsingInternalKey(k, "foo", in, true)
		assert.Nil(t, err)
		val, ok := common.GetFromInternalKey(k, in)
		assert.True(t, ok)
		assert.Equal(t, "foo", val)
		assert.True(t, common.ContainsReservedKey(in))
	}
}

func TestInternalKeyPrefixSetGet(t *testing.T) {
	in := make(map[string]interface{})

	for i, k := range reservedPrefixes {
		err := common.SetUsingInternalPrefix(k, fmt.Sprintf("%d", i), fmt.Sprintf("%d", i), in,
			false)
		assert.Nil(t, err)
		val, ok := common.GetFromInternalPrefix(k, fmt.Sprintf("%d", i), in)
		assert.True(t, ok)
		assert.Equal(t, fmt.Sprintf("%d", i), val)
		assert.True(t, common.ContainsReservedKey(in))
	}

	for i, k := range reservedPrefixes {
		err := common.SetUsingInternalPrefix(k, fmt.Sprintf("%d", i), fmt.Sprintf("%d", i), in,
			false)
		assert.Error(t, err)
	}

	for i, k := range reservedPrefixes {
		err := common.SetUsingInternalPrefix(k, fmt.Sprintf("%d", i), "foo", in,
			true)
		assert.Nil(t, err)
		val, ok := common.GetFromInternalPrefix(k, fmt.Sprintf("%d", i), in)
		assert.True(t, ok)
		assert.Equal(t, "foo", val)
		assert.True(t, common.ContainsReservedKey(in))
	}
}

func TestContainsNotReservedKey(t *testing.T) {
	in := map[string]interface{} {
		"buzz": "foo",
		"foo": "bar",
	}

	assert.False(t, common.ContainsReservedKey(in))
}
