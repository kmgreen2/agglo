package storage_test

import (
	"github.com/kmgreen2/agglo/pkg/serialization"
	"github.com/kmgreen2/agglo/pkg/storage"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewObjectDescriptorFromBytes(t *testing.T) {
	objectDescParams, err := storage.NewMemObjectStoreBackendParams(storage.MemObjectStoreBackend, "default")
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	objectDesc := storage.NewObjectDescriptor(objectDescParams, "foobar")

	objectDescBytes, err := serialization.Serialize(objectDesc)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	objectDescFromBytes, err := storage.NewObjectDescriptorFromBytes(objectDescBytes)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	assert.Equal(t, objectDesc.GetParams().GetBackendType(), objectDescFromBytes.GetParams().GetBackendType())
	assert.Equal(t, objectDesc.GetKey(), objectDescFromBytes.GetKey())
}

func TestNewObjectDescriptorFromBytesInvalid(t *testing.T) {
	_, err := storage.NewObjectDescriptorFromBytes([]byte{})
	assert.Error(t, err)
}

func TestNewObjectStore(t *testing.T) {
	objectStoreParams, err := storage.NewMemObjectStoreBackendParams(storage.MemObjectStoreBackend, "default")
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	_, err = storage.NewObjectStore(objectStoreParams)
	assert.Nil(t, err)
}

func TestNewObjectStoreParamsInvalid(t *testing.T) {
	_, err := storage.NewMemObjectStoreBackendParams(storage.UnknownBackend, "default")
	assert.Error(t, err)
}
