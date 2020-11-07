package crypto

import (
	"testing"
	"bytes"
)

func TestInavlidKey(t *testing.T) {
	_, err := InitAes([]byte("foo"))

	if err == nil {
		t.Errorf("'foo' should be an invalid key!")
	}

}

func TestEncryptDecrypt(t *testing.T) {
	orig := []byte("thisisacrappymessage")
	cleartext := []byte("thisisacrappymessage")
	ciphertext := make([]byte, len(cleartext))
	key, err := GenerateKey(AES128)

	if err != nil {
		t.Errorf("Error generating key")
	}

	block, err := InitAes(key)

	if err != nil {
		t.Errorf("Error initing AES")
	}

	block.Encrypt(ciphertext, cleartext)

	if bytes.Compare(cleartext, ciphertext) == 0 {
		t.Errorf("Error cleartext should not equal ciphertext")
	}

	block.Decrypt(cleartext, ciphertext)

	if bytes.Compare(cleartext, orig) != 0 {
		t.Errorf("Cleartext does not equal original message")
	}

}
