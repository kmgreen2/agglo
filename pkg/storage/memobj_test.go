package storage_test

import (
	"context"
	"crypto/rand"
	"crypto/sha1"
	"errors"
	"fmt"
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/pkg/common"
	"github.com/kmgreen2/agglo/pkg/storage"
	"github.com/stretchr/testify/assert"
	"hash"
	"io"
	"testing"
)

type RandomReader struct {
	len int
	offset int
	digest hash.Hash
}

func NewRandomReader(len int) *RandomReader {
	return &RandomReader{
		len: len,
		offset: 0,
		digest: sha1.New(),
	}
}

func (reader *RandomReader) Read(b []byte) (int, error) {
	if reader.offset >= reader.len {
		return -1, io.EOF
	}
	numLeft := reader.len - reader.offset
	bufRef := b
	if len(b) > numLeft {
		bufRef = b[:numLeft]
	}
	numRead, err := rand.Read(bufRef)
	if err != nil && !errors.Is(err, io.EOF) {
		return -1, err
	}
	_, err = reader.digest.Write(bufRef)
	if err != nil {
		return -1, err
	}
	reader.offset += numRead
	return numRead, nil
}

func (reader *RandomReader) Hash() []byte {
	return reader.digest.Sum(nil)
}

type BadReader struct {
}

func (reader *BadReader) Read(b []byte) (int ,error) {
	return -1, &common.InternalError{}
}

func RandomMemObjectStoreInstanceParams() (*storage.MemObjectStoreBackendParams, error) {
	uuid := gUuid.New()
	return storage.NewMemObjectStoreBackendParams(storage.MemObjectStoreBackend, uuid.String())
}

func TestHappyPath(t *testing.T) {
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

func TestPutConflictError(t *testing.T) {
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

func TestPutReadError(t *testing.T) {
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

func TestGetNotFound(t *testing.T) {
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

func TestHeadNotFound(t *testing.T) {
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

func putObjects(objStore storage.ObjectStore, prefix string, numWithPrefix int, numWithoutPrefix int) error {
	for i := 0; i < numWithPrefix; i++ {
		randomReader := NewRandomReader(1024)
		err := objStore.Put(context.Background(), fmt.Sprintf("%s%d", prefix, i), randomReader)
		if err != nil {
			return err
		}
	}
	for i := 0; i < numWithoutPrefix; i++ {
		randomReader := NewRandomReader(1024)
		err := objStore.Put(context.Background(), fmt.Sprintf("%d%s%d", i, prefix, i), randomReader)
		if err != nil {
			return err
		}
	}
	return nil
}

func TestListNone(t *testing.T) {
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

func TestListAll(t *testing.T) {
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

func TestListSome(t *testing.T) {
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
