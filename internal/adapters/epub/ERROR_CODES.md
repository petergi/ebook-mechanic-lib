# EPUB Container Validation Error Codes

This document provides a complete reference for all error codes used by the EPUB container validator, aligned with EPUB OCF specification section 3.1.

## Error Code Reference

### EPUB-CONTAINER-001: ZIP Invalid

**Severity:** Error  
**Description:** The file is not a valid ZIP archive.

**Common Causes:**
- File is corrupted
- File is not in ZIP format
- File has been truncated or incomplete

**Example:**
```json
{
  "code": "EPUB-CONTAINER-001",
  "message": "File is not a valid ZIP archive",
  "details": {
    "error": "zip: not a valid zip file"
  }
}
```

**Resolution:** Ensure the file is a properly formatted EPUB/ZIP file.

---

### EPUB-CONTAINER-002: Mimetype Invalid

**Severity:** Error  
**Description:** The mimetype file has incorrect content or compression.

**Common Causes:**
- Mimetype file contains wrong content (not "application/epub+zip")
- Mimetype file is compressed (must be stored uncompressed)
- Mimetype file is missing or empty

**Examples:**

Wrong content:
```json
{
  "code": "EPUB-CONTAINER-002",
  "message": "mimetype file must contain exactly 'application/epub+zip'",
  "details": {
    "expected": "application/epub+zip",
    "found": "application/wrong"
  }
}
```

Compressed mimetype:
```json
{
  "code": "EPUB-CONTAINER-002",
  "message": "mimetype file must be stored uncompressed",
  "details": {
    "compression_method": 8
  }
}
```

**Resolution:** 
- Ensure mimetype file contains exactly "application/epub+zip" with no extra whitespace
- Ensure mimetype file is stored with no compression (ZIP Store method)

---

### EPUB-CONTAINER-003: Mimetype Not First

**Severity:** Error  
**Description:** The mimetype file is not the first entry in the ZIP archive.

**Common Causes:**
- ZIP archive was created with files in wrong order
- EPUB was repacked incorrectly
- Archive was created with a tool that doesn't preserve file order

**Example:**
```json
{
  "code": "EPUB-CONTAINER-003",
  "message": "mimetype file must be first in ZIP archive, found 'other.txt' instead",
  "details": {
    "first_file": "other.txt"
  }
}
```

**Resolution:** Recreate the EPUB with mimetype as the first file in the ZIP archive.

---

### EPUB-CONTAINER-004: Container XML Missing

**Severity:** Error  
**Description:** The required META-INF/container.xml file is missing.

**Common Causes:**
- File was not included when creating the EPUB
- Incorrect directory structure
- File was deleted or corrupted

**Example:**
```json
{
  "code": "EPUB-CONTAINER-004",
  "message": "Required file 'META-INF/container.xml' is missing",
  "details": {
    "expected_path": "META-INF/container.xml"
  }
}
```

**Resolution:** Add the META-INF/container.xml file with proper rootfile declarations.

---

### EPUB-CONTAINER-005: Container XML Invalid

**Severity:** Error  
**Description:** The META-INF/container.xml file is malformed or invalid.

**Common Causes:**
- XML syntax errors
- Missing required elements
- Empty or invalid rootfile entries
- No rootfile declarations

**Examples:**

Invalid XML:
```json
{
  "code": "EPUB-CONTAINER-005",
  "message": "META-INF/container.xml is not valid XML",
  "details": {
    "error": "XML syntax error at line 1: unexpected EOF"
  }
}
```

No rootfiles:
```json
{
  "code": "EPUB-CONTAINER-005",
  "message": "META-INF/container.xml must contain at least one rootfile",
  "details": {}
}
```

Empty rootfile path:
```json
{
  "code": "EPUB-CONTAINER-005",
  "message": "Rootfile at index 0 has empty full-path attribute",
  "details": {
    "rootfile_index": 0
  }
}
```

**Resolution:** 
- Fix XML syntax errors
- Ensure at least one rootfile element exists
- Ensure all rootfile elements have non-empty full-path attributes

---

## OCF Specification Compliance

These error codes implement checks for the following OCF 3.0 requirements:

1. **Section 3.1**: OCF ZIP Container
   - EPUB containers must be valid ZIP archives
   
2. **Section 3.3**: The mimetype File
   - Must be first file in archive
   - Must be uncompressed
   - Must contain "application/epub+zip"
   
3. **Section 3.4**: META-INF Directory
   - Must contain container.xml file
   
4. **Section 3.5**: container.xml
   - Must be valid XML
   - Must declare at least one rootfile
   - Rootfiles must have valid full-path attributes

## Validation Flow

```
┌─────────────────────┐
│ Read EPUB File      │
└──────┬──────────────┘
       │
       ▼
┌─────────────────────┐
│ Check ZIP Validity  │───► EPUB-CONTAINER-001
└──────┬──────────────┘
       │
       ▼
┌─────────────────────┐
│ Validate Mimetype   │───► EPUB-CONTAINER-002
│ - First in archive  │───► EPUB-CONTAINER-003
│ - Uncompressed      │
│ - Correct content   │
└──────┬──────────────┘
       │
       ▼
┌─────────────────────┐
│ Validate container  │───► EPUB-CONTAINER-004
│ - Exists            │───► EPUB-CONTAINER-005
│ - Valid XML         │
│ - Has rootfiles     │
└─────────────────────┘
```

## Usage Example

```go
validator := epub.NewContainerValidator()
result, err := validator.ValidateFile("book.epub")

if err != nil {
    // I/O error occurred
    log.Fatal(err)
}

if !result.Valid {
    for _, validationError := range result.Errors {
        switch validationError.Code {
        case epub.ErrorCodeZIPInvalid:
            // Handle ZIP validation failure
        case epub.ErrorCodeMimetypeInvalid:
            // Handle mimetype validation failure
        case epub.ErrorCodeMimetypeNotFirst:
            // Handle mimetype order failure
        case epub.ErrorCodeContainerXMLMissing:
            // Handle missing container.xml
        case epub.ErrorCodeContainerXMLInvalid:
            // Handle invalid container.xml
        }
    }
}
```
