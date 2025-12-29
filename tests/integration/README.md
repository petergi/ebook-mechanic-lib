# Integration Test Suite

This directory contains comprehensive integration tests for the EPUB and PDF validation and repair functionality.

## Test Organization

### Core Integration Tests

- **epub_validator_integration_test.go**: Tests all EPUB error codes and validation scenarios
- **pdf_validator_integration_test.go**: Tests all PDF error codes and validation scenarios  
- **coverage_test.go**: Additional tests to maximize code coverage across all validators
- **real_world_scenarios_test.go**: Real-world usage patterns and workflows

### Benchmark Tests

- **benchmark_test.go**: Performance benchmarks for large file handling

## Running Tests

### All Integration Tests
```bash
cd tests/integration
go test -v
```

### Specific Test
```bash
go test -v -run TestEPUBValidatorIntegration_ValidMinimal
```

### With Coverage
```bash
go test -v -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Benchmark Tests
```bash
go test -v -bench=. -benchmem
```

### Run from Project Root
```bash
go test ./tests/integration/... -v
```

## Test Fixtures

Test fixtures are generated using the fixture generators in `testdata/`:

### Generate EPUB Fixtures
```bash
cd testdata/epub
go run generate_fixtures.go .
```

### Generate PDF Fixtures
```bash
cd testdata/pdf
go run generate_fixtures.go .
```

## Coverage Goals

The test suite aims for ≥80% code coverage across:

- EPUB validation (container, OPF, navigation, content)
- PDF validation (structure, header, trailer, xref, catalog)
- Repair services (EPUB and PDF)
- Reporter functionality (text, JSON, markdown)

## Test Patterns

### Error Code Validation
Each error code defined in the validators is tested with a specific fixture that triggers that error.

### Report Structure Validation
Tests verify that validation reports contain all expected fields and metadata.

### Real-World Scenarios
Tests simulate actual usage patterns including:
- Complete validation + repair workflows
- Batch validation of multiple files
- Progressive validation (structure → metadata → content)
- Multiple reporter formats
- Error recovery and graceful degradation

### Performance Benchmarks
Benchmarks measure performance with:
- Minimal files (baseline)
- Large files (100-500 chapters for EPUB, 100-1000 pages for PDF)
- Different validation modes (full, structure-only, metadata-only)

## Adding New Tests

When adding new validators or error codes:

1. Create corresponding test fixtures in `testdata/`
2. Add integration tests that verify the error code is properly triggered
3. Update the `AllErrorCodes` test case with the new error code
4. Add benchmark tests if the feature impacts performance
5. Update this README

## Notes

- Tests use `t.Skipf()` to gracefully handle missing fixtures
- All tests use relative paths from `tests/integration/` to `testdata/`
- Tests are designed to be run both individually and as a suite
- Benchmark tests only run when explicitly requested with `-bench` flag
