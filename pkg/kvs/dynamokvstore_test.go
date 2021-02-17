package kvs_test

import (
	"context"
	"fmt"
	"github.com/kmgreen2/agglo/pkg/kvs"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDynamoDBHappyPath(t *testing.T) {
	kvStore := kvs.NewDynamoKVStore("http://localhost:8000", "us-west-2", "localkvstore")

	err := kvStore.Put(context.Background(), "foo", []byte("bar"))
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	items, err := kvStore.List(context.Background(), "foo")
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	fmt.Print(items)
}