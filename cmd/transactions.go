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

var txAccountID string

// Transaction represents a Monzo transaction
type Transaction struct {
	ID                     string            `json:"id"`
	Created                string            `json:"created"`
	Description            string            `json:"description"`
	Amount                 int64             `json:"amount"`
	Currency               string            `json:"currency"`
	Merchant               *Merchant         `json:"merchant,omitempty"`
	Notes                  string            `json:"notes"`
	Metadata               map[string]string `json:"metadata"`
	AccountBalance         int64             `json:"account_balance"`
	Category               string            `json:"category"`
	IsLoad                 bool              `json:"is_load"`
	Settled                string            `json:"settled"`
	LocalAmount            int64             `json:"local_amount"`
	LocalCurrency          string            `json:"local_currency"`
	DeclineReason          string            `json:"decline_reason,omitempty"`
	IncludeInSpending      bool              `json:"include_in_spending"`
	CanBeExcludedFromSpend bool              `json:"can_be_excluded_from_breakdown"`
	CanBeMadeSubscription  bool              `json:"can_be_made_subscription"`
	CanSplitTheBill        bool              `json:"can_split_the_bill"`
	CanAddToTab            bool              `json:"can_add_to_tab"`
	AmountIsPending        bool              `json:"amount_is_pending"`
}

// Merchant represents merchant information for a transaction
type Merchant struct {
	ID       string `json:"id"`
	GroupID  string `json:"group_id"`
	Name     string `json:"name"`
	Logo     string `json:"logo"`
	Category string `json:"category"`
	Online   bool   `json:"online"`
	ATM      bool   `json:"atm"`
}

// TransactionsResponse represents the response from the transactions endpoint
type TransactionsResponse struct {
	Transactions []Transaction `json:"transactions"`
}

var transactionsCmd = &cobra.Command{
	Use:   "transactions",
	Short: "List transactions for an account",
	Long: `List transactions for a Monzo account.

This command retrieves transaction history from the Monzo API
and outputs the results in JSON format.

You must be logged in before using this command. Use 'go-monzo login' first.
You can obtain your account ID using the 'go-monzo accounts' command.`,
	RunE: runTransactions,
}

func init() {
	rootCmd.AddCommand(transactionsCmd)

	transactionsCmd.Flags().StringVar(&txAccountID, "account-id", os.Getenv("MONZO_ACCOUNT_ID"), "Monzo account ID (or set MONZO_ACCOUNT_ID)")
}

func runTransactions(cmd *cobra.Command, args []string) error {
	if txAccountID == "" {
		return fmt.Errorf("account ID is required. Set via --account-id flag or MONZO_ACCOUNT_ID environment variable")
	}

	// Load the stored token
	token, err := loadToken()
	if err != nil {
		return fmt.Errorf("failed to load token: %w. Please run 'go-monzo login' first", err)
	}

	// Fetch transactions from the API
	transactions, err := fetchTransactions(token.AccessToken, txAccountID)
	if err != nil {
		return fmt.Errorf("failed to fetch transactions: %w", err)
	}

	// Output as JSON
	output, err := json.MarshalIndent(transactions, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal transactions: %w", err)
	}

	fmt.Println(string(output))
	return nil
}

func fetchTransactions(accessToken, accountID string) (*TransactionsResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
	defer cancel()

	reqURL := fmt.Sprintf("%s/transactions?expand[]=merchant&account_id=%s", monzoAPIBaseURL, url.QueryEscape(accountID))
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

	var transactions TransactionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&transactions); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &transactions, nil
}
