# EPUB Adapter Package

This package provides EPUB validation adapters implementing the EPUB Open Container Format (OCF) specification.

## Structure

```
internal/adapters/epub/
├── README.md                    # This file
├── DOC.md                       # Detailed API documentation
├── ERROR_CODES.md               # Complete error code reference
├── container_validator.go       # Container validation implementation
├── container_validator_test.go  # Unit tests with fixtures
├── content_validator.go         # Content document validation
├── content_validator_test.go    # Content validation tests
├── opf_validator.go             # OPF package document validation
├── opf_validator_test.go        # OPF validation tests
├── nav_validator.go             # Navigation document validation
├── nav_validator_test.go        # Navigation validation tests
└── integration_test.go          # Integration tests
```

## Components

### Container Validator

Implements OCF 3.0 specification section 3.1 container validation:

- ✅ ZIP archive validation
- ✅ Mimetype file validation (first, uncompressed, correct content)
- ✅ META-INF/container.xml validation
- ✅ Rootfile path extraction

**Files:**
- `container_validator.go` - Implementation
- `container_validator_test.go` - Comprehensive unit tests

### Content Validator

Implements EPUB content document validation:

- ✅ XHTML well-formedness validation
- ✅ DOCTYPE validation
- ✅ Required elements (html, head, body)
- ✅ XHTML namespace validation

**Files:**
- `content_validator.go` - Implementation
- `content_validator_test.go` - Comprehensive unit tests

### OPF Validator

Implements EPUB package document (OPF) validation:

- ✅ OPF XML structure validation
- ✅ Required metadata validation
- ✅ Manifest validation
- ✅ Spine validation
- ✅ Navigation document reference validation

**Files:**
- `opf_validator.go` - Implementation
- `opf_validator_test.go` - Comprehensive unit tests

### Navigation Validator

Implements EPUB 3 navigation document validation:

- ✅ Navigation document well-formedness
- ✅ Required TOC validation (`<nav epub:type="toc">`)
- ✅ Optional landmarks validation
- ✅ Nested `<ol>` structure validation
- ✅ Relative link validation
- ✅ Link extraction

**Files:**
- `nav_validator.go` - Implementation
- `nav_validator_test.go` - Comprehensive unit tests

### Integration Tests

**Files:**
- `integration_test.go` - Integration tests with fixtures

## Quick Start

### Container Validation

```go
import "github.com/example/project/internal/adapters/epub"

validator := epub.NewContainerValidator()
result, err := validator.ValidateFile("path/to/book.epub")

if err != nil {
    // Handle I/O error
    log.Fatal(err)
}

if !result.Valid {
    // Handle validation errors
    for _, e := range result.Errors {
        fmt.Printf("[%s] %s\n", e.Code, e.Message)
    }
}

// Use extracted rootfiles
for _, rf := range result.Rootfiles {
    fmt.Printf("Rootfile: %s\n", rf.FullPath)
}
```

### Navigation Validation

```go
import "github.com/example/project/internal/adapters/epub"

validator := epub.NewNavValidator()
result, err := validator.ValidateFile("path/to/nav.xhtml")

if err != nil {
    log.Fatal(err)
}

if !result.Valid {
    for _, e := range result.Errors {
        fmt.Printf("[%s] %s\n", e.Code, e.Message)
    }
}

// Access navigation links
for _, link := range result.TOCLinks {
    fmt.Printf("TOC: %s -> %s\n", link.Text, link.Href)
}
```

## Error Codes

### Container Validation (`EPUB-CONTAINER-XXX`)

| Code | Description |
|------|-------------|
| 001 | ZIP Invalid |
| 002 | Mimetype Invalid |
| 003 | Mimetype Not First |
| 004 | Container XML Missing |
| 005 | Container XML Invalid |

### Navigation Validation (`EPUB-NAV-XXX`)

| Code | Description |
|------|-------------|
| 001 | Not Well-Formed |
| 002 | Missing TOC |
| 003 | Invalid TOC Structure |
| 004 | Invalid Links |
| 005 | Invalid Landmarks |
| 006 | Missing Nav Element |

See [ERROR_CODES.md](ERROR_CODES.md) for complete details.

## Testing

Run unit tests:
```bash
go test ./internal/adapters/epub/
```

Run with coverage:
```bash
go test -cover ./internal/adapters/epub/
```

Run integration tests:
```bash
# Generate fixtures first
go run testdata/epub/generate_fixtures.go testdata/epub/

# Run integration tests
go test ./internal/adapters/epub/ -run Integration
```

## Test Coverage

The test suite includes:

- ✅ Valid EPUB (single rootfile)
- ✅ Valid EPUB (multiple rootfiles)
- ✅ Invalid ZIP archive
- ✅ Wrong mimetype content
- ✅ Compressed mimetype
- ✅ Mimetype not first
- ✅ Missing container.xml
- ✅ Invalid container.xml XML
- ✅ No rootfiles declared
- ✅ Empty rootfile paths
- ✅ File and byte validation
- ✅ Error code verification

All test fixtures are generated programmatically for consistency and maintainability.

## Architecture

The implementation follows the hexagonal architecture pattern:

```
Domain Layer (ports)
       ↓
Adapter Layer (epub)
       ↓
External (file system, ZIP)
```

The validator is:
- **Stateless**: No internal state, thread-safe
- **Testable**: Pure functions with dependency injection
- **Spec-compliant**: Implements OCF 3.0 section 3.1
- **Error-accumulating**: Reports all errors, not just the first

## Documentation

- [DOC.md](DOC.md) - Complete API documentation with examples
- [ERROR_CODES.md](ERROR_CODES.md) - Detailed error code reference
- [docs/testdata/epub/README.md](../../testdata/epub/README.md) - Test fixture documentation

## Future Enhancements

Potential additions (not yet implemented):

- EPUB 2.0 vs 3.0 version detection
- Media type validation
- Fallback chain validation
- Encryption.xml handling
- Signature validation
- Cross-reference validation between OPF and actual files

## Contributing

When adding new validation rules:

1. Add appropriate error code constant
2. Update ERROR_CODES.md with new code
3. Implement validation logic
4. Add comprehensive unit tests
5. Update DOC.md if API changes
