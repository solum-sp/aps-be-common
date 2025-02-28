package loader

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// **Load function: Generic loader for ENV, JSON, and YAML files**
func Load(cfg interface{}, filePath ...string) error {
	// If no filePath is provided, fallback to default
	if len(filePath) == 0 || filePath[0] == "" {
		return fmt.Errorf("no file path provided")
	}

	path := filePath[0]
	ext := filepath.Ext(path)
	if ext == "" && filepath.Base(path) == ".env" { // ✅ Handle .env files correctly
		ext = ".env"
	}
	// Determine file type by extension
	switch ext {
	case ".json":
		return loadJSON(cfg, path)
	case ".yaml", ".yml":
		return loadYAML(cfg, path)
	case ".env":
		return loadEnv(path)
	default:
		return fmt.Errorf("unsupported file type: %s", ext)
	}
}

// **Load JSON Config**
func loadJSON(cfg interface{}, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read JSON file: %w", err)
	}
	return json.Unmarshal(data, cfg)
}

// **Load YAML Config**
func loadYAML(cfg interface{}, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read YAML file: %w", err)
	}
	return yaml.Unmarshal(data, cfg)
}

func loadEnv(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("env file not found: %s", path)
	}

	err := godotenv.Load(path)
	if err != nil {
		return fmt.Errorf("failed to load env file: %w", err)
	}

	return nil
}
// **Parse environment variables into struct**
func ParseConfig(cfg interface{}) error {
	return env.Parse(cfg)
}


