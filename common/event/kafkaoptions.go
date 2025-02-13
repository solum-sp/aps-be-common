package event

// KafkaProducerConfig holds Kafka producer settings
type KafkaProducerConfig struct {
	Brokers  string
	ClientID string
}

// KafkaConsumerConfig holds Kafka consumer settings
type KafkaConsumerConfig struct {
	Brokers             string
	GroupID             string
	AutoOffsetReset     string
	EnableAutoCommit    bool
	MaxPollIntervalMs   int
	SessionTimeoutMs    int
	HeartbeatIntervalMs int
	RetryBackoffMs      int
	FetchMinBytes       int
	FetchWaitMaxMs      int
}

// SchemaRegistryConfig holds Schema Registry settings
type SchemaRegistryConfig struct {
	URL string
}

// DefaultConfig holds the default Kafka settings
var DefaultConfig = struct {
	Producer KafkaProducerConfig
	Consumer KafkaConsumerConfig
	Schema   SchemaRegistryConfig
}{
	Producer: KafkaProducerConfig{
		Brokers:  "localhost:9092",
		ClientID: "default-client",
	},
	Consumer: KafkaConsumerConfig{
		Brokers:             "localhost:9092",
		GroupID:             "default-group",
		AutoOffsetReset:     "earliest",
		EnableAutoCommit:    false,
		MaxPollIntervalMs:   300000,
		SessionTimeoutMs:    45000,
		HeartbeatIntervalMs: 3000,
		RetryBackoffMs:      100,
		FetchMinBytes:       1,
		FetchWaitMaxMs:      500,
	},
	Schema: SchemaRegistryConfig{
		URL: "http://localhost:8081",
	},
}

// Option is a functional option for configuring Kafka
type KafkaOption func(*KafkaProducerConfig, *KafkaConsumerConfig, *SchemaRegistryConfig)

// WithBrokers sets Kafka brokers
func WithKafkaBrokers(brokers string) KafkaOption {
	return func(p *KafkaProducerConfig, c *KafkaConsumerConfig, _ *SchemaRegistryConfig) {
		p.Brokers = brokers
		c.Brokers = brokers
	}
}

// WithKafkaClientID sets Kafka client ID for producer
func WithKafkaClientID(clientID string) KafkaOption {
	return func(p *KafkaProducerConfig, _ *KafkaConsumerConfig, _ *SchemaRegistryConfig) {
		p.ClientID = clientID
	}
}

// WithKafkaConsumerGroupID sets Kafka consumer group ID
func WithKafkaConsumerGroupID(groupID string) KafkaOption {
	return func(_ *KafkaProducerConfig, c *KafkaConsumerConfig, _ *SchemaRegistryConfig) {
		c.GroupID = groupID
	}
}

// WithKafkaSchemaRegistryURL sets the schema registry URL
func WithKafkaSchemaRegistryURL(url string) KafkaOption {
	return func(_ *KafkaProducerConfig, _ *KafkaConsumerConfig, s *SchemaRegistryConfig) {
		s.URL = url
	}
}

// WithKafkaAutoOffsetReset sets the auto offset reset policy
func WithKafkaAutoOffsetReset(offset string) KafkaOption {
	return func(_ *KafkaProducerConfig, c *KafkaConsumerConfig, _ *SchemaRegistryConfig) {
		c.AutoOffsetReset = offset
	}
}

func WithKafkaEnableAutoCommit(enable bool) KafkaOption {
	return func(_ *KafkaProducerConfig, c *KafkaConsumerConfig, _ *SchemaRegistryConfig) {
		c.EnableAutoCommit = enable
	}
}

// WithKafkaMaxPollIntervalMs sets the max poll interval
func WithKafkaMaxPollIntervalMs(ms int) KafkaOption {
	return func(_ *KafkaProducerConfig, c *KafkaConsumerConfig, _ *SchemaRegistryConfig) {
		c.MaxPollIntervalMs = ms
	}
}

// WithKafkaSessionTimeoutMs sets the session timeout
func WithKafkaSessionTimeoutMs(ms int) KafkaOption {
	return func(_ *KafkaProducerConfig, c *KafkaConsumerConfig, _ *SchemaRegistryConfig) {
		c.SessionTimeoutMs = ms
	}
}

// WithKafkaHeartbeatIntervalMs sets the heartbeat interval
func WithKafkaHeartbeatIntervalMs(ms int) KafkaOption {
	return func(_ *KafkaProducerConfig, c *KafkaConsumerConfig, _ *SchemaRegistryConfig) {
		c.HeartbeatIntervalMs = ms
	}
}

// WithKafkaRetryBackoffMs sets the retry backoff time
func WithKafkaRetryBackoffMs(ms int) KafkaOption {
	return func(_ *KafkaProducerConfig, c *KafkaConsumerConfig, _ *SchemaRegistryConfig) {
		c.RetryBackoffMs = ms
	}
}

// WithKafkaFetchMinBytes sets the fetch minimum bytes
func WithKafkaFetchMinBytes(bytes int) KafkaOption {
	return func(_ *KafkaProducerConfig, c *KafkaConsumerConfig, _ *SchemaRegistryConfig) {
		c.FetchMinBytes = bytes
	}
}

// WithKafkaFetchWaitMaxMs sets the fetch wait max time
func WithKafkaFetchWaitMaxMs(ms int) KafkaOption {
	return func(_ *KafkaProducerConfig, c *KafkaConsumerConfig, _ *SchemaRegistryConfig) {
		c.FetchWaitMaxMs = ms
	}
}
