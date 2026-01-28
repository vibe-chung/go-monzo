#!/usr/bin/env python3
"""
Register a webhook URL to all Monzo accounts.

This script uses the go-monzo CLI to:
1. Fetch all accounts for the currently logged-in user
2. For each account, register the provided webhook URL

Prerequisites:
- go-monzo CLI must be installed and available in PATH
- User must be logged in (run: go-monzo login)
- Webhook URL must use HTTPS

Usage:
    python scripts/register_webhooks.py <webhook_url>

Example:
    python scripts/register_webhooks.py https://example.com/monzo/webhook
"""

import json
import subprocess
import sys


def run_command(cmd):
    """Run a shell command and return the output."""
    try:
        result = subprocess.run(
            cmd,
            shell=True,
            check=True,
            capture_output=True,
            text=True
        )
        return result.stdout.strip()
    except subprocess.CalledProcessError as e:
        print(f"Error running command: {cmd}", file=sys.stderr)
        print(f"Error output: {e.stderr}", file=sys.stderr)
        raise


def get_accounts():
    """Fetch all accounts using the go-monzo CLI."""
    print("Fetching accounts...")
    output = run_command("go-monzo accounts")
    
    try:
        data = json.loads(output)
        accounts = data.get("accounts", [])
        print(f"Found {len(accounts)} account(s)")
        return accounts
    except json.JSONDecodeError as e:
        print(f"Error parsing accounts JSON: {e}", file=sys.stderr)
        print(f"Output was: {output}", file=sys.stderr)
        sys.exit(1)


def register_webhook(account_id, webhook_url):
    """Register a webhook for a given account."""
    output = run_command(
        f"go-monzo webhook register --account-id={account_id} --url={webhook_url}"
    )
    
    try:
        data = json.loads(output)
        return data.get("webhook", {})
    except json.JSONDecodeError as e:
        print(f"Error parsing webhook JSON: {e}", file=sys.stderr)
        print(f"Output was: {output}", file=sys.stderr)
        return None


def main():
    """Main function to register webhooks."""
    # Check command line arguments
    if len(sys.argv) != 2:
        print("Usage: python scripts/register_webhooks.py <webhook_url>", file=sys.stderr)
        print("Example: python scripts/register_webhooks.py https://example.com/webhook", file=sys.stderr)
        sys.exit(1)
    
    webhook_url = sys.argv[1]
    
    # Validate webhook URL (must be HTTPS)
    if not webhook_url.startswith("https://"):
        print("Error: Webhook URL must use HTTPS", file=sys.stderr)
        sys.exit(1)
    
    print("=" * 60)
    print("Register Webhook Script")
    print("=" * 60)
    print(f"Webhook URL: {webhook_url}")
    print()
    
    # Get all accounts
    accounts = get_accounts()
    
    if not accounts:
        print("No accounts found. Exiting.")
        return
    
    total_registered = 0
    
    # Process each account
    for account in accounts:
        account_id = account.get("id")
        description = account.get("description", "Unknown")
        
        print()
        print(f"Processing account: {description} ({account_id})")
        print("-" * 60)
        
        try:
            webhook = register_webhook(account_id, webhook_url)
            
            if webhook:
                webhook_id = webhook.get("id", "Unknown")
                print(f"  ✓ Webhook registered successfully")
                print(f"    Webhook ID: {webhook_id}")
                total_registered += 1
            else:
                print(f"  ✗ Failed to register webhook (no response data)")
        except Exception as e:
            print(f"  ✗ Failed to register webhook: {e}", file=sys.stderr)
    
    print()
    print("=" * 60)
    print(f"Summary: Registered webhook to {total_registered} account(s)")
    print("=" * 60)


if __name__ == "__main__":
    main()
