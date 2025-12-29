# EPUB Validators

Implementation of EPUB validation according to the EPUB specifications.

## Features

### Container Validation
- **ZIP Format Validation**: Verifies the file is a valid ZIP archive
- **Mimetype File Validation**: 
  - Must be the first file in the ZIP archive
  - Must be stored uncompressed (using Store method)
  - Must contain exactly "application/epub+zip"
- **Container XML Validation**:
  - Checks for existence of META-INF/container.xml
  - Validates XML structure
  - Ensures at least one rootfile is declared
  - Validates rootfile paths are not empty
- **Rootfile Path Extraction**: Returns all rootfile paths from container.xml

## Error Codes

| Code | Description |
|------|-------------|
| `EPUB-CONTAINER-001` | File is not a valid ZIP archive |
| `EPUB-CONTAINER-002` | Mimetype file has incorrect content or is compressed |
| `EPUB-CONTAINER-003` | Mimetype file is not the first entry in the ZIP |
| `EPUB-CONTAINER-004` | META-INF/container.xml is missing |
| `EPUB-CONTAINER-005` | META-INF/container.xml is malformed or invalid |

## Usage

```go
import "github.com/example/project/internal/adapters/epub"

validator := epub.NewContainerValidator()

// Validate from file path
result, err := validator.ValidateFile("path/to/book.epub")
if err != nil {
    // Handle I/O errors
}

// Validate from byte slice
result, err := validator.ValidateBytes(epubData)
if err != nil {
    // Handle validation errors
}

// Check results
if !result.Valid {
    for _, validationError := range result.Errors {
        fmt.Printf("Error [%s]: %s\n", validationError.Code, validationError.Message)
        fmt.Printf("Details: %v\n", validationError.Details)
    }
}

// Extract rootfile paths
for _, rootfile := range result.Rootfiles {
    fmt.Printf("Rootfile: %s (%s)\n", rootfile.FullPath, rootfile.MediaType)
}
```

## API

### Types

#### `ContainerValidator`
Main validator struct with methods for validating EPUB containers.

#### `ValidationResult`
Contains validation results:
- `Valid` (bool): Overall validation status
- `Errors` ([]ValidationError): List of validation errors
- `Rootfiles` ([]Rootfile): Extracted rootfile entries

#### `ValidationError`
Represents a single validation error:
- `Code` (string): Error code (e.g., "EPUB-CONTAINER-001")
- `Message` (string): Human-readable error message
- `Details` (map[string]interface{}): Additional error details

#### `Rootfile`
Represents a rootfile entry from container.xml:
- `FullPath` (string): Path to the rootfile within the EPUB
- `MediaType` (string): Media type of the rootfile

### Methods

#### `NewContainerValidator() *ContainerValidator`
Creates a new container validator instance.

#### `ValidateFile(filePath string) (*ValidationResult, error)`
Validates an EPUB file from a file path. Returns an error only for I/O issues.

#### `Validate(reader io.ReaderAt, size int64) (*ValidationResult, error)`
Validates an EPUB container from a reader. Returns an error only for I/O issues.

#### `ValidateBytes(data []byte) (*ValidationResult, error)`
Validates an EPUB container from a byte slice. Returns an error only for I/O issues.

## Test Coverage

The implementation includes comprehensive unit tests covering:

1. ✅ Valid EPUB with single rootfile
2. ✅ Valid EPUB with multiple rootfiles
3. ✅ Invalid ZIP archive
4. ✅ Wrong mimetype content
5. ✅ Compressed mimetype file
6. ✅ Mimetype not first in archive
7. ✅ Missing META-INF/container.xml
8. ✅ Invalid XML in container.xml
9. ✅ Container.xml with no rootfiles
10. ✅ Container.xml with empty rootfile path
11. ✅ File path validation
12. ✅ Non-existent file handling
13. ✅ Error code constants

All tests use programmatically generated EPUB fixtures to ensure correctness and maintainability.

## OCF Specification Compliance

This implementation follows the EPUB Open Container Format (OCF) specification section 3.1:

- The mimetype file MUST be the first file in the ZIP archive
- The mimetype file MUST be uncompressed (stored)
- The mimetype file MUST NOT have an extra field in its ZIP header
- The mimetype file MUST contain the ASCII string "application/epub+zip" with no trailing padding
- The META-INF/container.xml file MUST exist
- The container.xml MUST contain at least one rootfile element
- Each rootfile element MUST have a full-path attribute

## Implementation Notes

- The validator uses Go's standard `archive/zip` package for ZIP handling
- XML parsing uses Go's standard `encoding/xml` package
- The validator is designed to be stateless and thread-safe
- Validation errors are accumulated rather than failing fast, providing complete feedback
- The API distinguishes between I/O errors (returned as `error`) and validation errors (in `ValidationResult`)

---

# EPUB Navigation Document Validator

Implementation of navigation document validation for EPUB 3 files according to the EPUB 3 specification.

## Features

### Navigation Validation
- **Well-Formedness Check**: Verifies the navigation document is valid XHTML
- **TOC Validation**:
  - Ensures presence of `<nav epub:type="toc">` element
  - Validates nested `<ol>` structure within TOC
  - Extracts all TOC links with their text
- **Landmarks Validation** (optional):
  - Validates `<nav epub:type="landmarks">` element if present
  - Ensures nested `<ol>` structure within landmarks
  - Extracts all landmark links
- **Relative Link Validation**:
  - Ensures all links are relative (no absolute URLs)
  - Rejects protocol-relative URLs (`//example.com`)
  - Rejects absolute paths (`/path/to/file`)
  - Rejects parent directory references (`../`)
  - Allows fragment-only links (`#section1`)
  - Allows subdirectory paths (`content/chapter1.xhtml`)

## Error Codes

| Code | Description |
|------|-------------|
| `EPUB-NAV-001` | Navigation document is not well-formed XHTML |
| `EPUB-NAV-002` | Missing required `<nav epub:type="toc">` element |
| `EPUB-NAV-003` | TOC navigation element missing required `<ol>` structure |
| `EPUB-NAV-004` | Navigation contains invalid relative links |
| `EPUB-NAV-005` | Landmarks navigation element missing required `<ol>` structure |
| `EPUB-NAV-006` | Navigation document missing any `<nav>` elements |

## Usage

```go
import "github.com/example/project/internal/adapters/epub"

validator := epub.NewNavValidator()

// Validate from file path
result, err := validator.ValidateFile("path/to/nav.xhtml")
if err != nil {
    // Handle I/O errors
}

// Validate from byte slice
result, err := validator.ValidateBytes(navData)
if err != nil {
    // Handle validation errors
}

// Check results
if !result.Valid {
    for _, validationError := range result.Errors {
        fmt.Printf("Error [%s]: %s\n", validationError.Code, validationError.Message)
        fmt.Printf("Details: %v\n", validationError.Details)
    }
}

// Access extracted navigation data
fmt.Printf("Has TOC: %v\n", result.HasTOC)
fmt.Printf("Has Landmarks: %v\n", result.HasLandmarks)

for _, link := range result.TOCLinks {
    fmt.Printf("TOC Link: %s -> %s\n", link.Text, link.Href)
}

for _, link := range result.LandmarkLinks {
    fmt.Printf("Landmark: %s -> %s\n", link.Text, link.Href)
}
```

## API

### Types

#### `NavValidator`
Main validator struct with methods for validating EPUB navigation documents.

#### `NavValidationResult`
Contains validation results:
- `Valid` (bool): Overall validation status
- `Errors` ([]ValidationError): List of validation errors
- `HasTOC` (bool): Whether a TOC navigation element was found
- `HasLandmarks` (bool): Whether a landmarks navigation element was found
- `TOCLinks` ([]NavLink): Extracted TOC links
- `LandmarkLinks` ([]NavLink): Extracted landmark links

#### `NavLink`
Represents a navigation link:
- `Href` (string): Link target (relative path)
- `Text` (string): Link text content

#### `ValidationError`
Represents a single validation error:
- `Code` (string): Error code (e.g., "EPUB-NAV-001")
- `Message` (string): Human-readable error message
- `Details` (map[string]interface{}): Additional error details

### Methods

#### `NewNavValidator() *NavValidator`
Creates a new navigation validator instance.

#### `ValidateFile(filePath string) (*NavValidationResult, error)`
Validates a navigation document from a file path. Returns an error only for I/O issues.

#### `Validate(reader io.Reader) (*NavValidationResult, error)`
Validates a navigation document from a reader. Returns an error only for I/O issues.

#### `ValidateBytes(data []byte) (*NavValidationResult, error)`
Validates a navigation document from a byte slice. Returns an error only for I/O issues.

## Test Coverage

The implementation includes comprehensive unit tests covering:

1. ✅ Valid navigation with TOC
2. ✅ Valid navigation with TOC and landmarks
3. ✅ Missing TOC (error)
4. ✅ Missing `<ol>` in TOC (error)
5. ✅ Invalid links (absolute URLs, protocol-relative, absolute paths)
6. ✅ Malformed XHTML
7. ✅ Missing `<nav>` element
8. ✅ Empty links (error)
9. ✅ Protocol-relative URLs (error)
10. ✅ Landmarks missing `<ol>` (error)
11. ✅ Complex nested navigation structure
12. ✅ File path validation
13. ✅ Non-existent file handling
14. ✅ Empty content handling
15. ✅ Link validation edge cases
16. ✅ Error code constants

All tests use programmatically generated XHTML fixtures to ensure correctness and maintainability.

## EPUB 3 Specification Compliance

This implementation follows the EPUB 3 specification for navigation documents:

- The navigation document MUST contain at least one `<nav>` element
- At least one `<nav>` element MUST have `epub:type="toc"`
- TOC and landmarks `<nav>` elements MUST contain an ordered list (`<ol>`)
- Navigation links MUST be relative within the EPUB package
- Landmarks are optional but validated if present

## Implementation Notes

- The validator uses `golang.org/x/net/html` for HTML parsing
- HTML parser is lenient with minor formatting issues but strict on structure
- The validator extracts and validates all links, including nested structures
- Link validation ensures EPUB portability by rejecting external references
- The validator is designed to be stateless and thread-safe
- Validation errors are accumulated rather than failing fast, providing complete feedback
- The API distinguishes between I/O errors (returned as `error`) and validation errors (in `NavValidationResult`)
