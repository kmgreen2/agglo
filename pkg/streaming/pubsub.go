package streaming

import (
	"context"
	"time"
)

type Publisher interface {
	Publish(ctx context.Context, b []byte) error
	Flush(ctx context.Context, timeout time.Duration) error
}

type Subscriber interface {
	Subscribe(handler func(ctx context.Context, payload []byte)) error
	Stop() error
	Status() SubscriberState
}

type Replayer interface {
	Replay(handler func(ctx context.Context, payload []byte) error) error
}