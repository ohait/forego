# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
make test          # vet + lint + test-long + test-race (full CI pipeline)
make test-short    # go test --short --trimpath ./...
make test-long     # go test --trimpath ./...
make test-race     # go test --short --race --trimpath ./...
make fmt           # go fmt ./...
make vet           # go vet ./...
make lint          # golangci-lint run --timeout 5m --skip-dirs '(example)'
make lint-setup    # install golangci-lint v1.53.2
```

To run a single test:
```bash
go test -run TestName ./pkg/...
```

## Architecture

**forego** is a Go framework (module `github.com/ohait/forego`) for building JSON/WebSocket services. The central pattern is: define a struct implementing `api.Op`, register it with the HTTP server, and exercise the same struct in tests — no HTTP glue needed.

### Core packages and their roles

**`ctx` / `ctx/log`** — Every context parameter is typed `ctx.C` (an alias for `context.Context`). Tags attached via `ctx.WithTag(c, key, value)` automatically appear in JSON log output. `ctx/log` emits JSONL. `ctx.Error` wraps errors with stack traces.

**`api`** — Binds Go structs to RPC-style operations via struct tags (`api:"in,required"`, `api:"out"`, `api:"both"`). The handler interface is just `Do(c ctx.C) error`. `api.Test(c, op)` invokes an op directly in tests. Also generates OpenAPI documentation.

**`http`** — Production HTTP server. `srv.RegisterAPI(c, path, op)` or `srv.MustRegisterAPI(...)` wires an `api.Op` to an endpoint. Built-in routes: `/live`, `/ready`, `/openapi.json`. Middleware via `srv.Use()`. The server automatically tags requests with user agent, path, and remote address.

**`http/ws`** — WebSocket RPC layer reusing the same struct/tag approach as REST.

**`enc`** — Intermediate JSON representation (`enc.Node`, `enc.Map`, `enc.Pairs`) used to avoid `json.RawMessage` gymnastics. Provides `Marshaler`/`Unmarshaler` interfaces for custom coercion.

**`test`** — Test helpers: `test.Context(t)` creates a context tied to `*testing.T`. `test.EqualsStr`, etc. provide AST-aware assertions with readable output.

**`shutdown`** — Graceful signal handling (INT/TERM/QUIT). `shutdown.Hold()`/`Release()` coordinate goroutine cleanup. `shutdown.WaitForSignal(c, cf)` blocks until a signal arrives.

### Typical request flow

1. Struct implements `api.Op` (or `api.StreamingOp` for bidirectional streaming)
2. Registered on `http.Server` via `RegisterAPI`
3. HTTP request → server tags context → deserializes JSON into struct fields tagged `api:"in"` → calls `Do(c)` → serializes fields tagged `api:"out"` back as JSON
4. Same `Do(c)` called directly in tests via `api.Test(c, op)`

### Conventions

- Context parameter is always named `c ctx.C`
- Tags: `ctx.WithTag(c, key, value)`
- Commit messages: short imperative sentences without prefixes (e.g., `Fix WebSocket graceful shutdown hang`)
- Tests use table-driven style with `github.com/ohait/forego/test` helpers
- Test listeners bind to `127.0.0.1:0`
- `make test-race` must pass for non-trivial changes
