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
	interCheckPointer := NewKVCheckPointer(kvStore)
	var intraCheckPointer IntraProcessCheckPointer = interCheckPointer

	in := map[string]interface{} {
		string(common.ResourceNameKey): pipelineName,
		string(common.MessageIDKey): messageID.String(),
	}

	ctx := context.Background()

	for i := 0; i < 10; i++ {
		intraCheckpoint := NewIntraProcessCheckpoint(nil, map[string]interface{}{"foo": i},
			"intracheckpointkey")
		ctx, err = intraCheckPointer.SetIntraProcessCheckpoint(ctx, intraCheckpoint)
		if err != nil {
			assert.FailNow(t, err.Error())
		}

		intraStateBytes, err := kvStore.Get(ctx, "intracheckpointkey")
		if err != nil {
			assert.FailNow(t, err.Error())
		}

		storedIntraCheckpoint, err := NewIntraProcessCheckpointFromBytes(intraStateBytes)
		if err != nil {
			assert.FailNow(t, err.Error())
		}
		assert.Equal(t, float64(i), storedIntraCheckpoint.Update["foo"])

		_, err = interCheckPointer.Process(ctx, in)
		if err != nil {
			assert.FailNow(t, err.Error())
		}

		_, err = kvStore.Get(ctx, "intracheckpointkey")
		if err == nil {
			assert.FailNow(t, "intra checkpoint state should have been cleared")
		}

		stateBytes, err := kvStore.Get(ctx, fmt.Sprintf("%s:%s", pipelineName, messageID))
		if err != nil {
			assert.FailNow(t, err.Error())
		}

		m, err := common.JsonToMap(stateBytes)
		if err != nil {
			assert.FailNow(t, err.Error())
		}
		assert.Equal(t, float64(i), m[CheckpointIdxKey].(float64))
	}
	err = interCheckPointer.Finalize(context.Background(), in)
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

	interCheckPointer, err := NewLocalFileCheckPointer(path)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	var intraCheckPointer IntraProcessCheckPointer = interCheckPointer

	in := map[string]interface{} {
		string(common.ResourceNameKey): pipelineName,
		string(common.MessageIDKey): messageID.String(),
	}

	ctx := context.Background()

	for i := 0; i < 10; i++ {
		intraCheckpoint := NewIntraProcessCheckpoint(nil, map[string]interface{}{"foo": i},
			"intracheckpointkey")
		ctx, err = intraCheckPointer.SetIntraProcessCheckpoint(ctx, intraCheckpoint)
		if err != nil {
			assert.FailNow(t, err.Error())
		}

		intraStateBytes, err := ioutil.ReadFile(fmt.Sprintf("%s/%s.json", path, "intracheckpointkey"))
		if err != nil {
			assert.FailNow(t, err.Error())
		}

		storedIntraCheckpoint, err := NewIntraProcessCheckpointFromBytes(intraStateBytes)
		if err != nil {
			assert.FailNow(t, err.Error())
		}
		assert.Equal(t, float64(i), storedIntraCheckpoint.Update["foo"])
		_, err = interCheckPointer.Process(ctx, in)
		if err != nil {
			assert.FailNow(t, err.Error())
		}

		_, err = ioutil.ReadFile(fmt.Sprintf("%s/%s.json", path, "intracheckpointkey"))
		if err == nil {
			assert.FailNow(t, "intra checkpoint state should have been cleared")
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
	err = interCheckPointer.Finalize(context.Background(), in)
	assert.Nil(t, err)
}
