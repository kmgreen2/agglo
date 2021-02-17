package storage

import (
	"fmt"
	"github.com/kmgreen2/agglo/pkg/config"
	"github.com/kmgreen2/agglo/pkg/util"
	"strings"
)

// ObjectStoreConfig an in-memory object that represents the configuration for all object stores
type ObjectStoreConfig struct {
	*config.ConfigBase
	streamingBufferSize int
}

// NewObjectStoreConfig will reconcile configuration values from the environment and return a complete config object
func NewObjectStoreConfig(baseConfig *config.ConfigBase) (*ObjectStoreConfig, error) {
	osConfig := &ObjectStoreConfig{
		ConfigBase: baseConfig,
	}
	v, err := osConfig.GetAndValidate("streamingBufferSize", config.NoValidate)
	if err != nil {
		return nil, err
	}

	err = config.GetIntOrError(v, &osConfig.streamingBufferSize)
	if err != nil {
		return nil, err
	}
	return osConfig, nil
}

// SetStreamingBufferSize is a setter for streamingBufferSize
func (osConfig *ObjectStoreConfig) SetStreamingBufferSize(size int) error {
	osConfig.streamingBufferSize = size
	return nil
}

// GetStreamingBufferSize is a getter for streamingBufferSize
func (osConfig *ObjectStoreConfig) GetStreamingBufferSize() int {
	return osConfig.streamingBufferSize
}

// GetAccessKeyID
func (osConfig *ObjectStoreConfig) GetAccessKeyID(backendType BackendType, bucketName string) (string, error) {
	var varName string
	var accessKeyID string

	switch backendType {
	case MemObjectStoreBackend: varName = fmt.Sprintf("MemObjectStoreAccessKeyID_%s", bucketName)
	case S3ObjectStoreBackend: varName = fmt.Sprintf("S3ObjectStoreAccessKeyID_%s", bucketName)
	default:
		return "", util.NewInvalidError(fmt.Sprintf("invalid backend type: %d", backendType))
	}
	v, err := osConfig.GetAndValidate(varName, config.NotNil)
	if err != nil {
		return "", err
	}

	err = config.GetStringOrError(v, &accessKeyID)
	if err != nil {
		return "", err
	}

	return accessKeyID, nil
}

// GetSecretAccessKey
func (osConfig *ObjectStoreConfig) GetSecretAccessKey(backendType BackendType, bucketName string) (string, error) {
	var varName string
	var secretAccessKey string

	switch backendType {
	case MemObjectStoreBackend: varName = fmt.Sprintf("MemObjectStoreSecretAccessKey_%s", bucketName)
	case S3ObjectStoreBackend: varName = fmt.Sprintf("S3ObjectStoreSecretAccessKey_%s", bucketName)
	default:
		return "", util.NewInvalidError(fmt.Sprintf("invalid backend type: %d", backendType))
	}
	v, err := osConfig.GetAndValidate(varName, config.NotNil)
	if err != nil {
		return "", err
	}

	err = config.GetStringOrError(v, &secretAccessKey)
	if err != nil {
		return "", err
	}

	return secretAccessKey, nil
}

// UseSSL
func (osConfig *ObjectStoreConfig) UseSSL(backendType BackendType, bucketName string) bool {
	var varName string
	var useSSL string

	switch backendType {
	case MemObjectStoreBackend: varName = fmt.Sprintf("MemObjectStoreUseSSL_%s", bucketName)
	case S3ObjectStoreBackend: varName = fmt.Sprintf("S3ObjectStoreUseSSL_%s", bucketName)
	default:
		return false
	}
	v, err := osConfig.GetAndValidate(varName, config.NoValidate)
	if err != nil {
		return false
	}

	err = config.GetStringOrError(v, &useSSL)
	if err != nil {
		return false
	}

	return strings.Compare(useSSL, "true") == 0
}

