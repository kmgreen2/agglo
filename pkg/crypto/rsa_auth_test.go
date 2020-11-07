package crypto_test

import (
	"testing"
	"crypto"
	"crypto/rsa"
	"crypto/rand"
	. "github.com/kmgreen2/agglo/pkg/crypto"
	. "github.com/kmgreen2/agglo/test"
)

func TestBasicAuth(t *testing.T) {
	message := []byte("thisisamessage")
	hashAlgorithms := []crypto.Hash{crypto.SHA1, crypto.SHA256}

	for _, hashAlgorithm := range hashAlgorithms {

		signer, authenticator, _, err := GetSignerAuthenticator(hashAlgorithm)

		if err != nil {
			t.Errorf("Error generating RSA keys: %s", err.Error())
		}

		signature, err := signer.Sign(message)

		if err != nil {
			t.Errorf("Error signing the message: %s", err.Error())
		}

		if !authenticator.Verify(message, signature) {
			t.Error("Expected signature to verify message!")
		}
	}
}

func TestAuthWithKeySerialization(t *testing.T) {
	message := []byte("thisisamessage")
	reader := rand.Reader
	bitSize := 2048

	key, err := rsa.GenerateKey(reader, bitSize)

	if err != nil {
		t.Errorf("Error generating RSA keys: %s", err.Error())
	}

	serializedPublicKey, err := SerialilzeRSAPublicKey(&key.PublicKey)

	if err != nil {
		t.Errorf("Error serializing public key!")
	}

	serializedPrivateKey, err := SerialilzeRSAPrivateKey(key)

	if err != nil {
		t.Errorf("Error serializing private key!")
	}

	publicKey, err := DeserialilzeRSAPublicKey(serializedPublicKey)

	if err != nil {
		t.Errorf("Error deserializing public key!")
	}

	privateKey, err := DeserialilzeRSAPrivateKey(serializedPrivateKey)

	if err != nil {
		t.Errorf("Error deserializing private key!")
	}


	signer := NewRSASigner(privateKey, crypto.SHA1)
	authenticator := NewRSAAuthenticator(publicKey, crypto.SHA1)

	signature, err := signer.Sign(message)

	if err != nil {
		t.Errorf("Error signing the message: %s", err.Error())
	}

	if !authenticator.Verify(message, signature) {
		t.Error("Expected signature to verify message!")
	}
}

func TestBasicFailedAuth(t *testing.T) {
	message := []byte("thisisamessage")
	message2 := []byte("thisisabadmessage")
	hashAlgorithms := []crypto.Hash{crypto.SHA1, crypto.SHA256}

	for _, hashAlgorithm := range hashAlgorithms {

		signer, authenticator, _, err := GetSignerAuthenticator(hashAlgorithm)

		if err != nil {
			t.Errorf("Error generating RSA keys: %s", err.Error())
		}

		signature, err := signer.Sign(message)

		if err != nil {
			t.Errorf("Error signing the message: %s", err.Error())
		}

		if authenticator.Verify(message2, signature) {
			t.Error("Expected signature to *not* verify message!")
		}
	}
}

func TestAlgorithmMismatch(t *testing.T) {
	message := []byte("thisisamessage")

	reader := rand.Reader
	bitSize := 2048

	key, err := rsa.GenerateKey(reader, bitSize)

	signer := NewRSASigner(key, crypto.SHA1)
	authenticator := NewRSAAuthenticator(&key.PublicKey, crypto.SHA256)

	if err != nil {
		t.Errorf("Error generating RSA keys: %s", err.Error())
	}

	signature, err := signer.Sign(message)

	if err != nil {
		t.Errorf("Error signing the message: %s", err.Error())
	}

	if authenticator.Verify(message, signature) {
		t.Error("Expected signature to *not* verify message!")
	}
}



