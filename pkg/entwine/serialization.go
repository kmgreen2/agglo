package entwine

import (
	"fmt"
	"github.com/kmgreen2/agglo/pkg/util"
)

// Serialize will serialize an object; otherwise return an error
func Serialize(object interface{}) ([]byte, error) {
	var err error
	switch v := object.(type) {
	case *StreamImmutableMessage:
		return SerializeStreamImmutableMessage(v)
	case *TickerImmutableMessage:
		return SerializeTickerImmutableMessage(v)
	case *ProofImpl:
		return SerializeProof(v)
	default:
		err = util.NewInvalidError(fmt.Sprintf("Serialize - invalid type: %T", v))
	}
	return nil, err
}

// Deserialize will deserialize an object from bytes; otherwise return an error
func Deserialize(objectBytes []byte, target interface{}) error {
	var err error
	switch v := target.(type) {
	case *StreamImmutableMessage:
		return DeserializeStreamImmutableMessage(objectBytes, v)
	case *TickerImmutableMessage:
		return DeserializeTickerImmutableMessage(objectBytes, v)
	case *ProofImpl:
		return DeserializeProof(objectBytes, v)
	default:
		err = util.NewInvalidError(fmt.Sprintf("Deserialize - invalid type: %T", v))
	}

	return err
}
