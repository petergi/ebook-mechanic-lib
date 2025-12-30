# Test Infrastructure

Complete testing infrastructure for ebm-lib validation and repair functionality.

## Quick Start

```bash
# 1. Generate test fixtures
make generate-fixtures

# 2. Run all tests
make test

# 3. View coverage
make coverage
open coverage.html

# 4. Run benchmarks
make test-bench
```

## Directory Structure

```
tests/
├── integration/                    # Integration test suite
│   ├── epub_validator_integration_test.go
│   ├── pdf_validator_integration_test.go
│   ├── coverage_test.go
│   ├── real_world_scenarios_test.go
│   ├── benchmark_test.go
│   ├── README.md                  # Integration test documentation
│   ├── TEST_SUMMARY.md            # Implementation summary
│   ├── run_tests.sh               # Complete test runner script
│   └── verify_coverage.sh         # Coverage verification script
└── README.md                       # This file
```

## Test Types

### Unit Tests
- Located alongside source code (`*_test.go`)
- Fast, isolated component testing
- Run with: `go test ./internal/...`

### Integration Tests  
- Located in `tests/integration/`
- End-to-end validation with real fixtures
- Run with: `make test-integration`

### Benchmark Tests
- Performance and memory profiling
- Large file handling validation
- Run with: `make test-bench`

## Coverage Goals

**Target: ≥80% overall coverage**

Key packages:
- `internal/adapters/epub` - EPUB validation
- `internal/adapters/pdf` - PDF validation
- `internal/adapters/reporter` - Report generation
- `internal/domain` - Core domain models

Run coverage analysis:
```bash
make coverage
# Opens coverage.html in browser
```

## Test Fixtures

### Location
- `testdata/epub/` - EPUB test files
- `testdata/pdf/` - PDF test files

### Generation
Test fixtures are generated programmatically:

```bash
# Generate all fixtures
make generate-fixtures

# Generate specific format
cd testdata/epub && go run generate_fixtures.go .
cd testdata/pdf && go run generate_fixtures.go .

# Clean and regenerate
make clean-fixtures generate-fixtures
```

### Fixture Coverage

**EPUB Fixtures:**
- 3 valid (minimal, 100 chapters, 500 chapters)
- 16 invalid (covering all error codes)

**PDF Fixtures:**
- 3 valid (minimal, 100 pages, 1000 pages)
- 9 invalid (covering all error codes)

See `docs/testdata/README.md` for complete fixture documentation.

## Make Targets

| Target | Description |
|--------|-------------|
| `make test` | Run all tests (unit + integration) |
| `make test-unit` | Run only unit tests |
| `make test-integration` | Run only integration tests |
| `make test-bench` | Run benchmark tests |
| `make coverage` | Generate coverage report |
| `make generate-fixtures` | Generate test fixtures |
| `make clean-fixtures` | Clean generated fixtures |

## Scripts

### `integration/run_tests.sh`
Complete test suite runner:
1. Checks and generates fixtures
2. Runs integration tests
3. Generates coverage report
4. Runs benchmarks

```bash
./tests/integration/run_tests.sh
```

### `integration/verify_coverage.sh`
Verifies test suite completeness:
1. Checks all error codes are tested
2. Verifies fixtures exist
3. Analyzes coverage percentage
4. Checks benchmark tests exist

```bash
./tests/integration/verify_coverage.sh
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Install dependencies
        run: make install
      
      - name: Generate fixtures
        run: make generate-fixtures
      
      - name: Run tests
        run: make test
      
      - name: Generate coverage
        run: make coverage
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.out
```

## Development Workflow

### Running Tests During Development

```bash
# Run specific test
cd tests/integration
go test -v -run TestEPUBValidatorIntegration_ValidMinimal

# Run with detailed output
go test -v ./tests/integration/...

# Run with race detector
go test -race ./...

# Quick coverage check
go test -cover ./internal/adapters/epub/...
```

### Adding New Tests

1. **Add error code** to validator
2. **Create fixture** in `testdata/*/generate_fixtures.go`
3. **Regenerate fixtures**: `make clean-fixtures generate-fixtures`
4. **Add test case** to integration tests
5. **Run tests**: `make test`
6. **Verify coverage**: `make coverage`
7. **Update documentation**

### Debugging Test Failures

```bash
# Run with verbose output
go test -v ./tests/integration/...

# Run single test
go test -v -run TestName

# Check which fixtures are missing
ls testdata/epub/valid/
ls testdata/epub/invalid/

# Regenerate all fixtures
make clean-fixtures generate-fixtures
```

## Performance Testing

### Running Benchmarks

```bash
# All benchmarks
make test-bench

# Specific benchmark
cd tests/integration
go test -bench=BenchmarkEPUBValidation_Large500 -benchmem

# With CPU profiling
go test -bench=. -cpuprofile=cpu.prof
go tool pprof cpu.prof

# With memory profiling
go test -bench=. -memprofile=mem.prof
go tool pprof mem.prof
```

### Performance Targets

- Minimal files: <10ms validation
- Large EPUB (500 chapters): <200ms
- Large PDF (1000 pages): <100ms
- Memory: <50MB for large files

## Troubleshooting

### "Test file not found"
```bash
# Generate missing fixtures
make generate-fixtures
```

### "Coverage below 80%"
```bash
# See which packages need more tests
make coverage
go tool cover -func=coverage.out | awk '$3 < 80.0'
```

### Tests timeout
```bash
# Increase timeout
go test -timeout 10m ./...
```

### Race conditions detected
```bash
# Run specific test without race detector
go test -run TestName ./...
```

## Best Practices

1. **Always generate fixtures before testing**
   ```bash
   make generate-fixtures
   ```

2. **Run full suite before committing**
   ```bash
   make test && make test-bench
   ```

3. **Check coverage**
   ```bash
   make coverage
   ```

4. **Use table-driven tests** for multiple scenarios

5. **Add benchmarks** for performance-sensitive code

6. **Document test intent** with clear names and comments

7. **Skip gracefully** when fixtures are missing

## Documentation

- `docs/tests/integration/README.md` - Integration test guide
- `docs/tests/integration/TEST_SUMMARY.md` - Implementation details
- `docs/testdata/README.md` - Fixture documentation
- `docs/TEST_SUITE.md` - Comprehensive test documentation

## Maintenance

### Weekly
- Run full test suite: `make test`
- Check coverage: `make coverage`
- Run benchmarks: `make test-bench`

### Before Release
- Verify coverage ≥80%: `./tests/integration/verify_coverage.sh`
- Run all tests with race detector: `go test -race ./...`
- Update fixture generators if needed
- Update documentation

### After Adding Features
- Add corresponding fixtures
- Add integration tests
- Add benchmarks if needed
- Update error code lists
- Update documentation

## Support

For issues or questions:
1. Check test documentation
2. Run verification script: `./tests/integration/verify_coverage.sh`
3. Check CI/CD logs
4. Review test output: `go test -v ./...`

## License

Same as main project license.
