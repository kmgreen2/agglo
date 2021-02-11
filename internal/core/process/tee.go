package process

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/internal/common"
	"github.com/kmgreen2/agglo/pkg/util"
	"github.com/kmgreen2/agglo/internal/core"
	"github.com/kmgreen2/agglo/pkg/kvs"
	"github.com/kmgreen2/agglo/pkg/streaming"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
)

var TeeMetadataKey string = string(common.TeeMetadataKey)

// Tee is a process processor that will send a provided mapping to
// a system (i.e. KVStore, Pubsub, etc.) and return the map
type Tee struct {
	outputFunc func(ctx context.Context, key string, in map[string]interface{}) (map[string]interface{}, error)
	condition *core.Condition
	transformer *Transformer
	outputType string
	connectionString string
	additionalBody map[string]interface{}
}

// NewKVTee will create a Tee processor that stores maps in the provided KVStore
// Note: the returned map will contain the UUID of the KV entry with key "_uuid_key"
func NewKVTee(kvStore kvs.KVStore, condition *core.Condition, transformer *Transformer,
	additionalBody map[string]interface{}) *Tee {
	outputFunc := func(ctx context.Context, key string, in map[string]interface{}) (map[string]interface{}, error) {
		var err error
		payload := in
		byteBuffer := bytes.NewBuffer([]byte{})
		encoder := json.NewEncoder(byteBuffer)

		if len(additionalBody) > 0 {
			payload, err = util.MergeMaps(in, additionalBody)
			if err != nil {
				return nil, err
			}
		}

		err = encoder.Encode(payload)
		if err != nil {
			return nil, err
		}
		return nil, kvStore.Put(ctx, key, byteBuffer.Bytes())
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
		additionalBody,
	}
}

// NewLocalfileTee will create a Tee processor that writes maps to a local file system
func NewLocalFileTee(path string, condition *core.Condition, transformer *Transformer,
	additionalBody map[string]interface{}) (*Tee, error) {
	if d, err := os.Stat(path); err != nil || !d.IsDir() {
		msg := fmt.Sprintf("'%s is not a valid path", path)
		return nil, util.NewInvalidError(msg)
	}
	outputFunc := func(ctx context.Context, key string, in map[string]interface{}) (map[string]interface{}, error) {
		var err error
		payload := in
		byteBuffer := bytes.NewBuffer([]byte{})
		encoder := json.NewEncoder(byteBuffer)
		if len(additionalBody) > 0 {
			payload, err = util.MergeMaps(in, additionalBody)
			if err != nil {
				return nil, err
			}
		}
		err = encoder.Encode(payload)
		if err != nil {
			return nil, err
		}
		return nil, ioutil.WriteFile(fmt.Sprintf("%s/%s.json", path, key), byteBuffer.Bytes(), 0644)
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
		additionalBody,
	}, nil
}

// NewPubSubTee will create a Tee processor that publishes maps using the provided
// publisher.
func NewPubSubTee(publisher streaming.Publisher, condition *core.Condition, transformer *Transformer,
	additionalBody map[string]interface{}) *Tee {
	outputFunc := func(ctx context.Context, key string, in map[string]interface{}) (map[string]interface{}, error) {
		var err error
		payload := in
		byteBuffer := bytes.NewBuffer([]byte{})
		encoder := json.NewEncoder(byteBuffer)
		if len(additionalBody) > 0 {
			payload, err = util.MergeMaps(in, additionalBody)
			if err != nil {
				return nil, err
			}
		}
		err = encoder.Encode(payload)
		if err != nil {
			return nil, err
		}
		return nil, publisher.Publish(ctx, byteBuffer.Bytes())
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
		additionalBody,
	}
}

// NewHttpTee will create a tee processor that posts JSON-encoded
// maps to a specified endpoint
func NewHttpTee(client common.HTTPClient, url string, condition *core.Condition, transformer *Transformer,
	additionalBody map[string]interface{}) *Tee {
	outputFunc := func(ctx context.Context, key string, in map[string]interface{}) (map[string]interface{}, error) {
		var err error
		payload := in
		byteBuffer := bytes.NewBuffer([]byte{})
		encoder := json.NewEncoder(byteBuffer)
		if len(additionalBody) > 0 {
			payload, err = util.MergeMaps(in, additionalBody)
			if err != nil {
				return nil, err
			}
		}
		err = encoder.Encode(payload)
		if err != nil {
			return nil, err
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, byteBuffer)
		if err != nil {
			return nil, err
		}
		headers := http.Header{}
		headers.Set("Content-Type", "application/json")
		req.Header = headers

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode >= 300 {
			msg := fmt.Sprintf("error status %d posting to url '%s'", resp.StatusCode, url)
			return nil, util.NewInternalError(msg)
		}

		if resp.Body != nil {
			bodyBytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}

			respMap, err := util.JsonToMap(bodyBytes)
			if err != nil {
				return nil, err
			}

			return respMap, nil
		} else {
			return nil, nil
		}
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
		additionalBody,
	}
}

// Process processes an input map by sending it to the appropriate system and
// returns a copy of the provided map annotated with information about the backing system
func (t Tee) Process(ctx context.Context, in map[string]interface{}) (map[string]interface{}, error) {
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

	out := util.CopyableMap(in).DeepCopy()

	teeOut, err := t.transformer.Process(ctx, in)
	if err != nil {
		return nil, err
	}

	respMap, err := t.outputFunc(ctx, uuid.String(), teeOut)
	if err != nil {
		return nil, err
	}


	if _, ok := out[TeeMetadataKey]; !ok {
		out[TeeMetadataKey] = make([]map[string]interface{}, 0)
	}

	switch outVal := out[TeeMetadataKey].(type) {
	case []map[string]interface{}:
		teeOutputMap := map[string]interface{}{
			"uuid": uuid.String(),
			"outputType": t.outputType,
			"connectionString": t.connectionString,
		}
		if respMap != nil && len(respMap) > 0 {
			teeOutputMap["response"] = respMap
		}
		out[TeeMetadataKey] = append(outVal, teeOutputMap)
	default:
		msg := fmt.Sprintf("detected corrupted %s in map when teeing.  expected []map[string]string, got %v",
			TeeMetadataKey, reflect.TypeOf(outVal))
		return nil, util.NewInternalError(msg)
	}

	return out, nil

}
