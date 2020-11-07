package crypto

import (
	"crypto"
	"crypto/rsa"
	"crypto/rand"
	"crypto/x509"
	"github.com/kmgreen2/agglo/pkg/common"
	"hash"
	"reflect"
)

///
// RSA implementations of Signature, Authenticator and Signer
///
type RSAAuthenticator struct {
	publicKey *rsa.PublicKey
	hashAlgorithm crypto.Hash
}

type RSASigner struct {
	privateKey *rsa.PrivateKey
	hashAlgorithm crypto.Hash
}

type RSASignature struct {
	signature []byte
	hashAlgorithm crypto.Hash
}

// Return the byte array representation of the signature
func (r *RSASignature) Bytes() []byte {
	return r.signature
}

// Return the hash algorithm used to compute the signature
func (r *RSASignature) HashAlgorithm() crypto.Hash {
	return r.hashAlgorithm
}

// Return the public-key algorithm used to compute the signature
// This will always return RSA
func (r *RSASignature) PKAlgorithm() x509.PublicKeyAlgorithm {
	return x509.RSA
}

// Helper function that ensured the authenticator's algorithms
// match those of the signature
func algorithmMatch(a *RSAAuthenticator, s Signature) bool {
	return (a.hashAlgorithm == s.HashAlgorithm() && s.PKAlgorithm() == x509.RSA)
}

// Verify the authenticity of a message using a provided signature.  Returns
// false if either the target algorithms do not match (hash+PK) or the signature
// does not match
func (a *RSAAuthenticator) Verify(message []byte, signature Signature) bool {
	var digest []byte
	var hash hash.Hash
	if !algorithmMatch(a, signature) {
		return false
	}

	switch(a.hashAlgorithm) {
	case crypto.SHA1:
		hash = common.InitHash(common.SHA1)
	case crypto.SHA256:
		hash = common.InitHash(common.SHA256)
	}
	hash.Write(message)
	digest = hash.Sum(nil)

	err := rsa.VerifyPKCS1v15(a.publicKey, a.hashAlgorithm, digest, signature.Bytes())

	if err != nil {
		return false
	}

	return true
}

// Sign a message using a RSASinger, which supports SHA1 and SHA256, depending on
// how the signer is instantiated.
func (s *RSASigner) Sign(message []byte) (Signature, error) {
	var digest []byte
	var hash hash.Hash

	switch(s.hashAlgorithm) {
	case crypto.SHA1:
		hash = common.InitHash(common.SHA1)
	case crypto.SHA256:
		hash = common.InitHash(common.SHA256)
	}
	hash.Write(message)
	digest = hash.Sum(nil)
	signature, err := rsa.SignPKCS1v15(rand.Reader, s.privateKey, s.hashAlgorithm, digest)

	if err != nil {
		return nil, err
	}

	return NewRSASignatureBytes(signature, s.hashAlgorithm), nil
}

// Helper function to create a new RSASigner with a specific private key and hash algorithm
func NewRSASigner(privateKey *rsa.PrivateKey, hashAlgorithm crypto.Hash) *RSASigner {
	s := &RSASigner {
		privateKey: privateKey,
		hashAlgorithm: hashAlgorithm,
	}
	return s
}

// Helper function to create a new RSAAuthenticator with a specific public key and hash algorithm
func NewRSAAuthenticator(publicKey *rsa.PublicKey, hashAlgorithm crypto.Hash) *RSAAuthenticator {
	a := &RSAAuthenticator{
		publicKey: publicKey,
		hashAlgorithm: hashAlgorithm,
	}
	return a
}

// Helper function to create a new signature from the byte representation of the signature
func NewRSASignatureBytes(bytes []byte, hashAlgorithm crypto.Hash) *RSASignature {
	s := &RSASignature{
		signature: bytes,
		hashAlgorithm: hashAlgorithm,
	}
	return s
}

// Helper function used to serialize RSA public key
func SerialilzeRSAPublicKey(publicKey *rsa.PublicKey) ([]byte, error) {
	bytes, err := x509.MarshalPKIXPublicKey(publicKey)

	if err != nil {
		return nil, err
	}

	return bytes, nil
}

// Helper function used to deserialize RSA public key
func DeserialilzeRSAPublicKey(bytes []byte) (*rsa.PublicKey, error) {
	publicKey, err := x509.ParsePKIXPublicKey(bytes)

	if err != nil {
		return nil, err
	}

	if reflect.TypeOf(publicKey) != reflect.TypeOf(&rsa.PublicKey{}) {
		return nil, err
	}

	return publicKey.(*rsa.PublicKey), err
}

// Helper function used to serialize RSA private key
func SerialilzeRSAPrivateKey(privateKey *rsa.PrivateKey) ([]byte, error) {
	bytes := x509.MarshalPKCS1PrivateKey(privateKey)

	return bytes, nil
}

// Helper function used to deserialize RSA private key
func DeserialilzeRSAPrivateKey(bytes []byte) (*rsa.PrivateKey, error) {
	privateKey, err := x509.ParsePKCS1PrivateKey(bytes)

	if err != nil {
		return nil, err
	}

	return privateKey, err
}
