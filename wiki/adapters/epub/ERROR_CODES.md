# EPUB Validation Error Codes

This document provides a complete reference for all error codes used by the EPUB validators, aligned with EPUB specifications.

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

---

## Navigation Document Error Codes

### EPUB-NAV-001: Not Well-Formed

**Severity:** Error  
**Description:** The navigation document is not well-formed XHTML.

**Common Causes:**
- XML/XHTML syntax errors
- Unclosed tags
- Invalid character encoding
- Malformed HTML structure

**Example:**
```json
{
  "code": "EPUB-NAV-001",
  "message": "Navigation document is not well-formed XHTML",
  "details": {
    "error": "XML syntax error at line 5"
  }
}
```

**Resolution:** Fix the XHTML syntax errors in the navigation document.

---

### EPUB-NAV-002: Missing TOC

**Severity:** Error  
**Description:** The navigation document does not contain a required `<nav epub:type="toc">` element.

**Common Causes:**
- Missing TOC navigation element
- Incorrect epub:type attribute value
- TOC element not marked with proper namespace

**Example:**
```json
{
  "code": "EPUB-NAV-002",
  "message": "Navigation document must contain <nav epub:type=\"toc\">",
  "details": {}
}
```

**Resolution:** Add a `<nav epub:type="toc">` element containing the table of contents.

---

### EPUB-NAV-003: Invalid TOC Structure

**Severity:** Error  
**Description:** The TOC navigation element does not have the required `<ol>` structure.

**Common Causes:**
- Missing `<ol>` element within TOC nav
- TOC nav contains only text or other elements
- Incorrect nesting of navigation structure

**Example:**
```json
{
  "code": "EPUB-NAV-003",
  "message": "TOC <nav> element must contain an <ol> element",
  "details": {}
}
```

**Resolution:** Ensure the `<nav epub:type="toc">` element contains an ordered list (`<ol>`) with navigation items.

---

### EPUB-NAV-004: Invalid Links

**Severity:** Error  
**Description:** The navigation document contains invalid or non-relative links.

**Common Causes:**
- Absolute URLs (http://, https://)
- Protocol-relative URLs (//)
- Absolute paths (starting with /)
- Links pointing outside the EPUB package (..)
- Empty href attributes

**Examples:**

Absolute URL:
```json
{
  "code": "EPUB-NAV-004",
  "message": "TOC contains invalid relative link: http://example.com/chapter.xhtml",
  "details": {
    "href": "http://example.com/chapter.xhtml",
    "text": "Chapter 1"
  }
}
```

Empty link:
```json
{
  "code": "EPUB-NAV-004",
  "message": "TOC contains invalid relative link: ",
  "details": {
    "href": "",
    "text": "Invalid Link"
  }
}
```

**Resolution:** Use only relative links within the EPUB package (e.g., "chapter1.xhtml", "content/chapter2.xhtml#section1").

---

### EPUB-NAV-005: Invalid Landmarks

**Severity:** Error  
**Description:** The landmarks navigation element does not have the required `<ol>` structure.

**Common Causes:**
- Missing `<ol>` element within landmarks nav
- Landmarks nav contains only text or other elements
- Incorrect nesting of landmarks structure

**Example:**
```json
{
  "code": "EPUB-NAV-005",
  "message": "Landmarks <nav> element must contain an <ol> element",
  "details": {}
}
```

**Resolution:** Ensure the `<nav epub:type="landmarks">` element contains an ordered list (`<ol>`) with landmark items.

---

### EPUB-NAV-006: Missing Nav Element

**Severity:** Error  
**Description:** The navigation document does not contain any `<nav>` elements.

**Common Causes:**
- Using `<div>` or other elements instead of `<nav>`
- Missing navigation structure entirely
- Navigation content not properly wrapped

**Example:**
```json
{
  "code": "EPUB-NAV-006",
  "message": "Navigation document must contain at least one <nav> element",
  "details": {}
}
```

**Resolution:** Add at least one `<nav>` element with proper epub:type attribute to the navigation document.

---

## Navigation Document Validation Flow

```
┌─────────────────────────┐
│ Parse Navigation Doc    │
└──────┬──────────────────┘
       │
       ▼
┌─────────────────────────┐
│ Check Well-Formedness   │───► EPUB-NAV-001
└──────┬──────────────────┘
       │
       ▼
┌─────────────────────────┐
│ Find <nav> Elements     │───► EPUB-NAV-006
└──────┬──────────────────┘
       │
       ▼
┌─────────────────────────┐
│ Validate TOC Present    │───► EPUB-NAV-002
└──────┬──────────────────┘
       │
       ▼
┌─────────────────────────┐
│ Validate TOC Structure  │───► EPUB-NAV-003
│ - Must have <ol>        │
│ - Extract links         │
└──────┬──────────────────┘
       │
       ▼
┌─────────────────────────┐
│ Validate Links          │───► EPUB-NAV-004
│ - Must be relative      │
│ - No absolute URLs      │
│ - No parent refs (..)   │
└──────┬──────────────────┘
       │
       ▼
┌─────────────────────────┐
│ Validate Landmarks      │───► EPUB-NAV-005
│ (if present)            │───► EPUB-NAV-004
│ - Must have <ol>        │
│ - Validate links        │
└─────────────────────────┘
```

## Navigation Document Usage Example

```go
validator := epub.NewNavValidator()
result, err := validator.ValidateFile("OEBPS/nav.xhtml")

if err != nil {
    // I/O error occurred
    log.Fatal(err)
}

if !result.Valid {
    for _, validationError := range result.Errors {
        switch validationError.Code {
        case epub.ErrorCodeNavNotWellFormed:
            // Handle malformed XHTML
        case epub.ErrorCodeNavMissingTOC:
            // Handle missing TOC
        case epub.ErrorCodeNavInvalidTOCStructure:
            // Handle invalid TOC structure
        case epub.ErrorCodeNavInvalidLinks:
            // Handle invalid links
        case epub.ErrorCodeNavInvalidLandmarks:
            // Handle invalid landmarks
        case epub.ErrorCodeNavMissingNavElement:
            // Handle missing nav element
        }
    }
}

// Access extracted navigation data
for _, link := range result.TOCLinks {
    fmt.Printf("TOC: %s -> %s\n", link.Text, link.Href)
}

for _, link := range result.LandmarkLinks {
    fmt.Printf("Landmark: %s -> %s\n", link.Text, link.Href)
}
```

---

## Accessibility Error Codes (WCAG 2.1 & EPUB Accessibility 1.1)

### EPUB-A11Y-001: Missing Language Declaration

**Severity:** Error  
**Description:** HTML element missing lang or xml:lang attribute.  
**WCAG 2.1:** 3.1.1 Language of Page (Level A)

**Resolution:** Add lang="en" to the <html> element.

---

### EPUB-A11Y-002: Invalid Language Code

**Severity:** Warning  
**Description:** Language code may not be valid.  
**WCAG 2.1:** 3.1.1 Language of Page (Level A)

**Resolution:** Use valid ISO 639 codes (e.g., "en", "fr", "en-US").

---

### EPUB-A11Y-003: Missing Semantic Structure

**Severity:** Warning  
**Description:** Document contains no HTML5 semantic elements.  
**WCAG 2.1:** 1.3.1 Info and Relationships (Level A)

**Resolution:** Use <article>, <section>, <nav>, <header>, <footer>, <aside>, <main>.

---

### EPUB-A11Y-005: Missing Alt Text

**Severity:** Error  
**Description:** Image missing alt attribute.  
**WCAG 2.1:** 1.1.1 Non-text Content (Level A)

**Resolution:** Add alt attribute to all images. Use alt="" for decorative images.

---

### EPUB-A11Y-007: Invalid ARIA Role

**Severity:** Error  
**Description:** Invalid ARIA role value.  
**WCAG 2.1:** 4.1.2 Name, Role, Value (Level A)

**Resolution:** Use valid ARIA roles from WAI-ARIA specification.

---

### EPUB-A11Y-009: Missing ARIA Label

**Severity:** Error  
**Description:** Element with role requires aria-label or aria-labelledby.  
**WCAG 2.1:** 4.1.2 Name, Role, Value (Level A)

**Resolution:** Add aria-label or aria-labelledby.

---

### EPUB-A11Y-011: Missing Table Headers

**Severity:** Error  
**Description:** Data table missing header cells.  
**WCAG 2.1:** 1.3.1 Info and Relationships (Level A)

**Resolution:** Add <th> elements or headers attribute.

---

### EPUB-A11Y-013: Missing Form Labels

**Severity:** Error  
**Description:** Form control missing label.  
**WCAG 2.1:** 3.3.2 Labels (Level A), 4.1.2 Name, Role, Value (Level A)

**Resolution:** Associate with <label> or add aria-label.

---

### EPUB-A11Y-019: Empty Heading

**Severity:** Error  
**Description:** Heading element is empty.  
**WCAG 2.1:** 1.3.1 Info and Relationships (Level A)

**Resolution:** Provide meaningful heading text.

---

### EPUB-A11Y-020: Skipped Heading Level

**Severity:** Error  
**Description:** Heading hierarchy skips levels.  
**WCAG 2.1:** 1.3.1 Info and Relationships (Level A)

**Resolution:** Follow h1 → h2 → h3 hierarchy without skipping.

---

## Accessibility Scoring (0-100)

- **Language Declaration (5%):** Valid lang/xml:lang
- **Semantic Structure (25%):** HTML5 semantic elements
- **ARIA Compliance (20%):** Proper ARIA roles/attributes
- **Alt Text (25%):** Images with appropriate alt text
- **Heading Hierarchy (15%):** Proper heading structure
- **Reading Order (10%):** No disruptive tabindex

**Compliance Levels:**
- 90-100: WCAG 2.1 AA
- 80-89: WCAG 2.1 A
- 60-79: Partial
- 0-59: Non-compliant
