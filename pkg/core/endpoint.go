package core

import (
	"github.com/kmgreen2/agglo/pkg/kvs"
	"github.com/kmgreen2/agglo/pkg/storage"
)

type Endpoint interface {
	Connect() error
}

type ObjectEndpoint struct {
	objectDescriptor storage.ObjectDescriptor
	objectStore storage.ObjectStore
}

func (objectEndpoint *ObjectEndpoint) Connect() error {
	objectStore, err := storage.NewObjectStore(objectEndpoint.objectDescriptor.GetParams())
	if err != nil {
		return err
	}
	objectEndpoint.objectStore = objectStore
	return nil
}

type KVEndpoint struct {
	kvEndpoint kvs.KVStore
}

func (kvEndpoint *KVEndpoint) Connect() error {
	return nil
}
