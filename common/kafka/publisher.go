package kafka

import (
	"context"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde/avrov2"
)

type kafkaPublisher struct {
	producer *kafka.Producer
	serde    serde.Serializer
	topic    string
}

var _ IPublisher = (*kafkaPublisher)(nil)

func NewKafkaPublisher(producer *kafka.Producer, sr *SchemaRegistry, schemaID int, topic string) (*kafkaPublisher, error) {
	serde, err := avrov2.NewSerializer(sr.client, serde.ValueSerde, &avrov2.SerializerConfig{
		SerializerConfig: serde.SerializerConfig{
			AutoRegisterSchemas: false,
			UseSchemaID:         schemaID,
			UseLatestVersion:    true,
			NormalizeSchemas:    true,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create avro serializer: %s", err)
	}
	return &kafkaPublisher{producer: producer, serde: serde, topic: topic}, nil
}

func (s *kafkaPublisher) SendMessage(ctx context.Context, value interface{}) error {
	deliveryChan := make(chan kafka.Event)

	payload, err := s.serde.Serialize(s.topic, &value)
	if err != nil {
		return fmt.Errorf("failed to serialize: %s", err)
	}

	err = s.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &s.topic, Partition: kafka.PartitionAny},
		Value:          payload,
	}, deliveryChan)
	if err != nil {
		return fmt.Errorf("produce failed: %v", err)
	}

	e := <-deliveryChan
	m := e.(*kafka.Message)

	if m.TopicPartition.Error != nil {
		return fmt.Errorf("delivery failed: %v", m.TopicPartition.Error)
	}

	fmt.Printf("Delivered message to topic: %s [%d] at offset: %v\n",
		*m.TopicPartition.Topic, m.TopicPartition.Partition, m.TopicPartition.Offset)

	return nil
}

func NewKafkaProducer(opts ...Option) (*kafka.Producer, error) {
	producerConfig := DefaultConfig.Producer
	consumerConfig := DefaultConfig.Consumer
	schemaConfig := DefaultConfig.Schema

	for _, opt := range opts {
		opt(&producerConfig, &consumerConfig, &schemaConfig)
	}
	return kafka.NewProducer(&kafka.ConfigMap{
		// "bootstrap.servers": cfg.Brokers,
		// "client.id":         cfg.ClientID,
		"bootstrap.servers": producerConfig.Brokers,
		"client.id":         producerConfig.ClientID,
	})
}

func NewAdminClientFromProducer(producer *kafka.Producer) (*kafka.AdminClient, error) {
	return kafka.NewAdminClientFromProducer(producer)
}
