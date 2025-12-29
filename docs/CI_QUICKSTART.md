# CI/CD Quick Start Guide

This guide will help you quickly set up and use the CI/CD pipeline for local development.

## Prerequisites

- Go 1.21+
- golangci-lint (optional for local linting)
- pre-commit (optional for pre-commit hooks)
- Docker (optional for local Docker builds)

## Quick Setup

### 1. Install Dependencies

```bash
make install
```

### 2. Install Pre-commit Hooks (Optional but Recommended)

```bash
# Install pre-commit tool
pip install pre-commit

# Run setup script
./scripts/setup-hooks.sh
```

Alternatively, install manually:
```bash
pre-commit install
```

### 3. Install golangci-lint (Optional but Recommended)

**macOS/Linux:**
```bash
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
```

**Or using Homebrew:**
```bash
brew install golangci-lint
```

## Local Development Workflow

### Before Committing

Run these commands to ensure your code will pass CI:

```bash
# 1. Format code
make fmt

# 2. Run linter
make lint

# 3. Run tests
make test

# 4. Check coverage
./scripts/check-coverage.sh
```

### Using Pre-commit Hooks

If you installed pre-commit hooks, they will run automatically on `git commit`. To run manually:

```bash
# Run all hooks on staged files
pre-commit run

# Run all hooks on all files
pre-commit run --all-files

# Run specific hook
pre-commit run golangci-lint --all-files
```

## CI Pipeline Overview

When you push code or create a PR, the following checks run automatically:

1. **Build** - Verifies code compiles
2. **Lint** - Runs golangci-lint with 30+ linters
3. **Test** - Runs tests with race detector and coverage check (≥80%)
4. **Docker** - Builds Docker image
5. **Quality Gate** - Ensures all checks passed

## Common Tasks

### Check Coverage

```bash
# Use default threshold (80%)
./scripts/check-coverage.sh

# Use custom threshold
./scripts/check-coverage.sh 85

# Or use make
make coverage
```

### Run Linter

```bash
make lint
```

### Build Project

```bash
make build
```

### Build Docker Image

```bash
make docker-build
# or
docker build -t app:latest .
```

### Run Tests with Race Detection

```bash
go test -race ./...
```

## Troubleshooting

### Coverage Below 80%

1. Generate HTML coverage report:
   ```bash
   make coverage
   ```

2. Open `coverage.html` in your browser to see uncovered lines

3. Add tests for uncovered code

### Linting Errors

1. Some issues can be auto-fixed:
   ```bash
   make fmt
   goimports -w .
   ```

2. For other issues, check the error message and fix manually

3. To see all available linters and their rules:
   ```bash
   golangci-lint linters
   ```

### Pre-commit Hooks Failing

1. Run hooks manually to see details:
   ```bash
   pre-commit run --all-files --verbose
   ```

2. Fix the issues reported

3. To skip hooks temporarily (not recommended):
   ```bash
   git commit --no-verify
   ```

### CI Failures

1. Check the GitHub Actions logs for detailed error messages

2. Run the same checks locally:
   ```bash
   make lint
   make test
   ./scripts/check-coverage.sh
   make build
   ```

3. Fix issues and push again

## Best Practices

✅ **DO:**
- Run `make fmt` and `make lint` before committing
- Write tests for new functionality
- Keep coverage above 80%
- Use meaningful commit messages
- Run pre-commit hooks

❌ **DON'T:**
- Skip pre-commit hooks (use `--no-verify` sparingly)
- Commit code that doesn't compile
- Ignore linting errors
- Push without running tests locally
- Decrease coverage threshold without good reason

## Next Steps

- Read the full [CI/CD documentation](CI.md)
- Review the [golangci-lint configuration](.golangci.yml)
- Check the [GitHub Actions workflow](.github/workflows/ci.yml)
- Explore the [pre-commit configuration](.pre-commit-config.yaml)

## Getting Help

If you encounter issues:

1. Check the [CI/CD documentation](CI.md) for detailed troubleshooting
2. Review the error messages carefully
3. Check if similar issues occurred in past CI runs
4. Ask the team for help
