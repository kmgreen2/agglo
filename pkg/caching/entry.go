package caching

import (
	"bytes"
	"fmt"
)

// Entry implementations will provide functions to (de)serialize values
// and compare caching entries
type Entry interface {
	ToBytes() ([]byte, error)
	ToString() string
	ToType(interface{}) error
	Compare(other Entry) int
}

// LRUEntry implementations will provide that of Entry and the ability
// to set/get create time for TTL functionality
type LRUEntry interface {
	Entry
	CTime() int64
	SetCTime(ctime int64)
}

// StringEntry is a wrapper around a string that implements Entry
type StringEntry struct {
	val string
}

// NewStringCacheEntry will create a new StringEntry
func NewStringCacheEntry(val string) *StringEntry {
	return &StringEntry{
		val,
	}
}

// ToBytes will return the byte array representation of the string entry
func (entry *StringEntry) ToBytes() ([]byte, error) {
	return []byte(entry.val), nil
}

// ToString will return the string
func (entry *StringEntry) ToString() string {
	return entry.val
}

// ToType will attempt to convert the string representation to a specified type
func (entry *StringEntry) ToType(elem interface{}) error {
	switch v := elem.(type) {
	case *string:
		elem = entry.val
		return nil
	default:
		return fmt.Errorf("Converting to %T is not supported for StringEntry", v)
	}
}

// Compare will compare *this* entry to another, specified entr
func (entry *StringEntry) Compare(other Entry) int {
	otherBytes, err := other.ToBytes()
	if err != nil {
		return -1
	}
	return bytes.Compare([]byte(entry.val), otherBytes)
}
