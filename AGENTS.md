# Repository Guidelines

## Project Structure & Module Organization
- `cmd/` holds application entrypoints; add new binaries here.
- `internal/` contains core hexagonal layers: `domain/` (business logic), `ports/` (interfaces), `adapters/` (implementations).
- `pkg/` exposes reusable public packages.
- `tests/` and `testdata/` contain integration tests and fixtures; `examples/` shows library usage.
- `build/` is generated output; avoid manual edits.

## Build, Test, and Development Commands
- `make install` downloads and tidies Go module dependencies.
- `make build` builds binaries into `build/`.
- `make run` runs all entrypoints in `cmd/` for local development.
- `make lint` runs `golangci-lint` across the codebase.
- `make test` runs all tests with race detection and auto-generates fixtures.
- `make coverage` produces `coverage.out` and `coverage.html` (CI enforces >=80%).

## Coding Style & Naming Conventions
- Format with `go fmt` (`make fmt`); use standard Go tabs and naming (exported `PascalCase`, unexported `camelCase`).
- Keep comments minimal and only for complex logic.
- Prefer small, focused packages that align to ports/adapters boundaries.
- Run `go vet` (`make vet`) and `golangci-lint` before submitting.

## Testing Guidelines
- Use `go test` via `make test`, which generates fixtures in `testdata/` if missing.
- Unit tests should be fast (`make test-unit`), integration tests live in `tests/integration/` and can be run with `make test-integration`.
- Name tests with Go conventions: `TestXxx`, `BenchmarkXxx`, `TestXxxIntegration` for integration focus.

## Commit & Pull Request Guidelines
- Use Conventional Commits (`feat:`, `fix:`, `docs:`, `refactor:`, `test:`, `chore:`, `perf:`, `ci:`). Example: `feat(epub): add container.xml validation`.
- Keep PRs scoped and link relevant issues. Include a short summary of behavior changes and how you tested.
- PRs should pass CI checks (build, lint, tests with coverage threshold, docker build) and have at least one approval.

## Architecture Overview
- The library follows a hexagonal architecture: domain logic is isolated from I/O, and adapters implement external integrations. Add new external concerns as adapters behind ports.
