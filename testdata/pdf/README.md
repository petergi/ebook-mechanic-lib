# PDF Test Fixtures

This directory contains test PDF files used for validation testing.

## Test Files

### Valid PDFs

- **valid_minimal.pdf**: Minimal valid PDF 1.4 with single page
- **valid_v1.0.pdf** through **valid_v1.7.pdf**: Valid PDFs for each supported version

### Invalid PDFs - Header Issues

- **invalid_header_missing.pdf**: Missing PDF header
- **invalid_header_wrong.pdf**: Wrong header format (not %PDF-)
- **invalid_version_too_high.pdf**: Version 1.9 (unsupported)
- **invalid_version_too_low.pdf**: Version 0.9 (invalid)

### Invalid PDFs - Trailer Issues

- **truncated_no_eof.pdf**: Missing %%EOF marker
- **truncated_mid_file.pdf**: File truncated in middle
- **missing_startxref.pdf**: Missing startxref keyword
- **invalid_startxref_value.pdf**: startxref with invalid offset

### Invalid PDFs - Cross-Reference Issues

- **damaged_xref_table.pdf**: Corrupted xref table
- **empty_xref_table.pdf**: Xref table with no entries
- **overlapping_xref.pdf**: Multiple objects at same offset

### Invalid PDFs - Catalog Issues

- **missing_catalog.pdf**: No catalog object
- **invalid_catalog_type.pdf**: Catalog with wrong /Type
- **missing_catalog_type.pdf**: Catalog missing /Type entry
- **missing_pages.pdf**: Catalog missing /Pages entry

## Generating Test Files

Test files are generated programmatically in the test code to ensure consistency and avoid binary file storage in the repository. The helper functions in `structure_validator_test.go` create the necessary test PDFs on-demand.

## Usage in Tests

```go
// Example: Testing with truncated file
func TestTruncatedFile(t *testing.T) {
    data := createTruncatedPDF()
    validator := NewStructureValidator()
    result, err := validator.ValidateBytes(data)
    // Assertions...
}
```

## Test Coverage

The test fixtures cover all error codes:

- PDF-HEADER-001: Invalid/missing header
- PDF-HEADER-002: Invalid version
- PDF-TRAILER-001: Invalid startxref
- PDF-TRAILER-002: Invalid trailer dictionary
- PDF-TRAILER-003: Missing EOF
- PDF-XREF-001: Damaged xref
- PDF-XREF-002: Empty xref
- PDF-XREF-003: Overlapping xref entries
- PDF-CATALOG-001: Missing/invalid catalog
- PDF-CATALOG-002: Wrong catalog type
- PDF-CATALOG-003: Missing pages
- PDF-STRUCTURE-012: General structure errors
