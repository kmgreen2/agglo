package kvs_test

import (
	"context"
	"github.com/kmgreen2/agglo/pkg/kvs"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDynamoDBHappyPath(t *testing.T) {
	kvStore := kvs.NewDynamoKVStore("http://localhost:8000", "us-west-2", "localkvstore", 2)

	err := kvStore.Put(context.Background(), "fuu", []byte("bar"))
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	err = kvStore.Put(context.Background(), "fizz", []byte("baz"))
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	err = kvStore.Put(context.Background(), "fuzz", []byte("buzz"))
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	val, err := kvStore.Get(context.Background(), "fuu")
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	assert.Equal(t, []byte("bar"), val)

	val, err = kvStore.Get(context.Background(), "fizz")
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	assert.Equal(t, []byte("baz"), val)

	val, err = kvStore.Get(context.Background(), "fuzz")
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	assert.Equal(t, []byte("buzz"), val)

	items, err := kvStore.List(context.Background(), "fu")
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	assert.Equal(t, 2, len(items))

	items, err = kvStore.List(context.Background(), "fuu")
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	assert.Equal(t, 1, len(items))

	items, err = kvStore.List(context.Background(), "fuz")
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	assert.Equal(t, 1, len(items))

	items, err = kvStore.List(context.Background(), "fi")
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	assert.Equal(t, 1, len(items))

	err = kvStore.Delete(context.Background(), "fuu")
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	err = kvStore.Delete(context.Background(), "fizz")
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	err = kvStore.Delete(context.Background(), "fuzz")
	if err != nil {
		assert.FailNow(t, err.Error())
	}
}

func TestDynamoConnectionString(t *testing.T) {
	kvStore := kvs.NewDynamoKVStore("http://localhost:8000", "us-west-2", "localkvstore", 2)
	connectionString := kvStore.ConnectionString()
	assert.Equal(t, "endpoint=http://localhost:8000,region=us-west-2,tableName=localkvstore,prefixLength=2",
		connectionString)
}

func TestDynamoBadConnectionString(t *testing.T) {
	_, err := kvs.NewDynamoKVStoreFromConnectionString(
		"endpoint=http://localhost:8000,region=us-west-2,tableName=localkvstore,prefixLength=2")
	assert.Nil(t, err)
	_, err = kvs.NewDynamoKVStoreFromConnectionString(
		"region=us-west-2,tableName=localkvstore,prefixLength=2")
	assert.Error(t, err)
	_, err = kvs.NewDynamoKVStoreFromConnectionString(
		"endpoint=http://localhost:8000,tableName=localkvstore,prefixLength=2")
	assert.Error(t, err)
	_, err = kvs.NewDynamoKVStoreFromConnectionString(
		"endpoint=http://localhost:8000,region=us-west-2,prefixLength=2")
	assert.Error(t, err)
	_, err = kvs.NewDynamoKVStoreFromConnectionString(
		"endpoint=http://localhost:8000,region=us-west-2")
	assert.Error(t, err)
}

func TestDynamoAtomicPut(t *testing.T) {
	kvStore := kvs.NewDynamoKVStore("http://localhost:8000", "us-west-2", "localkvstore", 2)

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

	err = kvStore.Delete(context.Background(), "foo")
	if err != nil {
		assert.FailNow(t, err.Error())
	}
}

func TestDynamoAtomicDelete(t *testing.T) {
	kvStore := kvs.NewDynamoKVStore("http://localhost:8000", "us-west-2", "localkvstore", 2)

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

func TestDynamoGetNotFound(t *testing.T) {
	kvStore := kvs.NewDynamoKVStore("http://localhost:8000", "us-west-2", "localkvstore", 2)
	err := kvStore.Put(context.Background(), "foo", []byte("foobar"))
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	_, err = kvStore.Get(context.Background(), "fizz")
	assert.Error(t, err)

	err = kvStore.Delete(context.Background(), "foo")
	if err != nil {
		assert.FailNow(t, err.Error())
	}
}

func TestDynamoHeadNotFound(t *testing.T) {
	kvStore := kvs.NewDynamoKVStore("http://localhost:8000", "us-west-2", "localkvstore", 2)
	err := kvStore.Put(context.Background(), "foo", []byte("foobar"))
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	err = kvStore.Head(context.Background(), "fizz")
	assert.Error(t, err)

	err = kvStore.Delete(context.Background(), "foo")
	if err != nil {
		assert.FailNow(t, err.Error())
	}
}

func TestDynamoDeleteNotFound(t *testing.T) {
	kvStore := kvs.NewDynamoKVStore("http://localhost:8000", "us-west-2", "localkvstore", 2)
	err := kvStore.Put(context.Background(), "foo", []byte("foobar"))
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	err = kvStore.Delete(context.Background(), "fizz")
	assert.Error(t, err)

	err = kvStore.Delete(context.Background(), "foo")
	if err != nil {
		assert.FailNow(t, err.Error())
	}
}