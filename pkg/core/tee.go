package core

import (
	"bytes"
	"context"
	"encoding/json"
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/pkg/kvs"
	"github.com/kmgreen2/agglo/pkg/streaming"
)

var uuidKey string = "_uuid_key"

// Tee is a pipeline processor that will send a provided mapping to
// a system (i.e. KVStore, Pubsub, etc.) and return the map
type Tee struct {
	outputFunc func(key string, in map[string]interface{}) error
}

// NewKVTee will create a Tee processor that stores maps in the provided KVStore
// Note: the returned map will contain the UUID of the KV entry with key "_uuid_key"
func NewKVTee(kvStore kvs.KVStore) *Tee {
	outputFunc := func(key string, in map[string]interface{}) error {
		byteBuffer := bytes.NewBuffer([]byte{})
		encoder := json.NewEncoder(byteBuffer)
		err := encoder.Encode(in)
		if err != nil {
			return err
		}
		return kvStore.Put(context.Background(), key, byteBuffer.Bytes())
	}
	return &Tee {
		outputFunc,
	}
}

// NewPubSubTee will create a Tee processor that publishes maps using the provided
// publisher.
func NewPubSubTee(publisher streaming.Publisher) *Tee {
	outputFunc := func(key string, in map[string]interface{}) error {
		byteBuffer := bytes.NewBuffer([]byte{})
		encoder := json.NewEncoder(byteBuffer)
		err := encoder.Encode(in)
		if err != nil {
			return err
		}
		return publisher.Publish(context.Background(), byteBuffer.Bytes())
	}
	return &Tee {
		outputFunc,
	}
}

// Process processes an input map by sending it to the appropriate system and
// returns a copy of the provided map annotated with information about the backing system
func (t Tee) Process(in map[string]interface{}) (map[string]interface{}, error) {
	uuid, err := gUuid.NewUUID()
	if err != nil {
		return nil, err
	}
	out := CopyableMap(in).DeepCopy()
	out[uuidKey] = uuid.String()

	err = t.outputFunc(uuid.String(), in)
	if err != nil {
		return nil, err
	}
	return out, nil
}
