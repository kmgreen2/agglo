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
	UnknownBackend = iota
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
	backendParams ObjectStoreBackendParams
	backendKey string
}

func NewObjectDescriptor(backendParams ObjectStoreBackendParams, backendKey string) *ObjectDescriptor {
	return &ObjectDescriptor{
		backendParams: backendParams,
		backendKey: backendKey,
	}
}

func NewObjectDescriptorFromBytes(payload []byte) (*ObjectDescriptor, error) {
	desc := &ObjectDescriptor{
	}
	err := desc.Deserialize(payload)
	if err != nil {
		return nil, err
	}
	return desc, nil
}

func (desc *ObjectDescriptor) GetParams() ObjectStoreBackendParams {
	return desc.backendParams
}

func (desc *ObjectDescriptor) GetKey() string {
	return desc.backendKey
}

func (desc *ObjectDescriptor) Serialize() ([]byte, error) {
	byteBuffer := bytes.NewBuffer(make([]byte, 0))
	gEncoder := gob.NewEncoder(byteBuffer)
	err := gEncoder.Encode(desc.backendKey)
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

func (desc *ObjectDescriptor) Deserialize(descBytes []byte) error {
	var backendKey string
	var paramsBytes []byte
	byteBuffer := bytes.NewBuffer(descBytes)
	gDecoder := gob.NewDecoder(byteBuffer)
	err := gDecoder.Decode(&backendKey)
	if err != nil {
		return err
	}
	err = gDecoder.Decode(&paramsBytes)
	if err != nil {
		return err
	}
	desc.backendKey = backendKey
	err = desc.backendParams.Deserialize(paramsBytes)
	if err != nil {
		return err
	}
	return nil
}

func NewObjectStore(params ObjectStoreBackendParams) (ObjectStore, error) {
	var objectStore ObjectStore
	var err error
	if params.GetBackendType() == MemObjectStoreBackend {
		objectStore, err = NewMemObjectStore(params)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, common.NewInvalidError(fmt.Sprintf("NewObjectStore - invalid backendType: %s",
			params.GetBackendType()))
	}
	return objectStore, nil
}
