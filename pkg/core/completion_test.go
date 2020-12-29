package core

import (
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/pkg/kvs"
	"github.com/kmgreen2/agglo/test"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestNewCompletion(t *testing.T) {
	partitionID, err := gUuid.NewUUID()
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	maps, joinedKeys := test.GetJoinedMaps(24, 16, partitionID, "foo")

	completion := NewCompletion("foo", partitionID, joinedKeys, -1, false)

	kvStore := kvs.NewMemKVStore()

	completer := NewCompleter(completion, kvStore)

	numTriggered := 0
	numComplete := 0
	for _, m := range maps {
		out, err := completer.Process(m)
		if err != nil {
			assert.FailNow(t, err.Error())
		}
		if val, ok := out["agglo:completion:foo"]; ok {
			if strings.Compare(val.(string), "complete") == 0 {
				numComplete++
			}
			if strings.Compare(val.(string), "triggered") == 0 {
				numTriggered++
			}
		}
	}
	assert.Equal(t, 1, numComplete)
	assert.Equal(t, 15, numTriggered)
}