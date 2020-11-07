package crypto

import (
	"crypto/rand"
)

type CryptoAlgorithm int

const  (
	AES128 CryptoAlgorithm = iota
	AES256
	AES512
)

func GenerateKey(cryptoAlgorithm CryptoAlgorithm) ([]byte, error) {
	var key []byte
	switch(cryptoAlgorithm) {
	case AES128: key = make([]byte, 16)
	case AES256: key = make([]byte, 32)
	case AES512: key = make([]byte, 64)
	}
	_, err := rand.Read(key)

	if err != nil {
		return nil, err
	}
	return key, nil
}

type Cipher interface {
	Encrypt(src []byte, dst []byte)
	Decrypt(src []byte, dst []byte)
}

