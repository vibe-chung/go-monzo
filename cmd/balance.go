package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/spf13/cobra"
)

var accountID string

// BalanceResponse represents the response from the balance endpoint
type BalanceResponse struct {
	Balance                         int64  `json:"balance"`
	TotalBalance                    int64  `json:"total_balance"`
	BalanceIncludingFlexibleSavings int64  `json:"balance_including_flexible_savings"`
	Currency                        string `json:"currency"`
	SpendToday                      int64  `json:"spend_today"`
}

var balanceCmd = &cobra.Command{
	Use:   "balance",
	Short: "Get the balance of an account",
	Long: `Get the balance of a Monzo account.

This command retrieves the current balance information from the Monzo API
and outputs the results in JSON format.

You must be logged in before using this command. Use 'go-monzo login' first.
You can obtain your account ID using the 'go-monzo accounts' command.`,
	RunE: runBalance,
}

func init() {
	rootCmd.AddCommand(balanceCmd)

	balanceCmd.Flags().StringVar(&accountID, "account-id", os.Getenv("MONZO_ACCOUNT_ID"), "Monzo account ID (or set MONZO_ACCOUNT_ID)")
}

func runBalance(cmd *cobra.Command, args []string) error {
	if accountID == "" {
		return fmt.Errorf("account ID is required. Set via --account-id flag or MONZO_ACCOUNT_ID environment variable")
	}

	// Load the stored token
	token, err := loadToken()
	if err != nil {
		return fmt.Errorf("failed to load token: %w. Please run 'go-monzo login' first", err)
	}

	// Fetch balance from the API
	balance, err := fetchBalance(token.AccessToken, accountID)
	if err != nil {
		return fmt.Errorf("failed to fetch balance: %w", err)
	}

	// Output as JSON
	output, err := json.MarshalIndent(balance, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal balance: %w", err)
	}

	fmt.Println(string(output))
	return nil
}

func fetchBalance(accessToken, accountID string) (*BalanceResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
	defer cancel()

	reqURL := fmt.Sprintf("%s/balance?account_id=%s", monzoAPIBaseURL, url.QueryEscape(accountID))
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
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
		// Error from ReadAll is intentionally ignored as we're in an error path
		// and want to include whatever body content we can read in the error message
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var balance BalanceResponse
	if err := json.NewDecoder(resp.Body).Decode(&balance); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &balance, nil
}
