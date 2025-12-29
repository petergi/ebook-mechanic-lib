# PDF Test Fixtures

This directory contains comprehensive PDF test fixtures covering all validation error codes, corruption scenarios, and edge cases.

## Quick Start

Generate all fixtures:

```bash
go run generate_fixtures.go
```

This creates the following directory structure:

```
pdf/
├── valid/
│   ├── minimal.pdf
│   ├── with_images.pdf
│   ├── large_100_pages.pdf
│   └── large_1000_pages.pdf
├── invalid/
│   ├── not_pdf.pdf
│   ├── no_header.pdf
│   ├── invalid_version.pdf
│   ├── no_eof.pdf
│   ├── no_startxref.pdf
│   ├── corrupt_xref.pdf
│   ├── no_catalog.pdf
│   ├── invalid_catalog.pdf
│   ├── corrupt.pdf
│   ├── truncated_stream.pdf
│   └── malformed_objects.pdf
└── edge_cases/
    ├── large_10mb_plus.pdf
    └── encrypted.pdf
```

## Fixture Details

### Valid PDFs

#### minimal.pdf
- **Purpose**: Baseline valid PDF 1.4
- **Structure**: Single page with "Hello World" text
- **Size**: ~410 bytes
- **Contents**:
  - Proper %PDF-1.4 header
  - Catalog object
  - Pages object
  - Single Page object
  - Content stream with text
  - Complete xref table
  - Trailer with startxref
  - %%EOF marker

#### with_images.pdf
- **Purpose**: Test PDF with embedded resources
- **Structure**: Single page with embedded image (XObject)
- **Size**: ~600 bytes
- **Contents**:
  - Image as XObject in resources
  - Content stream with image display operator
  - Complete structure with xref

#### large_100_pages.pdf
- **Purpose**: Performance testing - medium size
- **Structure**: 100 pages, each with unique text
- **Size**: ~50 KB
- **Use**: Benchmark validation speed

#### large_1000_pages.pdf
- **Purpose**: Performance testing - large size
- **Structure**: 1000 pages, each with unique text
- **Size**: ~500 KB
- **Use**: Test scalability

### Invalid PDFs

Each invalid fixture targets specific error codes:

#### Header Errors (PDF-HEADER-XXX)

**not_pdf.pdf**
- **Error**: PDF-HEADER-001
- **Issue**: Plain text file "This is not a PDF file..."
- **Detection**: Missing %PDF- signature

**no_header.pdf**
- **Error**: PDF-HEADER-001
- **Issue**: File starts with "This is not a PDF file" instead of %PDF-
- **Detection**: Header pattern matching fails

**invalid_version.pdf**
- **Error**: PDF-HEADER-002
- **Issue**: Header contains %PDF-2.0 (unsupported version)
- **Detection**: Version validation (only 1.0-1.7 valid)

#### Trailer Errors (PDF-TRAILER-XXX)

**no_eof.pdf**
- **Error**: PDF-TRAILER-003
- **Issue**: File ends without %%EOF marker
- **Detection**: EOF marker search fails

**no_startxref.pdf**
- **Error**: PDF-TRAILER-001
- **Issue**: Missing "startxref" keyword before %%EOF
- **Detection**: startxref pattern matching fails

#### Cross-Reference Errors (PDF-XREF-XXX)

**corrupt_xref.pdf**
- **Error**: PDF-XREF-001
- **Issue**: xref table replaced with "this is not a valid xref table"
- **Detection**: xref parsing error

#### Catalog Errors (PDF-CATALOG-XXX)

**no_catalog.pdf**
- **Error**: PDF-CATALOG-001
- **Issue**: Trailer missing /Root entry
- **Detection**: Catalog object not found in trailer

**invalid_catalog.pdf**
- **Error**: PDF-CATALOG-003
- **Issue**: Catalog object missing required /Pages entry
- **Detection**: Pages dictionary not found in catalog

#### Corruption Scenarios

**corrupt.pdf**
- **Error**: Various (PDF-HEADER-001 or structure errors)
- **Issue**: Valid PDF truncated by 100 bytes
- **Detection**: Structural parsing fails

**truncated_stream.pdf**
- **Error**: PDF-STRUCTURE-012
- **Issue**: Content stream declares Length 100 but contains only ~40 bytes
- **Detection**: Stream length mismatch or parsing error

**malformed_objects.pdf**
- **Error**: PDF-STRUCTURE-012
- **Issue**: Object dictionary has unclosed bracket: `/Kids [3 0 R` (missing `]`)
- **Detection**: Object syntax parsing error

### Edge Cases

**edge_cases/large_10mb_plus.pdf**
- **Purpose**: Stress testing with very large file
- **Structure**: 5000+ pages
- **Size**: ~10+ MB
- **Use**: Memory and performance limits testing

**edge_cases/encrypted.pdf**
- **Purpose**: Test handling of encrypted PDFs
- **Structure**: PDF with /Encrypt dictionary in trailer
- **Size**: ~300 bytes
- **Note**: Contains standard encryption dictionary; actual encryption not implemented

## Test Coverage Matrix

| Error Code | Fixture | Test Status |
|------------|---------|-------------|
| PDF-HEADER-001 | not_pdf.pdf, no_header.pdf | ✅ |
| PDF-HEADER-002 | invalid_version.pdf | ✅ |
| PDF-TRAILER-001 | no_startxref.pdf | ✅ |
| PDF-TRAILER-003 | no_eof.pdf | ✅ |
| PDF-XREF-001 | corrupt_xref.pdf | ✅ |
| PDF-CATALOG-001 | no_catalog.pdf | ✅ |
| PDF-CATALOG-003 | invalid_catalog.pdf | ✅ |
| PDF-STRUCTURE-012 | corrupt.pdf, truncated_stream.pdf, malformed_objects.pdf | ✅ |

## Integration Testing

Tests are located in `../../tests/integration/pdf_validator_integration_test.go`:

```go
// Table-driven test covering all error codes
func TestPDFValidatorIntegration_TableDriven_AllErrorCodes(t *testing.T)

// Valid file tests
func TestPDFValidatorIntegration_ValidFiles(t *testing.T)

// Performance tests
func TestPDFValidatorIntegration_PerformanceLargeFiles(t *testing.T)

// Corruption scenario tests
func TestPDFValidatorIntegration_CorruptionScenarios(t *testing.T)

// Systematic coverage tests
func TestPDFValidatorIntegration_Systematic_Coverage(t *testing.T)
```

## PDF Specification Compliance

Fixtures are designed to comply with or intentionally violate:

- **PDF 1.4 Specification** (ISO 32000-1:2008 compatible)
- **PDF 1.7 Specification** (ISO 32000-1:2008)

Key validation points:

### File Structure
- Header: `%PDF-1.x` where x is 0-7
- Body: Objects (dictionaries, streams, arrays, etc.)
- Cross-reference table: Object locations
- Trailer: Document catalog and metadata
- EOF marker: `%%EOF`

### Required Objects
- **Catalog**: Document root, must have /Type /Catalog and /Pages
- **Pages**: Page tree root, must have /Type /Pages and /Kids array
- **Page**: Individual pages, must have /Type /Page, /Parent, and /MediaBox

### Validation Rules
- Header must be first line
- Version must be 1.0 through 1.7
- Cross-reference table must be valid and complete
- Trailer must contain /Root pointing to Catalog
- Catalog must contain /Pages pointing to Pages tree
- %%EOF must be present at end of file
- startxref must point to xref table location

## Understanding PDF Structure

### Minimal Valid PDF Anatomy

```
%PDF-1.4                          ← Header (must be first)
1 0 obj                           ← Catalog object
<<
/Type /Catalog
/Pages 2 0 R
>>
endobj
2 0 obj                           ← Pages object
<<
/Type /Pages
/Kids [3 0 R]
/Count 1
>>
endobj
3 0 obj                           ← Page object
<<
/Type /Page
/Parent 2 0 R
/MediaBox [0 0 612 792]
/Contents 4 0 R
>>
endobj
4 0 obj                           ← Content stream
<<
/Length 44
>>
stream
BT
/F1 12 Tf
100 700 Td
(Hello World) Tj
ET
endstream
endobj
xref                              ← Cross-reference table
0 5
0000000000 65535 f 
0000000009 00000 n 
0000000058 00000 n 
0000000115 00000 n 
0000000317 00000 n 
trailer                           ← Trailer
<<
/Size 5
/Root 1 0 R
>>
startxref                         ← Pointer to xref
410
%%EOF                             ← End of file marker
```

## Extending Fixtures

To add a new fixture:

1. Create a function in `generate_fixtures.go`:
   ```go
   func createPDFNewCase() []byte {
       pdf := `%PDF-1.4
   1 0 obj
   << /Type /Catalog >>
   endobj
   ...
   %%EOF
   `
       return []byte(pdf)
   }
   ```

2. Add to the fixtures map:
   ```go
   fixtures := map[string][]byte{
       // ...
       "invalid/new_case.pdf": createPDFNewCase(),
   }
   ```

3. Regenerate: `go run generate_fixtures.go`

4. Add test case in `../../tests/integration/pdf_validator_integration_test.go`

## Corruption Scenarios Explained

### truncated_stream.pdf
Simulates incomplete file transfer or storage corruption where stream data is cut short.

### malformed_objects.pdf
Simulates parser-level errors with invalid PDF syntax (unclosed dictionaries, etc.).

### corrupt.pdf
Simulates file truncation, testing recovery from incomplete data.

### corrupt_xref.pdf
Simulates damage to the critical cross-reference table, making object lookup impossible.

## Performance Expectations

Based on the test fixtures:

| File Size | Pages | Expected Validation Time |
|-----------|-------|-------------------------|
| < 1 KB | 1 | < 10 ms |
| ~50 KB | 100 | < 100 ms |
| ~500 KB | 1000 | < 500 ms |
| ~10 MB | 5000 | < 2000 ms |

Note: Actual times depend on hardware and validator implementation.

## Validation with External Tools

To cross-validate fixtures:

```bash
# Using QPDF
qpdf --check valid/minimal.pdf

# Using PDF Toolkit
pdftk valid/minimal.pdf dump_data

# Using Poppler utils
pdfinfo valid/minimal.pdf
```

Compare results with our validator to ensure comprehensive coverage.
