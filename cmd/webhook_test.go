package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRunWebhookListWithoutAccountID(t *testing.T) {
	// Clear any environment variable
	originalAccountID := os.Getenv("MONZO_ACCOUNT_ID")
	os.Unsetenv("MONZO_ACCOUNT_ID")
	defer func() {
		if originalAccountID != "" {
			os.Setenv("MONZO_ACCOUNT_ID", originalAccountID)
		}
	}()

	// Reset the flag
	webhookAccountID = ""

	err := runWebhookList(nil, nil)
	if err == nil {
		t.Error("Expected error when account ID is not provided, got nil")
	}
}

func TestRunWebhookRegisterWithoutURL(t *testing.T) {
	// Set account ID but not URL
	webhookAccountID = "acc_123"
	webhookURL = ""

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

	// Create a valid token
	configDir := filepath.Join(tmpDir, ".go-monzo")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	storedToken := StoredToken{
		TokenResponse: TokenResponse{
			AccessToken:  "valid_access_token",
			TokenType:    "Bearer",
			ExpiresIn:    21600,
			RefreshToken: "valid_refresh_token",
			Scope:        "",
			UserID:       "user_123",
		},
		ExpiresAt: time.Now().Unix() + 21600,
	}

	data, err := json.MarshalIndent(storedToken, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal token: %v", err)
	}

	tokenPath := filepath.Join(configDir, "token.json")
	if err := os.WriteFile(tokenPath, data, 0600); err != nil {
		t.Fatalf("Failed to write token file: %v", err)
	}

	err = runWebhookRegister(nil, nil)
	if err == nil {
		t.Error("Expected error when webhook URL is not provided, got nil")
	}
}

func TestRunWebhookDeleteWithoutWebhookID(t *testing.T) {
	// Reset the flag
	webhookID = ""

	err := runWebhookDelete(nil, nil)
	if err == nil {
		t.Error("Expected error when webhook ID is not provided, got nil")
	}
}
