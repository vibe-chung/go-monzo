package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

const (
	monzoAPIBaseURL = "https://api.monzo.com"
	apiTimeout      = 30 * time.Second
)

// Account represents a Monzo account
type Account struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Created     string `json:"created"`
	Type        string `json:"type"`
	Closed      bool   `json:"closed"`
}

// AccountsResponse represents the response from the accounts endpoint
type AccountsResponse struct {
	Accounts []Account `json:"accounts"`
}

var accountsCmd = &cobra.Command{
	Use:   "accounts",
	Short: "List accounts for the authenticated user",
	Long: `List all accounts associated with the currently authenticated user.

This command retrieves account information from the Monzo API and outputs
the results in JSON format.

You must be logged in before using this command. Use 'go-monzo login' first.`,
	RunE: runAccounts,
}

func init() {
	rootCmd.AddCommand(accountsCmd)
}

func runAccounts(cmd *cobra.Command, args []string) error {
	// Load the stored token
	token, err := loadToken()
	if err != nil {
		return fmt.Errorf("failed to load token: %w. Please run 'go-monzo login' first", err)
	}

	// Fetch accounts from the API
	accounts, err := fetchAccounts(token.AccessToken)
	if err != nil {
		return fmt.Errorf("failed to fetch accounts: %w", err)
	}

	// Output as JSON
	output, err := json.MarshalIndent(accounts, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal accounts: %w", err)
	}

	fmt.Println(string(output))
	return nil
}

func loadToken() (*TokenResponse, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return nil, err
	}

	tokenPath := filepath.Join(configDir, "token.json")
	data, err := os.ReadFile(tokenPath)
	if err != nil {
		return nil, err
	}

	var token TokenResponse
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, err
	}

	return &token, nil
}

func fetchAccounts(accessToken string) (*AccountsResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", monzoAPIBaseURL+"/accounts", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var accounts AccountsResponse
	if err := json.NewDecoder(resp.Body).Decode(&accounts); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &accounts, nil
}
