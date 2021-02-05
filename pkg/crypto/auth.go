package crypto

import (
	"crypto"
	"crypto/x509"
	"bytes"
	"encoding/gob"
	"github.com/kmgreen2/agglo/pkg/util"
)

// A signature is a byte array and the algorithms used to construct and verify it
type Signature interface  {
	Bytes() []byte
	HashAlgorithm() crypto.Hash
	PKAlgorithm() x509.PublicKeyAlgorithm
}

func SerializeSignature(signature Signature) ([]byte, error) {
	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)

	err := enc.Encode(signature.Bytes())
	if err != nil {
		return nil, err
	}

	err = enc.Encode(signature.HashAlgorithm())
	if err != nil {
		return nil, err
	}

	err = enc.Encode(signature.PKAlgorithm())
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func DeserializeSignature(sigBytes []byte) (Signature, error) {
	buf := bytes.NewBuffer(sigBytes)
	dec := gob.NewDecoder(buf)

	var signature []byte
	var hashAlgorithm crypto.Hash
	var pkAlgorithm x509.PublicKeyAlgorithm

	err := dec.Decode(&signature)
	if err != nil {
		return nil, err
	}

	err = dec.Decode(&hashAlgorithm)
	if err != nil {
		return nil, err
	}

	err = dec.Decode(&pkAlgorithm)
	if err != nil {
		return nil, err
	}

	switch(pkAlgorithm) {
	case x509.RSA:
		return NewRSASignatureBytes(signature, hashAlgorithm), nil
	default:
		return nil, util.NewSignatureError("Currently only support RSA signatures")
	}
}

// An authenticator verifies that a mesasage is properly signed by a signature
type Authenticator interface {
	// Verify that the signature properly signs the provided message
	Verify([]byte, Signature) bool
}

// A signer creates a signature from a message
type Signer interface {
	// Generate a signature from a message
	Sign([]byte) (Signature, error)
}

