"""
Common utilities for go-monzo helper scripts.

This module provides shared functions used by the webhook management scripts.
"""

import json
import subprocess
import sys


def run_command(cmd_args):
    """
    Run a command and return the output.
    
    Args:
        cmd_args: List of command arguments (e.g., ["go-monzo", "accounts"])
    
    Returns:
        str: The command output
    
    Raises:
        subprocess.CalledProcessError: If the command fails
    """
    try:
        result = subprocess.run(
            cmd_args,
            check=True,
            capture_output=True,
            text=True
        )
        return result.stdout.strip()
    except subprocess.CalledProcessError as e:
        print(f"Error running command: {' '.join(cmd_args)}", file=sys.stderr)
        print(f"Error output: {e.stderr}", file=sys.stderr)
        raise


def get_accounts():
    """
    Fetch all accounts using the go-monzo CLI.
    
    Returns:
        list: List of account dictionaries
    
    Raises:
        SystemExit: If unable to fetch or parse accounts
    """
    print("Fetching accounts...")
    try:
        output = run_command(["go-monzo", "accounts"])
    except subprocess.CalledProcessError:
        sys.exit(1)
    
    try:
        data = json.loads(output)
        accounts = data.get("accounts", [])
        print(f"Found {len(accounts)} account(s)")
        return accounts
    except json.JSONDecodeError as e:
        print(f"Error parsing accounts JSON: {e}", file=sys.stderr)
        print(f"Output was: {output}", file=sys.stderr)
        sys.exit(1)
