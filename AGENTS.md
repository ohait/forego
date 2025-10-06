# Repository Guidelines

## Project Structure & Module Organization
- Root packages (`api`, `http`, `ctx`, `enc`, `shutdown`, `test`, `utils`) each expose a cohesive subsystem; tests sit beside code as `*_test.go` files.
- `example/` and `example/cmd/` showcase runnable demos; treat them as reference when wiring new services.
- Shared assets such as OpenAPI helpers live under package-specific subfolders (e.g., `api/openapi/`).

## Build, Test, and Development Commands
- `go test ./...` — run the full unit test suite; prefer this before commits.
- `make test-short` or `make test-long` — quick vs. exhaustive runs with consistent flags.
- `make test` — executes vet, lint, standard tests, and race checks in one go.
- `make fmt` / `make vet` / `make lint` — format, vet, or lint the tree; `lint` requires `golangci-lint` (see `make lint-setup`).
- `go mod tidy -v` — refresh module metadata after dependency changes.

## Coding Style & Naming Conventions
- Use Go defaults: tabs for indentation, `gofmt` (via `make fmt`) before sending patches, idiomatic CamelCase for exported identifiers, lowerCamel for locals.
- Keep files ASCII; log messages should stay concise JSON-friendly strings.
- Follow existing patterns for tagged logging (`ctx.WithTag`) and context parameters (`c ctx.C`).

## Testing Guidelines
- Prefer table-driven `TestXxx` functions and the helpers in `github.com/ohait/forego/test` (e.g., `test.Context`, `test.EqualsStr`).
- Aim for race-clean runs (`make test-race`) on non-trivial changes; use `make test-ci-cover` when producing coverage reports.
- Integration samples should live under `example/` unless they need to run in CI.

## Commit & Pull Request Guidelines
- Commit summaries follow the existing history: short imperative sentences without prefixes (e.g., `Improve request body handling`).
- Separate unrelated changes; include focused diffs and update docs/tests alongside code.
- Pull requests should describe intent, highlight risky areas, and note test commands executed (`go test ./...`, `make test`); link issues when applicable.
- Provide screenshots or curl transcripts only when UI/API behavior changes.

## Security & Configuration Tips
- Avoid hardcoding secrets; rely on environment variables or config packages under `config/`.
- When adding network listeners in tests, continue using `127.0.0.1:0` to remain sandbox-friendly.
