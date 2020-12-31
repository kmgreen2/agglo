package process

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/pkg/common"
	"github.com/kmgreen2/agglo/pkg/core"
	"github.com/kmgreen2/agglo/pkg/kvs"
	"github.com/kmgreen2/agglo/pkg/streaming"
	"net/http"
)

var uuidKey string = "_uuid_key"

// Tee is a process processor that will send a provided mapping to
// a system (i.e. KVStore, Pubsub, etc.) and return the map
type Tee struct {
	outputFunc func(key string, in map[string]interface{}) error
	condition *core.Condition
}

// NewKVTee will create a Tee processor that stores maps in the provided KVStore
// Note: the returned map will contain the UUID of the KV entry with key "_uuid_key"
func NewKVTee(kvStore kvs.KVStore, condition *core.Condition) *Tee {
	outputFunc := func(key string, in map[string]interface{}) error {
		byteBuffer := bytes.NewBuffer([]byte{})
		encoder := json.NewEncoder(byteBuffer)
		err := encoder.Encode(in)
		if err != nil {
			return err
		}
		return kvStore.Put(context.Background(), key, byteBuffer.Bytes())
	}
	return &Tee{
		outputFunc,
		condition,
	}
}

// NewPubSubTee will create a Tee processor that publishes maps using the provided
// publisher.
func NewPubSubTee(publisher streaming.Publisher, condition *core.Condition) *Tee {
	outputFunc := func(key string, in map[string]interface{}) error {
		byteBuffer := bytes.NewBuffer([]byte{})
		encoder := json.NewEncoder(byteBuffer)
		err := encoder.Encode(in)
		if err != nil {
			return err
		}
		return publisher.Publish(context.Background(), byteBuffer.Bytes())
	}
	return &Tee{
		outputFunc,
		condition,
	}
}

// NewHttpTee will create a tee processor that posts JSON-encoded
// maps to a specified endpoint
func NewHttpTee(client common.HTTPClient, url string, condition *core.Condition) *Tee {
	outputFunc := func(key string, in map[string]interface{}) error {
		byteBuffer := bytes.NewBuffer([]byte{})
		encoder := json.NewEncoder(byteBuffer)
		err := encoder.Encode(in)
		if err != nil {
			return err
		}
		req, err := http.NewRequest(http.MethodPost, url, byteBuffer)
		if err != nil {
			return err
		}
		headers := http.Header{}
		headers.Set("Content-Type", "application/json")
		req.Header = headers

		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		if resp.StatusCode >= 300 {
			msg := fmt.Sprintf("error status %d posting to url '%s'", resp.StatusCode, url)
			return common.NewInternalError(msg)
		}
		return nil
	}
	return &Tee{
		outputFunc,
		condition,
	}
}

// Process processes an input map by sending it to the appropriate system and
// returns a copy of the provided map annotated with information about the backing system
func (t Tee) Process(in map[string]interface{}) (map[string]interface{}, error) {
	shouldTee, err := t.condition.Evaluate(in)
	if err != nil {
		return in, err
	}

	if !shouldTee {
		return in, nil
	}

	uuid, err := gUuid.NewUUID()
	if err != nil {
		return nil, err
	}
	out := core.CopyableMap(in).DeepCopy()
	out[uuidKey] = uuid.String()

	err = t.outputFunc(uuid.String(), in)
	if err != nil {
		return nil, err
	}
	return out, nil
}
