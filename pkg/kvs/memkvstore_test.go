package kvs_test

import (
	"context"
	"github.com/kmgreen2/agglo/pkg/kvs"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestHappyPath(t *testing.T) {
	kvStore := kvs.NewMemKVStore()

	err := kvStore.Put(context.Background(), "foo", []byte("foobar"))
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	err = kvStore.Put(context.Background(), "fizz", []byte("foobar"))
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	value, err := kvStore.Get(context.Background(), "foo")
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	assert.Equal(t, "foobar", string(value))

	value, err = kvStore.Get(context.Background(), "fizz")
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	assert.Equal(t, "foobar", string(value))

	err = kvStore.Head(context.Background(), "foo")
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	err = kvStore.Head(context.Background(), "fizz")
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	keys, err := kvStore.List(context.Background(), "f")
	assert.Len(t, keys, 2)

	keys, err = kvStore.List(context.Background(), "fo")
	assert.Len(t, keys, 1)

	keys, err = kvStore.List(context.Background(), "f0")
	assert.Len(t, keys, 0)

	err = kvStore.Delete(context.Background(), "foo")
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	err = kvStore.Delete(context.Background(), "fizz")
	if err != nil {
		assert.FailNow(t, err.Error())
	}
}

func TestAtomicPut(t *testing.T) {
	kvStore := kvs.NewMemKVStore()

	err := kvStore.AtomicPut(context.Background(), "foo", nil, []byte("foobar"))
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	err = kvStore.Put(context.Background(), "foo", []byte("fizzbar"))
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	err = kvStore.AtomicPut(context.Background(), "foo", []byte("fizzbar"), []byte("barfoo"))
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	err = kvStore.AtomicPut(context.Background(), "foo", []byte("fizzbar"), []byte("barfoo"))
	assert.Error(t, err)
}

func TestAtomicDelete(t *testing.T) {
	kvStore := kvs.NewMemKVStore()

	err := kvStore.Put(context.Background(), "foo", []byte("foobar"))
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	err = kvStore.Put(context.Background(), "foo", []byte("fizzbar"))
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	err = kvStore.AtomicDelete(context.Background(), "foo", []byte("fizzbar"))
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	err = kvStore.AtomicDelete(context.Background(), "foo", []byte("fizzbar"))
	assert.Error(t, err)
}

func TestGetNotFound(t *testing.T) {
	kvStore := kvs.NewMemKVStore()
	err := kvStore.Put(context.Background(), "foo", []byte("foobar"))
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	_, err = kvStore.Get(context.Background(), "fizz")
	assert.Error(t, err)
}

func TestHeadNotFound(t *testing.T) {
	kvStore := kvs.NewMemKVStore()
	err := kvStore.Put(context.Background(), "foo", []byte("foobar"))
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	err = kvStore.Head(context.Background(), "fizz")
	assert.Error(t, err)
}

func TestDeleteNotFound(t *testing.T) {
	kvStore := kvs.NewMemKVStore()
	err := kvStore.Put(context.Background(), "foo", []byte("foobar"))
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	err = kvStore.Delete(context.Background(), "fizz")
	assert.Error(t, err)
}
