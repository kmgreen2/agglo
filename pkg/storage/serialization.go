package storage

import (
	"fmt"
	"github.com/kmgreen2/agglo/pkg/util"
)

// Serialize will serialize an object; otherwise return an error
func Serialize(object interface{}) ([]byte, error) {
	var err error
	switch v := object.(type) {
	case *ObjectDescriptor:
		return SerializeObjectDescriptor(v)
	case *MemObjectStoreBackendParams:
		return SerializeMemObjectStoreParams(v)
	default:
		err = util.NewInvalidError(fmt.Sprintf("Serialize - invalid type: %T", v))
	}
	return nil, err
}

// Deserialize will deserialize an object from bytes; otherwise return an error
func Deserialize(objectBytes []byte, target interface{}) error {
	var err error
	switch v := target.(type) {
	case *ObjectDescriptor:
		err = DeserializeObjectDescriptor(objectBytes, v)
	case *MemObjectStoreBackendParams:
		err = DeserializeMemObjectStoreParams(objectBytes, v)
	default:
		err = util.NewInvalidError(fmt.Sprintf("Deserialize - invalid type: %T", v))
	}

	return err
}
