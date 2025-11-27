# GitHub Copilot Instructions

This document provides instructions and guidelines for GitHub Copilot when working on the go-monzo repository.

## Project Overview

go-monzo is a Go command line interface (CLI) for interacting with the Monzo personal banking API. It uses the Cobra library for CLI structure and supports OAuth 2.0 authentication.

## Project Structure

```
go-monzo/
├── main.go          # Application entry point
├── cmd/             # CLI commands
│   ├── root.go      # Root command configuration
│   └── login.go     # OAuth login command
├── go.mod           # Go module definition
└── go.sum           # Dependency checksums
```

## Building and Testing

### Build the project

```bash
go build .
```

### Run tests

```bash
go test ./...
```

### Run with verbose output

```bash
go test -v ./...
```

### Format code

```bash
go fmt ./...
```

### Lint code

```bash
go vet ./...
```

## Coding Standards

### Go Conventions

- Follow standard Go conventions and idioms
- Use `gofmt` for code formatting
- Use meaningful variable and function names in camelCase
- Keep functions focused and small
- Handle errors explicitly - do not ignore errors
- Use context for cancellation and timeouts

### Error Handling

- Always return errors from functions that can fail
- Use `fmt.Errorf` with `%w` for error wrapping
- Provide descriptive error messages

### File Organization

- Each CLI command should be in its own file in the `cmd/` directory
- Keep related functionality grouped together
- Use clear, descriptive file names

## Dependencies

- Use the standard library whenever possible
- New dependencies should be added with `go get`
- Keep dependencies minimal

## Security Guidelines

- Never commit secrets, API keys, or tokens
- Use environment variables for sensitive configuration
- Token storage should use restrictive file permissions (0600)
- Validate all user input
- Use timeouts for HTTP requests and operations

## CLI Command Guidelines

When adding new commands:

1. Create a new file in `cmd/` directory
2. Define the command using `cobra.Command`
3. Register the command with `rootCmd.AddCommand()` in `init()`
4. Use flags for configurable options
5. Support environment variables as alternatives to flags
6. Provide clear help text with `Short` and `Long` descriptions

## Git Workflow

- Create focused, atomic commits
- Write clear commit messages
- Keep pull requests small and focused on a single feature or fix
