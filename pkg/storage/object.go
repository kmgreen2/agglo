package storage

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"github.com/kmgreen2/agglo/pkg/util"
	"io"
)

// ObjectStore is the interface for an object store
type ObjectStore interface {
	Put(ctx context.Context, key string, reader io.Reader)	error
	Get(ctx context.Context, key string) (io.Reader, error)
	Head(ctx context.Context, key string) error
	Delete(ctx context.Context, key string) error
	List(ctx context.Context, prefix string) ([]string, error)
}

// BackendType is the type of backend
type BackendType int

const (
	// UnknownBackend is an undefined object store backend
	UnknownBackend = iota
	// NilBackend
	NilBackend
	// MemObjectStoreBackend
	MemObjectStoreBackend
)

// ObjectStoreBackendParams is an interface whose implementation converts backend parameters into a map of strings
type ObjectStoreBackendParams interface {
	GetBackendType() BackendType
	Get() map[string]string
}

// NilObjectStoreBackendParams is a hack to get serialization to work for genesis messages
type NilObjectStoreBackendParams struct {
}

// GetBackendType returns NilBackend
func (backendParams *NilObjectStoreBackendParams) GetBackendType() BackendType {
	return NilBackend
}

// Get returns nil
func (backendParams *NilObjectStoreBackendParams) Get() map[string]string {
	return nil
}

// ObjectDescriptor contains all of the information needed to access an object from an object store
type ObjectDescriptor struct {
	backendType BackendType
	backendParams ObjectStoreBackendParams
	backendKey string
}

// NewObjectDescriptor returns a new object descriptor
func NewObjectDescriptor(backendParams ObjectStoreBackendParams, backendKey string) *ObjectDescriptor {
	return &ObjectDescriptor{
		backendType: backendParams.GetBackendType(),
		backendParams: backendParams,
		backendKey: backendKey,
	}
}

// NewObjectDescriptorFromBytes returns a new object descriptor deserialized from a byte slice.  If a descriptor
// cannot be deserialized, then it returns an error
func NewObjectDescriptorFromBytes(payload []byte) (*ObjectDescriptor, error) {
	desc := &ObjectDescriptor{
	}
	 err := DeserializeObjectDescriptor(payload, desc)
	if err != nil {
		return nil, err
	}
	return desc, nil
}

// GetParams will return the backend-specific parameters for this object
func (desc *ObjectDescriptor) GetParams() ObjectStoreBackendParams {
	return desc.backendParams
}

// GetKey will return the key used to address an object in the object store
func (desc *ObjectDescriptor) GetKey() string {
	return desc.backendKey
}

// SerializeObjectDescriptor will serialize the object descriptor and return an error, if it cannot serialize
func SerializeObjectDescriptor(desc *ObjectDescriptor) ([]byte, error) {
	byteBuffer := bytes.NewBuffer(make([]byte, 0))
	gEncoder := gob.NewEncoder(byteBuffer)
	err := gEncoder.Encode(desc.backendKey)
	if err != nil {
		return nil, err
	}
	err = gEncoder.Encode(desc.backendType)
	if err != nil {
		return nil, err
	}

	if _, ok := desc.backendParams.(*NilObjectStoreBackendParams); ok {
		return byteBuffer.Bytes(), nil
	}

	if params, ok := desc.backendParams.(*MemObjectStoreBackendParams); ok {
		paramsBytes, err := SerializeMemObjectStoreParams(params)
		if err != nil {
			return nil, err
		}
		err = gEncoder.Encode(paramsBytes)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, util.NewInvalidError(fmt.Sprintf("Deserialize - invalid backend params type: %T",
			desc.backendParams))
	}
	return byteBuffer.Bytes(), nil
}

// DeserializeObjectDescriptor will serialize the object descriptor and return an error, if it cannot deserialize
func DeserializeObjectDescriptor(descBytes []byte, desc *ObjectDescriptor) error {
	var backendKey string
	var backendType BackendType
	var paramsBytes []byte
	byteBuffer := bytes.NewBuffer(descBytes)
	gDecoder := gob.NewDecoder(byteBuffer)
	err := gDecoder.Decode(&backendKey)
	if err != nil {
		return err
	}
	desc.backendKey = backendKey

	err = gDecoder.Decode(&backendType)
	if err != nil {
		return err
	}
	desc.backendType = backendType

	if backendType == NilBackend {
		return nil
	}

	err = gDecoder.Decode(&paramsBytes)
	if err != nil {
		return err
	}
	if backendType == MemObjectStoreBackend {
		desc.backendParams, err = NewMemObjectStoreBackendParamsFromBytes(paramsBytes)
		if err != nil {
			return err
		}
	} else {
		return util.NewInvalidError(fmt.Sprintf("Deserialize - invalid backend type: %d", backendType))
	}
	return nil
}

// NewObjectStore will return an object that is used to access an object store
func NewObjectStore(params ObjectStoreBackendParams) (ObjectStore, error) {
	var objectStore ObjectStore
	var err error
	if params.GetBackendType() == MemObjectStoreBackend {
		objectStore, err = NewMemObjectStore(params)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, util.NewInvalidError(fmt.Sprintf("NewObjectStore - invalid backendType: %d",
			params.GetBackendType()))
	}
	return objectStore, nil
}
