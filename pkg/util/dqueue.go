package util

import (
	"bytes"
	"encoding/gob"
	"fmt"
	bolt "go.etcd.io/bbolt"
	"os"
	"sync"
	"time"
)

type BucketType string
const (
	UnprocessedQueue BucketType = "Unprocessed"
	InflightQueue = "Inflight"
	InternalBucket = "Internal"
)

var queueStateKey []byte = []byte("queueState")
type QueueState int
const (
	QueueUnknown QueueState = iota
	QueueOpening
	QueueRecovering
	QueueOpened
	QueueClosed
)

func validStateTransition(prev, curr QueueState) bool {
	switch curr {
	case QueueOpening:
		if prev == QueueClosed {
			return true
		}
		return false
	case QueueRecovering:
		if prev == QueueOpening {
			return true
		}
		return false
	case QueueOpened:
		if prev == QueueOpening || prev == QueueRecovering {
			return true
		}
		return false
	case QueueClosed:
		return true
	case QueueUnknown:
		return true
	}
	return false
}

type internalQueuePos struct {
	tail int64
	head int64
}

type DurableQueue struct {
	lock *sync.Mutex
	dbFile string
	db *bolt.DB
	state QueueState
	unprocessed internalQueuePos
	recoverFunc func(in []byte) error
	numInflight int64
}

type QueueItem struct {
	Data         []byte
	QueueTime    int64
	InflightTime int64
	Idx          int64
}

func NewQueueItem(data []byte, idx int64) *QueueItem {
	now := time.Now().UnixNano()
	return &QueueItem {
		Data:         data,
		QueueTime:    now,
		InflightTime: -1,
		Idx:          idx,
	}
}

func (item *QueueItem) inflight() {
	item.InflightTime = time.Now().UnixNano()
}

func queueItemToBytes(v *QueueItem) ([]byte, error) {
	byteBuffer := bytes.NewBuffer([]byte{})
	encoder := gob.NewEncoder(byteBuffer)
	err := encoder.Encode(v)
	if err != nil {
		return nil, err
	}
	return byteBuffer.Bytes(), nil
}

func bytesToQueueItem(itemBytes []byte) (*QueueItem, error) {
	queueItem  := &QueueItem{}
	byteBuffer := bytes.NewBuffer(itemBytes)
	decoder := gob.NewDecoder(byteBuffer)
	err := decoder.Decode(queueItem)
	if err != nil {
		return nil, err
	}
	return queueItem, nil
}

func int64ToBytes(v int64) ([]byte, error) {
	byteBuffer := bytes.NewBuffer([]byte{})
	encoder := gob.NewEncoder(byteBuffer)
	err := encoder.Encode(&v)
	if err != nil {
		return nil, err
	}
	return byteBuffer.Bytes(), nil
}

func bytesToInt64(intBytes []byte) (int64, error) {
	var v int64
	byteBuffer := bytes.NewBuffer(intBytes)
	decoder := gob.NewDecoder(byteBuffer)
	err := decoder.Decode(&v)
	if err != nil {
		return -1, err
	}
	return v, nil
}

func queueStateToBytes(v QueueState) ([]byte, error) {
	byteBuffer := bytes.NewBuffer([]byte{})
	encoder := gob.NewEncoder(byteBuffer)
	err := encoder.Encode(&v)
	if err != nil {
		return nil, err
	}
	return byteBuffer.Bytes(), nil
}

func bytesToQueueState(itemBytes []byte) (QueueState, error) {
	var v QueueState
	byteBuffer := bytes.NewBuffer(itemBytes)
	decoder := gob.NewDecoder(byteBuffer)
	err := decoder.Decode(&v)
	if err != nil {
		return -1, err
	}
	return v, nil
}

func OpenDurableQueue(dbFile string, recoverFunc func(in []byte) error, force bool) (*DurableQueue, error) {
	var err error
	q := &DurableQueue{
		lock: &sync.Mutex{},
		recoverFunc: recoverFunc,
	}
	q.dbFile = dbFile
	if q.db, err = bolt.Open(dbFile, 0644, &bolt.Options{Timeout: 1*time.Second}); err != nil {
		return nil, err
	}

	tx, err := q.db.Begin(true)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	// Atomically set the state to Opening
	err = q.setState(tx, QueueOpening, force)
	if err != nil {
		return nil, err
	}

	unprocessedBucket, err := tx.CreateBucketIfNotExists([]byte(UnprocessedQueue))
	if err != nil {
		return nil, err
	}

	err = unprocessedBucket.ForEach(q.recoverUnprocessed)
	if err != nil {
		return nil, err
	}

	// Atomically set the state to Recovering
	err = q.setState(tx, QueueRecovering, force)
	if err != nil {
		return nil, err
	}
	err = q.recoverInflight(tx)

	// Atomically set the state to Opened
	err = q.setState(tx, QueueOpened, force)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return q, nil
}

func (q *DurableQueue) recoverUnprocessed(k, v []byte) error {
	idx, err := bytesToInt64(k)
	if err != nil {
		return err
	}
	if idx < q.unprocessed.head || q.unprocessed.head == 0 {
		q.unprocessed.head = idx - 1
	}
	if idx > q.unprocessed.tail || q.unprocessed.tail == 0 {
		q.unprocessed.tail = idx
	}
	return nil
}

func (q *DurableQueue) getState() (QueueState, error) {
	var queueStateBytes []byte
	err := q.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(InternalBucket))
		if b == nil {
			return NewNotFoundError("cannot find internal bucket")
		}
		queueStateBytes = b.Get(queueStateKey)
		if queueStateBytes == nil {
			return NewNotFoundError("cannot find queue state in internal bucket")
		}
		return nil
	})
	if err != nil {
		return QueueUnknown, err
	}
	return bytesToQueueState(queueStateBytes)
}

func (q *DurableQueue) setState(tx *bolt.Tx, queueState QueueState, force bool) error {
	var b *bolt.Bucket
	var err error
	b, err = tx.CreateBucketIfNotExists([]byte(InternalBucket))
	if err != nil {
		return NewInternalError("cannot create or find internal bucket")
	}

	currQueueStateBytes := b.Get(queueStateKey)
	if !force && currQueueStateBytes != nil {
		var currQueueState QueueState
		currQueueState, err = bytesToQueueState(currQueueStateBytes)
		if !validStateTransition(currQueueState, queueState) {
			msg := fmt.Sprintf("cannot transition states from %v to %v", currQueueState, queueState)
			return NewInvalidError(msg)
		}
	}

	queueStateBytes, err := queueStateToBytes(queueState)
	if err != nil {
		return err
	}

	err = b.Put(queueStateKey, queueStateBytes)
	if err != nil {
		return err
	}
	q.state = queueState
	return nil
}

func (q *DurableQueue) getItem(bucketType BucketType, idx int64) (*QueueItem, error) {
	var queueItem *QueueItem
	err := q.db.View(func(tx *bolt.Tx) error {
		var err error
		b := tx.Bucket([]byte(bucketType))
		if b == nil {
			return NewNotFoundError(fmt.Sprintf("cannot find %v bucket", bucketType))
		}
		keyBytes, err := int64ToBytes(idx)
		if err != nil {
			return err
		}

		itemBytes := b.Get(keyBytes)
		if itemBytes == nil {
			return NewNotFoundError(fmt.Sprintf("cannot find %s in bucket %v", string(keyBytes), bucketType))
		}
		queueItem, err = bytesToQueueItem(itemBytes)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return queueItem, nil
}

func (q *DurableQueue) putItem(bucketType BucketType, item *QueueItem)  error {
	err := q.db.Update(func(tx *bolt.Tx) error {
		var err error
		b := tx.Bucket([]byte(bucketType))
		if b == nil {
			return NewNotFoundError(fmt.Sprintf("cannot find %v bucket", bucketType))
		}
		itemBytes, err := queueItemToBytes(item)
		if err != nil {
			return err
		}
		keyBytes, err := int64ToBytes(item.Idx)
		if err != nil {
			return err
		}

		err = b.Put(keyBytes, itemBytes)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (q *DurableQueue) deleteItem(bucketType BucketType, idx int64) error {
	err := q.db.Update(func(tx *bolt.Tx) error {
		var err error
		b := tx.Bucket([]byte(bucketType))
		if b == nil {
			return NewNotFoundError(fmt.Sprintf("cannot find %v bucket", bucketType))
		}
		keyBytes, err := int64ToBytes(idx)
		if err != nil {
			return err
		}

		err = b.Delete(keyBytes)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}


func (q *DurableQueue) Close() error {
	q.lock.Lock()
	defer q.lock.Unlock()
	_ = q.db.Update(func(tx *bolt.Tx) error {
		return q.setState(tx, QueueClosed, true)
	})
	return q.db.Close()
}

func (q *DurableQueue) Drop() error {
	q.lock.Lock()
	defer q.lock.Unlock()

	if q.state == QueueClosed || q.state == QueueUnknown {
		return os.Remove(q.dbFile)
	}
	return NewInvalidError(fmt.Sprintf("cannot drop the queue store when it is open"))
}

func (q *DurableQueue) checkOpened(errDetail string) error {
	if q.state != QueueOpened {
		msg := fmt.Sprintf("'%s' is closed, cannot '%s'", q.dbFile, errDetail)
		return NewClosedError(msg)
	}
	return nil
}

func (q *DurableQueue) Dequeue() (*QueueItem, error) {
	q.lock.Lock()
	defer q.lock.Unlock()
	if err := q.checkOpened("dequeue"); err != nil {
		return nil, err
	}

	// Nothing to process
	if q.unprocessed.head == q.unprocessed.tail {
		return nil, NewEmptyQueue("empty")
	}

	queueItem, err := q.getItem(UnprocessedQueue, q.unprocessed.head + 1)
	if err != nil {
		return nil, err
	}

	err = q.putItem(InflightQueue, queueItem)
	if err != nil {
		return nil, err
	}

	err = q.deleteItem(UnprocessedQueue, queueItem.Idx)
	if err != nil {
		return nil, err
	}

	q.numInflight++
	q.unprocessed.head++

	return queueItem, nil
}

func (q *DurableQueue) Enqueue(v []byte) error {
	q.lock.Lock()
	defer q.lock.Unlock()
	if err := q.checkOpened("enqueue"); err != nil {
		return err
	}

	queueItem := NewQueueItem(v, q.unprocessed.tail + 1)

	err := q.putItem(UnprocessedQueue, queueItem)
	if err != nil {
		return err
	}

	q.unprocessed.tail++

	return nil
}

func (q *DurableQueue) Ack(queueItem *QueueItem) error {
	q.lock.Lock()
	defer q.lock.Unlock()
	if err := q.checkOpened("ack"); err != nil {
		return err
	}

	err := q.deleteItem(InflightQueue, queueItem.Idx)
	if err != nil {
		return err
	}
	q.numInflight--
	return nil
}

func (q *DurableQueue) Length() int64 {
	return q.unprocessed.tail - q.unprocessed.head
}

func (q *DurableQueue) NumInflight() int64 {
	return q.numInflight
}

func (q *DurableQueue) GetUnprocessed() ([]*QueueItem, error) {
	q.lock.Lock()
	defer q.lock.Unlock()

	var unprocessedItems []*QueueItem

	err := q.db.View(func(tx *bolt.Tx) error {
		unprocessedBucket := tx.Bucket([]byte(UnprocessedQueue))
		if unprocessedBucket == nil {
			return NewInternalError(fmt.Sprintf("cannot find unprocessed queue"))
		}

		return unprocessedBucket.ForEach(func(k, v []byte) error {
			queueItem, err := bytesToQueueItem(v)
			if err != nil {
				return err
			}
			unprocessedItems = append(unprocessedItems, queueItem)
			return nil
		})
	})

	if err != nil {
		return nil, err
	}

	return unprocessedItems, nil
}

func (q *DurableQueue) GetInflight() ([]*QueueItem, error) {
	q.lock.Lock()
	defer q.lock.Unlock()

	var inflightItems []*QueueItem

	err := q.db.View(func(tx *bolt.Tx) error {
		inflightBucket := tx.Bucket([]byte(InflightQueue))
		if inflightBucket == nil {
			return NewInternalError(fmt.Sprintf("cannot find inflight queue"))
		}

		return inflightBucket.ForEach(func(k, v []byte) error {
			queueItem, err := bytesToQueueItem(v)
			if err != nil {
				return err
			}
			inflightItems = append(inflightItems, queueItem)
			return nil
		})
	})

	if err != nil {
		return nil, err
	}

	return inflightItems, nil
}

func (q *DurableQueue) recoverInflight(tx *bolt.Tx) error {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.numInflight = 0

	inflightRecoveryMap := make(map[int64]error)

	inflightBucket, err := tx.CreateBucketIfNotExists([]byte(InflightQueue))
	if err != nil {
		return err
	}

	err = inflightBucket.ForEach(func(k, v []byte) error {
		id, err := bytesToInt64(k)
		if err != nil {
			return err
		}
		inflightRecoveryMap[id] = q.recoverFunc(v)
		q.numInflight++
		return nil
	})

	if err != nil {
		return err
	}

	for k, v := range inflightRecoveryMap {
		idBytes, err := int64ToBytes(k)
		if err != nil {
			return err
		}
		if v == nil {
			err = inflightBucket.Delete(idBytes)
			if err != nil {
				return err
			}
			q.numInflight--
		}
	}

	return nil
}
