package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test configuration struct
type TestConfig struct {
	App struct {
		Name    string `env:"APP_NAME" envDefault:"test-app"`
		Version string `env:"APP_VERSION" envDefault:"v1.0.0"`
		Env     string `env:"APP_ENV" envDefault:"development"`
	}
	Database struct {
		URI      string `env:"DB_URI" envDefault:"postgresql://localhost:5432/testdb"`
		Username string `env:"DB_USERNAME" envDefault:"test"`
		Password string `env:"DB_PASSWORD"`
	}
	Custom struct {
		Feature string `env:"CUSTOM_FEATURE" envDefault:"test-feature"`
		Flag    bool   `env:"CUSTOM_FLAG" envDefault:"true"`
	}
}

func createTempEnvFile(t *testing.T, filename string, content string) string {
	dir := t.TempDir()
	envPath := filepath.Join(dir, filename)
	err := os.WriteFile(envPath, []byte(content), 0644)
	assert.NoError(t, err)
	return dir
}

func TestLoadEnv(t *testing.T) {
	tests := []struct {
		name       string
		env        string
		envContent string
		envFile    string
		wantErr    bool
	}{
		{
			name: "Development environment",
			env:  "development",
			envContent: `
APP_NAME=dev-app
APP_VERSION=v1.0.0
DB_URI=postgresql://localhost:5432/devdb
`,
			envFile: ".env",
			wantErr: false,
		},
		{
			name: "Production environment",
			env:  "production",
			envContent: `
APP_NAME=prod-app
APP_VERSION=v2.0.0
DB_URI=postgresql://prod-host:5432/proddb
`,
			envFile: ".env.production",
			wantErr: false,
		},
		{
			name: "Test environment",
			env:  "test",
			envContent: `
APP_NAME=test-app
APP_VERSION=v0.0.1
DB_URI=postgresql://localhost:5432/testdb
`,
			envFile: ".env.test",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			tmpDir := createTempEnvFile(t, tt.envFile, tt.envContent)
			os.Setenv("APP_ENV", tt.env)
			defer os.Unsetenv("APP_ENV")

			// Test
			err := LoadEnv(tmpDir)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseConfig(t *testing.T) {
	tests := []struct {
		name       string
		envContent string
		validate   func(*testing.T, *TestConfig)
		wantErr    bool
	}{
		{
			name: "Parse with all values set",
			envContent: `
APP_NAME=custom-app
APP_VERSION=v3.0.0
DB_URI=postgresql://custom:5432/customdb
DB_USERNAME=custom-user
DB_PASSWORD=secret
CUSTOM_FEATURE=special
CUSTOM_FLAG=true
`,
			validate: func(t *testing.T, cfg *TestConfig) {
				assert.Equal(t, "custom-app", cfg.App.Name)
				assert.Equal(t, "v3.0.0", cfg.App.Version)
				assert.Equal(t, "postgresql://custom:5432/customdb", cfg.Database.URI)
				assert.Equal(t, "custom-user", cfg.Database.Username)
				assert.Equal(t, "secret", cfg.Database.Password)
				assert.Equal(t, "special", cfg.Custom.Feature)
				assert.True(t, cfg.Custom.Flag)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			if tt.envContent != "" {
				tmpDir := createTempEnvFile(t, ".env", tt.envContent)
				err := LoadEnv(tmpDir)
				assert.NoError(t, err)
			}

			// Test
			cfg := &TestConfig{}
			err := ParseConfig(cfg)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			if tt.validate != nil {
				tt.validate(t, cfg)
			}
		})
	}
}

func TestNewAppConfig(t *testing.T) {
	tests := []struct {
		name       string
		envContent string
		validate   func(*testing.T, *TestConfig)
		wantErr    bool
	}{
		{
			name: "Create new config with custom values",
			envContent: `
APP_NAME=new-app
APP_VERSION=v4.0.0
DB_URI=postgresql://new:5432/newdb
CUSTOM_FEATURE=new-feature
`,
			validate: func(t *testing.T, cfg *TestConfig) {
				assert.Equal(t, "new-app", cfg.App.Name)
				assert.Equal(t, "v4.0.0", cfg.App.Version)
				assert.Equal(t, "postgresql://new:5432/newdb", cfg.Database.URI)
				assert.Equal(t, "new-feature", cfg.Custom.Feature)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			tmpDir := createTempEnvFile(t, ".env", tt.envContent)

			// Test
			cfg := &TestConfig{}
			err := NewAppConfig(tmpDir, cfg)

			// Assert
			if tt.wantErr {
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
}
