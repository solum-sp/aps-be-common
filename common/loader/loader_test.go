package loader

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// **Test struct matching a sample JSON/YAML file**
type TestConfig struct {
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
	APIKey   string `json:"api_key" yaml:"api_key"`
}

// **Test JSON Loading**
func TestLoadJSON(t *testing.T) {
	// Create a temporary JSON file
	jsonContent := `{"username": "testuser", "password": "testpass", "api_key": "123456"}`
	filePath := "test_config.json"
	os.WriteFile(filePath, []byte(jsonContent), 0644)
	defer os.Remove(filePath) // Cleanup after test

	// Load into struct
	var cfg TestConfig
	err := Load(&cfg, filePath)

	// Assert no error
	assert.NoError(t, err)
	assert.Equal(t, "testuser", cfg.Username)
	assert.Equal(t, "testpass", cfg.Password)
	assert.Equal(t, "123456", cfg.APIKey)
}

// **Test YAML Loading**
func TestLoadYAML(t *testing.T) {
	// Create a temporary YAML file
	yamlContent := `username: testuser
					password: testpass
					api_key: 123456`
	filePath := "test_config.yaml"
	os.WriteFile(filePath, []byte(yamlContent), 0644)
	defer os.Remove(filePath) // Cleanup

	// Load into struct
	var cfg TestConfig
	err := Load(&cfg, filePath)

	// Assert no error
	assert.NoError(t, err)
	assert.Equal(t, "testuser", cfg.Username)
	assert.Equal(t, "testpass", cfg.Password)
	assert.Equal(t, "123456", cfg.APIKey)
}


func TestLoadEnv(t *testing.T) {
	// ✅ Ensure the correct `.env` filename
	filePath := ".env"

	// ✅ Create a temporary .env file
	envContent := `DB_HOST=localhost
DB_USER=root`
	err := os.WriteFile(filePath, []byte(envContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test env file: %v", err)
	}
	defer os.Remove(filePath) // Cleanup after test

	// ✅ Clear any existing environment variables before testing
	os.Unsetenv("DB_HOST")
	os.Unsetenv("DB_USER")

	// ✅ Load .env file
	err = Load(nil, filePath)

	// ✅ Assert no error
	assert.NoError(t, err)

	// ✅ Explicitly check if environment variables are set correctly
	host, existsHost := os.LookupEnv("DB_HOST")
	user, existsUser := os.LookupEnv("DB_USER")

	// ✅ Ensure variables exist and match expected values
	assert.True(t, existsHost, "DB_HOST should be set")
	assert.True(t, existsUser, "DB_USER should be set")
	assert.Equal(t, "localhost", host)
	assert.Equal(t, "root", user)
}

// **Test Handling Missing Files**
func TestLoadMissingFile(t *testing.T) {
	var cfg TestConfig
	err := Load(&cfg, "nonexistent.json")

	// Assert an error occurs
	assert.Error(t, err)
}

// **Test Unsupported File Type**
func TestLoadUnsupportedFileType(t *testing.T) {
	var cfg TestConfig
	err := Load(&cfg, "config.txt") // .txt is unsupported

	// Assert an error occurs
	assert.Error(t, err)
}
