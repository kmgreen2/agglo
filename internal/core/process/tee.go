package process

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/internal/common"
	"github.com/kmgreen2/agglo/pkg/search"
	"github.com/kmgreen2/agglo/pkg/storage"
	"github.com/kmgreen2/agglo/pkg/util"
	"github.com/kmgreen2/agglo/internal/core"
	"github.com/kmgreen2/agglo/pkg/kvs"
	"github.com/kmgreen2/agglo/pkg/streaming"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"
)

var TeeMetadataKey string = string(common.TeeMetadataKey)

// Tee is a process processor that will send a provided mapping to
// a system (i.e. KVStore, Pubsub, etc.) and return the map
type Tee struct {
	name string
	outputFunc func(ctx context.Context, key string, in map[string]interface{}) (map[string]interface{}, error)
	condition *core.Condition
	transformer *Transformer
	outputType string
	connectionString string
	additionalBody map[string]interface{}
}

// NewKVTee will create a Tee processor that stores maps in the provided KVStore
// Note: the returned map will contain the UUID of the KV entry with key "_uuid_key"
func NewKVTee(name string, kvStore kvs.KVStore, condition *core.Condition, transformer *Transformer,
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
		name,
		outputFunc,
		condition,
		transformer,
		"kvstore",
		kvStore.ConnectionString(),
		additionalBody,
	}
}

// NewLocalfileTee will create a Tee processor that writes maps to a local file system
func NewLocalFileTee(name string ,path string, condition *core.Condition, transformer *Transformer,
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
		name,
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
func NewPubSubTee(name string, publisher streaming.Publisher, condition *core.Condition, transformer *Transformer,
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
		name,
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
func NewHttpTee(name string, client common.HTTPClient, url string, condition *core.Condition, transformer *Transformer,
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
				return nil, errors.Wrap(err, "NewHttpTee: error reading response body")
			}

			if len(bodyBytes) > 0 {
				respMap, err := util.JsonToMap(bodyBytes)
				if err != nil {
					return nil, errors.Wrap(err, "NewHttpTee: error converting response body to JSON")
				}
				return respMap, nil
			} else {
				return nil, nil
			}
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
		name,
		outputFunc,
		condition,
		transformer,
		"web",
		url,
		additionalBody,
	}
}

func NewSearchIndexTee(name string, searchIndex search.Index, condition *core.Condition, transformer *Transformer,
	additionalBody map[string]interface{}) *Tee {

	outputFunc := func(ctx context.Context, key string, in map[string]interface{}) (map[string]interface{}, error) {
		var err error
		blob := make(map[string]interface{})
		payload := in

		if len(additionalBody) > 0 {
			payload, err = util.MergeMaps(in, additionalBody)
			if err != nil {
				return nil, err
			}
		}

		builder := search.NewElasticIndexValueBuilder(key)

		// Only process first level of the dictionary as keyword or numeric values
		// All non-string/numeric keys beyond the first level will be added to
		// an un-indexed blob
		for k, v := range payload {
			switch _v := v.(type) {
			case string:
				// Detect if the string is a timestamp and, if so add as date
				if t, err := time.Parse(time.RFC3339, _v); err == nil {
					if strings.Compare(k, search.ElasticCreated) == 0 {
						builder = builder.SetCreated(t.UTC().Unix())
					} else {
						builder = builder.AddDate(k, t.UTC().Unix())
					}
				} else {
					// ToDo(KMG): May want to change the behavior here, but we should index as follows:
					//			  1. If the string is one token and less than 128 chars, then add as a keyword
					//			  2. If the string is multiple tokens, add as a text field
					//			  3. Else add as a blob (un-indexed)
					numTokens := len(strings.Split(_v, " "))
					if numTokens == 1 && len(_v) <= 128 {
						builder = builder.AddKeyword(k, _v)
					} else if numTokens > 1 {
						builder = builder.AddFreeText(k, _v)
					} else {
						blob[k] = _v
					}
				}
			case float64:
				builder = builder.AddNumeric(k, _v)
			case int:
			case int64:
			case uint:
			case uint64:
			case float32:
				builder = builder.AddNumeric(k, float64(_v))
			default:
				blob[k] = _v
			}
		}

		if len(blob) > 0 {
			blobBytes := bytes.NewBuffer([]byte{})
			encoder := json.NewEncoder(blobBytes)

			err = encoder.Encode(blob)
			if err != nil {
				return nil, err
			}

			builder.SetBlob(blobBytes.Bytes())
		}

		return nil, searchIndex.Put(ctx, key, builder.Get())
	}
	if transformer == nil {
		transformation := core.NewTransformation(
			[]core.FieldTransformation{&core.CopyTransformation{}},
			core.TrueCondition)
		transformer = DefaultTransformer()
		transformer.AddSpec("", "", transformation)
	}
	return &Tee{
		name,
		outputFunc,
		condition,
		transformer,
		"searchIndex",
		searchIndex.ConnectionString(),
		additionalBody,
	}
}

func NewObjectStoreTee(name string, objectStore storage.ObjectStore, condition *core.Condition,
	transformer *Transformer, additionalBody map[string]interface{}) *Tee {

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
		return nil, objectStore.Put(ctx, key, byteBuffer)
	}

	if transformer == nil {
		transformation := core.NewTransformation(
			[]core.FieldTransformation{&core.CopyTransformation{}},
			core.TrueCondition)
		transformer = DefaultTransformer()
		transformer.AddSpec("", "", transformation)
	}
	return &Tee{
		name,
		outputFunc,
		condition,
		transformer,
		"object",
		objectStore.ConnectionString(),
		additionalBody,
	}
}

func (t Tee) Name() string {
	return t.name
}

// Process processes an input map by sending it to the appropriate system and
// returns a copy of the provided map annotated with information about the backing system
func (t Tee) Process(ctx context.Context, in map[string]interface{}) (map[string]interface{}, error) {
	shouldTee, err := t.condition.Evaluate(in)
	if err != nil {
		return in, PipelineProcessError(t, err, "evaluating condition")
	}

	if !shouldTee {
		return in, nil
	}

	uuid, err := gUuid.NewRandom()
	if err != nil {
		return nil, PipelineProcessError(t, err, "generating UUID")
	}

	out := util.CopyableMap(in).DeepCopy()

	teeOut, err := t.transformer.Process(ctx, in)
	if err != nil {
		return nil, PipelineProcessError(t, err, "transforming fields")
	}

	respMap, err := t.outputFunc(ctx, uuid.String(), teeOut)
	if err != nil {
		return nil, PipelineProcessError(t, err, "running output function")
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
