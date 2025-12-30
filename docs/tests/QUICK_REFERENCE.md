# Test Suite Quick Reference

## One-Liners

```bash
# Setup and run everything
make install && make generate-fixtures && make test && make coverage

# Quick test
make test

# Integration tests only
make test-integration

# Benchmarks
make test-bench

# Coverage
make coverage && open coverage.html

# Verify suite is complete
./tests/integration/verify_coverage.sh

# Clean and regenerate fixtures
make clean-fixtures && make generate-fixtures
```

## Common Commands

### Running Tests
```bash
# All tests
go test ./...

# Specific package
go test ./internal/adapters/epub/...

# Specific test
go test -v -run TestEPUBValidatorIntegration_ValidMinimal ./tests/integration/

# With race detector
go test -race ./...

# Short mode (skip slow tests)
go test -short ./...

# With timeout
go test -timeout 10m ./...
```

### Coverage
```bash
# Generate coverage
go test -coverprofile=coverage.out ./...

# View in terminal
go tool cover -func=coverage.out

# View in browser
go tool cover -html=coverage.out

# Check specific package
go test -cover ./internal/adapters/epub/
```

### Benchmarks
```bash
# All benchmarks
go test -bench=. ./tests/integration/

# Specific benchmark
go test -bench=BenchmarkEPUBValidation_Large500 ./tests/integration/

# With memory stats
go test -bench=. -benchmem ./tests/integration/

# Save results
go test -bench=. ./tests/integration/ > bench.txt

# Compare results
benchstat old.txt new.txt
```

### Fixtures
```bash
# Generate EPUB fixtures
cd testdata/epub && go run generate_fixtures.go .

# Generate PDF fixtures
cd testdata/pdf && go run generate_fixtures.go .

# Generate all
make generate-fixtures

# Clean all
make clean-fixtures

# Check what's generated
ls -lh testdata/epub/valid/
ls -lh testdata/epub/invalid/
```

## File Locations

```
tests/integration/          # Integration tests
testdata/epub/valid/        # Valid EPUB fixtures
testdata/epub/invalid/      # Invalid EPUB fixtures
testdata/pdf/valid/         # Valid PDF fixtures
testdata/pdf/invalid/       # Invalid PDF fixtures
coverage.html               # Coverage report
```

## Error Codes

### EPUB
```
EPUB-CONTAINER-001  # Invalid ZIP
EPUB-CONTAINER-002  # Invalid mimetype
EPUB-CONTAINER-003  # Mimetype not first
EPUB-CONTAINER-004  # Missing container.xml
EPUB-CONTAINER-005  # Invalid container.xml
EPUB-OPF-001        # Invalid OPF XML
EPUB-OPF-002        # Missing title
EPUB-OPF-003        # Missing identifier
EPUB-OPF-004        # Missing language
EPUB-OPF-005        # Missing modified
EPUB-OPF-009        # Missing nav document
EPUB-NAV-002        # Missing TOC
EPUB-NAV-006        # Missing nav element
EPUB-CONTENT-002    # Missing DOCTYPE
EPUB-CONTENT-007    # Invalid namespace
```

### PDF
```
PDF-HEADER-001      # Invalid header
PDF-HEADER-002      # Invalid version
PDF-TRAILER-001     # Missing startxref
PDF-TRAILER-003     # Missing EOF
PDF-XREF-001        # Invalid xref
PDF-CATALOG-003     # Invalid catalog
PDF-STRUCTURE-012   # Structure error
```

## Troubleshooting

### Fixtures not found
```bash
make generate-fixtures
```

### Tests failing
```bash
# Regenerate fixtures
make clean-fixtures generate-fixtures

# Run with verbose output
go test -v ./tests/integration/...
```

### Coverage too low
```bash
# See what's missing
go tool cover -func=coverage.out | awk '$3 < 80.0'
```

### Benchmarks failing
```bash
# Ensure fixtures exist
make generate-fixtures

# Run with more time
go test -bench=. -timeout 10m ./tests/integration/
```

## Make Targets

```bash
make help              # Show all targets
make install           # Install dependencies
make test              # Run all tests
make test-unit         # Run unit tests
make test-integration  # Run integration tests
make test-bench        # Run benchmarks
make coverage          # Generate coverage
make generate-fixtures # Generate fixtures
make clean-fixtures    # Clean fixtures
make lint              # Run linter
make fmt               # Format code
```

## Test Patterns

### Table-driven test
```go
testCases := []struct {}
    name string
    file string
    expectedCode string
}{}
    {"NotZip", "invalid/not_zip.epub", "EPUB-CONTAINER-001"},
}

for _, tc := range testCases {}
    t.Run(tc.name, func(t *testing.T) {})
        // test logic
    })
}
```

### Benchmark
```go
func BenchmarkValidator(b *testing.B) {}
    validator := epub.NewEPUBValidator()
    b.ResetTimer()
    for i := 0; i < b.N; i++ {}
        validator.ValidateFile(ctx, file)
    }
}
```

## Documentation

- `docs/tests/README.md` - Main test documentation
- `docs/tests/integration/README.md` - Integration test guide
- `docs/testdata/README.md` - Fixture documentation
- `docs/TEST_SUITE.md` - Comprehensive guide

## CI/CD

```yaml
# GitHub Actions example
- name: Test
  run: |
    make generate-fixtures
    make test
    make coverage
```

## Performance Targets

- Minimal EPUB: <10ms
- Large EPUB (500ch): <200ms
- Minimal PDF: <5ms
- Large PDF (1000p): <100ms
- Memory: <50MB for large files

## Coverage Target

- Overall: ≥80%
- Validators: ≥85%
- Domain: ≥90%

---

**Pro Tip**: Bookmark this file for quick reference during development!
