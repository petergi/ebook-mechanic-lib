# EPUB Container Validator

Implementation of OCF (Open Container Format) validation for EPUB files according to the EPUB specification section 3.1.

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
