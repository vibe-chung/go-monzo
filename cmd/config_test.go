package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigNonExistent(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "go-monzo-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override the home directory for the test
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Load config when file doesn't exist
	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("Expected no error when config file doesn't exist, got: %v", err)
	}

	if config == nil {
		t.Fatal("Expected non-nil config")
	}

	if config.ClientID != "" {
		t.Errorf("Expected empty ClientID, got: %s", config.ClientID)
	}

	if config.ClientSecret != "" {
		t.Errorf("Expected empty ClientSecret, got: %s", config.ClientSecret)
	}
}

func TestLoadConfigValid(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "go-monzo-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override the home directory for the test
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Create the config directory
	configDir := filepath.Join(tmpDir, ".go-monzo")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	// Create a valid config file
	expectedConfig := Config{
		ClientID:     "test_client_id",
		ClientSecret: "test_client_secret",
	}

	data, err := json.MarshalIndent(expectedConfig, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	configPath := filepath.Join(configDir, "config.json")
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Load the config
	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if config.ClientID != expectedConfig.ClientID {
		t.Errorf("Expected ClientID %s, got %s", expectedConfig.ClientID, config.ClientID)
	}

	if config.ClientSecret != expectedConfig.ClientSecret {
		t.Errorf("Expected ClientSecret %s, got %s", expectedConfig.ClientSecret, config.ClientSecret)
	}
}

func TestLoadConfigInvalidJSON(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "go-monzo-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override the home directory for the test
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Create the config directory
	configDir := filepath.Join(tmpDir, ".go-monzo")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	// Create an invalid config file
	configPath := filepath.Join(configDir, "config.json")
	if err := os.WriteFile(configPath, []byte("invalid json"), 0600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Load the config - should fail
	_, err = LoadConfig()
	if err == nil {
		t.Error("Expected error when loading invalid JSON config, got nil")
	}
}

func TestGetClientCredentialsPriorityFlags(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "go-monzo-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override the home directory for the test
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Clear environment variables
	originalClientID := os.Getenv("MONZO_CLIENT_ID")
	originalClientSecret := os.Getenv("MONZO_CLIENT_SECRET")
	os.Setenv("MONZO_CLIENT_ID", "env_client_id")
	os.Setenv("MONZO_CLIENT_SECRET", "env_client_secret")
	defer func() {
		if originalClientID != "" {
			os.Setenv("MONZO_CLIENT_ID", originalClientID)
		} else {
			os.Unsetenv("MONZO_CLIENT_ID")
		}
		if originalClientSecret != "" {
			os.Setenv("MONZO_CLIENT_SECRET", originalClientSecret)
		} else {
			os.Unsetenv("MONZO_CLIENT_SECRET")
		}
	}()

	// Create a config file
	configDir := filepath.Join(tmpDir, ".go-monzo")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	configData := Config{
		ClientID:     "config_client_id",
		ClientSecret: "config_client_secret",
	}
	data, _ := json.Marshal(configData)
	configPath := filepath.Join(configDir, "config.json")
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Test: flags take priority over env vars and config file
	clientID, clientSecret := GetClientCredentials("flag_client_id", "flag_client_secret")

	if clientID != "flag_client_id" {
		t.Errorf("Expected clientID 'flag_client_id', got '%s'", clientID)
	}

	if clientSecret != "flag_client_secret" {
		t.Errorf("Expected clientSecret 'flag_client_secret', got '%s'", clientSecret)
	}
}

func TestGetClientCredentialsPriorityEnvVars(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "go-monzo-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override the home directory for the test
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Set environment variables
	originalClientID := os.Getenv("MONZO_CLIENT_ID")
	originalClientSecret := os.Getenv("MONZO_CLIENT_SECRET")
	os.Setenv("MONZO_CLIENT_ID", "env_client_id")
	os.Setenv("MONZO_CLIENT_SECRET", "env_client_secret")
	defer func() {
		if originalClientID != "" {
			os.Setenv("MONZO_CLIENT_ID", originalClientID)
		} else {
			os.Unsetenv("MONZO_CLIENT_ID")
		}
		if originalClientSecret != "" {
			os.Setenv("MONZO_CLIENT_SECRET", originalClientSecret)
		} else {
			os.Unsetenv("MONZO_CLIENT_SECRET")
		}
	}()

	// Create a config file
	configDir := filepath.Join(tmpDir, ".go-monzo")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	configData := Config{
		ClientID:     "config_client_id",
		ClientSecret: "config_client_secret",
	}
	data, _ := json.Marshal(configData)
	configPath := filepath.Join(configDir, "config.json")
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Test: env vars take priority over config file when flags are empty
	clientID, clientSecret := GetClientCredentials("", "")

	if clientID != "env_client_id" {
		t.Errorf("Expected clientID 'env_client_id', got '%s'", clientID)
	}

	if clientSecret != "env_client_secret" {
		t.Errorf("Expected clientSecret 'env_client_secret', got '%s'", clientSecret)
	}
}

func TestGetClientCredentialsPriorityConfigFile(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "go-monzo-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override the home directory for the test
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Clear environment variables
	originalClientID := os.Getenv("MONZO_CLIENT_ID")
	originalClientSecret := os.Getenv("MONZO_CLIENT_SECRET")
	os.Unsetenv("MONZO_CLIENT_ID")
	os.Unsetenv("MONZO_CLIENT_SECRET")
	defer func() {
		if originalClientID != "" {
			os.Setenv("MONZO_CLIENT_ID", originalClientID)
		}
		if originalClientSecret != "" {
			os.Setenv("MONZO_CLIENT_SECRET", originalClientSecret)
		}
	}()

	// Create a config file
	configDir := filepath.Join(tmpDir, ".go-monzo")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	configData := Config{
		ClientID:     "config_client_id",
		ClientSecret: "config_client_secret",
	}
	data, _ := json.Marshal(configData)
	configPath := filepath.Join(configDir, "config.json")
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Test: config file is used when flags and env vars are empty
	clientID, clientSecret := GetClientCredentials("", "")

	if clientID != "config_client_id" {
		t.Errorf("Expected clientID 'config_client_id', got '%s'", clientID)
	}

	if clientSecret != "config_client_secret" {
		t.Errorf("Expected clientSecret 'config_client_secret', got '%s'", clientSecret)
	}
}

func TestGetClientCredentialsPartialFromConfig(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "go-monzo-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Override the home directory for the test
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", originalHome)

	// Set only client ID in environment
	originalClientID := os.Getenv("MONZO_CLIENT_ID")
	originalClientSecret := os.Getenv("MONZO_CLIENT_SECRET")
	os.Setenv("MONZO_CLIENT_ID", "env_client_id")
	os.Unsetenv("MONZO_CLIENT_SECRET")
	defer func() {
		if originalClientID != "" {
			os.Setenv("MONZO_CLIENT_ID", originalClientID)
		} else {
			os.Unsetenv("MONZO_CLIENT_ID")
		}
		if originalClientSecret != "" {
			os.Setenv("MONZO_CLIENT_SECRET", originalClientSecret)
		}
	}()

	// Create a config file with both values
	configDir := filepath.Join(tmpDir, ".go-monzo")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	configData := Config{
		ClientID:     "config_client_id",
		ClientSecret: "config_client_secret",
	}
	data, _ := json.Marshal(configData)
	configPath := filepath.Join(configDir, "config.json")
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Test: env var for client ID, config file for client secret
	clientID, clientSecret := GetClientCredentials("", "")

	if clientID != "env_client_id" {
		t.Errorf("Expected clientID 'env_client_id', got '%s'", clientID)
	}

	if clientSecret != "config_client_secret" {
		t.Errorf("Expected clientSecret 'config_client_secret', got '%s'", clientSecret)
	}
}
