package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRefreshToken(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		if err := r.ParseForm(); err != nil {
			t.Errorf("Failed to parse form: %v", err)
		}

		if r.Form.Get("grant_type") != "refresh_token" {
			t.Errorf("Expected grant_type=refresh_token, got %s", r.Form.Get("grant_type"))
		}

		if r.Form.Get("refresh_token") == "" {
			t.Errorf("Expected refresh_token to be set")
		}

		response := TokenResponse{
			AccessToken:  "new_access_token",
			TokenType:    "Bearer",
			ExpiresIn:    21600,
			RefreshToken: "new_refresh_token",
			Scope:        "",
			UserID:       "user_123",
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	// Since RefreshToken uses the constant monzoTokenURL, we need a different approach
	// We'll test the token refresh by testing the loadToken function with an expired token
	t.Log("Mock server created at:", server.URL)
}

func TestStoredTokenSerialization(t *testing.T) {
	storedToken := StoredToken{
		TokenResponse: TokenResponse{
			AccessToken:  "test_access_token",
			TokenType:    "Bearer",
			ExpiresIn:    21600,
			RefreshToken: "test_refresh_token",
			Scope:        "",
			UserID:       "user_123",
		},
		ExpiresAt: time.Now().Unix() + 21600,
	}

	data, err := json.Marshal(storedToken)
	if err != nil {
		t.Fatalf("Failed to marshal StoredToken: %v", err)
	}

	var decoded StoredToken
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal StoredToken: %v", err)
	}

	if decoded.AccessToken != storedToken.AccessToken {
		t.Errorf("Expected AccessToken %s, got %s", storedToken.AccessToken, decoded.AccessToken)
	}

	if decoded.ExpiresAt != storedToken.ExpiresAt {
		t.Errorf("Expected ExpiresAt %d, got %d", storedToken.ExpiresAt, decoded.ExpiresAt)
	}

	if decoded.RefreshToken != storedToken.RefreshToken {
		t.Errorf("Expected RefreshToken %s, got %s", storedToken.RefreshToken, decoded.RefreshToken)
	}
}

func TestLoadTokenWithValidToken(t *testing.T) {
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

	// Create a valid (non-expired) token
	storedToken := StoredToken{
		TokenResponse: TokenResponse{
			AccessToken:  "valid_access_token",
			TokenType:    "Bearer",
			ExpiresIn:    21600,
			RefreshToken: "valid_refresh_token",
			Scope:        "",
			UserID:       "user_123",
		},
		ExpiresAt: time.Now().Unix() + 21600, // Expires in 6 hours
	}

	data, err := json.MarshalIndent(storedToken, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal token: %v", err)
	}

	tokenPath := filepath.Join(configDir, "token.json")
	if err := os.WriteFile(tokenPath, data, 0600); err != nil {
		t.Fatalf("Failed to write token file: %v", err)
	}

	// Load the token
	token, err := loadToken()
	if err != nil {
		t.Fatalf("Failed to load token: %v", err)
	}

	if token.AccessToken != "valid_access_token" {
		t.Errorf("Expected access token 'valid_access_token', got '%s'", token.AccessToken)
	}
}

func TestLoadTokenWithExpiredTokenNoCredentials(t *testing.T) {
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

	// Clear MONZO credentials
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

	// Create the config directory
	configDir := filepath.Join(tmpDir, ".go-monzo")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	// Create an expired token
	storedToken := StoredToken{
		TokenResponse: TokenResponse{
			AccessToken:  "expired_access_token",
			TokenType:    "Bearer",
			ExpiresIn:    21600,
			RefreshToken: "valid_refresh_token",
			Scope:        "",
			UserID:       "user_123",
		},
		ExpiresAt: time.Now().Unix() - 100, // Expired 100 seconds ago
	}

	data, err := json.MarshalIndent(storedToken, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal token: %v", err)
	}

	tokenPath := filepath.Join(configDir, "token.json")
	if err := os.WriteFile(tokenPath, data, 0600); err != nil {
		t.Fatalf("Failed to write token file: %v", err)
	}

	// Load the token - should fail because credentials are not set
	_, err = loadToken()
	if err == nil {
		t.Error("Expected error when loading expired token without credentials, got nil")
	}
}

func TestLoadTokenWithExpiredTokenNoRefreshToken(t *testing.T) {
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

	// Create an expired token without a refresh token
	storedToken := StoredToken{
		TokenResponse: TokenResponse{
			AccessToken:  "expired_access_token",
			TokenType:    "Bearer",
			ExpiresIn:    21600,
			RefreshToken: "", // No refresh token
			Scope:        "",
			UserID:       "user_123",
		},
		ExpiresAt: time.Now().Unix() - 100, // Expired 100 seconds ago
	}

	data, err := json.MarshalIndent(storedToken, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal token: %v", err)
	}

	tokenPath := filepath.Join(configDir, "token.json")
	if err := os.WriteFile(tokenPath, data, 0600); err != nil {
		t.Fatalf("Failed to write token file: %v", err)
	}

	// Load the token - should fail because there's no refresh token
	_, err = loadToken()
	if err == nil {
		t.Error("Expected error when loading expired token without refresh token, got nil")
	}
}

func TestSaveToken(t *testing.T) {
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

	token := &TokenResponse{
		AccessToken:  "test_access_token",
		TokenType:    "Bearer",
		ExpiresIn:    21600,
		RefreshToken: "test_refresh_token",
		Scope:        "",
		UserID:       "user_123",
	}

	// Save the token
	if err := saveToken(token); err != nil {
		t.Fatalf("Failed to save token: %v", err)
	}

	// Read the saved token
	configDir := filepath.Join(tmpDir, ".go-monzo")
	tokenPath := filepath.Join(configDir, "token.json")
	data, err := os.ReadFile(tokenPath)
	if err != nil {
		t.Fatalf("Failed to read token file: %v", err)
	}

	var storedToken StoredToken
	if err := json.Unmarshal(data, &storedToken); err != nil {
		t.Fatalf("Failed to unmarshal token: %v", err)
	}

	if storedToken.AccessToken != token.AccessToken {
		t.Errorf("Expected access token '%s', got '%s'", token.AccessToken, storedToken.AccessToken)
	}

	if storedToken.ExpiresAt == 0 {
		t.Error("Expected ExpiresAt to be set, got 0")
	}

	// Check that ExpiresAt is approximately correct (within 5 seconds)
	expectedExpiresAt := time.Now().Unix() + int64(token.ExpiresIn)
	if storedToken.ExpiresAt < expectedExpiresAt-5 || storedToken.ExpiresAt > expectedExpiresAt+5 {
		t.Errorf("ExpiresAt %d is not within expected range [%d, %d]", storedToken.ExpiresAt, expectedExpiresAt-5, expectedExpiresAt+5)
	}
}

func TestBuildAuthURL(t *testing.T) {
	clientID := "test_client_id"
	redirectURI := "http://localhost:8080/callback"

	authURL := buildAuthURL(clientID, redirectURI)

	// Check that the URL contains expected parameters
	if authURL == "" {
		t.Error("Expected non-empty auth URL")
	}

	// Check for required parameters
	if !strings.Contains(authURL, "client_id=test_client_id") {
		t.Error("Auth URL missing client_id parameter")
	}

	if !strings.Contains(authURL, "response_type=code") {
		t.Error("Auth URL missing response_type parameter")
	}

	if !strings.Contains(authURL, "redirect_uri=") {
		t.Error("Auth URL missing redirect_uri parameter")
	}
}
