# Helper Scripts

This directory contains helper scripts to automate common webhook management tasks.

## Scripts

### delete_webhooks.py

Delete all webhooks for all Monzo accounts.

**Prerequisites:**
- `go-monzo` CLI must be installed and available in PATH
- User must be logged in (run: `go-monzo login`)

**Usage:**
```bash
python scripts/delete_webhooks.py
```

**What it does:**
1. Fetches all accounts for the currently logged-in user
2. For each account, lists all webhooks
3. Deletes each webhook

### register_webhooks.py

Register a webhook URL to all Monzo accounts.

**Prerequisites:**
- `go-monzo` CLI must be installed and available in PATH
- User must be logged in (run: `go-monzo login`)
- Webhook URL must use HTTPS

**Usage:**
```bash
python scripts/register_webhooks.py <webhook_url>
```

**Example:**
```bash
python scripts/register_webhooks.py https://example.com/monzo/webhook
```

**What it does:**
1. Fetches all accounts for the currently logged-in user
2. For each account, registers the provided webhook URL
3. Displays the webhook ID for each successful registration

## Requirements

These scripts require Python 3.6 or later. They use only standard library modules (`json`, `subprocess`, `sys`), so no additional dependencies need to be installed.

## Workflow Example

A typical workflow to update webhooks across all accounts:

```bash
# 1. Make sure you're logged in
go-monzo login

# 2. Delete all existing webhooks
python scripts/delete_webhooks.py

# 3. Register a new webhook to all accounts
python scripts/register_webhooks.py https://your-webhook-url.com/webhook
```
