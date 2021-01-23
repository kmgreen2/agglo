package streaming

import (
	"context"
	"fmt"
	"github.com/kmgreen2/agglo/pkg/util"
)

// MemSubscriber is an object used to subscribe to a single topic
type MemReplayer struct {
	memPubSub *MemPubSub
	ctx *SubscriberContext
	state SubscriberState
	start int64
	end int64
	current int64
}

// NewMemReplayer will return a new MemReplayer object
func NewMemReplayer(memPubSub *MemPubSub, topic string, start, end int64) (*MemReplayer, error) {
	if memPubSub.HasTopic(topic) {
		return &MemReplayer {
			memPubSub: memPubSub,
			state: SubscriberStopped,
			start: start,
			end: end,
			current: start,
			ctx: &SubscriberContext{
				topic: topic,
				offset: start,
				maxOffset: end,
			},
		}, nil
	}

	return nil, util.NewInvalidError(fmt.Sprintf(
		"MemReplayer - cannot create replayer for non-existent topic: %s", topic))
}

// Replay will replay the stream based on the replayer config and the provided handler function
func (memReplayer *MemReplayer) Replay(ctx context.Context, handler func(ctx context.Context,
	payload []byte) error) error {
	for {
		payload, err := memReplayer.memPubSub.Next(memReplayer.ctx)
		if err != nil {
			return err
		}
		err = handler(ctx, payload)
		if err != nil {
			return err
		}
	}
	return nil
}

// ConnectionString will return a string that can be parsed to connect to the underlying pub/sub system
func (memReplayer *MemReplayer) ConnectionString() string {
	return "inMemPubSub"
}
