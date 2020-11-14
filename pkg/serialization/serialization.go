package serialization

import (
	"fmt"
	"github.com/kmgreen2/agglo/pkg/common"
	"github.com/kmgreen2/agglo/pkg/storage"
	"github.com/kmgreen2/agglo/pkg/entwine"
)

// Serialize will serialize an object; otherwise return an error
func Serialize(object interface{}) ([]byte, error) {
	var err error
	switch v := object.(type) {
	case *storage.ObjectDescriptor:
		return storage.SerializeObjectDescriptor(v)
	case *storage.MemObjectStoreBackendParams:
		return storage.SerializeMemObjectStoreParams(v)
	case *entwine.StreamImmutableMessage:
		return entwine.SerializeStreamImmutableMessage(v)
	case *entwine.TickerImmutableMessage:
		return entwine.SerializeTickerImmutableMessage(v)
	case *entwine.ProofImpl:
		return entwine.SerializeProof(v)
	default:
		err = common.NewInvalidError(fmt.Sprintf("Serialize - invalid type: %T", v))
	}
	return nil, err
}

// Deserialize will deserialize an object from bytes; otherwise return an error
func Deserialize(objectBytes []byte, target interface{}) error {
	var err error
	switch v := target.(type) {
	case *storage.ObjectDescriptor:
		err = storage.DeserializeObjectDescriptor(objectBytes, v)
	case *storage.MemObjectStoreBackendParams:
		err = storage.DeserializeMemObjectStoreParams(objectBytes, v)
	case *entwine.StreamImmutableMessage:
		return entwine.DeserializeStreamImmutableMessage(objectBytes, v)
	case *entwine.TickerImmutableMessage:
		return entwine.DeserializeTickerImmutableMessage(objectBytes, v)
	case *entwine.ProofImpl:
		return entwine.DeserializeProof(objectBytes, v)
	default:
		err = common.NewInvalidError(fmt.Sprintf("Deserialize - invalid type: %T", v))
	}

	return err
}
