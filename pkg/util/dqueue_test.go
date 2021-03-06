package util_test

import (
	"fmt"
	"github.com/kmgreen2/agglo/pkg/util"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func noOpRecoverFunc(in []byte) error {
	return nil
}

func failRecoverFunc(stride int) func(in []byte) error {
	counter := 0
	return func(in []byte) error {
		counter++
		if counter % stride == 0 {
			return fmt.Errorf("error")
		} else {
			return nil
		}
	}
}

func consumeQueue(q *util.DurableQueue, numProcessed, numAcked int) ([]*util.QueueItem, error) {
	var err error
	queueItems := make([]*util.QueueItem, numProcessed)

	for i := 0; i < numProcessed; i++ {
		queueItems[i], err = q.Dequeue()
		if err != nil {
			return nil, err
		}
	}

	for i := 0; i < numAcked; i++ {
		err := q.Ack(queueItems[i])
		if err != nil {
			return nil, err
		}
	}
	return queueItems, nil
}

func TestDurableQueueHappyPath(t *testing.T) {
	numItems := 100
	_ = os.Remove("/tmp/testdb")
	q, err := util.OpenDurableQueue("/tmp/testdb", noOpRecoverFunc, false)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	defer func() {
		err = q.Close()
		assert.Nil(t, err)

		err = q.Drop()
		assert.Nil(t, err)
	}()

	items := make([][]byte, numItems)
	for i := 0; i < numItems; i++ {
		items[i] = []byte(fmt.Sprintf("%d", i))
		err = q.Enqueue(items[i])
		if err != nil {
			assert.FailNow(t, err.Error())
		}
	}

	assert.Equal(t, int64(numItems), q.Length())

	_, err = consumeQueue(q, numItems, numItems)
	assert.Nil(t, err)

	assert.Equal(t, int64(0), q.Length())
	assert.Equal(t, int64(0), q.NumInflight())

}

func TestDurableQueueConcurrentOpen(t *testing.T) {
	_ = os.Remove("/tmp/testdb")
	q, err := util.OpenDurableQueue("/tmp/testdb", noOpRecoverFunc, false)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	defer func() {
		err = q.Close()
		assert.Nil(t, err)

		err = q.Drop()
		assert.Nil(t, err)
	}()

	_, err = util.OpenDurableQueue("/tmp/testdb", noOpRecoverFunc, false)
	assert.Error(t, err)
}

func TestDurableQueueCloseAndRestart(t *testing.T) {
	numItems := 100
	numProcessed := 80
	numAcked := 20
	_ = os.Remove("/tmp/testdb")
	q, err := util.OpenDurableQueue("/tmp/testdb", noOpRecoverFunc, false)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	defer func() {
		err = q.Close()
		assert.Nil(t, err)

		err = q.Drop()
		assert.Nil(t, err)
	}()

	items := make([][]byte, numItems)
	queueItems := make([]*util.QueueItem, numItems)
	for i := 0; i < numItems; i++ {
		items[i] = []byte(fmt.Sprintf("%d", i))
		err = q.Enqueue(items[i])
		if err != nil {
			assert.FailNow(t, err.Error())
		}
	}

	assert.Equal(t, int64(numItems), q.Length())

	queueItems, err = consumeQueue(q, numProcessed, numAcked)
	assert.Nil(t, err)

	for i := 0; i < numProcessed; i++ {
		assert.Equal(t, items[i], queueItems[i].Data)
	}

	assert.Equal(t, int64(numItems-numProcessed), q.Length())
	assert.Equal(t, int64(numProcessed-numAcked), q.NumInflight())

	err = q.Close()
	assert.Nil(t, err)

	q, err = util.OpenDurableQueue("/tmp/testdb", noOpRecoverFunc, false)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	queueItems, err = consumeQueue(q, numItems-numProcessed, numItems-numProcessed)
	assert.Nil(t, err)
	for i := 0; i < numItems-numProcessed; i++ {
		assert.Equal(t, items[numProcessed+i], queueItems[i].Data)
	}
	assert.Equal(t, int64(0), q.Length())
	assert.Equal(t, int64(0), q.NumInflight())
}

func TestDurableQueueFailAndFailRecovery(t *testing.T) {
	numItems := 100
	numProcessed := 80
	numAcked := 20
	_ = os.Remove("/tmp/testdb")
	q, err := util.OpenDurableQueue("/tmp/testdb", noOpRecoverFunc, false)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	defer func() {
		err = q.Close()
		assert.Nil(t, err)

		err = q.Drop()
		assert.Nil(t, err)
	}()

	items := make([][]byte, numItems)
	queueItems := make([]*util.QueueItem, numItems)
	for i := 0; i < numItems; i++ {
		items[i] = []byte(fmt.Sprintf("%d", i))
		err = q.Enqueue(items[i])
		if err != nil {
			assert.FailNow(t, err.Error())
		}
	}

	assert.Equal(t, int64(numItems), q.Length())

	queueItems, err = consumeQueue(q, numProcessed, numAcked)
	assert.Nil(t, err)

	for i := 0; i < numProcessed; i++ {
		assert.Equal(t, items[i], queueItems[i].Data)
	}

	assert.Equal(t, int64(numItems-numProcessed), q.Length())
	assert.Equal(t, int64(numProcessed-numAcked), q.NumInflight())

	err = q.Close()
	assert.Nil(t, err)

	q, err = util.OpenDurableQueue("/tmp/testdb", failRecoverFunc(5), false)
	assert.Equal(t, int64(12), q.NumInflight())
}

func TestUnprocessedAndInflightGetters(t *testing.T) {
	numItems := 100
	numProcessed := 80
	numAcked := 20
	_ = os.Remove("/tmp/testdb")
	q, err := util.OpenDurableQueue("/tmp/testdb", noOpRecoverFunc, false)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	defer func() {
		err = q.Close()
		assert.Nil(t, err)

		err = q.Drop()
		assert.Nil(t, err)
	}()

	items := make([][]byte, numItems)
	queueItems := make([]*util.QueueItem, numItems)
	for i := 0; i < numItems; i++ {
		items[i] = []byte(fmt.Sprintf("%d", i))
		err = q.Enqueue(items[i])
		if err != nil {
			assert.FailNow(t, err.Error())
		}
	}

	assert.Equal(t, int64(numItems), q.Length())

	queueItems, err = consumeQueue(q, numProcessed, numAcked)
	assert.Nil(t, err)

	for i := 0; i < numProcessed; i++ {
		assert.Equal(t, items[i], queueItems[i].Data)
	}

	assert.Equal(t, int64(numItems-numProcessed), q.Length())
	assert.Equal(t, int64(numProcessed-numAcked), q.NumInflight())

	inflight, err := q.GetInflight()
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	unprocessed, err := q.GetUnprocessed()
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	inflightMap := make(map[string]bool)
	for i := numProcessed + numAcked; i < numItems; i++ {
		inflightMap[fmt.Sprintf("%d", i)] = true
	}
	unprocessedMap := make(map[string]bool)
	for i := numProcessed; i < numItems; i++ {
		unprocessedMap[fmt.Sprintf("%d", i)] = true
	}

	for _, item := range inflight {
		delete(inflightMap, string(item.Data))
	}

	for _, item := range unprocessed {
		delete(unprocessedMap, string(item.Data))
	}

	assert.Len(t, inflightMap, 0)
	assert.Len(t, unprocessedMap, 0)
	err = q.Close()
	assert.Nil(t, err)
}
