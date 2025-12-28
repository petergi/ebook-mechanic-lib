# PDF Validation Error Codes

This document provides a complete reference for all error codes used by the PDF structure validators, aligned with PDF 1.7 specifications and ISO 32000-1:2008.

## Error Code Reference - Section 3: Basic Structure Validation

### PDF-HEADER-001: Invalid or Missing PDF Header

**Severity:** Critical  
**Description:** The file does not start with a valid PDF header signature.

**Common Causes:**
- File is not a PDF
- File is corrupted at the beginning
- File has been truncated or modified
- Wrong file format uploaded

**Example:**
```json
{
  "code": "PDF-HEADER-001",
  "message": "Invalid or missing PDF header",
  "details": {
    "expected": "%PDF-1.x where x=0-7"
  }
}
```

**Resolution:** Ensure the file starts with `%PDF-1.` followed by a valid version number (0-7).

---

### PDF-HEADER-002: Invalid PDF Version Number

**Severity:** Critical  
**Description:** The PDF header contains an unsupported or invalid version number.

**Common Causes:**
- Version number outside the range 1.0-1.7
- Malformed version string
- Future PDF version not yet supported
- Corrupted header

**Example:**
```json
{
  "code": "PDF-HEADER-002",
  "message": "Invalid PDF version number",
  "details": {
    "expected": "1.0 through 1.7",
    "found": "%PDF-1.9"
  }
}
```

**Resolution:** Use a PDF version between 1.0 and 1.7. PDF 2.0 and later require different validation.

---

### PDF-TRAILER-001: Invalid or Missing startxref

**Severity:** Error  
**Description:** The startxref keyword or cross-reference offset is missing or malformed.

**Common Causes:**
- File truncation
- Missing startxref keyword
- Invalid offset value
- Corrupted trailer section

**Example:**
```json
{
  "code": "PDF-TRAILER-001",
  "message": "Invalid or missing startxref",
  "details": {
    "expected": "startxref <offset> before %%EOF"
  }
}
```

**Resolution:** Ensure the file contains `startxref` followed by a valid byte offset, positioned before `%%EOF`.

---

### PDF-TRAILER-002: Invalid Trailer Dictionary

**Severity:** Error  
**Description:** The trailer dictionary is malformed or contains invalid entries.

**Common Causes:**
- Missing required trailer entries (Size, Root)
- Invalid dictionary syntax
- Corrupted trailer data
- Mismatched object references

**Example:**
```json
{
  "code": "PDF-TRAILER-002",
  "message": "Invalid trailer dictionary",
  "details": {
    "error": "Failed to parse trailer dictionary"
  }
}
```

**Resolution:** Ensure the trailer contains a valid dictionary with at least /Size and /Root entries.

---

### PDF-TRAILER-003: Missing %%EOF Marker

**Severity:** Error  
**Description:** The required %%EOF end-of-file marker is missing.

**Common Causes:**
- File truncation during download or transfer
- Incomplete file write
- Storage corruption
- Network interruption during file transfer

**Example:**
```json
{
  "code": "PDF-TRAILER-003",
  "message": "Missing %%EOF marker",
  "details": {
    "expected": "%%EOF at end of file"
  }
}
```

**Resolution:** Append `%%EOF` on a new line at the end of the file.

---

### PDF-XREF-001: Invalid or Damaged Cross-Reference Table

**Severity:** Error  
**Description:** The cross-reference table or stream is malformed or cannot be parsed.

**Common Causes:**
- Corrupted xref table entries
- Invalid xref stream
- Missing xref table
- Incorrect offset values
- Damaged PDF structure

**Example:**
```json
{
  "code": "PDF-XREF-001",
  "message": "Invalid or damaged cross-reference table",
  "details": {
    "error": "Failed to parse cross-reference section"
  }
}
```

**Resolution:** Rebuild the cross-reference table by scanning the file for object definitions. May require specialized PDF repair tools.

---

### PDF-XREF-002: Empty Cross-Reference Table

**Severity:** Error  
**Description:** The cross-reference table exists but contains no object entries.

**Common Causes:**
- Improperly generated PDF
- Corruption during PDF creation
- Invalid PDF writer implementation
- File truncation during xref generation

**Example:**
```json
{
  "code": "PDF-XREF-002",
  "message": "Empty cross-reference table",
  "details": {}
}
```

**Resolution:** Regenerate the PDF or rebuild the cross-reference table with proper object entries.

---

### PDF-XREF-003: Cross-Reference Table Has Overlapping Entries

**Severity:** Error  
**Description:** Multiple objects reference the same byte offset in the cross-reference table.

**Common Causes:**
- Incorrectly calculated offsets
- PDF writer bug
- Manual PDF editing errors
- Incremental update issues

**Example:**
```json
{
  "code": "PDF-XREF-003",
  "message": "Cross-reference table has overlapping entries",
  "details": {
    "offset": 1234,
    "objects": [5, 12]
  }
}
```

**Resolution:** Rebuild the cross-reference table with correct byte offsets for each object.

---

### PDF-CATALOG-001: Missing or Invalid Catalog Object

**Severity:** Critical  
**Description:** The document catalog (root object) is missing or cannot be accessed.

**Common Causes:**
- Invalid /Root entry in trailer
- Corrupted catalog object
- Missing catalog object definition
- Invalid object reference

**Example:**
```json
{
  "code": "PDF-CATALOG-001",
  "message": "Missing or invalid catalog object",
  "details": {}
}
```

**Resolution:** Ensure the trailer contains a valid /Root entry pointing to a catalog dictionary object.

---

### PDF-CATALOG-002: Catalog Missing /Type Entry or Invalid Type

**Severity:** Error  
**Description:** The catalog dictionary is missing the required /Type entry, or it's not set to /Catalog.

**Common Causes:**
- Missing /Type /Catalog entry
- Incorrect type value
- Malformed catalog dictionary
- PDF writer error

**Example:**
```json
{
  "code": "PDF-CATALOG-002",
  "message": "Catalog /Type must be /Catalog",
  "details": {
    "found": "/NotCatalog"
  }
}
```

**Resolution:** Add or correct the `/Type /Catalog` entry in the catalog dictionary.

---

### PDF-CATALOG-003: Catalog Missing /Pages Entry

**Severity:** Critical  
**Description:** The catalog dictionary is missing the required /Pages entry.

**Common Causes:**
- Incomplete catalog dictionary
- Missing page tree root
- Corrupted catalog
- Invalid PDF generation

**Example:**
```json
{
  "code": "PDF-CATALOG-003",
  "message": "Catalog missing /Pages entry",
  "details": {}
}
```

**Resolution:** Add a valid `/Pages` entry in the catalog pointing to the page tree root object.

---

### PDF-STRUCTURE-012: General Structure Error

**Severity:** Error  
**Description:** A general PDF structure error that doesn't fit into more specific categories.

**Common Causes:**
- Various structural issues
- Object numbering conflicts
- Duplicate object/generation pairs
- Parser failures
- Unexpected structure violations

**Example:**
```json
{
  "code": "PDF-STRUCTURE-012",
  "message": "Duplicate object number/generation pair",
  "details": {
    "object_number": 5,
    "generation": 0
  }
}
```

**Resolution:** Depends on the specific error. Review the details field for more information.

---

## PDF Structure Validation Flow

```
┌─────────────────────┐
│ Read PDF File       │
└──────┬──────────────┘
       │
       ▼
┌─────────────────────┐
│ Check Header        │───► PDF-HEADER-001
│ - %PDF-1.x format   │───► PDF-HEADER-002
│ - Version 1.0-1.7   │
└──────┬──────────────┘
       │
       ▼
┌─────────────────────┐
│ Check Trailer       │───► PDF-TRAILER-001
│ - %%EOF present     │───► PDF-TRAILER-002
│ - startxref valid   │───► PDF-TRAILER-003
└──────┬──────────────┘
       │
       ▼
┌─────────────────────┐
│ Parse with unipdf   │
└──────┬──────────────┘
       │
       ▼
┌─────────────────────┐
│ Validate Xref       │───► PDF-XREF-001
│ - Table exists      │───► PDF-XREF-002
│ - Not empty         │───► PDF-XREF-003
│ - No overlaps       │
└──────┬──────────────┘
       │
       ▼
┌─────────────────────┐
│ Validate Catalog    │───► PDF-CATALOG-001
│ - Exists            │───► PDF-CATALOG-002
│ - /Type /Catalog    │───► PDF-CATALOG-003
│ - Has /Pages        │
└──────┬──────────────┘
       │
       ▼
┌─────────────────────┐
│ Validate Objects    │───► PDF-STRUCTURE-012
│ - No duplicates     │
│ - Valid numbering   │
└─────────────────────┘
```

## PDF Specification Compliance

These error codes implement checks for the following PDF 1.7 / ISO 32000-1:2008 requirements:

1. **Section 7.5.2**: File Header
   - Must begin with `%PDF-1.n`
   - Version number must be valid

2. **Section 7.5.5**: File Trailer
   - Must contain trailer dictionary
   - Must have `startxref` keyword
   - Must end with `%%EOF`

3. **Section 7.5.4**: Cross-Reference Table
   - Must be valid table or stream
   - Must contain object references
   - No overlapping offsets

4. **Section 7.7.2**: Document Catalog
   - Must have `/Type /Catalog`
   - Must contain `/Pages` reference
   - Must be reachable from trailer

5. **Section 7.5.3**: Objects
   - Proper numbering and generation
   - No duplicate object identifiers

## Usage Example

```go
validator := pdf.NewStructureValidator()
result, err := validator.ValidateFile("document.pdf")

if err != nil {
    // I/O error occurred
    log.Fatal(err)
}

if !result.Valid {
    for _, validationError := range result.Errors {
        switch validationError.Code {
        case pdf.ErrorCodePDFHeader001:
            // Handle invalid header
        case pdf.ErrorCodePDFHeader002:
            // Handle invalid version
        case pdf.ErrorCodePDFTrailer001:
            // Handle missing startxref
        case pdf.ErrorCodePDFTrailer002:
            // Handle invalid trailer
        case pdf.ErrorCodePDFTrailer003:
            // Handle missing EOF
        case pdf.ErrorCodePDFXref001:
            // Handle damaged xref
        case pdf.ErrorCodePDFXref002:
            // Handle empty xref
        case pdf.ErrorCodePDFXref003:
            // Handle overlapping xref
        case pdf.ErrorCodePDFCatalog001:
            // Handle missing catalog
        case pdf.ErrorCodePDFCatalog002:
            // Handle invalid catalog type
        case pdf.ErrorCodePDFCatalog003:
            // Handle missing pages
        case pdf.ErrorCodePDFStructure012:
            // Handle general structure error
        }
    }
}
```

## Repair Strategy

For information about which errors can be safely repaired, see the main specification document at `docs/specs/EBMLib-PDF-SPEC.md`, Section 9.1.

**Automatically Repairable:**
- PDF-TRAILER-003: Can append `%%EOF`
- PDF-TRAILER-001: Can recompute startxref offset

**Conditionally Repairable (with caution):**
- PDF-XREF-001: Can attempt linear scan rebuild
- PDF-STRUCTURE-012: Depends on specific issue

**Not Automatically Repairable:**
- PDF-HEADER-001, PDF-HEADER-002: Require manual intervention
- PDF-CATALOG-001, PDF-CATALOG-002, PDF-CATALOG-003: Require structural rebuild
- PDF-XREF-002, PDF-XREF-003: Require xref regeneration
