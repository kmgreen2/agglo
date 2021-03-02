package storage

import (
	"bytes"
	"cloud.google.com/go/storage"
	"context"
	"encoding/gob"
	"fmt"
	"github.com/kmgreen2/agglo/pkg/config"
	"github.com/kmgreen2/agglo/pkg/util"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"io"
	"strings"
)

type GCSObjectStore struct {
	client *storage.Client
	bucketName string
	config *ObjectStoreConfig
}

type GCSObjectStoreBackendParams struct {
	backendType BackendType
	bucketName string
}

func NewGCSObjectStoreBackendParamsFromConnectionString(backendType BackendType,
	connectionString string) (*GCSObjectStoreBackendParams, error) {
	var bucketName string
	connectionStringAry := strings.Split(connectionString, ",")

	for _, entry := range connectionStringAry {
		entryAry := strings.Split(entry, "=")
		if len(entryAry) != 2 {
			return nil, util.NewInvalidError(fmt.Sprintf("invalid entry in connection string: %s", entry))
		}
		switch entryAry[0] {
		case "bucketName":
			bucketName = entryAry[1]
		}
	}

	missingEntries := ""

	if len(bucketName) == 0 {
		missingEntries += "bucketName "
	}

	if len(missingEntries) > 0 {
		return nil, util.NewInvalidError(fmt.Sprintf("missing entries in connection string: %s", missingEntries))
	}

	return NewGCSObjectStoreBackendParams(backendType, bucketName)
}

func NewGCSObjectStoreBackendParams(backendType BackendType, bucketName string) (*GCSObjectStoreBackendParams, error) {
	if backendType != GCSObjectStoreBackend {
		return nil, util.NewInvalidError(fmt.Sprintf("NewGCSObjectStoreBackendParams - Invalid backendType: %v",
			backendType))
	}
	return &GCSObjectStoreBackendParams{
		backendType: backendType,
		bucketName: bucketName,
	}, nil
}

// NewGCSObjectStoreBackendParamsFromBytes will deserialize backend params and return an error if the params cannot
// be deserialized
func NewGCSObjectStoreBackendParamsFromBytes(payload []byte) (*GCSObjectStoreBackendParams, error) {
	params := &GCSObjectStoreBackendParams{}
	err := DeserializeGCSObjectStoreParams(payload, params)
	if err != nil {
		return nil, err
	}
	return params, nil
}

// Get will get a map of backend params
func (gcsObjectStoreParams *GCSObjectStoreBackendParams) Get() map[string]string {
	params := make(map[string]string)
	params["bucketName"] = gcsObjectStoreParams.bucketName
	return params
}

// GetBackendType will return the object store backend type
func (gcsObjectStoreParams *GCSObjectStoreBackendParams) GetBackendType() BackendType {
	return gcsObjectStoreParams.backendType
}

//  ConnectionString will return the connection string for the parameters
func (gcsObjectStoreParams *GCSObjectStoreBackendParams) ConnectionString() string {
	return fmt.Sprintf("gcs:bucketName=%s", gcsObjectStoreParams.bucketName)
}

// SerializeGCSObjectStoreParams will serialize the backend params and return an error if the params cannot be serialized
func SerializeGCSObjectStoreParams(gcsObjectStoreParams *GCSObjectStoreBackendParams) ([]byte, error) {
	byteBuffer := bytes.NewBuffer(make([]byte, 0))
	gEncoder := gob.NewEncoder(byteBuffer)
	err := gEncoder.Encode(gcsObjectStoreParams.backendType)
	if err != nil {
		return nil, err
	}
	err = gEncoder.Encode(gcsObjectStoreParams.bucketName)
	if err != nil {
		return nil, err
	}
	return byteBuffer.Bytes(), nil
}

// DeserializeGCSObjectStoreParams will deserialize the backend params and return an error if the params cannot be
//deserialized
func DeserializeGCSObjectStoreParams(payload []byte, gcsObjectStoreParams *GCSObjectStoreBackendParams) error {
	var backendType BackendType
	var bucketName string

	byteBuffer := bytes.NewBuffer(payload)
	gDecoder := gob.NewDecoder(byteBuffer)
	err := gDecoder.Decode(&backendType)
	if err != nil {
		return err
	}
	err = gDecoder.Decode(&bucketName)
	if err != nil {
		return err
	}
	gcsObjectStoreParams.backendType = backendType
	gcsObjectStoreParams.bucketName = bucketName
	return nil
}

func NewGCSObjectStore(params ObjectStoreBackendParams, opts ...option.ClientOption) (*GCSObjectStore, error) {
	var err error
	objectStore := &GCSObjectStore{
	}

	configBase, err := config.NewConfigBase()
	if err != nil {
		return nil, err
	}

	objectStore.config, err = NewObjectStoreConfig(configBase)
	if err != nil {
		return nil, err
	}

	objectStoreParams, ok := params.(*GCSObjectStoreBackendParams)
	if !ok {
		return nil, util.NewInvalidError(fmt.Sprintf("NewGCSObjectStore - invalid params"))
	}

	objectStore.client, err = storage.NewClient(context.Background(), opts...)
	if err != nil {
		return nil, err
	}

	objectStore.bucketName = objectStoreParams.bucketName

	return objectStore, nil
}

func (objStore *GCSObjectStore) ConnectionString() string {
	return fmt.Sprintf("gcs:bucketName=%s", objStore.bucketName)
}

func (objStore *GCSObjectStore) ObjectStoreBackendParams() ObjectStoreBackendParams {
	return &GCSObjectStoreBackendParams{
		backendType: GCSObjectStoreBackend,
		bucketName: objStore.bucketName,
	}
}

func (objStore *GCSObjectStore) Put(ctx context.Context, key string, reader io.Reader) error {
	wc := objStore.client.Bucket(objStore.bucketName).Object(key).NewWriter(ctx)
	if _, err := io.Copy(wc, reader); err != nil {
		return fmt.Errorf("io.Copy: %v", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %v", err)
	}
	return nil
}

func (objStore *GCSObjectStore) Get(ctx context.Context, key string) (io.Reader, error) {
	rc, err := objStore.client.Bucket(objStore.bucketName).Object(key).NewReader(ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			return nil, util.NewNotFoundError(err.Error())
		}
		return nil, err
	}
	return rc, nil
}

func (objStore *GCSObjectStore) Head(ctx context.Context, key string) error {
	var err error
	rc, err := objStore.client.Bucket(objStore.bucketName).Object(key).NewReader(ctx)
	defer func() {
		if rc != nil {
			_ = rc.Close()
		}
	}()
	if err != nil {
		if err == storage.ErrObjectNotExist {
			return util.NewNotFoundError(err.Error())
		}
		return err
	}
	return nil
}

func (objStore *GCSObjectStore) Delete(ctx context.Context, key string) error {
	err := objStore.client.Bucket(objStore.bucketName).Object(key).Delete(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (objStore *GCSObjectStore) List(ctx context.Context, prefix string) ([]string, error) {
	var objectKeys []string
	objects := objStore.client.Bucket(objStore.bucketName).Objects(ctx, &storage.Query{
		Prefix: prefix,
	})

	for {
		object, err := objects.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		objectKeys = append(objectKeys, object.Name)
	}

	return objectKeys, nil
}
