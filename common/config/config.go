package config

import (
	"log"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

func NewAppConfig(path string, cfg interface{}) error {
	err := LoadEnv(path)
	if err != nil {
		return err
	}
	err = ParseConfig(cfg)
	if err != nil {
		return err
	}
	return nil
}

func LoadEnv(path string) error {
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

func ParseConfig(c interface{}) error {
	return env.Parse(c)
}
