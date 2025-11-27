package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const (
	monzoAuthURL  = "https://auth.monzo.com"
	monzoTokenURL = "https://api.monzo.com/oauth2/token"

	// Timeout durations
	authTimeout          = 5 * time.Minute
	tokenExchangeTimeout = 30 * time.Second
	serverShutdownTimeout = 5 * time.Second
)

var (
	clientID     string
	clientSecret string
	redirectURI  string
	port         int
)

// TokenResponse represents the OAuth token response from Monzo
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope"`
	UserID       string `json:"user_id"`
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Monzo API",
	Long: `Authenticate with the Monzo API using OAuth 2.0.

This command will:
1. Start a local HTTP server to receive the OAuth callback
2. Open your browser to the Monzo authorization page
3. Wait for you to authorize the application
4. Exchange the authorization code for an access token
5. Save the token for future use

You need to provide your OAuth client credentials, which you can obtain
from the Monzo Developer Portal at https://developers.monzo.com/`,
	RunE: runLogin,
}

func init() {
	rootCmd.AddCommand(loginCmd)

	loginCmd.Flags().StringVar(&clientID, "client-id", os.Getenv("MONZO_CLIENT_ID"), "Monzo OAuth client ID (or set MONZO_CLIENT_ID)")
	loginCmd.Flags().StringVar(&clientSecret, "client-secret", os.Getenv("MONZO_CLIENT_SECRET"), "Monzo OAuth client secret (or set MONZO_CLIENT_SECRET)")
	loginCmd.Flags().StringVar(&redirectURI, "redirect-uri", "", "OAuth redirect URI (default: http://localhost:<port>/callback)")
	loginCmd.Flags().IntVar(&port, "port", 8080, "Local server port for OAuth callback")
}

func runLogin(cmd *cobra.Command, args []string) error {
	if clientID == "" {
		return fmt.Errorf("client ID is required. Set via --client-id flag or MONZO_CLIENT_ID environment variable")
	}

	if clientSecret == "" {
		return fmt.Errorf("client secret is required. Set via --client-secret flag or MONZO_CLIENT_SECRET environment variable")
	}

	if redirectURI == "" {
		redirectURI = fmt.Sprintf("http://localhost:%d/callback", port)
	}

	// Channel to receive the authorization code
	codeChan := make(chan string, 1)
	errChan := make(chan error, 1)

	// Start local server
	server, err := startCallbackServer(port, codeChan, errChan)
	if err != nil {
		return fmt.Errorf("failed to start local server: %w", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), serverShutdownTimeout)
		defer cancel()
		_ = server.Shutdown(ctx)
	}()

	// Build authorization URL
	authURL := buildAuthURL(clientID, redirectURI)

	fmt.Println("Opening browser for Monzo authorization...")
	fmt.Printf("If the browser doesn't open, visit this URL:\n%s\n\n", authURL)

	// Open browser
	if err := openBrowser(authURL); err != nil {
		fmt.Printf("Warning: Could not open browser automatically: %v\n", err)
	}

	fmt.Println("Waiting for authorization...")

	// Wait for the authorization code
	var code string
	select {
	case code = <-codeChan:
		// Success
	case err := <-errChan:
		return fmt.Errorf("authorization failed: %w", err)
	case <-time.After(authTimeout):
		return fmt.Errorf("authorization timed out")
	}

	fmt.Println("Authorization received! Exchanging code for token...")

	// Exchange code for token
	token, err := exchangeCodeForToken(clientID, clientSecret, redirectURI, code)
	if err != nil {
		return fmt.Errorf("failed to exchange code for token: %w", err)
	}

	// Save token
	if err := saveToken(token); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	fmt.Println("\nLogin successful!")
	fmt.Printf("User ID: %s\n", token.UserID)
	fmt.Printf("Token expires in: %d seconds\n", token.ExpiresIn)
	fmt.Println("\nNote: You may need to approve access in the Monzo app for full API permissions.")

	return nil
}

func startCallbackServer(port int, codeChan chan<- string, errChan chan<- error) (*http.Server, error) {
	mux := http.NewServeMux()

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		errorMsg := r.URL.Query().Get("error")

		if errorMsg != "" {
			errorDesc := r.URL.Query().Get("error_description")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = fmt.Fprintf(w, "<html><body><h1>Authorization Failed</h1><p>%s: %s</p></body></html>", errorMsg, errorDesc)
			errChan <- fmt.Errorf("%s: %s", errorMsg, errorDesc)
			return
		}

		if code == "" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = fmt.Fprint(w, "<html><body><h1>Authorization Failed</h1><p>No authorization code received</p></body></html>")
			errChan <- fmt.Errorf("no authorization code received")
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, "<html><body><h1>Authorization Successful!</h1><p>You can close this window and return to the terminal.</p></body></html>")
		codeChan <- code
	})

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}

	server := &http.Server{Handler: mux}

	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	return server, nil
}

func buildAuthURL(clientID, redirectURI string) string {
	params := url.Values{
		"client_id":     {clientID},
		"redirect_uri":  {redirectURI},
		"response_type": {"code"},
		"state":         {fmt.Sprintf("%d", time.Now().UnixNano())},
	}
	return fmt.Sprintf("%s/?%s", monzoAuthURL, params.Encode())
}

func exchangeCodeForToken(clientID, clientSecret, redirectURI, code string) (*TokenResponse, error) {
	data := url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"redirect_uri":  {redirectURI},
		"code":          {code},
	}

	ctx, cancel := context.WithTimeout(context.Background(), tokenExchangeTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", monzoTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
			return nil, fmt.Errorf("token exchange failed: %v", errResp)
		}
		return nil, fmt.Errorf("token exchange failed with status: %d", resp.StatusCode)
	}

	var token TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	return &token, nil
}

func saveToken(token *TokenResponse) error {
	configDir, err := getConfigDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(configDir, 0700); err != nil {
		return err
	}

	tokenPath := filepath.Join(configDir, "token.json")

	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(tokenPath, data, 0600)
}

func getConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".go-monzo"), nil
}

func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "linux":
		cmd = "xdg-open"
		args = []string{url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return exec.Command(cmd, args...).Start()
}
