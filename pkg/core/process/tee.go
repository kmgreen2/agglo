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
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
)

var TeeMetadataKey string = "agglo:tee:output"

// Tee is a process processor that will send a provided mapping to
// a system (i.e. KVStore, Pubsub, etc.) and return the map
type Tee struct {
	outputFunc func(key string, in map[string]interface{}) error
	condition *core.Condition
	transformer *Transformer
	outputType string
	connectionString string
}

// NewKVTee will create a Tee processor that stores maps in the provided KVStore
// Note: the returned map will contain the UUID of the KV entry with key "_uuid_key"
func NewKVTee(kvStore kvs.KVStore, condition *core.Condition, transformer *Transformer) *Tee {
	outputFunc := func(key string, in map[string]interface{}) error {
		byteBuffer := bytes.NewBuffer([]byte{})
		encoder := json.NewEncoder(byteBuffer)
		err := encoder.Encode(in)
		if err != nil {
			return err
		}
		return kvStore.Put(context.Background(), key, byteBuffer.Bytes())
	}

	if transformer == nil {
		transformation := core.NewTransformation(
			[]core.FieldTransformation{&core.CopyTransformation{}},
			core.TrueCondition)
		transformer = DefaultTransformer()
		transformer.AddSpec("", "", transformation)
	}
	return &Tee{
		outputFunc,
		condition,
		transformer,
		"kvstore",
		kvStore.ConnectionString(),
	}
}

// NewLocalfileTee will create a Tee processor that writes maps to a local file system
func NewLocalFileTee(path string, condition *core.Condition, transformer *Transformer) (*Tee, error) {
	if d, err := os.Stat(path); err != nil || !d.IsDir() {
		msg := fmt.Sprintf("'%s is not a valid path", path)
		return nil, common.NewInvalidError(msg)
	}
	outputFunc := func(key string, in map[string]interface{}) error {
		byteBuffer := bytes.NewBuffer([]byte{})
		encoder := json.NewEncoder(byteBuffer)
		err := encoder.Encode(in)
		if err != nil {
			return err
		}
		return ioutil.WriteFile(fmt.Sprintf("%s/%s.json", path, key), byteBuffer.Bytes(), 0644)
	}

	if transformer == nil {
		transformation := core.NewTransformation(
			[]core.FieldTransformation{&core.CopyTransformation{}},
			core.TrueCondition)
		transformer = DefaultTransformer()
		transformer.AddSpec("", "", transformation)
	}
	return &Tee{
		outputFunc,
		condition,
		transformer,
		"localfile",
		path,
	}, nil
}

// NewPubSubTee will create a Tee processor that publishes maps using the provided
// publisher.
func NewPubSubTee(publisher streaming.Publisher, condition *core.Condition, transformer *Transformer) *Tee {
	outputFunc := func(key string, in map[string]interface{}) error {
		byteBuffer := bytes.NewBuffer([]byte{})
		encoder := json.NewEncoder(byteBuffer)
		err := encoder.Encode(in)
		if err != nil {
			return err
		}
		return publisher.Publish(context.Background(), byteBuffer.Bytes())
	}

	if transformer == nil {
		transformation := core.NewTransformation(
			[]core.FieldTransformation{&core.CopyTransformation{}},
			core.TrueCondition)
		transformer = DefaultTransformer()
		transformer.AddSpec("", "", transformation)
	}
	return &Tee{
		outputFunc,
		condition,
		transformer,
		"pubsub",
		publisher.ConnectionString(),
	}
}

// NewHttpTee will create a tee processor that posts JSON-encoded
// maps to a specified endpoint
func NewHttpTee(client common.HTTPClient, url string, condition *core.Condition, transformer *Transformer) *Tee {
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

	if transformer == nil {
		transformation := core.NewTransformation(
			[]core.FieldTransformation{&core.CopyTransformation{}},
			core.TrueCondition)
		transformer = DefaultTransformer()
		transformer.AddSpec("", "", transformation)
	}
	return &Tee{
		outputFunc,
		condition,
		transformer,
		"web",
		url,
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

	uuid, err := gUuid.NewRandom()
	if err != nil {
		return nil, err
	}

	out := core.CopyableMap(in).DeepCopy()

	teeOut, err := t.transformer.Process(in)
	if err != nil {
		return nil, err
	}

	err = t.outputFunc(uuid.String(), teeOut)
	if err != nil {
		return nil, err
	}

	if _, ok := out[TeeMetadataKey]; !ok {
		out[TeeMetadataKey] = make([]map[string]string, 0)
	}

	switch outVal := out[TeeMetadataKey].(type) {
	case []map[string]string:
		out[TeeMetadataKey] = append(outVal, map[string]string{
			"uuid": uuid.String(),
			"outputType": t.outputType,
			"connectionString": t.connectionString,
		})
	default:
		msg := fmt.Sprintf("detected corrupted %s in map when teeing.  expected []map[string]string, got %v",
			TeeMetadataKey, reflect.TypeOf(outVal))
		return nil, common.NewInternalError(msg)
	}

	return out, nil

}
