# CI/CD Pipeline Documentation

## Overview

This document describes the Continuous Integration and Continuous Deployment (CI/CD) pipeline for the project. The pipeline is implemented using GitHub Actions and enforces strict quality gates to ensure code quality, test coverage, and build integrity.

## Pipeline Architecture

The CI pipeline consists of four main jobs that run in parallel, followed by a quality gate check:

1. **Build** - Compiles the Go application
2. **Lint** - Runs static code analysis
3. **Test** - Executes tests with race detection and coverage
4. **Docker** - Builds Docker image (runs after other jobs pass)
5. **Quality Gate** - Final validation (runs after other jobs pass)

## Workflow Triggers

The CI pipeline is triggered on:
- Push to `main` or `develop` branches
- Pull requests targeting `main` or `develop` branches

## Jobs

### 1. Build Job

**Purpose**: Verify that the code compiles successfully.

**Steps**:
- Checkout code
- Set up Go 1.21
- Download and verify dependencies
- Build all packages

**Success Criteria**: All packages must compile without errors.

### 2. Lint Job

**Purpose**: Enforce code quality and style standards.

**Steps**:
- Checkout code
- Set up Go 1.21
- Run golangci-lint with custom configuration

**Linters Enabled**:
- **Error Detection**: errcheck, gosec, errorlint, nilerr
- **Code Quality**: gocritic, gocyclo, dupl, unparam, wastedassign
- **Style**: gofmt, goimports, revive, stylecheck, whitespace
- **Static Analysis**: staticcheck, govet, ineffassign, unused
- **Best Practices**: gosimple, unconvert, nakedret, prealloc
- **Correctness**: exportloopref, bodyclose, noctx, exhaustive, makezero, predeclared, thelper

**Custom Rules**:
- Maximum cyclomatic complexity: 15
- Duplication threshold: 100 tokens
- Named returns: max 30 lines
- Security checks enabled (gosec)
- Type assertions must be checked (errcheck)

**Success Criteria**: No linting errors. Code must pass all enabled linters.

### 3. Test Job

**Purpose**: Validate functionality and enforce coverage standards.

**Steps**:
- Checkout code
- Set up Go 1.21
- Download dependencies
- Run tests with race detector and coverage profiling
- Generate coverage report
- Check coverage threshold (≥80%)
- Upload coverage artifacts

**Test Flags**:
- `-race`: Enable data race detection
- `-coverprofile=coverage.out`: Generate coverage profile
- `-covermode=atomic`: Use atomic coverage mode for race detection

**Coverage Threshold**: **80% minimum**

**Success Criteria**:
- All tests must pass
- No data races detected
- Total coverage must be ≥80%

**Coverage Report**: Uploaded as artifact and retained for 30 days.

### 4. Docker Job

**Purpose**: Verify that the application can be containerized.

**Dependencies**: Requires build, lint, and test jobs to pass.

**Steps**:
- Checkout code
- Set up Docker Buildx
- Build Docker image with caching

**Success Criteria**: Docker image builds successfully.

### 5. Quality Gate Job

**Purpose**: Final validation that all quality checks have passed.

**Dependencies**: Requires build, lint, and test jobs to pass.

**Success Criteria**: All dependent jobs complete successfully.

## Quality Gates

The following quality gates must pass for a successful CI run:

| Gate | Requirement | Enforcement |
|------|-------------|-------------|
| **Build** | Code compiles without errors | ✅ Required |
| **Linting** | Passes all golangci-lint checks | ✅ Required |
| **Tests** | All tests pass with race detection | ✅ Required |
| **Coverage** | ≥80% code coverage | ✅ Required |
| **Docker** | Image builds successfully | ✅ Required |

## Pre-commit Hooks

Pre-commit hooks are configured to run local checks before code is committed. This helps catch issues early and reduces CI failures.

### Installation

Install pre-commit hooks:

```bash
# Install pre-commit tool
pip install pre-commit

# Install hooks
pre-commit install
```

### Hooks Enabled

1. **Standard Checks**:
   - Trailing whitespace removal
   - End-of-file fixer
   - YAML validation
   - Large file detection (max 1MB)
   - Merge conflict detection
   - Private key detection
   - Case conflict detection

2. **Go-specific Checks**:
   - go fmt
   - go vet
   - go imports
   - go cyclo (complexity check)
   - go critic
   - golangci-lint
   - Unit tests (with race detector)
   - go build
   - go mod tidy

### Manual Hook Execution

Run all hooks on all files:

```bash
pre-commit run --all-files
```

Run specific hook:

```bash
pre-commit run golangci-lint --all-files
```

## Local Development

### Running CI Checks Locally

Before pushing code, run these commands to verify your changes:

```bash
# Format code
make fmt

# Run linter
make lint

# Run tests with coverage
make test

# Check coverage
make coverage

# Build application
make build
```

### Coverage Check Script

To manually check coverage threshold:

```bash
go test -race -coverprofile=coverage.out -covermode=atomic ./...
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
echo "Total coverage: ${COVERAGE}%"
if (( $(echo "$COVERAGE < 80" | bc -l) )); then
  echo "❌ Coverage below 80%"
  exit 1
fi
echo "✅ Coverage meets threshold"
```

## golangci-lint Configuration

The project uses a comprehensive golangci-lint configuration (`.golangci.yml`) with:

- **30+ linters enabled** for comprehensive code analysis
- **Timeout**: 5 minutes
- **Complexity threshold**: 15
- **Custom rules** for revive and stylecheck
- **Security scanning** with gosec
- **Test file exemptions** for dupl and gosec

### Configuration Highlights

```yaml
linters-settings:
  gocyclo:
    min-complexity: 15
  
  dupl:
    threshold: 100
  
  nakedret:
    max-func-lines: 30
  
  gosec:
    severity: medium
    confidence: medium
```

## Troubleshooting

### Coverage Below Threshold

If coverage is below 80%, you need to add more tests:

1. Identify uncovered code:
   ```bash
   make coverage
   # Open coverage.html in browser
   ```

2. Write tests for uncovered functions/lines
3. Verify improvement:
   ```bash
   go test -coverprofile=coverage.out ./...
   go tool cover -func=coverage.out
   ```

### Linting Failures

If linting fails:

1. Run linter locally:
   ```bash
   make lint
   ```

2. Fix issues automatically where possible:
   ```bash
   make fmt
   goimports -w .
   ```

3. For complex issues, check the linter documentation:
   ```bash
   golangci-lint linters
   golangci-lint run --help
   ```

### Race Conditions

If tests fail with race detection:

1. Run tests with race detector locally:
   ```bash
   go test -race ./...
   ```

2. Fix data races by using proper synchronization (mutexes, channels, atomic operations)

3. Verify fix:
   ```bash
   go test -race ./...
   ```

### Docker Build Failures

If Docker build fails:

1. Build locally:
   ```bash
   docker build -t app:test .
   ```

2. Check Dockerfile syntax and dependencies
3. Ensure all required files are present and not gitignored

## Performance Optimization

The CI pipeline uses several optimizations:

- **Go module caching**: Speeds up dependency downloads
- **Docker layer caching**: Speeds up image builds (GitHub Actions cache)
- **Parallel job execution**: Build, lint, and test run concurrently
- **Allow parallel runners**: golangci-lint can run multiple instances

## Maintenance

### Updating Dependencies

1. Update GitHub Actions versions in `.github/workflows/ci.yml`
2. Update golangci-lint linters in `.golangci.yml`
3. Update pre-commit hooks in `.pre-commit-config.yaml`

### Adjusting Coverage Threshold

To change the coverage threshold, update the `THRESHOLD` variable in `.github/workflows/ci.yml`:

```yaml
- name: Check coverage threshold
  run: |
    COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
    THRESHOLD=80  # Change this value
    ...
```

### Adding New Linters

To add new linters, update `.golangci.yml`:

```yaml
linters:
  enable:
    - errcheck
    - new-linter-name  # Add here
```

## Best Practices

1. **Always run pre-commit hooks** before pushing
2. **Write tests for new functionality** to maintain coverage
3. **Fix linting issues** immediately; don't accumulate technical debt
4. **Monitor CI failures** and fix them promptly
5. **Keep dependencies updated** for security and performance
6. **Review coverage reports** to identify weak test areas
7. **Use meaningful commit messages** for better CI logs

## CI Status Badge

Add this badge to README.md to show CI status:

```markdown
[![CI](https://github.com/YOUR-USERNAME/YOUR-REPO/actions/workflows/ci.yml/badge.svg)](https://github.com/YOUR-USERNAME/YOUR-REPO/actions/workflows/ci.yml)
```

## Further Reading

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [golangci-lint Documentation](https://golangci-lint.run/)
- [pre-commit Documentation](https://pre-commit.com/)
- [Go Testing Guide](https://golang.org/doc/tutorial/add-a-test)
- [Go Coverage Documentation](https://go.dev/blog/cover)
