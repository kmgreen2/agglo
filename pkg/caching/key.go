package caching

import (
	"strings"
)

// Key implementation will provide functions to serialize and compare keys
type Key interface {
	ToBytes() ([]byte, error)
	ToString() string
	Compare(other Key) int
}

// LRUKey is the interface to implement for keys for a LRU caching
type LRUKey interface {
	Key
	ATime() int64
	SetATime(atime int64)
	Index() int
	SetIndex(index int)
}

// StringKey is a wrapper around a string that implements the Key interface
type StringKey struct {
	val string
}

// NewStringCacheKey will create a new StringKey
func NewStringCacheKey(val string) *StringKey {
	return &StringKey{val}
}

// ToBytes will return the byte array represntation of the underlying string
func (key *StringKey) ToBytes() ([]byte, error) {
	return []byte(key.val), nil
}

// ToString will return the underlying string
func (key *StringKey) ToString() string {
	return key.val
}

// Compare will compare *this* key to another Key
func (key *StringKey) Compare(other Key) int {
	switch v := other.(type) {
	case *StringKey:
		return strings.Compare(key.val, v.ToString())
	default:
		return -1
	}
}
