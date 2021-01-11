package process_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/kmgreen2/agglo/pkg/core"
	"github.com/kmgreen2/agglo/pkg/core/process"
	"github.com/kmgreen2/agglo/pkg/kvs"
	"github.com/kmgreen2/agglo/pkg/streaming"
	"github.com/kmgreen2/agglo/test"
	mocks "github.com/kmgreen2/agglo/test/mocks"
	"github.com/stretchr/testify/assert"
	"net/http"
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
		if !core.CopyableMap(bodyMap).DeepCompare(jsonMap) {
			checkErr = fmt.Errorf("maps do not keepMatched: %v \n!=\n %v", bodyMap, jsonMap)
		}
		checkErr = nil
	}

	httpClient := test.NewMockHttpClient(nil, 200, checkFunc)

	httpTee := process.NewHttpTee(httpClient, "foo", core.TrueCondition, nil)

	out, err := httpTee.Process(context.Background(), jsonMap)
	assert.Nil(t, err)
	assert.Nil(t, checkErr)
	delete(out, process.TeeMetadataKey)
	assert.True(t, core.CopyableMap(out).DeepCompare(jsonMap))
}

func TestHttpTeeError(t *testing.T) {
	jsonMap := test.TestJson()
	httpClient := test.NewMockHttpClient(fmt.Errorf("error"),500, func(req *http.Request){})
	httpTee := process.NewHttpTee(httpClient, "foo", core.TrueCondition, nil)

	out, err := httpTee.Process(context.Background(), jsonMap)
	assert.Nil(t, out)
	assert.Error(t, err)

	httpClient = test.NewMockHttpClient(nil, 500, func(req *http.Request){})
	httpTee = process.NewHttpTee(httpClient, "foo", core.TrueCondition, nil)
	out, err = httpTee.Process(context.Background(), jsonMap)
	assert.Nil(t, out)
	assert.Error(t, err)
}

func TestHttpFalseCondition(t *testing.T) {
	jsonMap := test.TestJson()
	httpClient := test.NewMockHttpClient(nil, 200, func(req *http.Request){})
	httpTee := process.NewHttpTee(httpClient, "foo", core.FalseCondition, nil)

	out, err := httpTee.Process(context.Background(), jsonMap)
	assert.Equal(t, jsonMap, out)
	assert.Nil(t, err)
}

func TestKVTee(t *testing.T) {
	var storedMap map[string]interface{}
	jsonMap := test.TestJson()
	kvStore := kvs.NewMemKVStore()
	kvTee := process.NewKVTee(kvStore, core.TrueCondition, nil)

	out, err := kvTee.Process(context.Background(), jsonMap)
	assert.Nil(t, err)

	teeMetadata := out[process.TeeMetadataKey].([]map[string]string)

	storedMapBytes, err := kvStore.Get(context.Background(), teeMetadata[0]["uuid"])
	assert.Nil(t, err)

	decoder := json.NewDecoder(bytes.NewBuffer(storedMapBytes))
	err = decoder.Decode(&storedMap)
	assert.Nil(t, err)

	assert.True(t, core.CopyableMap(storedMap).DeepCompare(jsonMap))
	delete(out, process.TeeMetadataKey)
	assert.True(t, core.CopyableMap(out).DeepCompare(jsonMap))
}

func TestKVTeeError(t *testing.T) {
	ctrl := gomock.NewController(t)
	jsonMap := test.TestJson()
	kvStore := mocks.NewMockKVStore(ctrl)
	kvStore.EXPECT().Put(context.Background(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("error"))
	kvStore.EXPECT().ConnectionString().Return("")

	kvTee := process.NewKVTee(kvStore, core.TrueCondition,nil)

	out, err := kvTee.Process(context.Background(), jsonMap)
	assert.Nil(t, out)
	assert.Error(t, err)
}

func TestKVTeeFalseCondition(t *testing.T) {
	ctrl := gomock.NewController(t)
	jsonMap := test.TestJson()
	kvStore := mocks.NewMockKVStore(ctrl)
	kvStore.EXPECT().ConnectionString().Return("")
	kvTee := process.NewKVTee(kvStore, core.FalseCondition, nil)

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

	publisherTee := process.NewPubSubTee(publisher, core.TrueCondition, nil)

	out, err := publisherTee.Process(context.Background(), jsonMap)
	assert.Nil(t, err)

	// Wait for subscriber
	lock.Lock()
	assert.True(t, core.CopyableMap(storedMap).DeepCompare(jsonMap))
	delete(out, process.TeeMetadataKey)
	assert.True(t, core.CopyableMap(out).DeepCompare(jsonMap))
}

func TestPublisherTeeError(t *testing.T) {
	ctrl := gomock.NewController(t)
	jsonMap := test.TestJson()
	publisher := mocks.NewMockPublisher(ctrl)
	publisher.EXPECT().Publish(context.Background(), gomock.Any()).Return(fmt.Errorf("error"))
	publisher.EXPECT().ConnectionString().Return("")

	publisherTee := process.NewPubSubTee(publisher, core.TrueCondition, nil)
	out, err := publisherTee.Process(context.Background(), jsonMap)
	assert.Nil(t, out)
	assert.Error(t, err)
}

func TestPublisherTeeFalseCondition(t *testing.T) {
	ctrl := gomock.NewController(t)
	jsonMap := test.TestJson()
	publisher := mocks.NewMockPublisher(ctrl)
	publisher.EXPECT().ConnectionString().Return("")
	publisherTee := process.NewPubSubTee(publisher, core.FalseCondition, nil)

	out, err := publisherTee.Process(context.Background(), jsonMap)
	assert.Equal(t, jsonMap, out)
	assert.Nil(t, err)
}
