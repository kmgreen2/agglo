package streaming

import (
	"context"
	"fmt"
	"github.com/kmgreen2/agglo/pkg/common"
)

// MemSubscriber is an object used to subscribe to a single topic
type MemSubscriber struct {
	memPubSub *MemPubSub
	ctx *SubscriberContext
	state SubscriberState
}

// NewMemSubscriber will return a new MemSubscriber object
func NewMemSubscriber(memPubSub *MemPubSub, topic string) (*MemSubscriber, error) {
	if memPubSub.HasTopic(topic) {
		return &MemSubscriber {
			memPubSub: memPubSub,
			state: SubscriberStopped,
			ctx: &SubscriberContext{
				topic: topic,
				offset: 0,
				maxOffset: int64(^uint64(0) >> 1),
			},
		}, nil
	}

	return nil, common.NewInvalidError(fmt.Sprintf(
		"MemReplayer - cannot create replayer for non-existent topic: %s", topic))
}

// terminateSubscriber will return true if the error is fatal and false otherwise
func terminateSubscriber(err error) bool {
	return false
}

// Subscribe will spawn a go routine that subscribes from a topic using the provided handler
func (memSubscriber *MemSubscriber) Subscribe(handler func(ctx context.Context, payload []byte)) error {
	if memSubscriber.state == SubscriberRunning {
		return nil
	}
	memSubscriber.state = SubscriberRunning
	go func() {
		for {
			if memSubscriber.state == SubscriberStopped {
				break
			}
			payload, err := memSubscriber.memPubSub.Next(memSubscriber.ctx)
			if err != nil {
				if terminateSubscriber(err) {
					break
				}
			} else {
				handler(common.ExtractPubSubContext(payload), payload)
			}
		}
	}()
	return nil
}

// Stop will stop the go routine that is consuming from the topic
func (memSubscriber *MemSubscriber) Stop() error {
	memSubscriber.state = SubscriberStopped
	return nil
}

// Status will return the status of the subscriber
func (memSubscriber *MemSubscriber) Status() SubscriberState {
	return memSubscriber.state
}

// ConnectionString will return a string that can be parsed to connect to the underlying pub/sub system
func (memSubscriber *MemSubscriber) ConnectionString() string {
	return "inMemPubSub"
}
