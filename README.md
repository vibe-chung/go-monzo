# go-monzo

Go command line client for accessing the Monzo Personal API.

## Installation

```bash
go install github.com/vibe-chung/go-monzo@latest
```

Or build from source:

```bash
git clone https://github.com/vibe-chung/go-monzo.git
cd go-monzo
go build .
```

## Prerequisites

Before using this CLI, you need to register an OAuth client with Monzo:

1. Go to the [Monzo Developer Portal](https://developers.monzo.com/)
2. Sign in with your Monzo account
3. Create a new OAuth client
4. Note your `client_id` and `client_secret`
5. Set the redirect URI to `http://localhost:8080/callback` (or customize with `--port` flag)

## Usage

### Login

Authenticate with the Monzo API:

```bash
# Using flags
go-monzo login --client-id=YOUR_CLIENT_ID --client-secret=YOUR_CLIENT_SECRET

# Using environment variables
export MONZO_CLIENT_ID=your_client_id
export MONZO_CLIENT_SECRET=your_client_secret
go-monzo login
```

This will:
1. Start a local HTTP server to receive the OAuth callback
2. Open your browser to the Monzo authorization page
3. Wait for you to authorize the application
4. Exchange the authorization code for an access token
5. Save the token to `~/.go-monzo/token.json`

**Note:** After initial authorization, you may need to approve access in the Monzo app for full API permissions.

## Configuration

The CLI stores tokens in `~/.go-monzo/token.json`.

Environment variables:
- `MONZO_CLIENT_ID` - Your OAuth client ID
- `MONZO_CLIENT_SECRET` - Your OAuth client secret

## License

MIT
