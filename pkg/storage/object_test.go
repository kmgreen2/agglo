package storage_test

import (
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/kmgreen2/agglo/pkg/storage"
	"github.com/stretchr/testify/assert"
	"testing"
	"github.com/kmgreen2/agglo/test/mocks"
)

func TestNewObjectDescriptorFromBytes(t *testing.T) {
	objectDescParams, err := storage.NewMemObjectStoreBackendParams(storage.MemObjectStoreBackend, "default")
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
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockObjectStoreParams := test.NewMockObjectStoreBackendParams(ctrl)

	mockObjectStoreParams.
		EXPECT().
		Serialize().
		Return(nil, fmt.Errorf("error"))

	mockObjectStoreParams.
		EXPECT().
		GetBackendType().
		Return(storage.BackendType(storage.MemObjectStoreBackend))

	objectDesc := storage.NewObjectDescriptor(mockObjectStoreParams, "foobar")

	_, err := objectDesc.Serialize()
	assert.Errorf(t, err, "error")
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
