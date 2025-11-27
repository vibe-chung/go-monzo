package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "go-monzo",
	Short: "A CLI client for the Monzo personal API",
	Long: `go-monzo is a command line interface for interacting with the Monzo
personal banking API. It allows you to access your account information,
transactions, and other banking features from the terminal.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
