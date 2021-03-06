package streaming

import (
	"context"
	"time"
)

type Publisher interface {
	Publish(ctx context.Context, b []byte) error
	Flush(ctx context.Context, timeout time.Duration) error
	Close() error
	ConnectionString() string
}

type Subscriber interface {
	Subscribe(handler func(ctx context.Context, payload []byte)) error
	Stop() error
	Status() SubscriberState
	ConnectionString() string
}

type Replayer interface {
	Replay(ctx context.Context, handler func(ctx context.Context, payload []byte) error) error
	ConnectionString() string
}

// StreamingCtrl is really only for testing purposes.  This package is in no place to be creating infrastructure
// resources.
type StreamingCtrl interface {
	CreateTopic(ctx context.Context, topicName string) error
}