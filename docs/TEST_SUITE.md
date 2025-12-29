# Comprehensive Test Suite Documentation

This document describes the complete test suite for ebm-lib validation and repair functionality.

## Overview

The test suite provides comprehensive coverage of:
- EPUB validation (all error codes)
- PDF validation (all error codes)
- Repair services
- Reporter functionality
- Performance benchmarks
- Real-world scenarios

## Test Organization

### Unit Tests
Located alongside implementation files (`*_test.go`)
- Component-level validation
- Fast execution (<1s)
- No external dependencies

### Integration Tests
Located in `tests/integration/`
- End-to-end validation workflows
- Tests against generated fixtures
- Coverage target: ≥80%

### Benchmark Tests
Located in `tests/integration/benchmark_test.go`
- Performance measurement
- Large file handling
- Memory profiling

## Running Tests

### Quick Start
```bash
# Install dependencies
make install

# Generate test fixtures
make generate-fixtures

# Run all tests
make test

# Run with coverage
make coverage

# Run integration tests only
make test-integration

# Run benchmarks
make test-bench
```

### Advanced Usage

#### Run specific test
```bash
cd tests/integration
go test -v -run TestEPUBValidatorIntegration_ValidMinimal
```

#### Run with race detector
```bash
go test -race ./...
```

#### Run with memory profiling
```bash
go test -bench=. -benchmem -memprofile=mem.out ./tests/integration/
go tool pprof mem.out
```

#### Run with CPU profiling
```bash
go test -bench=. -cpuprofile=cpu.out ./tests/integration/
go tool pprof cpu.out
```

## Test Fixtures

### Generation
Test fixtures are generated programmatically using:
- `testdata/epub/generate_fixtures.go` - EPUB test files
- `testdata/pdf/generate_fixtures.go` - PDF test files

```bash
# Generate all fixtures
make generate-fixtures

# Clean fixtures
make clean-fixtures

# Regenerate
make clean-fixtures generate-fixtures
```

### Fixture Types

#### Valid EPUBs
- `minimal.epub` - Minimal EPUB3 (baseline)
- `large_100_chapters.epub` - 100 chapters (~1MB)
- `large_500_chapters.epub` - 500 chapters (~5MB)

#### Invalid EPUBs
Each fixture triggers a specific error code:

| File | Error Code | Description |
|------|-----------|-------------|
| `not_zip.epub` | EPUB-CONTAINER-001 | Not a ZIP file |
| `corrupt_zip.epub` | EPUB-CONTAINER-001 | Truncated ZIP |
| `wrong_mimetype.epub` | EPUB-CONTAINER-002 | Wrong mimetype content |
| `mimetype_not_first.epub` | EPUB-CONTAINER-003 | Mimetype not first |
| `mimetype_compressed.epub` | EPUB-CONTAINER-002 | Compressed mimetype |
| `no_container.epub` | EPUB-CONTAINER-004 | Missing container.xml |
| `invalid_container_xml.epub` | EPUB-CONTAINER-005 | Malformed XML |
| `no_rootfile.epub` | EPUB-CONTAINER-005 | Empty rootfiles |
| `invalid_opf.epub` | EPUB-OPF-001 | Malformed OPF |
| `missing_title.epub` | EPUB-OPF-002 | Missing dc:title |
| `missing_identifier.epub` | EPUB-OPF-003 | Missing dc:identifier |
| `missing_language.epub` | EPUB-OPF-004 | Missing dc:language |
| `missing_modified.epub` | EPUB-OPF-005 | Missing dcterms:modified |
| `missing_nav_document.epub` | EPUB-OPF-009 | No nav in manifest |
| `invalid_nav_document.epub` | EPUB-NAV-002/006 | Invalid nav structure |
| `invalid_content_document.epub` | EPUB-CONTENT-002/007 | Invalid XHTML |

#### Valid PDFs
- `minimal.pdf` - Minimal PDF 1.4 (baseline)
- `large_100_pages.pdf` - 100 pages (~200KB)
- `large_1000_pages.pdf` - 1000 pages (~2MB)

#### Invalid PDFs
Each fixture triggers a specific error code:

| File | Error Code | Description |
|------|-----------|-------------|
| `not_pdf.pdf` | PDF-HEADER-001 | Not a PDF file |
| `no_header.pdf` | PDF-HEADER-001 | Missing header |
| `invalid_version.pdf` | PDF-HEADER-002 | Invalid version |
| `no_eof.pdf` | PDF-TRAILER-003 | Missing %%EOF |
| `no_startxref.pdf` | PDF-TRAILER-001 | Missing startxref |
| `corrupt_xref.pdf` | PDF-XREF-001 | Corrupt xref table |
| `no_catalog.pdf` | PDF-STRUCTURE-012 | Missing catalog |
| `invalid_catalog.pdf` | PDF-CATALOG-003 | Invalid catalog |
| `corrupt.pdf` | Various | General corruption |

## Coverage Goals

### Target: ≥80% Coverage

#### Current Coverage by Package:
Run `make coverage` to see current coverage metrics.

Expected high-coverage packages:
- `internal/adapters/epub` - EPUB validation
- `internal/adapters/pdf` - PDF validation
- `internal/domain` - Domain models
- `internal/adapters/reporter` - Report generation

#### Coverage Exclusions:
- Example code (`examples/`)
- Main entry points (`cmd/`)
- Generated code
- Third-party integrations (mocked)

## Test Patterns

### Integration Test Pattern
```go
func TestValidator_Scenario(t *testing.T) {}
    // Setup
    validator := epub.NewEPUBValidator()
    ctx := context.Background()
    testFile := filepath.Join("..", "..", "testdata", "epub", "valid", "minimal.epub")
    
    // Skip if fixture missing
    if _, err := os.Stat(testFile); os.IsNotExist(err) {}
        t.Skipf("Test file not found: %s", testFile)
    }
    
    // Execute
    report, err := validator.ValidateFile(ctx, testFile)
    
    // Assert
    if err != nil {}
        t.Fatalf("ValidateFile failed: %v", err)
    }
    if !report.IsValid {}
        t.Errorf("Expected valid, got invalid")
    }
}
```

### Benchmark Pattern
```go
func BenchmarkValidator_Scenario(b *testing.B) {}
    validator := epub.NewEPUBValidator()
    ctx := context.Background()
    testFile := "..."
    
    // Check fixture exists
    if _, err := os.Stat(testFile); os.IsNotExist(err) {}
        b.Skipf("Test file not found: %s", testFile)
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {}
        _, err := validator.ValidateFile(ctx, testFile)
        if err != nil {}
            b.Fatalf("ValidateFile failed: %v", err)
        }
    }
}
```

## Performance Targets

### Validation Performance
- Minimal EPUB (<10KB): <10ms
- Minimal PDF (<5KB): <5ms
- Large EPUB (100 chapters): <100ms
- Large PDF (100 pages): <50ms

### Memory Usage
- Minimal files: <5MB allocated
- Large files: <50MB allocated
- No memory leaks (constant memory in benchmarks)

## CI/CD Integration

### GitHub Actions
```yaml
- name: Run Tests
  run: |
    make generate-fixtures
    make test
    make coverage

- name: Run Benchmarks
  run: make test-bench
```

### Coverage Reporting
Coverage reports are automatically generated and can be uploaded to coverage services:
```bash
go test -coverprofile=coverage.out -covermode=atomic ./...
# Upload to codecov, coveralls, etc.
```

## Troubleshooting

### Tests Fail to Find Fixtures
```bash
# Regenerate fixtures
make clean-fixtures generate-fixtures
```

### Tests Timeout
```bash
# Increase timeout
go test -timeout 10m ./...
```

### Race Detector Issues
```bash
# Run without race detector
go test ./...
```

### Memory Issues with Large Files
```bash
# Skip large file tests
go test -short ./...
```

## Maintenance

### Adding New Error Codes
1. Add error code constant to validator
2. Update `generate_fixtures.go` with new fixture
3. Regenerate fixtures: `make generate-fixtures`
4. Add test case to `AllErrorCodes` test
5. Update this documentation

### Adding New Validators
1. Implement validator with tests
2. Add integration tests in `tests/integration/`
3. Add benchmark tests if performance-sensitive
4. Update coverage targets

### Updating Fixtures
1. Modify `generate_fixtures.go`
2. Run `make clean-fixtures generate-fixtures`
3. Verify tests still pass: `make test`
4. Update documentation

## Best Practices

1. **Always skip when fixtures missing** - Tests should be robust
2. **Use table-driven tests** - Easier to add scenarios
3. **Test error paths** - Not just happy paths
4. **Measure performance** - Use benchmarks for critical code
5. **Mock external dependencies** - Tests should be fast and reliable
6. **Document test intent** - Clear test names and comments
7. **Clean up resources** - Close files, remove temp files
8. **Use helpers** - Reduce duplication in tests

## Resources

- [Go Testing Documentation](https://pkg.go.dev/testing)
- [Table Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)
- [Testify Assertions](https://github.com/stretchr/testify)
- [Go Benchmark Guide](https://dave.cheney.net/2013/06/30/how-to-write-benchmarks-in-go)
