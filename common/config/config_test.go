package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func createTempEnvFile(t *testing.T, filename string, content string, env string) string {
	dir := t.TempDir()
	envPath := filepath.Join(dir, filename)
	err := os.Setenv("APP_ENV", env)
	assert.NoError(t, err)
	err = os.WriteFile(envPath, []byte(content), 0644)

	assert.NoError(t, err)
	return dir
}

func TestNewAppConfig(t *testing.T) {
	tt := struct {
		name        string
		envContent  string
		envFile     string
		expectError bool
		env         string
		validate    func(*testing.T, *AppConfig)
	}{
		name: "Valid development config",
		envContent: `
							APP_NAME=test-app
							APP_VERSION=v1.0.0
							APP_ENV=development
							COCKROACH_URI=postgresql://root@localhost:26257/defaultdb
							MONGO_URI=mongodb://localhost:27017
					`,
		envFile:     ".env",
		expectError: false,
		env:         "development",
		validate: func(t *testing.T, cfg *AppConfig) {
			assert.Equal(t, "test-app", cfg.App.Name)
			assert.Equal(t, "v1.0.0", cfg.App.Version)
			assert.Equal(t, "development", cfg.App.Env)
		},
	}

	t.Run(tt.name, func(t *testing.T) {
		// Create temporary environment file
		tmpDir := createTempEnvFile(t, tt.envFile, tt.envContent, tt.env)

		// Test configuration loading
		cfg, err := NewAppConfig(tmpDir)

		if tt.expectError {
			assert.Error(t, err)
			return
		}

		assert.NoError(t, err)
		assert.NotNil(t, cfg)

		if tt.validate != nil {
			tt.validate(t, cfg)
		}
	})
}

func TestNewAppConfigProduction(t *testing.T) {
	tt := struct {
		name        string
		envContent  string
		envFile     string
		expectError bool
		env         string
		validate    func(*testing.T, *AppConfig)
	}{
		name: "Valid production config",
		envContent: `
							APP_NAME=test-app
							APP_VERSION=v1.0.0
							APP_ENV=production
							COCKROACH_URI=postgresql://root@localhost:26257/defaultdb
							MONGO_URI=mongodb://localhost:27017
					`,
		envFile:     ".env.production",
		expectError: false,
		env:         "production",
		validate: func(t *testing.T, cfg *AppConfig) {
			assert.Equal(t, "test-app", cfg.App.Name)
			assert.Equal(t, "v1.0.0", cfg.App.Version)
			assert.Equal(t, "production", cfg.App.Env)
		},
	}

	t.Run(tt.name, func(t *testing.T) {
		// Create temporary environment file
		tmpDir := createTempEnvFile(t, tt.envFile, tt.envContent, tt.env)

		// Test configuration loading
		cfg, err := NewAppConfig(tmpDir)

		if tt.expectError {
			assert.Error(t, err)
			return
		}

		assert.NoError(t, err)
		assert.NotNil(t, cfg)

		if tt.validate != nil {
			tt.validate(t, cfg)
		}
	})
}

func TestParseConfig(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		expectError bool
	}{
		{
			name: "Valid configuration",
			envVars: map[string]string{
				"APP_NAME":      "test-app",
				"COCKROACH_URI": "postgresql://root@localhost:26257/defaultdb",
				"MONGO_URI":     "mongodb://localhost:27017",
			},
			expectError: false,
		},
		{
			name: "Missing required fields",
			envVars: map[string]string{
				"APP_NAME": "test-app",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			cfg := &AppConfig{}
			err := ParseConfig(cfg)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Clean up environment variables
			for k := range tt.envVars {
				os.Unsetenv(k)
			}
		})
	}
}
