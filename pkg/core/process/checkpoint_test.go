package process

import (
	"context"
	"fmt"
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/pkg/common"
	"github.com/kmgreen2/agglo/pkg/kvs"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestKVCheckPointer(t *testing.T) {
	messageID, err := gUuid.NewRandom()
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	pipelineName := "foo"
	kvStore := kvs.NewMemKVStore()
	checkpointer := NewKVCheckPointer(kvStore)

	in := map[string]interface{} {
		string(common.ResourceNameKey): pipelineName,
		string(common.MessageIDKey): messageID.String(),
	}

	for i := 0; i < 10; i++ {

		_, err = checkpointer.Process(context.Background(), in)
		if err != nil {
			assert.FailNow(t, err.Error())
		}

		stateBytes, err := kvStore.Get(context.Background(), fmt.Sprintf("%s:%s", pipelineName, messageID))
		if err != nil {
			assert.FailNow(t, err.Error())
		}

		m, err := common.JsonToMap(stateBytes)
		if err != nil {
			assert.FailNow(t, err.Error())
		}
		assert.Equal(t, float64(i), m[CheckpointIdxKey].(float64))
	}
	err = checkpointer.Finalize(context.Background(), in)
	assert.Nil(t, err)
}

func TestLocalFileCheckPointer(t *testing.T) {
	messageID, err := gUuid.NewRandom()
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	pipelineName := "foo"
	path := "/tmp/checkpoints"
	// Hack central: try to make dir then stat it.
	// ToDo(KMG): The best thing to do is refactor the local file
	// implementations to use a mock file system
	_ = os.Mkdir(path, 0755)

	if _, err := os.Stat(path); err != nil {
		assert.FailNow(t, err.Error())
	}

	checkpointer, err := NewLocalFileCheckPointer(path)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	in := map[string]interface{} {
		string(common.ResourceNameKey): pipelineName,
		string(common.MessageIDKey): messageID.String(),
	}

	for i := 0; i < 10; i++ {
		_, err = checkpointer.Process(context.Background(), in)
		if err != nil {
			assert.FailNow(t, err.Error())
		}

		stateBytes, err := ioutil.ReadFile(fmt.Sprintf("%s/%s:%s.json", path, pipelineName, messageID))
		if err != nil {
			assert.FailNow(t, err.Error())
		}

		m, err := common.JsonToMap(stateBytes)
		if err != nil {
			assert.FailNow(t, err.Error())
		}
		assert.Equal(t, float64(i), m[CheckpointIdxKey].(float64))
	}
	err = checkpointer.Finalize(context.Background(), in)
	assert.Nil(t, err)
}
