package kafka

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde/avrov2"
	"github.com/solum-sp/aps-be-common/common/utils"
)

type kafkaSubscriber struct {
	consumer *kafka.Consumer
	serde    serde.Deserializer
	topic    string
}

var _ ISubscriber = (*kafkaSubscriber)(nil)

func NewKafkaSubscriber(consumer *kafka.Consumer, sr *SchemaRegistry, topic string) (*kafkaSubscriber, error) {
	serde, err := avrov2.NewDeserializer(sr.client, serde.ValueSerde, &avrov2.DeserializerConfig{
		DeserializerConfig: serde.DeserializerConfig{},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create avro serializer: %s", err)
	}
	return &kafkaSubscriber{consumer: consumer, serde: serde, topic: topic}, nil
}

func (s *kafkaSubscriber) SubscribeToTopic(ctx context.Context) error {
	err := s.consumer.SubscribeTopics([]string{s.topic}, nil)
	if err != nil {
		return fmt.Errorf("failed to subscribe to topic: %s", err)
	}
	return nil
}

func (s *kafkaSubscriber) ConsumeMessages(
	ctx context.Context,

	msgTypeConstructor func() ConsumerMessage,
) (<-chan ConsumerMessage, <-chan error, chan<- bool) {
	chMsg := make(chan ConsumerMessage)
	chCommitRequest := make(chan bool)
	chErr := make(chan error)
	go func() {
		defer close(chMsg)
		defer close(chErr)
		defer close(chCommitRequest)

		for {
			select {
			case <-ctx.Done():
				return
			default:
				msg, err := s.consumer.ReadMessage(100 * time.Millisecond)
				if err != nil {
					if err.(kafka.Error).Code() == kafka.ErrTimedOut {
						continue // Normal timeout, just retry
					}

					// Log and potentially retry or send to error channel
					chErr <- fmt.Errorf("consumer read error: %v", err)
					continue
				}

				// Deserialize and process message
				msgObj := msgTypeConstructor()
				err = s.serde.DeserializeInto(s.topic, msg.Value, &msgObj)
				if err != nil {
					chErr <- fmt.Errorf("deserialization error: %v", err)
					continue
				}
				log.Printf("Message on Topic: %s, Offset: %+v\n", *msg.TopicPartition.Topic, msg.TopicPartition.Offset)
				chMsg <- msgObj

				// Manual offset commit
				if <-chCommitRequest {
					_, err := s.consumer.CommitMessage(msg)
					if err != nil {
						chErr <- fmt.Errorf("offset commit error: %v", err)
					}
				}
			}
		}
	}()
	return chMsg, chErr, chCommitRequest
}

/*
USAGE EXAMPLE:
- DEFAULT DECLARATION

consumer, err := kafka.NewKafkaConsumer()

	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}

- CUSTOM DECLARATION:

consumer, err := service.NewKafkaConsumer(

	kafka.WithBrokers("custom-addr-of-broker"), //default addr == 'localhost:9092'
	kafka.WithClientID("custom-client-id"),

)
*/

func NewKafkaConsumer(opts ...Option) (*kafka.Consumer, error) {
	producerConfig := DefaultConfig.Producer
	consumerConfig := DefaultConfig.Consumer
	schemaConfig := DefaultConfig.Schema

	// Apply functional options
	for _, opt := range opts {
		opt(&producerConfig, &consumerConfig, &schemaConfig)
	}

	c, err := utils.Retry(10, 1*time.Second, func() (*kafka.Consumer, error) {
		return kafka.NewConsumer(&kafka.ConfigMap{
			"bootstrap.servers":     consumerConfig.Brokers,
			"group.id":              consumerConfig.GroupID,
			"auto.offset.reset":     consumerConfig.AutoOffsetReset,
			"enable.auto.commit":    consumerConfig.EnableAutoCommit,
			"max.poll.interval.ms":  consumerConfig.MaxPollIntervalMs,
			"session.timeout.ms":    consumerConfig.SessionTimeoutMs,
			"heartbeat.interval.ms": consumerConfig.HeartbeatIntervalMs,
			"retry.backoff.ms":      consumerConfig.RetryBackoffMs,
			"fetch.min.bytes":       consumerConfig.FetchMinBytes,
			"fetch.wait.max.ms":     consumerConfig.FetchWaitMaxMs,
		})
	})
	if err != nil {
		log.Printf("Failed to create kafka consumer: %s", err)
		return nil, err
	}
	return c, nil
}

func NewAdminClientFromConsumer(consumer *kafka.Consumer) (*kafka.AdminClient, error) {
	return kafka.NewAdminClientFromConsumer(consumer)
}
