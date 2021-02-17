package storage_test

import (
	"context"
	"crypto/sha1"
	"errors"
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/pkg/storage"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

func RandomMemObjectStoreInstanceParams() (*storage.MemObjectStoreBackendParams, error) {
	uuid := gUuid.New()
	return storage.NewMemObjectStoreBackendParams(storage.MemObjectStoreBackend, uuid.String())
}

func TestMemHappyPath(t *testing.T) {
	fileSize := 1024
	params, err := RandomMemObjectStoreInstanceParams()
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	objStore, err := storage.NewMemObjectStore(params)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	randomReader := NewRandomReader(fileSize)

	err = objStore.Put(context.Background(), "foo", randomReader)
	assert.Nil(t, err)

	objHash := randomReader.Hash()

	err = objStore.Head(context.Background(), "foo")
	assert.Nil(t, err)

	reader, err := objStore.Get(context.Background(), "foo")
	assert.Nil(t, err)

	readBytes := make([]byte, 1024)
	readDigest := sha1.New()
	isDone := false
	for {
		if isDone {
			break
		}
		numRead, err := reader.Read(readBytes)
		if errors.Is(err, io.EOF) {
			isDone = true
		} else if err != nil {
			break
		}
		_, err = readDigest.Write(readBytes[:numRead])
		if err != nil {
			break
		}
	}
	assert.Equal(t, objHash, readDigest.Sum(nil))

	err = objStore.Delete(context.Background(), "foo")
	assert.Nil(t, err)
}

func TestMemPutConflictError(t *testing.T) {
	params, err := RandomMemObjectStoreInstanceParams()
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	objStore, err := storage.NewMemObjectStore(params)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	randomReader := NewRandomReader(1024)

	err = objStore.Put(context.Background(), "foo", randomReader)
	assert.Nil(t, err)

	randomReader = NewRandomReader(1024)

	err = objStore.Put(context.Background(), "foo", randomReader)
	assert.Error(t, err)
}

func TestMemPutReadError(t *testing.T) {
	params, err := RandomMemObjectStoreInstanceParams()
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	objStore, err := storage.NewMemObjectStore(params)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	badReader := &BadReader{}

	err = objStore.Put(context.Background(), "foo", badReader)
	assert.Error(t, err)
}

func TestMemGetNotFound(t *testing.T) {
	params, err := RandomMemObjectStoreInstanceParams()
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	objStore, err := storage.NewMemObjectStore(params)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	_, err = objStore.Get(context.Background(), "baz")
	assert.Error(t, err)
}

func TestMemHeadNotFound(t *testing.T) {
	params, err := RandomMemObjectStoreInstanceParams()
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	objStore, err := storage.NewMemObjectStore(params)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	err = objStore.Head(context.Background(), "baz")
	assert.Error(t, err)
}

func TestMemListNone(t *testing.T) {
	params, err := RandomMemObjectStoreInstanceParams()
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	objStore, err := storage.NewMemObjectStore(params)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	err = putObjects(objStore, "testprefix", 10, 0)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	results, err := objStore.List(context.Background(), "notprefix")
	assert.Equal(t, len(results), 0)
}

func TestMemListAll(t *testing.T) {
	params, err := RandomMemObjectStoreInstanceParams()
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	objStore, err := storage.NewMemObjectStore(params)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	err = putObjects(objStore, "testprefix", 10, 10)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	results, err := objStore.List(context.Background(), "")
	assert.Equal(t, len(results), 20)
}

func TestMemListSome(t *testing.T) {
	params, err := RandomMemObjectStoreInstanceParams()
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	objStore, err := storage.NewMemObjectStore(params)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	err = putObjects(objStore, "testprefix", 10, 10)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	results, err := objStore.List(context.Background(), "testprefix")
	assert.Equal(t, len(results), 10)
}
