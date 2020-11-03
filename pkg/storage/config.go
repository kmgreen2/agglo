package storage

import "github.com/kmgreen2/agglo/pkg/config"

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
