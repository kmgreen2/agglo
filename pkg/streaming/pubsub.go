package streaming

import "time"

type Publisher interface {
	Publish(b []byte) error
	Flush(timeout time.Duration) error
}

type Subscriber interface {
	Subscribe(handler func(payload []byte)) error
	Stop() error
	Status() SubscriberState
}

type Replayer interface {
	Replay(handler func(payload []byte) error) error
}