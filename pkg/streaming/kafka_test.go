package streaming_test

import (
	"github.com/kmgreen2/agglo/pkg/streaming"
	"github.com/stretchr/testify/assert"
	"testing"
	"context"
	"fmt"
)

func TestKafkaHappyPath(t *testing.T) {
	kafkaCtrl, err := streaming.NewKafkaCtrl("localhost:9092")
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	err = kafkaCtrl.CreateTopic(context.Background(), "testing")
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	kafkaPublisher, err := streaming.NewKafkaPublisher([]string{"localhost:9092"}, "testing", true)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	kafkaSubscriber, err := streaming.NewKafkaSubscriber([]string{"localhost:9092"}, "testing", "testingGroup", 10000)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	numMessages := 10

	go func() {
		for i := 0; i < numMessages; i++ {
			kafkaPublisher.Publish(context.Background(), []byte(fmt.Sprintf("%d", i)))
		}
	}()

	numConsumed := 0

	consumeChannel := make(chan string, numMessages)

	handler := func(ctx context.Context, payload []byte) {
		consumeChannel <- string(payload)
		numConsumed++
		if numConsumed == numMessages {
			close(consumeChannel)
			kafkaSubscriber.Stop()
		}
	}

	err = kafkaSubscriber.Subscribe(handler)
	assert.Nil(t, err)

	i := 0
	for val := range consumeChannel {
		assert.Equal(t, fmt.Sprintf("%d", i), val)
		i++
	}
}