package kvs_test

import (
	"github.com/kmgreen2/agglo/pkg/kvs"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewKVStoreFromConnectionString(t *testing.T) {
	_, err := kvs.NewKVStoreFromConnectionString("mem:foo")
	assert.Nil(t, err)

	_, err = kvs.NewKVStoreFromConnectionString(
		"dynamo:endpoint=http://localhost:8000,region=us-west-2,tableName=localkvstore,prefixLength=2")

	assert.Nil(t, err)
}

func TestNewKVStoreFromConnectionStringError(t *testing.T) {
	_, err := kvs.NewKVStoreFromConnectionString("mem")
	assert.Error(t, err)

	_, err = kvs.NewKVStoreFromConnectionString(
		"dynamo:region=us-west-2,tableName=localkvstore,prefixLength=2")
	assert.Error(t, err)

	_, err = kvs.NewKVStoreFromConnectionString("fizz:buzz")
	assert.Error(t, err)
}
