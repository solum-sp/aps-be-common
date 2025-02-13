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

### Kafka Package
- Full Kafka producer and consumer implementations
- Support for Schema Registry with Avro serialization
- Configurable consumer groups and auto-commit settings
- Robust error handling and retry mechanisms
- Supports both synchronous and asynchronous message processing
- Built-in admin client functionality

### Logger Package
- Structured logging with multiple log levels (Debug, Info, Warn, Error, Fatal)
- Context-aware logging
- OpenTelemetry integration for distributed tracing
- Field-based logging with sanitization of sensitive data
- Stack trace capture for error logging
- Service name tagging for multi-service environments

### Redis Package
- Redis client implementation with connection pooling
- Support for key-value operations with expiration time
- Pattern-based key operations (get, delete)
- Service-specific key prefixing
- Bulk operations support (clear all, clear by pattern)
- Error handling and connection management

### Utils Package
- Common utility functions and helpers
- Shared types and constants
- Error handling utilities
- Helper functions for common operations

## Configuration

### Redis Configuration
```go
redisConfig := redis.Config{
    Addr:     "localhost:6379",
    Password: "",
    DB:       0,
    Service:  "your-service-name",
}
```

### Logger Configuration
```go
loggerConfig := logger.Config{
    Service: "your-service-name",
    Level:   logger.InfoLv,
}
```

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