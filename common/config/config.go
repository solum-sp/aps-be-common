package config

import (
	"log"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type AppConfig struct {
	App struct {
		Name    string `env:"APP_NAME" envDefault:"microservice-base"`
		Version string `env:"APP_VERSION" envDefault:"v0.0.0"`
		Env     string `env:"APP_ENV" envDefault:"development"`
	}
	Server struct {
		Port string `env:"HTTP_PORT" envDefault:"8080"`
		Host string `env:"HTTP_HOST" envDefault:"localhost"`
	}

	Logger struct {
		Level string `env:"LOG_LEVEL" envDefault:"info"`
	}

	CockroachDB struct {
		URI string `env:"COCKROACH_URI,required"`
	}

	MongoDB struct {
		URI      string `env:"MONGO_URI,required"`
		Database string `env:"MONGO_DATABASE" envDefault:"defaultdb"`
	}

	Redis struct {
		Addr     string `env:"REDIS_HOST" envDefault:"localhost:6379"`
		Password string `env:"REDIS_PASSWORD" envDefault:""`
		DB       int    `env:"REDIS_DB" envDefault:"0"`
	}
}

func NewAppConfig(path string) (*AppConfig, error) {
	cfg := &AppConfig{}
	err := loadEnv(path)
	if err != nil {
		return nil, err
	}
	err = ParseConfig(cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func loadEnv(path string) error {
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "development"
	}
	envFileName := ""

	switch env {
	case "development":
		envFileName = ".env"
	case "test":
		envFileName = ".env.test"
	case "production":
		envFileName = ".env.production"
	default:
		envFileName = ".env"
	}
	envPath := path + "/" + envFileName
	if envFileName != "" {
		log.Printf("Loading config from file:%s\n", envPath)
		_ = godotenv.Load(envPath)
	}
	log.Printf("Loading config from environment\n")
	_ = godotenv.Load()
	return nil
}

func ParseConfig(c *AppConfig) error {
	return env.Parse(c)
}
