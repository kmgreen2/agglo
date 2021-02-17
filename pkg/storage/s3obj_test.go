package storage_test

import (
	"context"
	"crypto/sha1"
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/pkg/storage"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"io"
	"testing"
)

func LocalS3ObjectStoreParams() (*storage.S3ObjectStoreBackendParams, error) {
	return storage.NewS3ObjectStoreBackendParams(storage.S3ObjectStoreBackend,
		"localhost:9000", "localtest")
}

func TestS3HappyPath(t *testing.T) {
	fileSize := 1024
	params, err := LocalS3ObjectStoreParams()
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	objStore, err := storage.NewS3ObjectStore(params)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	randomReader := NewRandomReader(fileSize)

	key := gUuid.New().String()

	err = objStore.Put(context.Background(), key, randomReader)
	assert.Nil(t, err)

	objHash := randomReader.Hash()

	err = objStore.Head(context.Background(), key)
	assert.Nil(t, err)

	reader, err := objStore.Get(context.Background(), key)
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

	err = objStore.Delete(context.Background(), key)
	assert.Nil(t, err)
}

func TestS3PutOverwrite(t *testing.T) {
	params, err := LocalS3ObjectStoreParams()
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	objStore, err := storage.NewS3ObjectStore(params)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	randomReader := NewRandomReader(1024)

	key := gUuid.New().String()

	err = objStore.Put(context.Background(), key, randomReader)
	assert.Nil(t, err)

	randomReader = NewRandomReader(1024)

	err = objStore.Put(context.Background(), key, randomReader)
	assert.Nil(t, err)
}

func TestS3PutReadError(t *testing.T) {
	params, err := LocalS3ObjectStoreParams()
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	objStore, err := storage.NewS3ObjectStore(params)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	badReader := &BadReader{}

	err = objStore.Put(context.Background(), "foo", badReader)
	assert.Error(t, err)
}

func TestS3GetNotFound(t *testing.T) {
	params, err := LocalS3ObjectStoreParams()
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	objStore, err := storage.NewS3ObjectStore(params)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	_, err = objStore.Get(context.Background(), "baz")
	assert.Error(t, err)
}

func TestS3HeadNotFound(t *testing.T) {
	params, err := LocalS3ObjectStoreParams()
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	objStore, err := storage.NewS3ObjectStore(params)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	err = objStore.Head(context.Background(), "baz")
	assert.Error(t, err)
}

func TestS3ListNone(t *testing.T) {
	params, err := LocalS3ObjectStoreParams()
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	objStore, err := storage.NewS3ObjectStore(params)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	key := gUuid.New().String()

	err = putObjects(objStore, key, 10, 0)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	results, err := objStore.List(context.Background(), "notprefix")
	assert.Equal(t, len(results), 0)
}

func TestS3ListAll(t *testing.T) {
	params, err := LocalS3ObjectStoreParams()
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	objStore, err := storage.NewS3ObjectStore(params)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	key := gUuid.New().String()

	err = putObjects(objStore, key, 10, 10)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	results, err := objStore.List(context.Background(), "")
	// Gotta do >= because the bucket may have a bunch of objects
	assert.True(t, len(results) >= 20)
}

func TestS3ListSome(t *testing.T) {
	params, err := LocalS3ObjectStoreParams()
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	objStore, err := storage.NewS3ObjectStore(params)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	key := gUuid.New().String()

	err = putObjects(objStore, key, 10, 10)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	results, err := objStore.List(context.Background(), key)
	assert.Equal(t, len(results), 10)
}

func TestSerDe(t *testing.T) {
	var objDescDeser storage.ObjectDescriptor
	params, err := LocalS3ObjectStoreParams()
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	objDesc := storage.NewObjectDescriptor(params, "foo")

	objDescBytes, err := storage.SerializeObjectDescriptor(objDesc)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	err = storage.DeserializeObjectDescriptor(objDescBytes, &objDescDeser)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	assert.Equal(t, objDesc.GetParams(), objDescDeser.GetParams())
	assert.Equal(t, objDesc.GetKey(), objDescDeser.GetKey())
}

func TestSerDeDifferent(t *testing.T) {
	var objDescDeser storage.ObjectDescriptor
	params, err := LocalS3ObjectStoreParams()
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	objDesc := storage.NewObjectDescriptor(params, "foo")

	objDescBytes, err := storage.SerializeObjectDescriptor(objDesc)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	err = storage.DeserializeObjectDescriptor(objDescBytes, &objDescDeser)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	objDesc = storage.NewObjectDescriptor(params, "bar")

	assert.Equal(t, objDesc.GetParams(), objDescDeser.GetParams())
	assert.NotEqual(t, objDesc.GetKey(), objDescDeser.GetKey())
}

func TestSerDeFail(t *testing.T) {
	var objDescDeser storage.ObjectDescriptor

	err := storage.DeserializeObjectDescriptor([]byte("bad"), &objDescDeser)
	assert.Error(t, err)

}