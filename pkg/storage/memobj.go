package storage

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/kmgreen2/agglo/pkg/util"
	"github.com/kmgreen2/agglo/pkg/config"
	"io"
	"strings"
	"sync"
)

// MemObjectStore is a simple map-based implementation of an ObjectStore
type MemObjectStore struct {
	blobs map[string][]byte
	config *ObjectStoreConfig
	instanceName string
	lock *sync.Mutex
}

// MemObjectStoreBackendParams are parameters used to access a MemObjectStore
type MemObjectStoreBackendParams struct {
	backendType BackendType
	instanceName string
}

// NewMemObjectStoreBackendParams will create and return a MemObjectStoreBackendParams object
func NewMemObjectStoreBackendParams(backendType BackendType, instanceName string) (*MemObjectStoreBackendParams,
	error) {
	if backendType != MemObjectStoreBackend {
		return nil, util.NewInvalidError(fmt.Sprintf("NewMemObjectStoreBackendParams - Invalid backendType: %v",
			backendType))
	}
	return &MemObjectStoreBackendParams {
		backendType: backendType,
		instanceName: instanceName,
	}, nil
}

// NewMemObjectStoreBackendParamsFromBytes will deserialize backend params and return an error if the params cannot
// be deserialized
func NewMemObjectStoreBackendParamsFromBytes(payload []byte) (*MemObjectStoreBackendParams, error) {
	params := &MemObjectStoreBackendParams{}
	err := DeserializeMemObjectStoreParams(payload, params)
	if err != nil {
		return nil, err
	}
	return params, nil
}

// Get will get a map of backend params
func (memObjectStoreParams *MemObjectStoreBackendParams) Get() map[string]string {
	params := make(map[string]string)
	params["instanceName"] = memObjectStoreParams.instanceName
	return params
}

// GetBackendType will return the object store backend type
func (memObjectStoreParams *MemObjectStoreBackendParams) GetBackendType() BackendType {
	return memObjectStoreParams.backendType
}

//  ConnectionString will return the connection string for the parameters
func (memObjectStoreParams *MemObjectStoreBackendParams) ConnectionString() string {
	return fmt.Sprintf("mem:instanceName=%s", memObjectStoreParams.instanceName)
}

// SerializeMemObjectStoreParams will serialize the backend params and return an error if the params cannot be serialized
func SerializeMemObjectStoreParams(memObjectStoreParams *MemObjectStoreBackendParams) ([]byte, error) {
	byteBuffer := bytes.NewBuffer(make([]byte, 0))
	gEncoder := gob.NewEncoder(byteBuffer)
	err := gEncoder.Encode(memObjectStoreParams.backendType)
	if err != nil {
		return nil, err
	}
	err = gEncoder.Encode(memObjectStoreParams.instanceName)
	if err != nil {
		return nil, err
	}
	return byteBuffer.Bytes(), nil
}

// DeserializeMemObjectStoreParams will deserialize the backend params and return an error if the params cannot be
//deserialized
func DeserializeMemObjectStoreParams(payload []byte, memObjectStoreParams *MemObjectStoreBackendParams) error {
	var backendType BackendType
	var instanceName string
	byteBuffer := bytes.NewBuffer(payload)
	gDecoder := gob.NewDecoder(byteBuffer)
	err := gDecoder.Decode(&backendType)
	if err != nil {
		return err
	}
	err = gDecoder.Decode(&instanceName)
	if err != nil {
		return err
	}
	memObjectStoreParams.backendType = backendType
	memObjectStoreParams.instanceName = instanceName
	return nil
}

// We implement the reference to the object store as a singleton.  This allows for two things:
// 1. Shared reference to the same in-memory object store in the same process (testing)
// 2. The individual object store implementations can manage their own object store references, which
//    allows us to instantiate object stores on-demand.  That is, each object has the object store parameters
//    stored as metadata and will instantiate via NewObjectStore(params ObjectStoreBackendParams), which is
//    only needed when accessing the object
var (
	memObjectStoreInstance map[string]*MemObjectStore
)

var instanceLock = &sync.Mutex{}

// NewMemObjectStore returns a MemObjectStore object
func NewMemObjectStore(params ObjectStoreBackendParams) (*MemObjectStore, error) {
	instanceLock.Lock()
	defer instanceLock.Unlock()

	if memObjectStoreInstance == nil {
		memObjectStoreInstance = make(map[string]*MemObjectStore)
	}

	memObjectStoreParams, ok := params.(*MemObjectStoreBackendParams)
	if !ok {
		return nil, util.NewInvalidError(fmt.Sprintf("NewMemObjectStore - invalid params"))
	}

	if _, ok := memObjectStoreInstance[memObjectStoreParams.instanceName]; !ok  {
		configBase, err := config.NewConfigBase()
		if err != nil {
			return nil, err
		}
		osConfig, err := NewObjectStoreConfig(configBase)
		memObjectStoreInstance[memObjectStoreParams.instanceName] = &MemObjectStore{
			blobs: make(map[string][]byte),
			instanceName: memObjectStoreParams.instanceName,
			config: osConfig,
			lock: &sync.Mutex{},
		}
	}
	return memObjectStoreInstance[memObjectStoreParams.instanceName], nil
}

func (objStore *MemObjectStore) ConnectionString() string {
	return fmt.Sprintf("mem:%s", objStore.instanceName)
}

func (objStore *MemObjectStore) ObjectStoreBackendParams() ObjectStoreBackendParams {
	return &MemObjectStoreBackendParams{
		backendType: MemObjectStoreBackend,
		instanceName: objStore.instanceName,
	}
}

// Put will map the content read from a stream to a provided key and store the stream as a blob
func (objStore *MemObjectStore) Put(ctx context.Context, key string, reader io.Reader)	error {
	objStore.lock.Lock()
	defer objStore.lock.Unlock()

	length := 0
	buf := make([]byte, objStore.config.streamingBufferSize)
	if _, ok := objStore.blobs[key]; ok {
		return util.NewConflictError(fmt.Sprintf("MemObjectStore - key exists: %s", key))
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
			return util.NewInternalError(fmt.Sprintf("MemObjectStore - error writing key: %s", key))
		}
		length += numRead
	}
	objStore.blobs[key] = byteBuffer.Bytes()[:length]
	return nil
}

// Get will return a reader that reads the stream of bytes associated with a keyed blob
func (objStore *MemObjectStore) Get(ctx context.Context, key string) (io.Reader, error) {
	objStore.lock.Lock()
	defer objStore.lock.Unlock()
	if payload, ok := objStore.blobs[key]; ok {
		return bytes.NewBuffer(payload), nil
	}
	return nil, util.NewNotFoundError(fmt.Sprintf("MemObjectStore - key: %s", key))
}

// Head will return nil if the key maps a blob exists and an error otherwise
func (objStore *MemObjectStore) Head(ctx context.Context, key string) error {
	objStore.lock.Lock()
	defer objStore.lock.Unlock()
	if _, ok := objStore.blobs[key]; ok {
		return nil
	}
	return util.NewNotFoundError(fmt.Sprintf("MemObjectStore - key: %s", key))
}

// Delete will delete the key and blob from the underlying map
func (objStore *MemObjectStore) Delete(ctx context.Context, key string) error {
	objStore.lock.Lock()
	defer objStore.lock.Unlock()
	if _, ok := objStore.blobs[key]; ok {
		delete(objStore.blobs, key)
		return nil
	}
	return util.NewNotFoundError(fmt.Sprintf("MemObjectStore - key: %s", key))
}

// List will return a slice of keys that are prefixed on the provided prefix
func (objStore *MemObjectStore) List(ctx context.Context, prefix string) ([]string, error) {
	var result []string
	prefixLength := len(prefix)
	objStore.lock.Lock()
	defer objStore.lock.Unlock()
	for s, _ := range objStore.blobs {
		if len(s) >= prefixLength && strings.Compare(prefix, s[:prefixLength]) == 0 {
			result = append(result, s)
		}
	}
	return result, nil
}

