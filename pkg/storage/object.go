package storage

import (
	"io"
)

// ObjectStore is the interface for an object store
type ObjectStore interface {
	Put(key string, reader io.Reader)	error
	Get(key string) (io.Reader, error)
	Head(key string) error
	Delete(key string) error
	List(prefix string) ([]string, error)
}

