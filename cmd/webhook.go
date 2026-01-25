package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	webhookAccountID string
	webhookURL       string
	webhookID        string
)

// Webhook represents a Monzo webhook
type Webhook struct {
	ID        string `json:"id"`
	AccountID string `json:"account_id"`
	URL       string `json:"url"`
}

// WebhooksResponse represents the response from the list webhooks endpoint
type WebhooksResponse struct {
	Webhooks []Webhook `json:"webhooks"`
}

// WebhookResponse represents the response from the register webhook endpoint
type WebhookResponse struct {
	Webhook Webhook `json:"webhook"`
}

var webhookCmd = &cobra.Command{
	Use:   "webhook",
	Short: "Manage webhooks for your Monzo account",
	Long: `Manage webhooks for your Monzo account.

This command allows you to list, register, and delete webhooks using the Monzo API.
Webhooks allow you to receive real-time notifications when transactions occur.

You must be logged in before using this command. Use 'go-monzo login' first.`,
}

var webhookListCmd = &cobra.Command{
	Use:   "list",
	Short: "List registered webhooks for an account",
	Long: `List all registered webhooks for a Monzo account.

This command retrieves webhook information from the Monzo API
and outputs the results in JSON format.

You must be logged in before using this command. Use 'go-monzo login' first.
You can obtain your account ID using the 'go-monzo accounts' command.`,
	RunE: runWebhookList,
}

var webhookRegisterCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new webhook for an account",
	Long: `Register a new webhook for a Monzo account.

This command registers a webhook URL that will receive notifications when
transactions occur on the specified account. The webhook URL must be HTTPS.

You must be logged in before using this command. Use 'go-monzo login' first.
You can obtain your account ID using the 'go-monzo accounts' command.`,
	RunE: runWebhookRegister,
}

var webhookDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a webhook",
	Long: `Delete a registered webhook.

This command deletes a webhook and stops notifications from being sent to the URL.

You must be logged in before using this command. Use 'go-monzo login' first.
You can obtain webhook IDs using the 'go-monzo webhook list' command.`,
	RunE: runWebhookDelete,
}

func init() {
	rootCmd.AddCommand(webhookCmd)
	webhookCmd.AddCommand(webhookListCmd)
	webhookCmd.AddCommand(webhookRegisterCmd)
	webhookCmd.AddCommand(webhookDeleteCmd)

	// List command flags
	webhookListCmd.Flags().StringVar(&webhookAccountID, "account-id", os.Getenv("MONZO_ACCOUNT_ID"), "Monzo account ID (or set MONZO_ACCOUNT_ID)")

	// Register command flags
	webhookRegisterCmd.Flags().StringVar(&webhookAccountID, "account-id", os.Getenv("MONZO_ACCOUNT_ID"), "Monzo account ID (or set MONZO_ACCOUNT_ID)")
	webhookRegisterCmd.Flags().StringVar(&webhookURL, "url", "", "Webhook URL (must be HTTPS)")

	// Delete command flags
	webhookDeleteCmd.Flags().StringVar(&webhookID, "webhook-id", "", "Webhook ID to delete")
}

func runWebhookList(cmd *cobra.Command, args []string) error {
	if webhookAccountID == "" {
		return fmt.Errorf("account ID is required. Set via --account-id flag or MONZO_ACCOUNT_ID environment variable")
	}

	// Load the stored token
	token, err := loadToken()
	if err != nil {
		return fmt.Errorf("failed to load token: %w. Please run 'go-monzo login' first", err)
	}

	// Fetch webhooks from the API
	webhooks, err := fetchWebhooks(token.AccessToken, webhookAccountID)
	if err != nil {
		return fmt.Errorf("failed to fetch webhooks: %w", err)
	}

	// Output as JSON
	output, err := json.MarshalIndent(webhooks, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal webhooks: %w", err)
	}

	fmt.Println(string(output))
	return nil
}

func runWebhookRegister(cmd *cobra.Command, args []string) error {
	if webhookAccountID == "" {
		return fmt.Errorf("account ID is required. Set via --account-id flag or MONZO_ACCOUNT_ID environment variable")
	}

	if webhookURL == "" {
		return fmt.Errorf("webhook URL is required. Set via --url flag")
	}

	// Load the stored token
	token, err := loadToken()
	if err != nil {
		return fmt.Errorf("failed to load token: %w. Please run 'go-monzo login' first", err)
	}

	// Register webhook with the API
	webhook, err := registerWebhook(token.AccessToken, webhookAccountID, webhookURL)
	if err != nil {
		return fmt.Errorf("failed to register webhook: %w", err)
	}

	// Output as JSON
	output, err := json.MarshalIndent(webhook, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal webhook: %w", err)
	}

	fmt.Println(string(output))
	return nil
}

func runWebhookDelete(cmd *cobra.Command, args []string) error {
	if webhookID == "" {
		return fmt.Errorf("webhook ID is required. Set via --webhook-id flag")
	}

	// Load the stored token
	token, err := loadToken()
	if err != nil {
		return fmt.Errorf("failed to load token: %w. Please run 'go-monzo login' first", err)
	}

	// Delete webhook via the API
	if err := deleteWebhook(token.AccessToken, webhookID); err != nil {
		return fmt.Errorf("failed to delete webhook: %w", err)
	}

	fmt.Printf("Webhook %s deleted successfully\n", webhookID)
	return nil
}

func fetchWebhooks(accessToken, accountID string) (*WebhooksResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
	defer cancel()

	reqURL := fmt.Sprintf("%s/webhooks?account_id=%s", monzoAPIBaseURL, url.QueryEscape(accountID))
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

	var webhooks WebhooksResponse
	if err := json.NewDecoder(resp.Body).Decode(&webhooks); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &webhooks, nil
}

func registerWebhook(accessToken, accountID, webhookURL string) (*WebhookResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
	defer cancel()

	data := url.Values{
		"account_id": {accountID},
		"url":        {webhookURL},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", monzoAPIBaseURL+"/webhooks", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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

	var webhook WebhookResponse
	if err := json.NewDecoder(resp.Body).Decode(&webhook); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &webhook, nil
}

func deleteWebhook(accessToken, webhookID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
	defer cancel()

	reqURL := fmt.Sprintf("%s/webhooks/%s", monzoAPIBaseURL, url.PathEscape(webhookID))
	req, err := http.NewRequestWithContext(ctx, "DELETE", reqURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Error from ReadAll is intentionally ignored as we're in an error path
		// and want to include whatever body content we can read in the error message
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
