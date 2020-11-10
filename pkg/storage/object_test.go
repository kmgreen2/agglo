package storage_test

import (
	"github.com/kmgreen2/agglo/pkg/storage"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewObjectDescriptorFromBytes(t *testing.T) {
	objectDescParams, err := storage.NewMemObjectStoreBackendParams(storage.MemObjectStoreBackend)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	objectDesc := storage.NewObjectDescriptor(objectDescParams, "foobar")

	objectDescBytes, err := objectDesc.Serialize()
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

func TestObjectDescriptorSerializeInvalid(t *testing.T) {
}

func TestNewObjectStore(t *testing.T) {
}

func TestNewObjectStoreInvalid(t *testing.T) {
}
