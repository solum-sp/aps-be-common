package kafka

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
type Option func(*KafkaProducerConfig, *KafkaConsumerConfig, *SchemaRegistryConfig)

// WithBrokers sets Kafka brokers
func WithBrokers(brokers string) Option {
	return func(p *KafkaProducerConfig, c *KafkaConsumerConfig, _ *SchemaRegistryConfig) {
		p.Brokers = brokers
		c.Brokers = brokers
	}
}

// WithClientID sets Kafka client ID for producer
func WithClientID(clientID string) Option {
	return func(p *KafkaProducerConfig, _ *KafkaConsumerConfig, _ *SchemaRegistryConfig) {
		p.ClientID = clientID
	}
}

// WithConsumerGroupID sets Kafka consumer group ID
func WithConsumerGroupID(groupID string) Option {
	return func(_ *KafkaProducerConfig, c *KafkaConsumerConfig, _ *SchemaRegistryConfig) {
		c.GroupID = groupID
	}
}

// WithSchemaRegistryURL sets the schema registry URL
func WithSchemaRegistryURL(url string) Option {
	return func(_ *KafkaProducerConfig, _ *KafkaConsumerConfig, s *SchemaRegistryConfig) {
		s.URL = url
	}
}

// WithAutoOffsetReset sets the auto offset reset policy
func WithAutoOffsetReset(offset string) Option {
	return func(_ *KafkaProducerConfig, c *KafkaConsumerConfig, _ *SchemaRegistryConfig) {
		c.AutoOffsetReset = offset
	}
}

func WithEnableAutoCommit(enable bool) Option {
	return func(_ *KafkaProducerConfig, c *KafkaConsumerConfig, _ *SchemaRegistryConfig) {
		c.EnableAutoCommit = enable
	}
}

// WithMaxPollIntervalMs sets the max poll interval
func WithMaxPollIntervalMs(ms int) Option {
	return func(_ *KafkaProducerConfig, c *KafkaConsumerConfig, _ *SchemaRegistryConfig) {
		c.MaxPollIntervalMs = ms
	}
}

// WithSessionTimeoutMs sets the session timeout
func WithSessionTimeoutMs(ms int) Option {
	return func(_ *KafkaProducerConfig, c *KafkaConsumerConfig, _ *SchemaRegistryConfig) {
		c.SessionTimeoutMs = ms
	}
}

// WithHeartbeatIntervalMs sets the heartbeat interval
func WithHeartbeatIntervalMs(ms int) Option {
	return func(_ *KafkaProducerConfig, c *KafkaConsumerConfig, _ *SchemaRegistryConfig) {
		c.HeartbeatIntervalMs = ms
	}
}

// WithRetryBackoffMs sets the retry backoff time
func WithRetryBackoffMs(ms int) Option {
	return func(_ *KafkaProducerConfig, c *KafkaConsumerConfig, _ *SchemaRegistryConfig) {
		c.RetryBackoffMs = ms
	}
}

// WithFetchMinBytes sets the fetch minimum bytes
func WithFetchMinBytes(bytes int) Option {
	return func(_ *KafkaProducerConfig, c *KafkaConsumerConfig, _ *SchemaRegistryConfig) {
		c.FetchMinBytes = bytes
	}
}

// WithFetchWaitMaxMs sets the fetch wait max time
func WithFetchWaitMaxMs(ms int) Option {
	return func(_ *KafkaProducerConfig, c *KafkaConsumerConfig, _ *SchemaRegistryConfig) {
		c.FetchWaitMaxMs = ms
	}
}
