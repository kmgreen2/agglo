package streaming_test

import (
	"fmt"
	"github.com/kmgreen2/agglo/pkg/streaming"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBasicHappyPath(t *testing.T) {
	memPubSub := streaming.NewMemPubSub()
	memPubSub.CreateTopic("foo")
	fooPub, err := streaming.NewMemPublisher(memPubSub, "foo")
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	fooSub, err := streaming.NewMemSubscriber(memPubSub, "foo")
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	numMessages := 10

	go func() {
		for i := 0; i < numMessages; i++ {
			fooPub.Publish([]byte(fmt.Sprintf("%d", i)))
		}
	}()

	numConsumed := 0

	consumeChannel := make(chan string)

	handler := func(payload []byte) {
		consumeChannel <- string(payload)
		numConsumed++
		if numConsumed == numMessages {
			close(consumeChannel)
			fooSub.Stop()
		}
	}

	err = fooSub.Subscribe(handler)
	assert.Nil(t, err)

	i := 0
	for val := range consumeChannel {
		assert.Equal(t, fmt.Sprintf("%d", i), val)
		i++
	}
}

func TestMultipleConsumers(t *testing.T) {

}

func TestReplayer(t *testing.T) {

}

func TestCreateSubscriberNonTopic(t *testing.T) {

}

func TestCreatePublisherNonTopic(t *testing.T) {

}

func TestSubscribeNonTopic(t *testing.T) {

}

func TestPublishNonTopic(t *testing.T) {

}
