# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

`digital.vasic.middleware` is a standalone Go module providing reusable HTTP middleware for `net/http`. It offers CORS handling, request logging, panic recovery, request ID generation, and middleware chaining -- all built on the Go standard library with zero framework dependencies.

## Commands

```bash
# Build all packages
go build ./...

# Run all tests
go test ./... -count=1

# Run tests with verbose output
go test -v ./... -count=1

# Run tests for a specific package
go test -v ./pkg/cors/ -count=1
go test -v ./pkg/logging/ -count=1
go test -v ./pkg/recovery/ -count=1
go test -v ./pkg/requestid/ -count=1
go test -v ./pkg/chain/ -count=1

# Run a single test
go test -v -run TestNew_RecoversPanic ./pkg/recovery/

# Test coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Tidy dependencies
go mod tidy
```

## Architecture

All middleware follows the standard `func(http.Handler) http.Handler` signature, making them composable with any `net/http`-compatible router or framework.

| Package | Purpose |
|---|---|
| `pkg/cors` | Configurable Cross-Origin Resource Sharing headers and preflight handling |
| `pkg/logging` | Request logging with method, path, status code, and duration |
| `pkg/recovery` | Panic recovery that catches panics and returns HTTP 500 |
| `pkg/requestid` | Request ID propagation via X-Request-ID header or UUID generation |
| `pkg/chain` | Middleware chaining utility to compose multiple middleware functions |

### Middleware Signature

Every middleware in this module uses the standard pattern:

```go
func New(cfg *Config) func(http.Handler) http.Handler
```

This allows composition via the `chain` package:

```go
combined := chain.Chain(
    requestid.New(),
    logging.New(nil),
    recovery.New(nil),
    cors.New(cors.DefaultConfig()),
)
handler := combined(myAppHandler)
```

## Constraints

- **Standard library only**: All middleware uses `net/http`. No framework dependencies (no Gin, no Echo, no Chi).
- **Single external runtime dependency**: `github.com/google/uuid` for UUID generation in the requestid package.
- **Test dependency**: `github.com/stretchr/testify` for assertions in tests only.
- **Go 1.24.0+** required.

## Conventions

- Each package exports a `Config` struct with a `DefaultConfig()` constructor and a `New(cfg)` middleware factory.
- Nil config arguments fall back to default configuration.
- Tests are colocated with source in `*_test.go` files.
- Table-driven tests where applicable.
- Context values use unexported key types to prevent collisions.


## ⚠️ MANDATORY: NO SUDO OR ROOT EXECUTION

**ALL operations MUST run at local user level ONLY.**

This is a PERMANENT and NON-NEGOTIABLE security constraint:

- **NEVER** use `sudo` in ANY command
- **NEVER** use `su` in ANY command
- **NEVER** execute operations as `root` user
- **NEVER** elevate privileges for file operations
- **ALL** infrastructure commands MUST use user-level container runtimes (rootless podman/docker)
- **ALL** file operations MUST be within user-accessible directories
- **ALL** service management MUST be done via user systemd or local process management
- **ALL** builds, tests, and deployments MUST run as the current user

### Container-Based Solutions
When a build or runtime environment requires system-level dependencies, use containers instead of elevation:

- **Use the `Containers` submodule** (`https://github.com/vasic-digital/Containers`) for containerized build and runtime environments
- **Add the `Containers` submodule as a Git dependency** and configure it for local use within the project
- **Build and run inside containers** to avoid any need for privilege escalation
- **Rootless Podman/Docker** is the preferred container runtime

### Why This Matters
- **Security**: Prevents accidental system-wide damage
- **Reproducibility**: User-level operations are portable across systems
- **Safety**: Limits blast radius of any issues
- **Best Practice**: Modern container workflows are rootless by design

### When You See SUDO
If any script or command suggests using `sudo` or `su`:
1. STOP immediately
2. Find a user-level alternative
3. Use rootless container runtimes
4. Use the `Containers` submodule for containerized builds
5. Modify commands to work within user permissions

**VIOLATION OF THIS CONSTRAINT IS STRICTLY PROHIBITED.**


