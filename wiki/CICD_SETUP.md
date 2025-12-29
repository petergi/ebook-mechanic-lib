# CI/CD Pipeline Setup

This document provides a complete overview of the CI/CD pipeline setup for this project.

## 📋 Overview

A comprehensive CI/CD pipeline has been implemented with:

- ✅ GitHub Actions workflow with parallel jobs
- ✅ golangci-lint with 30+ linters and custom rules
- ✅ Pre-commit hooks for local validation
- ✅ 80% code coverage threshold enforcement
- ✅ Docker image build validation
- ✅ Race detection in tests
- ✅ Detailed documentation

## 📁 Files Created

### GitHub Actions Workflow
- `.github/workflows/ci.yml` - Main CI pipeline with build, lint, test, Docker, and quality gate jobs

### Linting Configuration
- `.golangci.yml` - Comprehensive golangci-lint configuration with 30+ linters

### Pre-commit Hooks
- `.pre-commit-config.yaml` - Pre-commit hooks configuration for local validation

### Docker
- `Dockerfile` - Multi-stage Docker build for production-ready images
- `.dockerignore` - Docker ignore rules to optimize build context

### Scripts
- `scripts/setup-hooks.sh` - Helper script to install pre-commit hooks
- `scripts/check-coverage.sh` - Script to check coverage threshold locally
- `docs/scripts/README.md` - Documentation for scripts

### Documentation
- `docs/CI.md` - Comprehensive CI/CD pipeline documentation
- `docs/CI_QUICKSTART.md` - Quick start guide for developers
- `docs/CI_BADGE.md` - Instructions for adding CI status badges

### Configuration Updates
- `.gitignore` - Updated with pre-commit backup files

## 🚀 Quick Start

### For Developers

1. **Install dependencies:**
   ```bash
   make install
   ```

2. **Install golangci-lint (optional but recommended):**
   ```bash
   # macOS
   brew install golangci-lint
   
   # Linux/macOS
   curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
   ```

3. **Install pre-commit hooks (optional but recommended):**
   ```bash
   pip install pre-commit
   ./scripts/setup-hooks.sh
   ```

4. **Before committing, run:**
   ```bash
   make fmt
   make lint
   make test
   ./scripts/check-coverage.sh
   ```

### For Repository Administrators

1. **Enable GitHub Actions:**
   - Go to repository Settings → Actions → General
   - Enable "Allow all actions and reusable workflows"

2. **Protect branches:**
   - Go to Settings → Branches
   - Add branch protection rule for `main` and `develop`
   - Enable "Require status checks to pass before merging"
   - Select the CI workflow jobs as required checks

3. **Add CI status badge to README.md:**
   ```markdown
   [![CI](https://github.com/USERNAME/REPOSITORY/actions/workflows/ci.yml/badge.svg)](https://github.com/USERNAME/REPOSITORY/actions/workflows/ci.yml)
   ```

## 🔍 CI Pipeline Jobs

### 1. Build Job
- Compiles the Go application
- Verifies all packages build successfully
- Uses Go module caching for speed

### 2. Lint Job
- Runs golangci-lint with custom configuration
- Checks code quality, style, and security
- Uses 30+ linters including gosec, gocritic, and revive

### 3. Test Job
- Runs tests with race detector (`-race`)
- Generates coverage profile (`-coverprofile=coverage.out`)
- Enforces 80% minimum coverage threshold
- Uploads coverage report as artifact (30-day retention)

### 4. Docker Job
- Builds Docker image using multi-stage build
- Verifies containerization works correctly
- Uses GitHub Actions cache for layer caching
- Only runs after build, lint, and test pass

### 5. Quality Gate Job
- Final validation step
- Only runs after build, lint, and test pass
- Indicates all quality checks succeeded

## 📊 Quality Gates

| Gate | Requirement | Enforcement |
|------|-------------|-------------|
| Build | Code compiles without errors | ✅ Required |
| Linting | Passes all enabled linters | ✅ Required |
| Tests | All tests pass with `-race` | ✅ Required |
| Coverage | ≥80% code coverage | ✅ Required |
| Docker | Image builds successfully | ✅ Required |

## 🛠️ golangci-lint Configuration

The `.golangci.yml` file configures:

**Enabled Linters (30+):**
- Error detection: errcheck, gosec, errorlint, nilerr
- Code quality: gocritic, gocyclo, dupl, unparam, wastedassign
- Style: gofmt, goimports, revive, stylecheck, whitespace
- Static analysis: staticcheck, govet, ineffassign, unused
- Best practices: gosimple, unconvert, nakedret, prealloc
- Correctness: exportloopref, bodyclose, noctx, exhaustive, makezero

**Custom Rules:**
- Max cyclomatic complexity: 15
- Duplication threshold: 100 tokens
- Named returns: max 30 lines
- Security checks with gosec (medium severity)
- Test file exemptions for dupl and gosec

## 🪝 Pre-commit Hooks

The `.pre-commit-config.yaml` configures hooks that run before each commit:

**Standard Checks:**
- Trailing whitespace removal
- End-of-file fixer
- YAML validation
- Large file detection (1MB limit)
- Merge conflict detection
- Private key detection

**Go-specific Checks:**
- go fmt, go vet, go imports
- Cyclomatic complexity check (threshold: 15)
- go critic
- golangci-lint
- Unit tests with race detector
- go build
- go mod tidy

## 🐳 Docker Configuration

The `Dockerfile` implements a secure multi-stage build:

**Build Stage:**
- Uses golang:1.21-alpine
- Downloads and verifies dependencies
- Compiles static binary with optimizations

**Final Stage:**
- Uses minimal alpine:latest base
- Runs as non-root user (appuser)
- Includes CA certificates for HTTPS
- Exposes port 8080

## 📚 Documentation

- **[CI.md](CI.md)** - Comprehensive documentation covering all aspects of the CI/CD pipeline
- **[CI_QUICKSTART.md](CI_QUICKSTART.md)** - Quick start guide for developers
- **[CI_BADGE.md](CI_BADGE.md)** - Instructions for adding status badges

## 🔧 Local Development

### Run All CI Checks Locally

```bash
# Format code
make fmt

# Run linter
make lint

# Run tests
make test

# Check coverage
./scripts/check-coverage.sh

# Build binary
make build

# Build Docker image
make docker-build
```

### Manual Commands

```bash
# Run specific test package
go test -race -v ./internal/domain/...

# Run linter on specific path
golangci-lint run ./internal/...

# Generate coverage HTML report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## 🐛 Troubleshooting

### Coverage Below Threshold

```bash
# Generate HTML report to see uncovered lines
make coverage
# Open coverage.html in browser
```

### Linting Errors

```bash
# Auto-fix formatting issues
make fmt
goimports -w .

# See detailed linter output
golangci-lint run --verbose
```

### Pre-commit Hooks Issues

```bash
# Run hooks manually with verbose output
pre-commit run --all-files --verbose

# Update hook versions
pre-commit autoupdate

# Reinstall hooks
pre-commit uninstall
pre-commit install
```

### CI Failures

1. Check GitHub Actions logs for detailed error messages
2. Run the same commands locally
3. Ensure all dependencies are up to date
4. Verify Go version matches (1.21)

## 🔄 Maintenance

### Update GitHub Actions Versions

Edit `.github/workflows/ci.yml` and update action versions:
- `actions/checkout@v4`
- `actions/setup-go@v5`
- `golangci/golangci-lint-action@v4`
- `docker/setup-buildx-action@v3`
- `docker/build-push-action@v5`

### Update Pre-commit Hooks

```bash
pre-commit autoupdate
```

### Update golangci-lint

```bash
# Update to latest version
brew upgrade golangci-lint  # macOS
# or
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### Adjust Coverage Threshold

Edit `.github/workflows/ci.yml` and change the `THRESHOLD` value:

```yaml
- name: Check coverage threshold
  run: |
    THRESHOLD=85  # Change from 80 to desired value
```

Also update documentation to reflect the new threshold.

## 📈 Best Practices

1. **Always run pre-commit hooks** - Catch issues before they reach CI
2. **Write tests for new code** - Maintain coverage above threshold
3. **Fix linting issues immediately** - Don't accumulate technical debt
4. **Review coverage reports** - Identify gaps in test coverage
5. **Keep dependencies updated** - Security and performance improvements
6. **Use meaningful commit messages** - Help with debugging CI failures
7. **Monitor CI performance** - Optimize if jobs take too long

## 🎯 Next Steps

1. Review and merge this setup into your main branch
2. Enable branch protection rules
3. Add CI status badge to README.md
4. Set up team notifications for CI failures
5. Train team members on CI/CD workflow
6. Consider adding additional quality gates (e.g., dependency scanning, SAST)

## 📞 Support

For issues or questions:
- Read the full documentation in `docs/CI.md`
- Check troubleshooting sections
- Review GitHub Actions logs
- Consult with the team

---

**Pipeline Version:** 1.0  
**Last Updated:** 2024  
**Go Version:** 1.21+
