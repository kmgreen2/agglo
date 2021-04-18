package streaming

import (
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/kmgreen2/agglo/pkg/util"
	"strconv"
	"strings"
	"time"
)

type KafkaCtrl struct {
	client *kafka.AdminClient
}

func NewKafkaCtrl(broker string) (*KafkaCtrl, error) {
	client, err := kafka.NewAdminClient(&kafka.ConfigMap{
		"bootstrap.servers": broker,
	})
	if err != nil {
		return nil, err
	}
	return &KafkaCtrl {
		client: client,
	}, nil
}

// CreateTopic will create a Kafka topic with 1 partition and replication factor of 1
// Note: This is only for testing!  Topic creation should be done out of band of this
// library and framework
func (ctrl *KafkaCtrl) CreateTopic(ctx context.Context, topicName string) error {
	_, err := ctrl.client.CreateTopics(ctx, []kafka.TopicSpecification{{
		Topic: topicName,
		NumPartitions: 1,
		ReplicationFactor: 1,
	}})
	return err
}

type KafkaPublisher struct {
	producer *kafka.Producer
	servers []string
	topicName string
	isSync bool
}

func NewKafkaPublisher(servers []string, topicName string, isSync bool) (*KafkaPublisher, error) {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": strings.Join(servers, ","),
	})
	if err != nil {
		return nil, err
	}
	return &KafkaPublisher{
		producer,
		servers,
		topicName,
		isSync,
	}, nil
}

func NewKafkaPublisherFromConnectionString(connectionString string) (*KafkaPublisher, error) {
	var servers []string
	var topicName string
	var isSync bool

	connectionStringAry := strings.Split(connectionString, ",")
	for _, entry := range connectionStringAry {
		entryAry := strings.Split(entry, "=")
		if len(entryAry) != 2 {
			return nil, util.NewInvalidError(fmt.Sprintf("invalid entry in connection string: %s", entry))
		}
		switch entryAry[0] {
		case "servers":
			servers = strings.Split(entryAry[1], ",")
		case "topicName":
			topicName = entryAry[1]
		case "isSync":
			if strings.Compare("true", strings.ToLower(entryAry[1])) == 0 {
				isSync = true
			} else if strings.Compare("false", strings.ToLower(entryAry[1])) == 0 {
				isSync = false
			} else {
				return nil, util.NewInvalidError(
					fmt.Sprintf("isSync should be 'true' or 'false', got %s", entryAry[1]))
			}
		}
	}

	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": servers,
	})
	if err != nil {
		return nil, err
	}

	return &KafkaPublisher{
		producer,
		servers,
		topicName,
		isSync,
	}, nil
}

func (publisher *KafkaPublisher) Publish(ctx context.Context, b []byte) error {
	var deliveryChan chan kafka.Event = nil

	if publisher.isSync {
		deliveryChan = make(chan kafka.Event, 1)
	}

	err := publisher.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &publisher.topicName, Partition: kafka.PartitionAny},
		Value: b,
	}, deliveryChan)
	if err != nil {
		return err
	}

	// If synchronous, then wait for timeout or for producer to signal completion
	if deliveryChan != nil {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case e := <-deliveryChan:
			switch ev := e.(type) {
			case *kafka.Message:
				return nil
			case kafka.Error:
				return ev
			default:
				return nil
			}
		}
	}
	return nil
}

func (publisher *KafkaPublisher) Flush(ctx context.Context, timeout time.Duration) error {
	numUnFlushedMessages := publisher.producer.Flush(int(timeout.Milliseconds()))
	if numUnFlushedMessages > 0 {
		return util.NewFlushDidNotCompleteError(fmt.Sprintf("%d messages were not flushed", numUnFlushedMessages))
	}
	return nil
}

func (publisher *KafkaPublisher) Close() error {
	publisher.producer.Close()
	return nil
}

func (publisher *KafkaPublisher) ConnectionString() string {
	return fmt.Sprintf("servers=%s,topicName=%s,isSync=%s", strings.Join(publisher.servers, ","), publisher.topicName,
		strconv.FormatBool(publisher.isSync))
}

type KafkaSubscriber struct {
	consumer *kafka.Consumer
	servers []string
	group string
	topicName string
	state SubscriberState
}

func NewKafkaSubscriber(servers []string, topicName, group string, sessionTimeoutMs int) (*KafkaSubscriber, error) {
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": strings.Join(servers, ","),
		"broker.address.family": "v4",
		"session.timeout.ms": sessionTimeoutMs,
		"auto.offset.reset": "earliest",
		"group.id": group,
	})
	if err != nil {
		return nil, err
	}
	return &KafkaSubscriber {
		consumer,
		servers,
		group,
		topicName,
		SubscriberStopped,
	}, nil
}

func (subscriber *KafkaSubscriber) Subscribe(handler func(ctx context.Context, payload []byte)) error {
	err := subscriber.consumer.SubscribeTopics([]string{subscriber.topicName}, nil)
	if err != nil {
		return err
	}

	subscriber.state = SubscriberRunning

	for subscriber.state == SubscriberRunning {
		event := subscriber.consumer.Poll(100)
		switch e := event.(type) {
		case *kafka.Message:
			handler(context.Background(), e.Value)
		case kafka.Error:
			if e.Code() == kafka.ErrAllBrokersDown {
				_ = subscriber.Stop()
				return util.NewInternalError(e.Error())
			}
		}
	}
	return nil
}

func (subscriber *KafkaSubscriber) Stop() error {
	subscriber.state = SubscriberStopped
	return nil
}

func (subscriber *KafkaSubscriber)  Status() SubscriberState {
	return subscriber.state
}

func (subscriber *KafkaSubscriber)  ConnectionString() string {
	return fmt.Sprintf("servers=%s,topicName=%s,group=%s", strings.Join(subscriber.servers, ","), subscriber.topicName,
		subscriber.group)
}
