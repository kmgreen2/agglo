package storage_test

import (
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

	objectDescBytes, err := storage.Serialize(objectDesc)
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

func TestNewObjectStoreFromConnectionString(t *testing.T) {
	objectStore, err := storage.NewObjectStoreFromConnectionString("mem:foo")
	assert.Nil(t, err)
	assert.Equal(t, "mem:foo", objectStore.ConnectionString())
	objectStore, err = storage.NewObjectStoreFromConnectionString("s3:endpoint=localhost:9000,bucketName=localtest")
	assert.Nil(t, err)
	assert.Equal(t, "s3:endpoint=localhost:9000,bucketName=localtest", objectStore.ConnectionString())
	objectStore, err = storage.NewObjectStoreFromConnectionString("s3:invalid")
	assert.Error(t, err)
	objectStore, err = storage.NewObjectStoreFromConnectionString("invalid:foo")
	assert.Error(t, err)
	objectStore, err = storage.NewObjectStoreFromConnectionString("invalid")
	assert.Error(t, err)
}
