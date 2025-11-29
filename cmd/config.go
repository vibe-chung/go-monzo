package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config represents the configuration stored in the config file
type Config struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

// LoadConfig loads the configuration from the config file at ~/.go-monzo/config.json
// Returns an empty Config if the file doesn't exist or can't be read
func LoadConfig() (*Config, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return nil, err
	}

	configPath := filepath.Join(configDir, "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty config if file doesn't exist
			return &Config{}, nil
		}
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// GetClientCredentials returns the client ID and secret with the following priority:
// 1. Command line flags (passed as parameters)
// 2. Environment variables
// 3. Config file (~/.go-monzo/config.json)
func GetClientCredentials(flagClientID, flagClientSecret string) (string, string) {
	clientID := flagClientID
	clientSecret := flagClientSecret

	// If flag is empty, try environment variable (already handled by cobra default)
	// This is for cases where we need to get credentials outside of command context
	if clientID == "" {
		clientID = os.Getenv("MONZO_CLIENT_ID")
	}
	if clientSecret == "" {
		clientSecret = os.Getenv("MONZO_CLIENT_SECRET")
	}

	// If still empty, try config file (lowest priority)
	if clientID == "" || clientSecret == "" {
		config, err := LoadConfig()
		if err == nil && config != nil {
			if clientID == "" {
				clientID = config.ClientID
			}
			if clientSecret == "" {
				clientSecret = config.ClientSecret
			}
		}
	}

	return clientID, clientSecret
}
