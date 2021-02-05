package streaming

import (
	"context"
	"fmt"
	"github.com/kmgreen2/agglo/pkg/util"
	"time"
)

// MemPublisher is an object used to publish messages to a single topic
type MemPublisher struct {
	memPubSub *MemPubSub
	topic string
}

// NewMemPublisher returns a new publisher object
func NewMemPublisher(memPubSub *MemPubSub, topic string) (*MemPublisher, error) {
	if memPubSub.HasTopic(topic) {
		return &MemPublisher{
			memPubSub: memPubSub,
			topic:     topic,
		}, nil
	}

	return nil, util.NewInvalidError(fmt.Sprintf(
		"MemPublisher - cannot create publisher for non-existent topic: %s", topic))
}

// Publish will publish a payload
func (memPublisher *MemPublisher) Publish(ctx context.Context, payload []byte) error {
	return memPublisher.memPubSub.Publish(memPublisher.topic, payload)
}

// Flush is a no-op, since all changes immediately take effect
func (memPublisher *MemPublisher) Flush(ctx context.Context, timeout time.Duration) error {
	return nil
}

// ConnectionString will return a string that can be parsed to connect to the underlying pub/sub system
func (memPublisher *MemPublisher) ConnectionString() string {
	return "inMemPubSub"
}

// Close is a no-op, since all changes immediately take effect
func (memPublisher *MemPublisher) Close() error {
	return nil
}