# Scripts

This directory contains utility scripts for development and CI/CD operations.

## Available Scripts

### setup-hooks.sh

Sets up pre-commit hooks for the repository.

**Usage:**
```bash
./scripts/setup-hooks.sh
```

**Prerequisites:**
- pre-commit must be installed (`pip install pre-commit` or `brew install pre-commit`)

### check-coverage.sh

Runs tests with coverage and checks if coverage meets the threshold.

**Usage:**
```bash
# Use default threshold (80%)
./scripts/check-coverage.sh

# Use custom threshold
./scripts/check-coverage.sh 75
```

**Features:**
- Runs tests with race detector
- Generates coverage report
- Checks coverage against threshold
- Provides guidance on viewing detailed coverage

## Making Scripts Executable

If scripts are not executable, run:
```bash
chmod +x scripts/*.sh
```
