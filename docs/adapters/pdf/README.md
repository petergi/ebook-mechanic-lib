# PDF Adapter

This package implements PDF validation functionality using the hexagonal architecture pattern. It provides comprehensive validation of PDF structure and well-formedness according to PDF 1.7 specifications (ISO 32000-1:2008).

## Components

### Structure Validator

**File:** `structure_validator.go`

Validates the basic well-formedness of PDF files, including:

- **Header Validation**: Checks for valid `%PDF-1.x` header where x is 0-7
- **Trailer Validation**: Ensures presence of `%%EOF` marker and valid `startxref`
- **Cross-Reference Validation**: Validates xref table/stream integrity and checks for overlapping entries
- **Catalog Validation**: Verifies catalog object exists with `/Type /Catalog` and `/Pages` entry
- **Object Numbering**: Ensures no duplicate object number/generation pairs

### Error Codes

All validation errors follow a structured format with specific error codes from `PDF-HEADER-001` through `PDF-STRUCTURE-012`. See `ERROR_CODES.md` for complete documentation.

## Architecture

This adapter follows the hexagonal architecture pattern:

```
┌─────────────────────────────────────┐
│         Ports (Interfaces)          │
│    internal/ports/validator.go      │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│           Adapters                  │
│   internal/adapters/pdf/            │
│   - structure_validator.go          │
└──────────────┬──────────────────────┘
               │
               ▼
┌─────────────────────────────────────┐
│        External Library             │
│    github.com/unidoc/unipdf/v3      │
└─────────────────────────────────────┘
```

## Usage

### Basic Validation

```go
import "github.com/petergi/ebook-mechanic-lib/internal/adapters/pdf"

validator := pdf.NewStructureValidator()
result, err := validator.ValidateFile("document.pdf")

if err != nil {
    // Handle I/O error
    log.Fatal(err)
}

if !result.Valid {
    for _, validationError := range result.Errors {
        fmt.Printf("[%s] %s\n", validationError.Code, validationError.Message)
        if validationError.Details != nil {
            fmt.Printf("Details: %+v\n", validationError.Details)
        }
    }
}
```

### Validating from Reader

```go
file, err := os.Open("document.pdf")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

validator := pdf.NewStructureValidator()
result, err := validator.ValidateReader(file)
// Handle result...
```

### Validating Bytes

```go
data, err := ioutil.ReadFile("document.pdf")
if err != nil {
    log.Fatal(err)
}

validator := pdf.NewStructureValidator()
result, err := validator.ValidateBytes(data)
// Handle result...
```

## Testing

The package includes comprehensive tests covering:

- Valid PDF files (all versions 1.0-1.7)
- Invalid headers and version numbers
- Missing EOF markers
- Truncated files
- Missing or invalid startxref
- Damaged cross-reference tables
- Missing or invalid catalog objects
- Missing catalog type or pages entries
- Empty files
- Multiple simultaneous errors

Run tests with:
```bash
go test ./internal/adapters/pdf/...
```

## Error Code Categories

| Code Range | Category | Description |
|------------|----------|-------------|
| PDF-HEADER-001 to PDF-HEADER-002 | Header | PDF file header issues |
| PDF-TRAILER-001 to PDF-TRAILER-003 | Trailer | File trailer and EOF issues |
| PDF-XREF-001 to PDF-XREF-003 | Cross-Reference | Xref table/stream issues |
| PDF-CATALOG-001 to PDF-CATALOG-003 | Catalog | Document catalog issues |
| PDF-STRUCTURE-012 | General | Other structural issues |

## Dependencies

- **unipdf v3**: PDF parsing and object model
  - Used for robust PDF structure parsing
  - Handles cross-reference tables and streams
  - Provides access to catalog and object tree

## Validation Workflow

1. **Pre-Parse Validation**
   - Check file size and basic structure
   - Validate header format and version
   - Check for EOF marker and startxref

2. **Structure Parsing**
   - Parse PDF using unipdf
   - Extract cross-reference table
   - Locate catalog object

3. **Deep Validation**
   - Validate cross-reference integrity
   - Check catalog structure
   - Verify object numbering

4. **Result Aggregation**
   - Collect all validation errors
   - Determine overall validity
   - Return structured result

## Future Enhancements

- PDF/A validation (archival conformance)
- PDF/UA validation (accessibility)
- Font embedding validation
- Color space validation
- Metadata validation
- Encryption and security validation
- Incremental update validation
- Linearization validation

## References

- ISO 32000-1:2008 (PDF 1.7)
- Adobe PDF Reference, version 1.7
- ebm-lib PDF Specification: `docs/specs/ebm-lib-PDF-SPEC.md`
- veraPDF validation rules
