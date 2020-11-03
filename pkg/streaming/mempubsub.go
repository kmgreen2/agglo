package streaming

import (
	"errors"
	"fmt"
	"github.com/kmgreen2/agglo/pkg/common"
	"sync"
)

type SubscriberState int

const (
	SubscriberRunning = iota
	SubscriberStopped
)

// SubscriberContext contains the context for a subscriber
type SubscriberContext struct {
	topic string
	offset int64
	maxOffset int64
}

// MemTopic is an in-memory queue for a single topic
type MemTopic struct {
	messageQueue [][]byte
	lock sync.Locker
	cond *sync.Cond
}

// MemPubSub is an in-memory pubsub system (mostly used for testing)
type MemPubSub struct {
	memTopics map[string]*MemTopic
	lock sync.Locker
}

// NewMemPubSub will return a new connection object to the in-memory pubsub system
func NewMemPubSub() *MemPubSub {
	return &MemPubSub{
		memTopics: make(map[string]*MemTopic),
		lock: &sync.Mutex{},
	}
}

// NewMemTopic will create a new MemTopic object
func NewMemTopic() *MemTopic {
	return &MemTopic {
		messageQueue: make([][]byte, 0),
		lock: &sync.Mutex{},
		cond: sync.NewCond(&sync.Mutex{}),
	}
}

// Publish will append the payload to the topic
func (memTopic *MemTopic) Publish(payload []byte) error {
	memTopic.lock.Lock()
	defer memTopic.lock.Unlock()

	memTopic.messageQueue = append(memTopic.messageQueue, payload)

	return nil
}

// Next will return the message at a given element
func (memTopic *MemTopic) Get(index int64) ([]byte, error) {
	numMessages := len(memTopic.messageQueue)
	if numMessages >= int(index) {
		return nil, common.NewOutOfBoundsError(fmt.Sprintf("MemPubSub - index out of bounds: %d >= %d", index,
			numMessages))
	}
	return memTopic.messageQueue[index], nil
}

// CreateTopic will create a new topic
func (pubSub *MemPubSub) CreateTopic(name string) error {
	pubSub.lock.Lock()
	defer pubSub.lock.Unlock()

	if _, ok := pubSub.memTopics[name]; ok {
		return common.NewConflictError(fmt.Sprintf("MemPubSub - cannot create topic that exists: %s", name))
	}
	pubSub.memTopics[name] = NewMemTopic()
	return nil
}

// Publish will publish a message to the specified topic
func (pubSub *MemPubSub) Publish(topic string, payload []byte) error {
	if _, ok := pubSub.memTopics[topic]; !ok {
		return common.NewNotFoundError(fmt.Sprintf("MemPubSub - cannot publish to non-existent topic: %s",
			topic))
	}
	err := pubSub.memTopics[topic].Publish(payload)
	if err != nil {
		return err
	}
	pubSub.memTopics[topic].cond.Broadcast()
	return nil
}

// HasTopic will return true if the topic exists and false otherwise
func (pubSub *MemPubSub) HasTopic(topic string) bool {
	_, ok := pubSub.memTopics[topic]
	return ok
}

// Next will get the next message based on the provided context and topic
func (pubSub *MemPubSub) Next(ctx *SubscriberContext) ([]byte, error) {
	if _, ok := pubSub.memTopics[ctx.topic]; !ok {
		return nil, common.NewNotFoundError(fmt.Sprintf("MemPubSub - cannot subscribe from non-existent topic: %s",
			ctx.topic))
	}
	if ctx.offset > ctx.maxOffset {
		return nil, common.NewEndOfStreamError(fmt.Sprintf("MemPubSub - end of stream"))
	}
	message, err := pubSub.memTopics[ctx.topic].Get(ctx.offset)
	if errors.Is(err, &common.OutOfBoundsError{}) {
		pubSub.memTopics[ctx.topic].cond.Wait()
	}
	if err != nil {
		return nil, err
	}
	ctx.offset++

	return message, nil
}

