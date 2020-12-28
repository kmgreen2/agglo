package core_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/kmgreen2/agglo/pkg/core"
	"github.com/kmgreen2/agglo/pkg/kvs"
	"github.com/kmgreen2/agglo/pkg/streaming"
	"github.com/kmgreen2/agglo/test/mocks"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestKVTee(t *testing.T) {
	var storedMap map[string]interface{}
	jsonMap := testJson()
	kvStore := kvs.NewMemKVStore()
	kvTee := core.NewKVTee(kvStore)

	out, err := kvTee.Process(jsonMap)
	assert.Nil(t, err)

	storedMapBytes, err := kvStore.Get(context.Background(), out["_uuid_key"].(string))
	assert.Nil(t, err)

	decoder := json.NewDecoder(bytes.NewBuffer(storedMapBytes))
	err = decoder.Decode(&storedMap)
	assert.Nil(t, err)

	assert.True(t, core.CopyableMap(storedMap).DeepCompare(jsonMap))
}

func TestKVTeeError(t *testing.T) {
	ctrl := gomock.NewController(t)
	jsonMap := testJson()
	kvStore := test.NewMockKVStore(ctrl)
	kvTee := core.NewKVTee(kvStore)

	kvStore.EXPECT().Put(context.Background(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("error"))
	out, err := kvTee.Process(jsonMap)
	assert.Nil(t, out)
	assert.Error(t, err)
}

func TestPublisherTee(t *testing.T) {
	var storedMap map[string]interface{}
	lock := sync.Mutex{}
	jsonMap := testJson()
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

	publisherTee := core.NewPubSubTee(publisher)

	_, err = publisherTee.Process(jsonMap)
	assert.Nil(t, err)

	// Wait for subscriber
	lock.Lock()
	assert.True(t, core.CopyableMap(storedMap).DeepCompare(jsonMap))
}

func TestPublisherTeeError(t *testing.T) {
	ctrl := gomock.NewController(t)
	jsonMap := testJson()
	publisher := test.NewMockPublisher(ctrl)
	publisherTee := core.NewPubSubTee(publisher)

	publisher.EXPECT().Publish(context.Background(), gomock.Any()).Return(fmt.Errorf("error"))
	out, err := publisherTee.Process(jsonMap)
	assert.Nil(t, out)
	assert.Error(t, err)
}
