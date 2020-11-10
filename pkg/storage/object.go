package storage

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/kmgreen2/agglo/pkg/common"
	"io"
)

// ObjectStore is the interface for an object store
type ObjectStore interface {
	Put(key string, reader io.Reader)	error
	Get(key string) (io.Reader, error)
	Head(key string) error
	Delete(key string) error
	List(prefix string) ([]string, error)
}

// BackendType is the type of backend
type BackendType int

const (
	// UnknownBackend is an undefined object store backend
	UnknownBackend = iota
	// MemObjectStoreBackend
	MemObjectStoreBackend
)

// ObjectStoreBackendParams is an interface whose implementation converts backend parameters into a map of strings
type ObjectStoreBackendParams interface {
	GetBackendType() BackendType
	Get() map[string]string
	Serialize() ([]byte, error)
	Deserialize([]byte) error
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
	err := desc.deserialize(payload)
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

// Serialize will serialize the object descriptor and return an error, if it cannot serialize
func (desc *ObjectDescriptor) Serialize() ([]byte, error) {
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
	paramsBytes, err := desc.backendParams.Serialize()
	if err != nil {
		return nil, err
	}
	err = gEncoder.Encode(paramsBytes)
	if err != nil {
		return nil, err
	}
	return byteBuffer.Bytes(), nil
}

// deserialize will serialize the object descriptor and return an error, if it cannot deserialize
func (desc *ObjectDescriptor) deserialize(descBytes []byte) error {
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
		return common.NewInvalidError(fmt.Sprintf("Deserialize - invalid backend type: %d", backendType))
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
		return nil, common.NewInvalidError(fmt.Sprintf("NewObjectStore - invalid backendType: %d",
			params.GetBackendType()))
	}
	return objectStore, nil
}
