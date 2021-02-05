package caching

import (
	"container/heap"
	"fmt"
	"github.com/kmgreen2/agglo/pkg/util"
)

// LRUItem will implement the functions needed to maintain an item
// in a LRU priority queue
type LRUItem interface {
	ATime() int64
	Index() int
	SetIndex(index int)
}

// LRUQueue implements heap.Interface and holds LRUItems.
type LRUQueue []LRUItem

// Len will return the current length of the queue
func (lruQueue LRUQueue) Len() int { return len(lruQueue) }

// Less will return true if the entry at index i is less than the
// entry at index j
func (lruQueue LRUQueue) Less(i, j int) bool {
	return lruQueue[i].ATime() < lruQueue[j].ATime()
}

// Swap will swap two entries in the queue
func (lruQueue LRUQueue) Swap(i, j int) {
	if len(lruQueue) == 0 {
		return
	}
	lruQueue[i], lruQueue[j] = lruQueue[j], lruQueue[i]
	lruQueue[i].SetIndex(i)
	lruQueue[j].SetIndex(j)
}

// Push is called when a new item is added to the queue
func (lruQueue *LRUQueue) Push(x interface{}) {
	n := len(*lruQueue)
	item := x.(LRUItem)
	item.SetIndex(n)
	*lruQueue = append(*lruQueue, item)
}

// Pop is called to retrieve the item with the smallest atime value
func (lruQueue *LRUQueue) Pop() interface{} {
	old := *lruQueue
	n := len(old)
	if n == 0 {
		return nil
	}
	item := old[n-1]
	old[n-1] = nil
	item.SetIndex(-1)
	*lruQueue = old[0 : n-1]
	return item
}

// Update will reorder an item in the LRU queue
func (lruQueue *LRUQueue) Update(item LRUItem) {
	heap.Fix(lruQueue, item.Index())
}

// Delete will delete an LRU item from the queue
func (lruQueue *LRUQueue) Delete(item LRUItem) error {
	idx := item.Index()
	n := len(*lruQueue)

	if n == 0 || idx > n - 1 {
		return util.WithStack(util.NewNotFoundError(fmt.Sprintf("No iten at index: %d", idx)))
	}

	if item != (*lruQueue)[idx] {
		return util.WithStack(util.NewInconsistentStateError(fmt.Sprintf("Incorrect state at %d", idx)))
	}

	removedItem := heap.Remove(lruQueue, idx)

	if removedItem != item {
		return util.WithStack(util.NewInconsistentStateError(fmt.Sprintf("Mismatch pop item at %d", idx)))
	}

	return nil
}

