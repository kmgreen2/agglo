package storage

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"github.com/kmgreen2/agglo/pkg/config"
	"github.com/kmgreen2/agglo/pkg/util"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
	"net/http"
	"strings"
)

type S3ObjectStore struct {
	client *minio.Client
	endpoint string
	bucketName string
	config *ObjectStoreConfig
}

type S3ObjectStoreCredentials struct {
	accessKeyID string
	secretAccessKey string
}

type S3ObjectStoreBackendParams struct {
	backendType BackendType
	endpoint string
	bucketName string
}

func NewS3ObjectStoreBackendParamsFromConnectionString(backendType BackendType,
	connectionString string) (*S3ObjectStoreBackendParams, error) {
	var endpoint, bucketName string
	connectionStringAry := strings.Split(connectionString, ",")

	for _, entry := range connectionStringAry {
		entryAry := strings.Split(entry, "=")
		if len(entryAry) != 2 {
			return nil, util.NewInvalidError(fmt.Sprintf("invalid entry in connection string: %s", entry))
		}
		switch entryAry[0] {
		case "endpoint":
			endpoint = entryAry[1]
		case "bucketName":
			bucketName = entryAry[1]
		}
	}

	missingEntries := ""

	if len(endpoint) == 0 {
		missingEntries += "endpoint "
	}

	if len(bucketName) == 0 {
		missingEntries += "bucketName "
	}

	if len(missingEntries) > 0 {
		return nil, util.NewInvalidError(fmt.Sprintf("missing entries in connection string: %s", missingEntries))
	}

	return NewS3ObjectStoreBackendParams(backendType, endpoint, bucketName)
}

func NewS3ObjectStoreBackendParams(backendType BackendType, endpoint,
	bucketName string) (*S3ObjectStoreBackendParams, error) {
	if backendType != S3ObjectStoreBackend {
		return nil, util.NewInvalidError(fmt.Sprintf("NewS3ObjectStoreBackendParams - Invalid backendType: %v",
			backendType))
	}
	return &S3ObjectStoreBackendParams{
		backendType: backendType,
		endpoint: endpoint,
		bucketName: bucketName,
	}, nil
}

// NewS3ObjectStoreBackendParamsFromBytes will deserialize backend params and return an error if the params cannot
// be deserialized
func NewS3ObjectStoreBackendParamsFromBytes(payload []byte) (*S3ObjectStoreBackendParams, error) {
	params := &S3ObjectStoreBackendParams{}
	err := DeserializeS3ObjectStoreParams(payload, params)
	if err != nil {
		return nil, err
	}
	return params, nil
}

// Get will get a map of backend params
func (s3ObjectStoreParams *S3ObjectStoreBackendParams) Get() map[string]string {
	params := make(map[string]string)
	params["bucketName"] = s3ObjectStoreParams.bucketName
	params["endpoint"] = s3ObjectStoreParams.endpoint
	return params
}

// GetBackendType will return the object store backend type
func (s3ObjectStoreParams *S3ObjectStoreBackendParams) GetBackendType() BackendType {
	return s3ObjectStoreParams.backendType
}

// SerializeS3ObjectStoreParams will serialize the backend params and return an error if the params cannot be serialized
func SerializeS3ObjectStoreParams(s3ObjectStoreParams *S3ObjectStoreBackendParams) ([]byte, error) {
	byteBuffer := bytes.NewBuffer(make([]byte, 0))
	gEncoder := gob.NewEncoder(byteBuffer)
	err := gEncoder.Encode(s3ObjectStoreParams.backendType)
	if err != nil {
		return nil, err
	}
	err = gEncoder.Encode(s3ObjectStoreParams.bucketName)
	if err != nil {
		return nil, err
	}
	err = gEncoder.Encode(s3ObjectStoreParams.endpoint)
	if err != nil {
		return nil, err
	}
	return byteBuffer.Bytes(), nil
}

// DeserializeS3ObjectStoreParams will deserialize the backend params and return an error if the params cannot be
//deserialized
func DeserializeS3ObjectStoreParams(payload []byte, s3ObjectStoreParams *S3ObjectStoreBackendParams) error {
	var backendType BackendType
	var endpoint, bucketName string

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
	err = gDecoder.Decode(&endpoint)
	if err != nil {
		return err
	}
	s3ObjectStoreParams.backendType = backendType
	s3ObjectStoreParams.endpoint = endpoint
	s3ObjectStoreParams.bucketName = bucketName
	return nil
}

func NewS3ObjectStore(params ObjectStoreBackendParams) (*S3ObjectStore, error) {
	var err error
	objectStore := &S3ObjectStore{
	}

	configBase, err := config.NewConfigBase()
	if err != nil {
		return nil, err
	}

	objectStore.config, err = NewObjectStoreConfig(configBase)
	if err != nil {
		return nil, err
	}

	objectStoreParams, ok := params.(*S3ObjectStoreBackendParams)
	if !ok {
		return nil, util.NewInvalidError(fmt.Sprintf("NewS3ObjectStore - invalid params"))
	}

	accessKeyID, err := objectStore.config.GetAccessKeyID(objectStoreParams.backendType, objectStoreParams.bucketName)
	if err != nil {
		return nil, err
	}

	secretAccessKey, err := objectStore.config.GetSecretAccessKey(objectStoreParams.backendType,
		objectStoreParams.bucketName)
	if err != nil {
		return nil, err
	}

	objectStore.client, err = minio.New(objectStoreParams.endpoint, &minio.Options{
		Creds: credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: objectStore.config.UseSSL(objectStoreParams.backendType, objectStoreParams.bucketName),
	})
	if err != nil {
		return nil, err
	}

	objectStore.bucketName = objectStoreParams.bucketName
	objectStore.endpoint = objectStoreParams.endpoint
	return objectStore, nil
}

func (objStore *S3ObjectStore) ConnectionString() string {
	return fmt.Sprintf("s3:endpoint=%s,bucketName=%s", objStore.endpoint, objStore.bucketName)
}

func (objStore *S3ObjectStore) Put(ctx context.Context, key string, reader io.Reader) error {
	_, err := objStore.client.PutObject(ctx, objStore.bucketName, key, reader, -1, minio.PutObjectOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (objStore *S3ObjectStore) Get(ctx context.Context, key string) (io.Reader, error) {
	objectInfo, err := objStore.client.GetObject(ctx, objStore.bucketName, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	_, err = objectInfo.Stat()
	if err != nil {
		if errResp, ok := err.(minio.ErrorResponse); ok {
			if errResp.StatusCode == http.StatusNotFound {
				return nil, util.NewNotFoundError(err.Error())
			}
		}
		return nil, err
	}
	return objectInfo, nil
}

func (objStore *S3ObjectStore) Head(ctx context.Context, key string) error {
	objectInfo, err := objStore.client.GetObject(ctx, objStore.bucketName, key, minio.GetObjectOptions{})
	if err != nil {
		return err
	}
	_, err = objectInfo.Stat()
	if err != nil {
		if errResp, ok := err.(minio.ErrorResponse); ok {
			if errResp.StatusCode == http.StatusNotFound {
				return util.NewNotFoundError(err.Error())
			}
		}
		return err
	}
	return objectInfo.Close()
}

func (objStore *S3ObjectStore) Delete(ctx context.Context, key string) error {
	err := objStore.client.RemoveObject(ctx, objStore.bucketName, key, minio.RemoveObjectOptions{})
	if err != nil {
		return  err
	}
	return nil
}

func (objStore *S3ObjectStore) List(ctx context.Context, prefix string) ([]string, error) {
	var objectKeys []string
	for objectInfo := range objStore.client.ListObjects(ctx, objStore.bucketName,
		minio.ListObjectsOptions{Prefix: prefix}) {
		objectKeys = append(objectKeys, objectInfo.Key)
	}
	return objectKeys, nil
}
