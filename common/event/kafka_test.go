package event

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test publisher

func TestNewKafkaProducer(t *testing.T) {
	prod, err := NewKafkaProducer()
	assert.NoError(t, err)
	assert.NotNil(t, prod)
}

func TestKafkaProducerOptions(t *testing.T) {
	producerConfig := DefaultConfig.Producer
	consumerConfig := DefaultConfig.Consumer
	schemaConfig := DefaultConfig.Schema

	opts := []KafkaOption{
		WithKafkaBrokers("kafka:9092"),
		WithKafkaClientID("test-client"),
		WithKafkaConsumerGroupID("test-group"),
		WithKafkaSchemaRegistryURL("http://schema-registry:8081"),
	}

	for _, opt := range opts {
		opt(&producerConfig, &consumerConfig, &schemaConfig)
	}

	assert.Equal(t, "kafka:9092", producerConfig.Brokers)
	assert.Equal(t, "test-client", producerConfig.ClientID)
	assert.Equal(t, "test-group", consumerConfig.GroupID)
	assert.Equal(t, "http://schema-registry:8081", schemaConfig.URL)
}

// Test subscriber
func TestNewKafkaConsumer(t *testing.T) {
	cons, err := NewKafkaConsumer()
	assert.NoError(t, err)
	assert.NotNil(t, cons)
}

func TestKafkaConsumerOptions(t *testing.T) {
	producerConfig := DefaultConfig.Producer
	consumerConfig := DefaultConfig.Consumer
	schemaConfig := DefaultConfig.Schema

	opts := []KafkaOption{
		WithKafkaBrokers("kafka:9092"),
		WithKafkaClientID("test-client"),
		WithKafkaConsumerGroupID("test-group"),
		WithKafkaSchemaRegistryURL("http://schema-registry:8081"),
	}

	for _, opt := range opts {
		opt(&producerConfig, &consumerConfig, &schemaConfig)
	}

	assert.Equal(t, "kafka:9092", producerConfig.Brokers)
	assert.Equal(t, "test-client", producerConfig.ClientID)
	assert.Equal(t, "test-group", consumerConfig.GroupID)
	assert.Equal(t, "http://schema-registry:8081", schemaConfig.URL)
}
