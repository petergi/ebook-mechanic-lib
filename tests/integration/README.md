# Integration Test Suite

This directory contains comprehensive integration tests for EPUB and PDF validation with systematic coverage of all error codes and edge cases.

## Overview

The integration test suite provides:
- **Systematic coverage** of all error codes (100%)
- **Table-driven tests** for maintainability and clarity
- **Performance benchmarks** for large files (>10MB)
- **Edge case validation** for robustness
- **Oracle comparison** methodology (epubcheck for EPUB)

## Test Organization

### Core Integration Tests

- **epub_validator_integration_test.go**: Comprehensive EPUB validation tests
  - Table-driven error code tests (all EPUB-XXX-XXX codes)
  - Valid file tests (minimal, complex nested, multiple rootfiles)
  - Performance tests (100, 500, 2000+ chapters)
  - Edge case tests (>10MB files, compression issues, corruption)
  
- **pdf_validator_integration_test.go**: Comprehensive PDF validation tests
  - Table-driven error code tests (all PDF-XXX-XXX codes)
  - Valid file tests (minimal, with images, multi-page)
  - Performance tests (100, 1000, 5000+ pages)
  - Corruption scenario tests (header, trailer, xref, catalog)
  
- **benchmark_test.go**: Performance benchmarks for large file handling

## Running Tests

### All Integration Tests
```bash
# From project root
make test-integration

# Or directly
cd tests/integration
go test -v
```

### Specific Test Suites
```bash
# EPUB tests only
go test -v -run TestEPUB

# PDF tests only
go test -v -run TestPDF

# Performance tests only
go test -v -run Performance

# Edge cases only
go test -v -run EdgeCases

# Table-driven error code tests
go test -v -run TableDriven
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

## Test Fixtures

Test fixtures are auto-generated from `testdata/{epub,pdf}/generate_fixtures.go`.

### Generate All Fixtures
```bash
# EPUB fixtures
cd testdata/epub && go run generate_fixtures.go

# PDF fixtures
cd testdata/pdf && go run generate_fixtures.go

# Or use make
make test  # Auto-generates if missing
```

### Fixture Organization

See detailed documentation:
- `../../testdata/README.md` - Overall test corpus documentation
- `../../testdata/epub/README.md` - EPUB fixture details
- `../../testdata/pdf/README.md` - PDF fixture details

## Test Coverage

### EPUB Error Codes (100% Coverage)

| Category | Codes | Test Status |
|----------|-------|-------------|
| Container | EPUB-CONTAINER-001 through 005 | ✅ Covered |
| OPF | EPUB-OPF-001 through 015 | ✅ Covered |
| Navigation | EPUB-NAV-001 through 006 | ✅ Covered |
| Content | EPUB-CONTENT-001 through 008 | ✅ Covered |

### PDF Error Codes (100% Coverage)

| Category | Codes | Test Status |
|----------|-------|-------------|
| Header | PDF-HEADER-001, 002 | ✅ Covered |
| Trailer | PDF-TRAILER-001, 002, 003 | ✅ Covered |
| Cross-Reference | PDF-XREF-001, 002, 003 | ✅ Covered |
| Catalog | PDF-CATALOG-001, 002, 003 | ✅ Covered |
| Structure | PDF-STRUCTURE-012 | ✅ Covered |

## Test Methodology

### Table-Driven Testing

All error codes use table-driven tests for consistency:

```go
testCases := []struct {
    name         string
    file         string
    expectedCode string
    shouldFail   bool
    description  string
}{
    {
        name:         "Container_NotZip",
        file:         "invalid/not_zip.epub",
        expectedCode: "EPUB-CONTAINER-001",
        shouldFail:   true,
        description:  "File is not a valid ZIP archive",
    },
    // ...
}
```

### Fixture Auto-Generation

Tests skip gracefully if fixtures are missing:
```go
if _, err := os.Stat(testFile); os.IsNotExist(err) {
    t.Skipf("Test file not found: %s", testFile)
}
```

### Report Validation

Each test validates:
- **Validity determination**: Correct true/false
- **Error code presence**: Expected codes detected
- **Error structure**: All fields populated (Code, Message, Severity, Timestamp, Location)
- **Report metadata**: FilePath, FileType, ValidationTime, Duration

### Performance Validation

Performance tests track:
- Validation duration vs. file size
- Memory usage (via successful completion)
- Scalability (linear vs. exponential time)

## Key Test Functions

### EPUB Tests

#### TestEPUBValidatorIntegration_TableDriven_AllErrorCodes
- Systematically tests all EPUB error codes
- Uses descriptive test case structure
- Validates expected error code appears in results
- Logs all errors for debugging

#### TestEPUBValidatorIntegration_ValidFiles
- Tests valid EPUB recognition
- Covers minimal, complex nested, multiple rootfiles
- Ensures no false positives

#### TestEPUBValidatorIntegration_PerformanceLargeFiles
- Tests 100-chapter and 500-chapter EPUBs
- Validates completion within time limits
- Logs duration for monitoring

#### TestEPUBValidatorIntegration_EdgeCases
- Very large files (>10MB)
- Compressed mimetype
- Corrupt ZIP files

### PDF Tests

#### TestPDFValidatorIntegration_TableDriven_AllErrorCodes
- Systematically tests all PDF error codes
- Structured test cases with descriptions
- Validates error detection

#### TestPDFValidatorIntegration_CorruptionScenarios
- Comprehensive corruption testing
- Header, trailer, xref, catalog errors
- Stream and object malformation

#### TestPDFValidatorIntegration_Systematic_Coverage
- Explicit error code to fixture mapping
- Ensures every code has a test

## Oracle Comparison

### EPUB: epubcheck

Compare with the reference validator:

```bash
# Install epubcheck (Java required)
# Download from: https://github.com/w3c/epubcheck/releases

# Validate fixture
java -jar epubcheck.jar testdata/epub/valid/minimal.epub

# Compare with our validator
go test -v -run TestEPUBValidatorIntegration_ValidMinimal
```

### PDF: Standard Tools

Compare with standard PDF tools:

```bash
# QPDF
qpdf --check testdata/pdf/valid/minimal.pdf

# Poppler
pdfinfo testdata/pdf/valid/minimal.pdf

# PDF Toolkit
pdftk testdata/pdf/valid/minimal.pdf dump_data
```

## Adding New Tests

### For a New Error Code

1. **Create fixture** in `testdata/{epub|pdf}/generate_fixtures.go`:
   ```go
   func createNewErrorCase() []byte {
       // Build file with specific error
   }
   ```

2. **Add to fixtures map**:
   ```go
   "invalid/new_error.{epub|pdf}": createNewErrorCase(),
   ```

3. **Regenerate fixtures**:
   ```bash
   cd testdata/{epub|pdf} && go run generate_fixtures.go
   ```

4. **Add test case** to table:
   ```go
   {
       name:         "NewError",
       file:         "invalid/new_error.epub",
       expectedCode: "EPUB-NEW-XXX",
       shouldFail:   true,
       description:  "Description of the error",
   },
   ```

5. **Run test**:
   ```bash
   go test -v ./tests/integration/ -run TableDriven
   ```

### For Performance Testing

1. Create large fixture in generator
2. Add to performance test function
3. Set reasonable time expectations
4. Monitor for regressions

## Continuous Integration

CI runs:
- All integration tests
- Fixture auto-generation
- Coverage reporting (≥80% target)
- Performance regression detection

## Performance Expectations

Based on comprehensive test corpus:

| File Type | Size | Pages/Chapters | Expected Time |
|-----------|------|----------------|---------------|
| EPUB (small) | ~1 KB | 1 chapter | < 10 ms |
| EPUB (medium) | ~500 KB | 100 chapters | < 500 ms |
| EPUB (large) | ~2.5 MB | 500 chapters | < 1500 ms |
| EPUB (very large) | ~10+ MB | 2000+ chapters | < 5000 ms |
| PDF (small) | ~500 bytes | 1 page | < 10 ms |
| PDF (medium) | ~50 KB | 100 pages | < 100 ms |
| PDF (large) | ~500 KB | 1000 pages | < 500 ms |
| PDF (very large) | ~10+ MB | 5000 pages | < 2000 ms |

## Troubleshooting

### Fixture Not Found
```
Test file not found: ../../testdata/epub/valid/minimal.epub
```
**Solution**: Generate fixtures with `go run generate_fixtures.go` in the testdata directory.

### Unexpected Error Code
```
Expected error code EPUB-OPF-002
Got errors: [EPUB-OPF-001]
```
**Solution**: 
1. Check if fixture generates intended error
2. Verify validator implementation
3. Update test expectation if behavior is correct

### Performance Regression
```
Validation took 5500 ms, expected < 5000 ms
```
**Solution**:
1. Profile with `go test -cpuprofile`
2. Identify bottlenecks
3. Optimize or adjust expectations

## References

- [EPUB 3.3 Specification](https://www.w3.org/TR/epub-33/)
- [PDF 1.7 Specification (ISO 32000-1:2008)](https://www.adobe.com/content/dam/acom/en/devnet/pdf/pdfs/PDF32000_2008.pdf)
- [epubcheck](https://github.com/w3c/epubcheck) - EPUB reference validator
- [QPDF](http://qpdf.sourceforge.net/) - PDF structure checker
