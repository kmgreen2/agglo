package kvs_test

import (
	"github.com/kmgreen2/agglo/pkg/kvs"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHappyPath(t *testing.T) {
	kvStore := kvs.NewMemKVStore()

	err := kvStore.Put("foo", []byte("foobar"))
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	err = kvStore.Put("fizz", []byte("foobar"))
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	value, err := kvStore.Get("foo")
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	assert.Equal(t, "foobar", string(value))

	value, err = kvStore.Get("fizz")
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	assert.Equal(t, "foobar", string(value))

	err = kvStore.Head("foo")
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	err = kvStore.Head("fizz")
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	keys, err := kvStore.List("f")
	assert.Len(t, keys, 2)

	keys, err = kvStore.List("fo")
	assert.Len(t, keys, 1)

	keys, err = kvStore.List("f0")
	assert.Len(t, keys, 0)

	err = kvStore.Delete("foo")
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	err = kvStore.Delete("fizz")
	if err != nil {
		assert.FailNow(t, err.Error())
	}
}

func TestPutConflict(t *testing.T) {
	kvStore := kvs.NewMemKVStore()
	err := kvStore.Put("foo", []byte("foobar"))
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	err = kvStore.Put("foo", []byte("foobar"))
	assert.Error(t, err)
}

func TestGetNotFound(t *testing.T) {
	kvStore := kvs.NewMemKVStore()
	err := kvStore.Put("foo", []byte("foobar"))
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	_, err = kvStore.Get("fizz")
	assert.Error(t, err)
}

func TestHeadNotFound(t *testing.T) {
	kvStore := kvs.NewMemKVStore()
	err := kvStore.Put("foo", []byte("foobar"))
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	err = kvStore.Head("fizz")
	assert.Error(t, err)
}

func TestDeleteNotFound(t *testing.T) {
	kvStore := kvs.NewMemKVStore()
	err := kvStore.Put("foo", []byte("foobar"))
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	err = kvStore.Delete("fizz")
	assert.Error(t, err)
}
