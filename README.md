# aps-be-common

Common Go packages and utilities for APS backend services.

## Installation

To use these common packages in your Go project, add this repository as a dependency:

```bash
go get -u github.com/solum-sp/aps-be-common
```

## Available Packages

The common packages are organized in the `common` directory and include:

### Config Package
- Environment-based configuration management
- Supports development, test, and production environments
- Automatic loading of `.env`, `.env.test`, and `.env.production` files
- Structured configuration using `AppConfig` with support for various service settings
- Define your configuration struct with environment variable tags:
    #### Basic usage
    ```go
    // Load the configuration
    cfg := &MyConfig{}
    config, err := config.NewAppConfig("./config", cfg)
    if err != nil {
        log.Fatal(err)
    }
    myConfig := config.(*MyConfig)
    ```

    Example `.env` file:
    ```env
    APP_NAME=microservice-base
    APP_ENV=development
    APP_VERSION=0.0.1

    HTTP_PORT=8000
    HTTP_HOST=localhost

    COCKROACH_URI=postgresql://root@localhost:26257/defaultdb?sslmode=disable

    MONGO_URI=mongodb://localhost:27017
    MONGO_DB=defaultdb

    REDIS_HOST=localhost:6379
    REDIS_PASSWORD=
    REDIS_DB=0
    ``` 

    #### Error Handling

    The configuration loader will return an error if:
    - A required field is missing
    - Environment file cannot be read
    - Type conversion fails

    Example error handling:
    ```go
    cfg := &MyConfig{}
    config, err := config.NewAppConfig("./config", cfg)
    if err != nil {
        switch {
        case strings.Contains(err.Error(), "required"):
            log.Fatal("Missing required configuration")
        case strings.Contains(err.Error(), "failed to load env"):
            log.Fatal("Failed to load environment file")
        default:
            log.Fatal("Configuration error:", err)
        }
    }
    ```

### Event Package
- Full Kafka producer and consumer implementations
- Support for Schema Registry with Avro serialization
- Configurable consumer groups and auto-commit settings
- Robust error handling and retry mechanisms
- Supports both synchronous and asynchronous message processing
- Built-in admin client functionality
    #### Basic usage
    ```go
    // create a new schema registry
    sr, err := event.NewSchemaRegistry(
        event.WithKafkaSchemaRegistryURL("http://localhost:8081"),
    )
    if err != nil {
        log.Fatal(err)
    }

    /// Create a new Kafka producer
    producer, err := event.NewKafkaProducer(event.WithKafkaBrokers("localhost:9092"), event.WithKafkaClientID("my-client"))
    if err != nil {
        log.Fatal(err)
    }

    publisher, err := event.NewKafkaPublisher(producer, sr, 1, "my-topic")
    if err != nil {
        log.Fatal(err)
    }
    publisher.SendMessage(context.Background(), "hello")

    /// Create a new Kafka consumer
    consumer, err := event.NewKafkaConsumer(
        event.WithKafkaBrokers("custom-addr-of-broker"), //default addr == 'localhost:9092'
        event.WithKafkaClientID("custom-client-id"),
    )
    if err != nil {
        log.Fatal(err)
    }

    subscriber, err := event.NewKafkaSubscriber(consumer, sr,1, "my-topic")
    if err != nil {
        log.Fatal(err)
    }

    subscriber.ConsumeMessages(context.Background(), func() event.ConsumerMessage {
		...
	})
    ```

### Logger Package
- Structured logging with multiple log levels (Debug, Info, Warn, Error, Fatal)
- Context-aware logging
- OpenTelemetry integration for distributed tracing
- Field-based logging with sanitization of sensitive data
- Stack trace capture for error logging
- Service name tagging for multi-service environments

    #### Basic usage
    ```go
    // Create a new logger
    logger := logger.NewLogger(logger.Config{
        Service: "my-service",
        Level:   logger.InfoLv,
    })

    // Log a message
    logger.Info(context.Background(), "Hello, World!")
    ```
    
### Cache Package
- Redis client implementation with connection pooling
- Support for key-value operations with expiration time
- Pattern-based key operations (get, delete)
- Service-specific key prefixing
- Bulk operations support (clear all, clear by pattern)
- Error handling and connection management

    #### Basic usage
    ```go
    // Create a new Redis client
    redisClient, err := cache.NewRedisCache(cache.RedisConfig{
        Addr:     "localhost:6379",
        Password: "",
        DB:       0,
        Service:  "my-service",
    })
    if err != nil {
        log.Fatal(err)
    }

    // Set a key-value pair
    err = redisClient.Set(context.Background(), "key", "value", 0)
    if err != nil {
        log.Fatal(err)
    }

    mutex, err := redisClient.Lock("test_key", 5*time.Second)
    if err != nil {
		fmt.Println("Could not acquire lock:", err)
		return
	}
	fmt.Println("Lock acquired")

    // Release the lock
	err = cache.Unlock(mutex)
	if err != nil {
		fmt.Println("Could not release lock:", err)
		return
	}
	fmt.Println("Lock released successfully")

    ```

### Token Package
- Secure token management using PASETO (Platform-Agnostic Security Tokens)
- Asymmetric key-based token generation and validation
- Built-in claims management (subject, user ID, session ID)
- Support for token expiration and issuance time
- Separate interfaces for token generation and validation

    #### Basic usage
    ```go
    // Initialize token manager (for auth service)
    tokenManager, err := token.NewPasetoTokenManager(privateKey)
    if err != nil {
        log.Fatal(err)
    }

    // Generate a token
    claims := token.TokenClaims{
        Sub:       "user123",
        UserId:    "user123",
        SessionId: "sess123",
        IssuedAt:  time.Now(),
        ExpiresAt: time.Now().Add(24 * time.Hour),
    }
    tokenString, err := tokenManager.GenerateToken(claims)

    // Initialize token parser (for other services)
    tokenParser, err := token.NewPasetoTokenParser(publicKey)
    if err != nil {
        log.Fatal(err)
    }

    // Validate a token
    claims, err := tokenParser.ParseToken(tokenString)
    if err != nil {
        log.Fatal(err)
    }
    ```

### Cron Package
- Production-ready cron job scheduling and management
- Job status tracking with execution metadata
- Thread-safe operations with proper concurrency handling
- Configuration-based job registry with dependency injection
- Real-time monitoring and logging integration
- Support for both cron expressions and time intervals
- Built-in job examples for common use cases

    #### Key Features
    - **Job Management**: Add, remove, and monitor cron jobs with metadata tracking
    - **Status Monitoring**: Real-time job status (idle, running, success, failed)
    - **Execution Tracking**: Run count, failure count, execution time, and duration
    - **Registry System**: Factory pattern for job creation and configuration management
    - **Thread Safety**: Full concurrent access protection with proper mutex handling
    - **Flexible Scheduling**: Support for cron expressions and time.Duration intervals

    #### Basic Usage
    ```go
    import "github.com/solum-sp/aps-be-common/common/cron"

    // Create cron manager
    manager := cron.NewCronManager(logger)

    // Create a simple function-based job
    job := cron.NewFuncJob("heartbeat", func(ctx context.Context) error {
        logger.Info("System heartbeat check")
        return nil
    })

    // Add job with cron schedule (every 5 minutes)
    err := manager.AddJob("*/5 * * * *", job)
    if err != nil {
        log.Fatal(err)
    }

    // Start the cron manager
    manager.Start()
    defer manager.Stop()
    ```

    #### Advanced Usage with Registry
    ```go
    // Create job factory
    factory := cron.NewExampleJobFactory(logger)

    // Setup complete cron system with registry
    registry := cron.NewJobRegistryBuilder(logger).
        WithJob("database_cleanup", factory.CreateDatabaseCleanupJob(30)).
        WithJob("cache_warmup", factory.CreateCacheWarmupJob([]string{"users", "products"})).
        WithJob("health_check", factory.CreateHealthCheckJob([]string{"http://api.example.com/health"})).
        WithConfig("database_cleanup", "0 2 * * *", true).       // Daily at 2 AM
        WithConfig("cache_warmup", "*/15 * * * *", true).        // Every 15 minutes
        WithConfig("health_check", "*/5 * * * *", true).         // Every 5 minutes
        Build()

    // Register all jobs with manager
    manager := cron.NewCronManager(logger)
    err := registry.RegisterJobsWithManager(manager)
    if err != nil {
        log.Fatal(err)
    }

    manager.Start()
    ```

    #### Custom Job Implementation
    ```go
    // Implement the Job interface
    type MyCustomJob struct {
        *cron.BaseJob
        database Database
        config   Config
    }

    func NewMyCustomJob(db Database, cfg Config, logger logger.ILogger) *MyCustomJob {
        return &MyCustomJob{
            BaseJob:  cron.NewBaseJob("my_custom_job"),
            database: db,
            config:   cfg,
        }
    }

    func (j *MyCustomJob) Run(ctx context.Context) error {
        // Your business logic here
        records, err := j.database.GetExpiredRecords(ctx)
        if err != nil {
            return fmt.Errorf("failed to get expired records: %w", err)
        }

        for _, record := range records {
            if err := j.database.DeleteRecord(ctx, record.ID); err != nil {
                return fmt.Errorf("failed to delete record %s: %w", record.ID, err)
            }
        }

        return nil
    }
    ```

    #### Job Monitoring and Status
    ```go
    // Get information about a specific job
    info, exists := manager.GetJobInfo("database_cleanup")
    if exists {
        fmt.Printf("Job: %s\n", info.Name)
        fmt.Printf("Status: %s\n", info.Status)
        fmt.Printf("Last Run: %v\n", info.LastRun)
        fmt.Printf("Run Count: %d\n", info.RunCount)
        fmt.Printf("Fail Count: %d\n", info.FailCount)
        fmt.Printf("Last Duration: %v\n", info.LastDuration)
    }

    // Get all running jobs
    runningJobs := manager.GetRunningJobs()
    fmt.Printf("Currently running: %v\n", runningJobs)

    // Get complete status of all jobs
    allJobs := manager.GetAllJobsInfo()
    for name, info := range allJobs {
        fmt.Printf("%s: %s (runs: %d, fails: %d)\n", 
            name, info.Status, info.RunCount, info.FailCount)
    }
    ```

    #### Configuration File Support
    ```go
    // Load job configurations from JSON file
    registry := cron.NewJobRegistryBuilder(logger).
        WithConfigFile("./config/cron_jobs.json").
        Build()
    ```

    Example `cron_jobs.json`:
    ```json
    [
        {
            "name": "database_cleanup",
            "schedule": "0 2 * * *",
            "enabled": true
        },
        {
            "name": "cache_warmup",
            "schedule": "*/15 * * * *",
            "enabled": true
        },
        {
            "name": "health_check",
            "schedule": "*/5 * * * *",
            "enabled": false
        }
    ]
    ```

    #### Built-in Job Examples
    The package includes several ready-to-use job implementations:

    - **DatabaseCleanupJob**: Cleanup old database records with configurable retention
    - **CacheWarmupJob**: Preload cache keys for better performance
    - **HealthCheckJob**: Monitor system health endpoints
    - **ReportGenerationJob**: Generate periodic reports

    ```go
    // Use built-in jobs
    dbCleanup := cron.NewDatabaseCleanupJob(30, logger) // 30 days retention
    cacheWarmup := cron.NewCacheWarmupJob([]string{"users", "products"}, logger)
    healthCheck := cron.NewHealthCheckJob([]string{"http://api.example.com/health"}, logger)

    manager.AddJob("0 2 * * *", dbCleanup)     // Daily at 2 AM
    manager.AddJob("*/15 * * * *", cacheWarmup) // Every 15 minutes
    manager.AddJob("*/5 * * * *", healthCheck)  // Every 5 minutes
    ```

    #### Error Handling and Recovery
    ```go
    // Jobs automatically capture and log errors
    // Failed jobs are marked with status and error message
    info, _ := manager.GetJobInfo("failing_job")
    if info.Status == cron.JobStatusFailed {
        fmt.Printf("Job failed with error: %s\n", info.LastError)
    }

    // Job execution is isolated - one failure doesn't affect others
    // Manager continues running even if individual jobs fail
    ```

### Utils Package
- Common utility functions and helpers
- Shared types and constants
- Error handling utilities
- Helper functions for common operations

## Requirements

- Go 1.22 or higher
- Dependencies listed in `go.mod`

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is proprietary and confidential. Unauthorized copying of files in this repository, via any medium, is strictly prohibited.

## Support

For support or questions, please contact the development team.
