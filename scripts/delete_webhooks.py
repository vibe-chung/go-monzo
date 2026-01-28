#!/usr/bin/env python3
"""
Delete all webhooks for all Monzo accounts.

This script uses the go-monzo CLI to:
1. Fetch all accounts for the currently logged-in user
2. For each account, list all webhooks
3. Delete each webhook

Prerequisites:
- go-monzo CLI must be installed and available in PATH
- User must be logged in (run: go-monzo login)

Usage:
    python scripts/delete_webhooks.py
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
        sys.exit(1)


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


def get_webhooks(account_id):
    """Fetch all webhooks for a given account."""
    output = run_command(f"go-monzo webhook list --account-id={account_id}")
    
    try:
        data = json.loads(output)
        webhooks = data.get("webhooks", [])
        return webhooks
    except json.JSONDecodeError as e:
        print(f"Error parsing webhooks JSON: {e}", file=sys.stderr)
        print(f"Output was: {output}", file=sys.stderr)
        return []


def delete_webhook(webhook_id):
    """Delete a webhook by ID."""
    run_command(f"go-monzo webhook delete --webhook-id={webhook_id}")


def main():
    """Main function to delete all webhooks."""
    print("=" * 60)
    print("Delete All Webhooks Script")
    print("=" * 60)
    print()
    
    # Get all accounts
    accounts = get_accounts()
    
    if not accounts:
        print("No accounts found. Exiting.")
        return
    
    total_deleted = 0
    
    # Process each account
    for account in accounts:
        account_id = account.get("id")
        description = account.get("description", "Unknown")
        
        print()
        print(f"Processing account: {description} ({account_id})")
        print("-" * 60)
        
        # Get webhooks for this account
        webhooks = get_webhooks(account_id)
        
        if not webhooks:
            print("  No webhooks found for this account")
            continue
        
        print(f"  Found {len(webhooks)} webhook(s)")
        
        # Delete each webhook
        for webhook in webhooks:
            webhook_id = webhook.get("id")
            webhook_url = webhook.get("url", "Unknown URL")
            
            print(f"  Deleting webhook: {webhook_id}")
            print(f"    URL: {webhook_url}")
            
            try:
                delete_webhook(webhook_id)
                print(f"    ✓ Deleted successfully")
                total_deleted += 1
            except Exception as e:
                print(f"    ✗ Failed to delete: {e}", file=sys.stderr)
    
    print()
    print("=" * 60)
    print(f"Summary: Deleted {total_deleted} webhook(s) in total")
    print("=" * 60)


if __name__ == "__main__":
    main()
