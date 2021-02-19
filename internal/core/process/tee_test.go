package process_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/kmgreen2/agglo/internal/core"
	"github.com/kmgreen2/agglo/internal/core/process"
	"github.com/kmgreen2/agglo/pkg/kvs"
	"github.com/kmgreen2/agglo/pkg/storage"
	"github.com/kmgreen2/agglo/pkg/streaming"
	"github.com/kmgreen2/agglo/pkg/util"
	"github.com/kmgreen2/agglo/test"
	mocks "github.com/kmgreen2/agglo/test/mocks"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"reflect"
	"sync"
	"testing"
)


func TestHttpTee(t *testing.T) {
	jsonMap := test.TestJson()
	var checkErr error

	checkFunc := func(req *http.Request) {
		var bodyMap map[string]interface{}
		byteBuffer := bytes.NewBuffer([]byte{})
		_, err := byteBuffer.ReadFrom(req.Body)
		if err != nil {
			checkErr = err
		}
		decoder := json.NewDecoder(byteBuffer)
		err = decoder.Decode(&bodyMap)
		if err != nil {
			checkErr = err
		}
		if !util.CopyableMap(bodyMap).DeepCompare(jsonMap) {
			checkErr = fmt.Errorf("maps do not keepMatched: %v \n!=\n %v", bodyMap, jsonMap)
		}
		checkErr = nil
	}

	httpClient := test.NewMockHttpClient(nil, 200, checkFunc)

	httpTee := process.NewHttpTee("httpTee", httpClient, "foo", core.TrueCondition, nil, nil)

	out, err := httpTee.Process(context.Background(), jsonMap)
	assert.Nil(t, err)
	assert.Nil(t, checkErr)
	delete(out, process.TeeMetadataKey)
	assert.True(t, util.CopyableMap(out).DeepCompare(jsonMap))
}

func TestHttpTeeWithAdditional(t *testing.T) {
	jsonMap := test.TestJson()
	var checkErr error

	additionalBody := map[string]interface{}{
		"foo": "bar",
		"buzz": 1,
	}

	requestMap, err := util.MergeMaps(jsonMap, additionalBody)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	checkFunc := func(req *http.Request) {
		var bodyMap map[string]interface{}
		byteBuffer := bytes.NewBuffer([]byte{})
		_, err := byteBuffer.ReadFrom(req.Body)
		if err != nil {
			checkErr = err
		}
		decoder := json.NewDecoder(byteBuffer)
		err = decoder.Decode(&bodyMap)
		if err != nil {
			checkErr = err
		}
		if !util.CopyableMap(bodyMap).DeepCompare(requestMap) {
			checkErr = fmt.Errorf("maps do not keepMatched: %v \n!=\n %v", bodyMap, jsonMap)
		}
		checkErr = nil
	}

	httpClient := test.NewMockHttpClient(nil, 200, checkFunc)

	httpTee := process.NewHttpTee("httpTee", httpClient, "foo", core.TrueCondition, nil, additionalBody)

	out, err := httpTee.Process(context.Background(), jsonMap)
	assert.Nil(t, err)
	assert.Nil(t, checkErr)
	delete(out, process.TeeMetadataKey)
	assert.True(t, util.CopyableMap(out).DeepCompare(jsonMap))
}

func TestHttpTeeWithResponse(t *testing.T) {
	jsonMap := test.TestJson()
	var checkErr error

	checkFunc := func(req *http.Request) {
		var bodyMap map[string]interface{}
		byteBuffer := bytes.NewBuffer([]byte{})
		_, err := byteBuffer.ReadFrom(req.Body)
		if err != nil {
			checkErr = err
		}
		decoder := json.NewDecoder(byteBuffer)
		err = decoder.Decode(&bodyMap)
		if err != nil {
			checkErr = err
		}
		if !util.CopyableMap(bodyMap).DeepCompare(jsonMap) {
			checkErr = fmt.Errorf("maps do not keepMatched: %v \n!=\n %v", bodyMap, jsonMap)
		}
		checkErr = nil
	}

	respMap := map[string]interface{} {
		"foo": "bar",
		"bizz": "buzz",
	}

	respBytes, err := util.MapToJson(respMap)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	httpClient := test.NewMockHttpClientWithResponseBody(nil, 200,
		ioutil.NopCloser(bytes.NewBuffer(respBytes)),
		checkFunc)

	httpTee := process.NewHttpTee("httpTee", httpClient, "foo", core.TrueCondition, nil, nil)

	out, err := httpTee.Process(context.Background(), jsonMap)
	assert.Nil(t, err)
	assert.Nil(t, checkErr)

	if teeMetadata, ok := out[process.TeeMetadataKey].([]map[string]interface{}); ok {
		if len(teeMetadata) == 0 {
			assert.FailNow(t, "expected non-zero tee metadata array")
		}
		assert.True(t, util.CopyableMap(teeMetadata[0]["response"].(map[string]interface{})).DeepCompare(respMap))
	} else {
		assert.FailNow(t, fmt.Sprintf("invalid tee metadata: %v", reflect.TypeOf(out[process.TeeMetadataKey])))
	}
	delete(out, process.TeeMetadataKey)
	assert.True(t, util.CopyableMap(out).DeepCompare(jsonMap))
}

func TestHttpTeeError(t *testing.T) {
	jsonMap := test.TestJson()
	httpClient := test.NewMockHttpClient(fmt.Errorf("error"),500, func(req *http.Request){})
	httpTee := process.NewHttpTee("httpTee", httpClient, "foo", core.TrueCondition, nil, nil)

	out, err := httpTee.Process(context.Background(), jsonMap)
	assert.Nil(t, out)
	assert.Error(t, err)

	httpClient = test.NewMockHttpClient(nil, 500, func(req *http.Request){})
	httpTee = process.NewHttpTee("httpTee", httpClient, "foo", core.TrueCondition, nil, nil)
	out, err = httpTee.Process(context.Background(), jsonMap)
	assert.Nil(t, out)
	assert.Error(t, err)
}

func TestHttpFalseCondition(t *testing.T) {
	jsonMap := test.TestJson()
	httpClient := test.NewMockHttpClient(nil, 200, func(req *http.Request){})
	httpTee := process.NewHttpTee("httpTee", httpClient, "foo", core.FalseCondition, nil, nil)

	out, err := httpTee.Process(context.Background(), jsonMap)
	assert.Equal(t, jsonMap, out)
	assert.Nil(t, err)
}

func TestKVTee(t *testing.T) {
	var storedMap map[string]interface{}
	jsonMap := test.TestJson()
	kvStore := kvs.NewMemKVStore()
	kvTee := process.NewKVTee("kvTee", kvStore, core.TrueCondition, nil, nil)

	out, err := kvTee.Process(context.Background(), jsonMap)
	assert.Nil(t, err)

	teeMetadata := out[process.TeeMetadataKey].([]map[string]interface{})

	storedMapBytes, err := kvStore.Get(context.Background(), teeMetadata[0]["uuid"].(string))
	assert.Nil(t, err)

	decoder := json.NewDecoder(bytes.NewBuffer(storedMapBytes))
	err = decoder.Decode(&storedMap)
	assert.Nil(t, err)

	assert.True(t, util.CopyableMap(storedMap).DeepCompare(jsonMap))
	delete(out, process.TeeMetadataKey)
	assert.True(t, util.CopyableMap(out).DeepCompare(jsonMap))
}

func TestKVTeeWithAdditional(t *testing.T) {
	var storedMap map[string]interface{}
	jsonMap := test.TestJson()
	kvStore := kvs.NewMemKVStore()

	additionalBody := map[string]interface{}{
		"foo": "bar",
		"buzz": float64(1),
	}

	requestMap, err := util.MergeMaps(jsonMap, additionalBody)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	kvTee := process.NewKVTee("kvTee", kvStore, core.TrueCondition, nil, additionalBody)

	out, err := kvTee.Process(context.Background(), jsonMap)
	assert.Nil(t, err)

	teeMetadata := out[process.TeeMetadataKey].([]map[string]interface{})

	storedMapBytes, err := kvStore.Get(context.Background(), teeMetadata[0]["uuid"].(string))
	assert.Nil(t, err)

	decoder := json.NewDecoder(bytes.NewBuffer(storedMapBytes))
	err = decoder.Decode(&storedMap)
	assert.Nil(t, err)

	assert.True(t, util.CopyableMap(storedMap).DeepCompare(requestMap))
	delete(out, process.TeeMetadataKey)
	assert.True(t, util.CopyableMap(out).DeepCompare(jsonMap))
}

func TestKVTeeError(t *testing.T) {
	ctrl := gomock.NewController(t)
	jsonMap := test.TestJson()
	kvStore := mocks.NewMockKVStore(ctrl)
	kvStore.EXPECT().Put(context.Background(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("error"))
	kvStore.EXPECT().ConnectionString().Return("")

	kvTee := process.NewKVTee("kvTee", kvStore, core.TrueCondition,nil, nil)

	out, err := kvTee.Process(context.Background(), jsonMap)
	assert.Nil(t, out)
	assert.Error(t, err)
}

func TestKVTeeFalseCondition(t *testing.T) {
	ctrl := gomock.NewController(t)
	jsonMap := test.TestJson()
	kvStore := mocks.NewMockKVStore(ctrl)
	kvStore.EXPECT().ConnectionString().Return("")
	kvTee := process.NewKVTee("kvTee", kvStore, core.FalseCondition, nil, nil)

	out, err := kvTee.Process(context.Background(), jsonMap)
	assert.Equal(t, jsonMap, out)
	assert.Nil(t, err)
}

func TestPublisherTee(t *testing.T) {
	var storedMap map[string]interface{}
	lock := sync.Mutex{}
	jsonMap := test.TestJson()
	pubSub := streaming.NewMemPubSub()
	err := pubSub.CreateTopic("testing")
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	publisher, err := streaming.NewMemPublisher(pubSub, "testing")
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	subscriber, err := streaming.NewMemSubscriber(pubSub, "testing")
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	// Sync test checks with subscriber loop
	lock.Lock()

	handler := func (ctx context.Context, payload []byte) {
		decoder := json.NewDecoder(bytes.NewBuffer(payload))
		err = decoder.Decode(&storedMap)
		lock.Unlock()
	}

	err = subscriber.Subscribe(handler)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	publisherTee := process.NewPubSubTee("httpTee", publisher, core.TrueCondition, nil, nil)

	out, err := publisherTee.Process(context.Background(), jsonMap)
	assert.Nil(t, err)

	// Wait for subscriber
	lock.Lock()
	assert.True(t, util.CopyableMap(storedMap).DeepCompare(jsonMap))
	delete(out, process.TeeMetadataKey)
	assert.True(t, util.CopyableMap(out).DeepCompare(jsonMap))
}

func TestPublisherTeeWithAdditional(t *testing.T) {
	var storedMap map[string]interface{}
	lock := sync.Mutex{}
	jsonMap := test.TestJson()
	pubSub := streaming.NewMemPubSub()
	err := pubSub.CreateTopic("testing")
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	publisher, err := streaming.NewMemPublisher(pubSub, "testing")
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	subscriber, err := streaming.NewMemSubscriber(pubSub, "testing")
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	// Sync test checks with subscriber loop
	lock.Lock()

	handler := func (ctx context.Context, payload []byte) {
		decoder := json.NewDecoder(bytes.NewBuffer(payload))
		err = decoder.Decode(&storedMap)
		lock.Unlock()
	}

	err = subscriber.Subscribe(handler)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	additionalBody := map[string]interface{}{
		"foo": "bar",
		"buzz": float64(1),
	}

	requestMap, err := util.MergeMaps(jsonMap, additionalBody)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	publisherTee := process.NewPubSubTee("pubSubTee", publisher, core.TrueCondition, nil, additionalBody)

	out, err := publisherTee.Process(context.Background(), jsonMap)
	assert.Nil(t, err)

	// Wait for subscriber
	lock.Lock()
	assert.True(t, util.CopyableMap(storedMap).DeepCompare(requestMap))
	delete(out, process.TeeMetadataKey)
	assert.True(t, util.CopyableMap(out).DeepCompare(jsonMap))
}

func TestPublisherTeeError(t *testing.T) {
	ctrl := gomock.NewController(t)
	jsonMap := test.TestJson()
	publisher := mocks.NewMockPublisher(ctrl)
	publisher.EXPECT().Publish(context.Background(), gomock.Any()).Return(fmt.Errorf("error"))
	publisher.EXPECT().ConnectionString().Return("")

	publisherTee := process.NewPubSubTee("pubSubTee", publisher, core.TrueCondition, nil, nil)
	out, err := publisherTee.Process(context.Background(), jsonMap)
	assert.Nil(t, out)
	assert.Error(t, err)
}

func TestPublisherTeeFalseCondition(t *testing.T) {
	ctrl := gomock.NewController(t)
	jsonMap := test.TestJson()
	publisher := mocks.NewMockPublisher(ctrl)
	publisher.EXPECT().ConnectionString().Return("")
	publisherTee := process.NewPubSubTee("pubSubTee", publisher, core.FalseCondition, nil, nil)

	out, err := publisherTee.Process(context.Background(), jsonMap)
	assert.Equal(t, jsonMap, out)
	assert.Nil(t, err)
}

func TestObjectStoreTee(t *testing.T) {
	var storedMap map[string]interface{}
	jsonMap := test.TestJson()
	params, err := storage.NewMemObjectStoreBackendParams(storage.MemObjectStoreBackend, "test")
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	objectStore, err := storage.NewMemObjectStore(params)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	objectStoreTee := process.NewObjectStoreTee("objectStoreTee", objectStore, core.TrueCondition, nil, nil)

	out, err := objectStoreTee.Process(context.Background(), jsonMap)
	assert.Nil(t, err)

	teeMetadata := out[process.TeeMetadataKey].([]map[string]interface{})

	storedMapByteBuffer, err := objectStore.Get(context.Background(), teeMetadata[0]["uuid"].(string))
	assert.Nil(t, err)

	decoder := json.NewDecoder(storedMapByteBuffer)
	err = decoder.Decode(&storedMap)
	assert.Nil(t, err)

	assert.True(t, util.CopyableMap(storedMap).DeepCompare(jsonMap))
	delete(out, process.TeeMetadataKey)
	assert.True(t, util.CopyableMap(out).DeepCompare(jsonMap))
}

func TestObjectStoreTeeWithAdditional(t *testing.T) {
	var storedMap map[string]interface{}
	jsonMap := test.TestJson()
	params, err := storage.NewMemObjectStoreBackendParams(storage.MemObjectStoreBackend, "test")
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	objectStore, err := storage.NewMemObjectStore(params)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	additionalBody := map[string]interface{}{
		"foo": "bar",
		"buzz": float64(1),
	}

	requestMap, err := util.MergeMaps(jsonMap, additionalBody)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	objectStoreTee := process.NewObjectStoreTee("objectStoreTee", objectStore, core.TrueCondition, nil, additionalBody)

	out, err := objectStoreTee.Process(context.Background(), jsonMap)
	assert.Nil(t, err)

	teeMetadata := out[process.TeeMetadataKey].([]map[string]interface{})

	storedMapByteBuffer, err := objectStore.Get(context.Background(), teeMetadata[0]["uuid"].(string))
	assert.Nil(t, err)

	decoder := json.NewDecoder(storedMapByteBuffer)
	err = decoder.Decode(&storedMap)
	assert.Nil(t, err)

	assert.True(t, util.CopyableMap(storedMap).DeepCompare(requestMap))
	delete(out, process.TeeMetadataKey)
	assert.True(t, util.CopyableMap(out).DeepCompare(jsonMap))
}

func TestObjectStoreTeeError(t *testing.T) {
	ctrl := gomock.NewController(t)
	jsonMap := test.TestJson()
	objectStore := mocks.NewMockObjectStore(ctrl)
	objectStore.EXPECT().Put(context.Background(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("error"))
	objectStore.EXPECT().ConnectionString().Return("")

	objectStoreTee := process.NewObjectStoreTee("objectStoreTee", objectStore, core.TrueCondition,nil, nil)

	out, err := objectStoreTee.Process(context.Background(), jsonMap)
	assert.Nil(t, out)
	assert.Error(t, err)
}

func TestObjectStoreTeeFalseCondition(t *testing.T) {
	ctrl := gomock.NewController(t)
	jsonMap := test.TestJson()
	objectStore := mocks.NewMockObjectStore(ctrl)
	objectStore.EXPECT().ConnectionString().Return("")
	objectStoreTee := process.NewObjectStoreTee("objectStoreTee", objectStore, core.FalseCondition, nil, nil)

	out, err := objectStoreTee.Process(context.Background(), jsonMap)
	assert.Equal(t, jsonMap, out)
	assert.Nil(t, err)
}
