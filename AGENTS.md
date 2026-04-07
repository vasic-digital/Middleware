# AGENTS.md

Instructions for AI agents working on this codebase.

## Project Context

This is `digital.vasic.middleware`, a standalone Go module providing HTTP middleware for `net/http`. It is designed to be imported by other Go projects that need CORS, logging, recovery, request ID, or middleware chaining functionality.

## Key Rules

1. **Standard library only**: All middleware MUST use `net/http.Handler` and `net/http.HandlerFunc`. Do NOT introduce framework dependencies (Gin, Echo, Chi, etc.).

2. **Middleware signature**: All middleware MUST follow the `func(http.Handler) http.Handler` pattern. Configuration is passed via a `Config` struct with a `DefaultConfig()` constructor and a `New(cfg)` factory.

3. **Nil safety**: `New(nil)` MUST work by falling back to `DefaultConfig()`.

4. **Tests required**: Every package MUST have corresponding `*_test.go` files. Use `github.com/stretchr/testify` for assertions.

5. **No global state**: Middleware must be safe for concurrent use. Use closures over configuration rather than package-level variables.

6. **Context usage**: When storing values in request context, use unexported key types to prevent key collisions across packages.

## Adding New Middleware

To add a new middleware package:

1. Create `pkg/<name>/<name>.go` with a `Config` struct, `DefaultConfig()`, and `New()` function.
2. Create `pkg/<name>/<name>_test.go` with comprehensive tests.
3. Update `CLAUDE.md` architecture table.
4. Run `go mod tidy && go build ./... && go test ./... -count=1` to verify.

## Testing

```bash
go test ./... -count=1        # all tests
go test -v ./pkg/<name>/      # single package, verbose
go test -race ./...            # race condition detection
```

## File Structure

```
pkg/
  cors/          - CORS middleware
  logging/       - Request logging middleware
  recovery/      - Panic recovery middleware
  requestid/     - Request ID middleware
  chain/         - Middleware chaining utility
```


## ⚠️ MANDATORY: NO SUDO OR ROOT EXECUTION

**ALL operations MUST run at local user level ONLY.**

This is a PERMANENT and NON-NEGOTIABLE security constraint:

- **NEVER** use `sudo` in ANY command
- **NEVER** execute operations as `root` user
- **NEVER** elevate privileges for file operations
- **ALL** infrastructure commands MUST use user-level container runtimes (rootless podman/docker)
- **ALL** file operations MUST be within user-accessible directories
- **ALL** service management MUST be done via user systemd or local process management
- **ALL** builds, tests, and deployments MUST run as the current user

### Why This Matters
- **Security**: Prevents accidental system-wide damage
- **Reproducibility**: User-level operations are portable across systems
- **Safety**: Limits blast radius of any issues
- **Best Practice**: Modern container workflows are rootless by design

### When You See SUDO
If any script or command suggests using `sudo`:
1. STOP immediately
2. Find a user-level alternative
3. Use rootless container runtimes
4. Modify commands to work within user permissions

**VIOLATION OF THIS CONSTRAINT IS STRICTLY PROHIBITED.**

