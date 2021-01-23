package state_test

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"github.com/kmgreen2/agglo/pkg/state"
	"github.com/stretchr/testify/assert"
	"testing"
)

func intToBytes(x int) []byte {
	byteBuffer := bytes.NewBuffer([]byte{})
	encoder := gob.NewEncoder(byteBuffer)
	_ = encoder.Encode(x)
	return byteBuffer.Bytes()
}

func bytesToInt(intBytes []byte) int {
	var result int
	byteBuffer := bytes.NewBuffer(intBytes)
	decoder := gob.NewDecoder(byteBuffer)
	_ = decoder.Decode(&result)
	return result
}

func mapFunction(curr, val []byte) ([]byte, error) {
	if curr == nil {
		return val, nil
	}
	currInt := bytesToInt(curr)
	valInt := bytesToInt(val)
	return intToBytes(currInt+valInt), nil
}

func TestMemStateStoreHappyPath(t *testing.T) {
	stateStore := state.NewMemStateStore()

	for i := 0; i < 10; i ++ {
		key := fmt.Sprintf("%d", i)
		for j := 0; j < 10; j++ {
			for k := 0; k < 10; k++ {
				err := stateStore.Append(context.Background(), key, intToBytes(k + (j*10)))
				if err != nil {
					assert.FailNow(t, err.Error())
				}
			}
			err := stateStore.Checkpoint(context.Background(), key, mapFunction)
			if err != nil {
				assert.FailNow(t, err.Error())
			}
		}
		intBytes, err := stateStore.Get(context.Background(), key)
		if err != nil {
			assert.FailNow(t, err.Error())
		}
		assert.Equal(t, 4950, bytesToInt(intBytes))
	}
}
