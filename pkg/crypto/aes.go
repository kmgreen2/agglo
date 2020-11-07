package crypto

import (
	"crypto/cipher"
	"crypto/aes"
)

type AesCipher struct {
	block cipher.Block
}

func InitAes(key []byte) (*AesCipher, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return &AesCipher {
		block: block,
	}, nil
}

func (aesCipher *AesCipher) Encrypt(dst []byte, src []byte) {
	aesCipher.block.Encrypt(dst, src)
}

func (aesCipher *AesCipher) Decrypt(dst []byte, src []byte) {
	aesCipher.block.Decrypt(dst, src)
}
