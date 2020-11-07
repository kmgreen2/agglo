package test

import (
	gocrypto "crypto"
	"crypto/rand"
	"crypto/rsa"
	"github.com/kmgreen2/agglo/pkg/crypto"
)

// GetSignerAuthenticator will return a signer/authenticator pair for a
// given hash algorithm
func GetSignerAuthenticator(hashAlgorithm gocrypto.Hash) (*crypto.RSASigner,
	*crypto.RSAAuthenticator, *rsa.PublicKey, error) {
	reader := rand.Reader
	bitSize := 2048

	key, err := rsa.GenerateKey(reader, bitSize)

	if err != nil {
		return nil, nil, nil, err
	}

	signer := crypto.NewRSASigner(key, hashAlgorithm)
	authenticator := crypto.NewRSAAuthenticator(&key.PublicKey, hashAlgorithm)

	return signer, authenticator, &key.PublicKey, nil
}
