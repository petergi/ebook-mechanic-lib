# Test Data Fixtures

This directory contains comprehensive test fixtures for EPUB and PDF validation, including valid files, invalid files with specific error conditions, and edge cases for performance testing.

## Directory Structure

```
testdata/
├── epub/
│   ├── valid/              # Valid EPUB files
│   ├── invalid/            # Invalid EPUB files with specific errors
│   ├── edge_cases/         # Edge cases (large files, etc.)
│   ├── generate_fixtures.go
│   └── README.md
├── pdf/
│   ├── valid/              # Valid PDF files
│   ├── invalid/            # Invalid PDF files with specific errors
│   ├── edge_cases/         # Edge cases (large files, encrypted, etc.)
│   ├── generate_fixtures.go
│   └── README.md
└── README.md               # This file
```

## Generating Fixtures

To generate all test fixtures:

```bash
# Generate EPUB fixtures
cd testdata/epub
go run generate_fixtures.go

# Generate PDF fixtures
cd testdata/pdf
go run generate_fixtures.go

# Or use the Makefile from the project root
make test  # Auto-generates fixtures if missing
```

## EPUB Test Fixtures

### Valid EPUBs

| Fixture | Description | Size | Error Codes Tested |
|---------|-------------|------|-------------------|
| `valid/minimal.epub` | Minimal valid EPUB 3.0 | ~1 KB | None (valid) |
| `valid/multiple_rootfiles.epub` | EPUB with multiple rootfiles in container.xml | ~2 KB | None (valid) |
| `valid/complex_nested.epub` | Complex nested directory structure with images and CSS | ~3 KB | None (valid) |
| `valid/large_100_chapters.epub` | 100 chapters for performance testing | ~500 KB | None (valid) |
| `valid/large_500_chapters.epub` | 500 chapters for performance testing | ~2.5 MB | None (valid) |

### Invalid EPUBs - Container Errors

| Fixture | Description | Error Code | Category |
|---------|-------------|------------|----------|
| `invalid/not_zip.epub` | Not a ZIP file | EPUB-CONTAINER-001 | Container |
| `invalid/corrupt_zip.epub` | Truncated/corrupt ZIP | EPUB-CONTAINER-001 | Container |
| `invalid/wrong_mimetype.epub` | Incorrect mimetype content | EPUB-CONTAINER-002 | Container |
| `invalid/mimetype_not_first.epub` | mimetype not first in ZIP | EPUB-CONTAINER-003 | Container |
| `invalid/mimetype_compressed.epub` | mimetype stored compressed | EPUB-CONTAINER-002 | Container |
| `invalid/no_container.epub` | Missing META-INF/container.xml | EPUB-CONTAINER-004 | Container |
| `invalid/invalid_container_xml.epub` | Malformed container.xml | EPUB-CONTAINER-005 | Container |
| `invalid/no_rootfile.epub` | container.xml with no rootfiles | EPUB-CONTAINER-005 | Container |

### Invalid EPUBs - OPF Errors

| Fixture | Description | Error Code | Category |
|---------|-------------|------------|----------|
| `invalid/invalid_opf.epub` | Malformed OPF XML | EPUB-OPF-001 | OPF |
| `invalid/missing_title.epub` | OPF missing dc:title | EPUB-OPF-002 | OPF |
| `invalid/missing_identifier.epub` | OPF missing dc:identifier | EPUB-OPF-003 | OPF |
| `invalid/missing_language.epub` | OPF missing dc:language | EPUB-OPF-004 | OPF |
| `invalid/missing_modified.epub` | OPF missing dcterms:modified | EPUB-OPF-005 | OPF |
| `invalid/missing_nav_document.epub` | OPF manifest missing nav item | EPUB-OPF-009 | OPF |

### Invalid EPUBs - Navigation Errors

| Fixture | Description | Error Code | Category |
|---------|-------------|------------|----------|
| `invalid/invalid_nav_document.epub` | Nav document missing nav element | EPUB-NAV-006 | Navigation |

### Invalid EPUBs - Content Document Errors

| Fixture | Description | Error Code | Category |
|---------|-------------|------------|----------|
| `invalid/invalid_content_document.epub` | Content missing DOCTYPE/namespace | EPUB-CONTENT-002 | Content |

### Edge Cases

| Fixture | Description | Size | Notes |
|---------|-------------|------|-------|
| `edge_cases/large_10mb_plus.epub` | Very large EPUB (2000+ chapters) | ~10+ MB | Performance testing |

## PDF Test Fixtures

### Valid PDFs

| Fixture | Description | Size | Error Codes Tested |
|---------|-------------|------|-------------------|
| `valid/minimal.pdf` | Minimal valid PDF 1.4 | ~500 bytes | None (valid) |
| `valid/with_images.pdf` | PDF with embedded images | ~1 KB | None (valid) |
| `valid/large_100_pages.pdf` | 100 pages for testing | ~50 KB | None (valid) |
| `valid/large_1000_pages.pdf` | 1000 pages for testing | ~500 KB | None (valid) |

### Invalid PDFs - Header Errors

| Fixture | Description | Error Code | Category |
|---------|-------------|------------|----------|
| `invalid/not_pdf.pdf` | Not a PDF file | PDF-HEADER-001 | Header |
| `invalid/no_header.pdf` | Missing PDF header | PDF-HEADER-001 | Header |
| `invalid/invalid_version.pdf` | Invalid version number (2.0) | PDF-HEADER-002 | Header |

### Invalid PDFs - Trailer Errors

| Fixture | Description | Error Code | Category |
|---------|-------------|------------|----------|
| `invalid/no_eof.pdf` | Missing %%EOF marker | PDF-TRAILER-003 | Trailer |
| `invalid/no_startxref.pdf` | Missing startxref | PDF-TRAILER-001 | Trailer |

### Invalid PDFs - Cross-Reference Errors

| Fixture | Description | Error Code | Category |
|---------|-------------|------------|----------|
| `invalid/corrupt_xref.pdf` | Corrupt xref table | PDF-XREF-001 | Cross-reference |

### Invalid PDFs - Catalog Errors

| Fixture | Description | Error Code | Category |
|---------|-------------|------------|----------|
| `invalid/no_catalog.pdf` | Missing catalog object | PDF-CATALOG-001 | Catalog |
| `invalid/invalid_catalog.pdf` | Catalog missing /Pages | PDF-CATALOG-003 | Catalog |

### Invalid PDFs - Corruption Scenarios

| Fixture | Description | Error Code | Category |
|---------|-------------|------------|----------|
| `invalid/corrupt.pdf` | Truncated file | PDF-HEADER-001/Various | Corruption |
| `invalid/truncated_stream.pdf` | Truncated stream object | PDF-STRUCTURE-012 | Corruption |
| `invalid/malformed_objects.pdf` | Malformed object syntax | PDF-STRUCTURE-012 | Corruption |

### Edge Cases

| Fixture | Description | Size | Notes |
|---------|-------------|------|-------|
| `edge_cases/large_10mb_plus.pdf` | Very large PDF (5000+ pages) | ~10+ MB | Performance testing |
| `edge_cases/encrypted.pdf` | PDF with encryption | ~1 KB | Encryption handling |

## Error Code Coverage

### EPUB Error Codes

#### Container (EPUB-CONTAINER-XXX)
- ✅ EPUB-CONTAINER-001: File is not a valid ZIP archive
- ✅ EPUB-CONTAINER-002: Invalid mimetype content
- ✅ EPUB-CONTAINER-003: mimetype file not first in ZIP
- ✅ EPUB-CONTAINER-004: Missing META-INF/container.xml
- ✅ EPUB-CONTAINER-005: Invalid container.xml (malformed XML or no rootfiles)

#### OPF Package (EPUB-OPF-XXX)
- ✅ EPUB-OPF-001: OPF file is not valid XML
- ✅ EPUB-OPF-002: Missing dc:title
- ✅ EPUB-OPF-003: Missing dc:identifier
- ✅ EPUB-OPF-004: Missing dc:language
- ✅ EPUB-OPF-005: Missing dcterms:modified
- ✅ EPUB-OPF-006: Invalid unique-identifier reference (tested via validation)
- ✅ EPUB-OPF-007: Missing manifest (tested via validation)
- ✅ EPUB-OPF-008: Missing spine (tested via validation)
- ✅ EPUB-OPF-009: Missing nav document in manifest
- ✅ EPUB-OPF-010: Invalid manifest item (tested via validation)
- ✅ EPUB-OPF-011: Invalid spine item (tested via validation)

#### Navigation (EPUB-NAV-XXX)
- ✅ EPUB-NAV-001: Not well-formed XHTML (tested via validation)
- ✅ EPUB-NAV-002: Missing TOC nav (tested via validation)
- ✅ EPUB-NAV-006: Missing nav element

#### Content Document (EPUB-CONTENT-XXX)
- ✅ EPUB-CONTENT-001: Not well-formed XHTML (tested via validation)
- ✅ EPUB-CONTENT-002: Missing DOCTYPE
- ✅ EPUB-CONTENT-007: Invalid namespace (tested via validation)

### PDF Error Codes

#### Header (PDF-HEADER-XXX)
- ✅ PDF-HEADER-001: Invalid or missing PDF header
- ✅ PDF-HEADER-002: Invalid PDF version number

#### Trailer (PDF-TRAILER-XXX)
- ✅ PDF-TRAILER-001: Invalid or missing startxref
- ✅ PDF-TRAILER-002: Invalid trailer dictionary
- ✅ PDF-TRAILER-003: Missing %%EOF marker

#### Cross-Reference (PDF-XREF-XXX)
- ✅ PDF-XREF-001: Invalid or damaged cross-reference table
- ✅ PDF-XREF-002: Empty cross-reference table (tested via validation)
- ✅ PDF-XREF-003: Overlapping xref entries (tested via validation)

#### Catalog (PDF-CATALOG-XXX)
- ✅ PDF-CATALOG-001: Missing or invalid catalog object
- ✅ PDF-CATALOG-002: Catalog missing /Type entry (tested via validation)
- ✅ PDF-CATALOG-003: Catalog missing /Pages entry

#### Structure (PDF-STRUCTURE-XXX)
- ✅ PDF-STRUCTURE-012: General structure parsing errors

## Usage in Tests

### Integration Tests

Integration tests are located in `tests/integration/`:

```go
// EPUB validation
func TestEPUBValidatorIntegration_TableDriven_AllErrorCodes(t *testing.T) {
    // Systematically tests all EPUB error codes
}

// PDF validation
func TestPDFValidatorIntegration_TableDriven_AllErrorCodes(t *testing.T) {
    // Systematically tests all PDF error codes
}
```

### Running Tests

```bash
# Run all integration tests
make test-integration

# Run only EPUB integration tests
go test -v ./tests/integration -run TestEPUB

# Run only PDF integration tests
go test -v ./tests/integration -run TestPDF

# Run all tests (includes auto-generation of fixtures)
make test
```

## Performance Testing

Large file fixtures are provided for performance benchmarking:

### EPUB
- `large_100_chapters.epub`: ~500 KB, tests mid-size document handling
- `large_500_chapters.epub`: ~2.5 MB, tests large document handling
- `large_10mb_plus.epub`: ~10+ MB, stress testing

### PDF
- `large_100_pages.pdf`: ~50 KB, tests mid-size document handling
- `large_1000_pages.pdf`: ~500 KB, tests large document handling
- `large_10mb_plus.pdf`: ~10+ MB, stress testing

Performance tests validate that:
- Files > 10 MB can be validated without errors
- Validation completes in reasonable time
- Memory usage remains acceptable

## Adding New Fixtures

To add a new test fixture:

1. Edit the appropriate `generate_fixtures.go` file
2. Add a function to create the fixture (e.g., `createEPUBNewCase()`)
3. Add an entry to the `fixtures` map in `main()`
4. Run the generator: `go run generate_fixtures.go`
5. Update this README with fixture documentation
6. Add corresponding test cases in `tests/integration/`

## Notes

- All fixtures are programmatically generated for consistency and reproducibility
- Fixtures are NOT committed to the repository (too large)
- The test suite auto-generates fixtures if they don't exist
- Fixture generation is idempotent and can be run repeatedly
- Binary fixtures use minimal valid structures to keep sizes small
- Large files use repeated content patterns to achieve target sizes efficiently

## Validation Oracle

For EPUB validation, consider cross-validation with epubcheck when available:

```bash
# Install epubcheck (if available)
# Compare results with our validator
epubcheck testdata/epub/valid/minimal.epub
```

This helps ensure our validation aligns with industry-standard tools.
