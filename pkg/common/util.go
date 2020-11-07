package common

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"hash"
)

// Used to distinguish between different Digest algorithms
type DigestType int

const (
	SHA1 DigestType = iota
	SHA256
	MD5
)

// Construct a hash object using a supported Digest
// type.  If the Digest type is not supported, return
// nil.
func InitHash(digestType DigestType) hash.Hash {
	switch digestType {
	case SHA1:
		return sha1.New()
	case SHA256:
		return sha256.New()
	case MD5:
		return md5.New()
	default:
		return nil
	}
}

