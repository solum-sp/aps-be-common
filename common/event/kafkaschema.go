package event

import (
	"embed"
	"fmt"
	"log"

	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde"
	"github.com/confluentinc/confluent-kafka-go/v2/schemaregistry/serde/avrov2"
	"github.com/solum-sp/aps-be-common/common/utils"
)

const (
	retryCount    = 10
	retryInterval = 1 * time.Second
)

type SchemaRegistry struct {
	client schemaregistry.Client
}

func NewSchemaRegistry(opts ...KafkaOption) (*SchemaRegistry, error) {

	producerConfig := DefaultConfig.Producer
	consumerConfig := DefaultConfig.Consumer
	schemaConfig := DefaultConfig.Schema

	// Apply functional options
	for _, opt := range opts {
		opt(&producerConfig, &consumerConfig, &schemaConfig)
	}
	sr, err := utils.Retry(retryCount, retryInterval, func() (schemaregistry.Client, error) {
		return schemaregistry.NewClient(schemaregistry.NewConfig(schemaConfig.URL))
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create schema registry client: %s", err)
	}
	return &SchemaRegistry{
		client: sr,
	}, nil
}

func (s *SchemaRegistry) Close() {
	s.client.Close()
}

func (s *SchemaRegistry) FindOrCreateArvoSchema(topic string, baseFS embed.FS, fullFileName string) (int, error) {
	schema, err := baseFS.ReadFile(fullFileName)
	if err != nil {
		return 0, fmt.Errorf("failed to read avro schema file: %s", err)
	}
	id, err := s.createAvroSchema(topic, schema)
	if err != nil {
		return 0, fmt.Errorf("failed to create avro schema: %s", err)
	}
	return id, nil

}

func (s *SchemaRegistry) createAvroSchema(topic string, schema []byte) (int, error) {
	id, err := utils.Retry(retryCount, retryInterval, func() (int, error) {
		return s.client.Register(topic+"-value", schemaregistry.SchemaInfo{
			Schema: string(schema),
		}, false)
	})
	if err != nil {
		return 0, fmt.Errorf("failed to register schema: %s", err)
	}
	return id, nil
}

func (s *SchemaRegistry) CreateAvroSerializer(schemaConfig avrov2.SerializerConfig) (*avrov2.Serializer, error) {
	serde, err := utils.Retry(retryCount, retryInterval, func() (*avrov2.Serializer, error) {
		return avrov2.NewSerializer(s.client, serde.ValueSerde, &schemaConfig)
	})
	if err != nil {
		log.Printf("Failed to create avro serializer: %s", err)
		return nil, err
	}
	return serde, nil
}

func (s *SchemaRegistry) CreateAvroDeserializer(schemaConfig avrov2.DeserializerConfig) (*avrov2.Deserializer, error) {
	serde, err := utils.Retry(retryCount, retryInterval, func() (*avrov2.Deserializer, error) {
		return avrov2.NewDeserializer(s.client, serde.ValueSerde, &schemaConfig)
	})
	if err != nil {
		log.Printf("Failed to create avro deserializer: %s", err)
		return nil, err
	}
	return serde, nil
}
