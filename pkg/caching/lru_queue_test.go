package caching_test

import (
	"container/heap"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/kmgreen2/agglo/pkg/caching"
	"testing"
)

func fillLRUQueue(numEntries int) (*caching.LRUQueue, []*caching.LRUStringKey) {
	lruStringKeys := make([]*caching.LRUStringKey, numEntries)
	lruQueue := &caching.LRUQueue{}
	heap.Init(lruQueue)

	for i := 0; i < 10; i++ {
		lruStringKeys[i] = caching.NewLRUStringKey(fmt.Sprint(i))
		lruStringKeys[i].SetATime(int64(i))
		heap.Push(lruQueue, lruStringKeys[i])
	}
	return lruQueue, lruStringKeys
}

func TestLRUQueuePushPop(t *testing.T) {
	lruQueue, _ := fillLRUQueue(10)
	for i := 0; i < 10; i++ {
		entry := heap.Pop(lruQueue).(*caching.LRUStringKey)
		assert.Equal(t, fmt.Sprint(i), entry.ToString())
	}
	assert.Nil(t, heap.Pop(lruQueue))
}

func TestLRUQueueUpdate(t *testing.T) {
	lruQueue, lruStringKeys := fillLRUQueue(10)
	lruStringKeys[0].SetATime(100)
	lruQueue.Update(lruStringKeys[0])

	entry := heap.Pop(lruQueue).(*caching.LRUStringKey)

	assert.Equal(t, fmt.Sprint(1), entry.ToString())
}

func TestLRUQueueDelete(t *testing.T) {

	for i := 0; i < 10; i++ {
		lruQueue, lruStringKeys := fillLRUQueue(10)

		err := lruQueue.Delete(lruStringKeys[i])
		assert.Nil(t, err)

		for j := 0; j < 10; j++ {
			if i == j {
				continue
			}
			item := heap.Pop(lruQueue).(*caching.LRUStringKey)
			assert.Equal(t, fmt.Sprint(j), item.ToString())
		}
	}
}

func TestLRUQueueDeleteErrors(t *testing.T) {
	lruQueue := &caching.LRUQueue{}
	lruStringKey1 := caching.NewLRUStringKey(fmt.Sprint(0))
	lruStringKey2 := caching.NewLRUStringKey(fmt.Sprint(1))

	err := lruQueue.Delete(lruStringKey1)
	assert.Error(t, err)

	heap.Push(lruQueue, lruStringKey1)
	heap.Push(lruQueue, lruStringKey2)

	lruStringKey1.SetIndex(1)

	err = lruQueue.Delete(lruStringKey1)
	assert.Error(t, err)

	lruStringKey1.SetIndex(100)
	err = lruQueue.Delete(lruStringKey1)
	assert.Error(t, err)
}
