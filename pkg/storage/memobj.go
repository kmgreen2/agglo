package storage

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/kmgreen2/agglo/pkg/common"
	"github.com/kmgreen2/agglo/pkg/config"
	"io"
	"strings"
)

// MemObjectStore is a simple map-based implementation of an ObjectStore
type MemObjectStore struct {
	blobs map[string][]byte
	config *ObjectStoreConfig
}

type MemObjectStoreBackendParams struct {
	backendType BackendType
}

func NewMemObjectStoreBackendParams(backendType BackendType) (*MemObjectStoreBackendParams, error) {
	if backendType != MemObjectStoreBackend {
		return nil, common.NewInvalidError(fmt.Sprintf("NewMemObjectStoreBackendParams - Invalid backendType: %v",
			backendType))
	}
	return &MemObjectStoreBackendParams {
		backendType: backendType,
	}, nil
}

func NewMemObjectStoreBackendParamsFromBytes(payload []byte) (*MemObjectStoreBackendParams, error) {
	params := &MemObjectStoreBackendParams {
	}
	err := params.Deserialize(payload)
	if err != nil {
		return nil, err
	}
	return params, nil
}

func (memObjectStoreParams *MemObjectStoreBackendParams) Get() map[string]string {
	return nil
}

func (memObjectStoreParams *MemObjectStoreBackendParams) GetBackendType() BackendType {
	return memObjectStoreParams.backendType
}

func (memObjectStoreParams *MemObjectStoreBackendParams) Serialize() ([]byte, error) {
	byteBuffer := bytes.NewBuffer(make([]byte, 0))
	gEncoder := gob.NewEncoder(byteBuffer)
	err := gEncoder.Encode(memObjectStoreParams.backendType)
	if err != nil {
		return nil, err
	}
	return byteBuffer.Bytes(), nil
}

func (memObjectStoreParams *MemObjectStoreBackendParams) Deserialize(payload []byte) error {
	var backendType BackendType
	byteBuffer := bytes.NewBuffer(payload)
	gDecoder := gob.NewDecoder(byteBuffer)
	err := gDecoder.Decode(&backendType)
	if err != nil {
		return err
	}
	memObjectStoreParams.backendType = backendType
	return nil
}

// NewMemObjectStore returns a MemObjectStore object
func NewMemObjectStore(params ObjectStoreBackendParams) (*MemObjectStore, error) {
	configBase, err := config.NewConfigBase()
	if err != nil {
		return nil, err
	}
	osConfig, err := NewObjectStoreConfig(configBase)
	return &MemObjectStore{
		blobs: make(map[string][]byte),
		config: osConfig,
	}, nil
}

// Put will map the content read from a stream to a provided key and store the stream as a blob
func (objStore *MemObjectStore) Put(key string, reader io.Reader)	error {
	length := 0
	buf := make([]byte, objStore.config.streamingBufferSize)
	if _, ok := objStore.blobs[key]; ok {
		return common.NewConflictError(fmt.Sprintf("MemObjectStore - key exists: %s", key))
	}
	byteBuffer := bytes.NewBuffer(make([]byte, 0))
	for {
		numRead, err := reader.Read(buf)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
		numWritten, err := byteBuffer.Write(buf[:numRead])
		if err != nil {
			return err
		}
		if numWritten != numRead {
			return common.NewInternalError(fmt.Sprintf("MemObjectStore - error writing key: %s", key))
		}
		length += numRead
	}
	objStore.blobs[key] = byteBuffer.Bytes()[:length]
	return nil
}

// Get will return a reader that reads the stream of bytes associated with a keyed blob
func (objStore *MemObjectStore) Get(key string) (io.Reader, error) {
	if payload, ok := objStore.blobs[key]; ok {
		return bytes.NewBuffer(payload), nil
	}
	return nil, common.NewNotFoundError(fmt.Sprintf("MemObjectStore - key: %s", key))
}

// Head will return nil if the key maps a blob exists and an error otherwise
func (objStore *MemObjectStore) Head(key string) error {
	if _, ok := objStore.blobs[key]; ok {
		return nil
	}
	return common.NewNotFoundError(fmt.Sprintf("MemObjectStore - key: %s", key))
}

// Delete will delete the key and blob from the underlying map
func (objStore *MemObjectStore) Delete(key string) error {
	if _, ok := objStore.blobs[key]; ok {
		delete(objStore.blobs, key)
		return nil
	}
	return common.NewNotFoundError(fmt.Sprintf("MemObjectStore - key: %s", key))
}

// List will return a slice of keys that are prefixed on the provided prefix
func (objStore *MemObjectStore) List(prefix string) ([]string, error) {
	var result []string
	prefixLength := len(prefix)
	for s, _ := range objStore.blobs {
		if strings.Compare(prefix, s[:prefixLength]) == 0 {
			result = append(result, s)
		}
	}
	return result, nil
}
