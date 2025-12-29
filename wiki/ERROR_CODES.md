# ebm-lib Error Code Catalog

**Version:** 1.0  
**Last Updated:** December 2025  
**Specification Alignment:** EPUB 3.3, PDF 1.7 (ISO 32000-1:2008)

## Table of Contents

1. [Overview](#overview)
2. [Error Format Structure](#error-format-structure)
3. [EPUB Error Codes](#epub-error-codes)
4. [PDF Error Codes](#pdf-error-codes)
5. [Severity Levels](#severity-levels)
6. [Error Handling Guidelines](#error-handling-guidelines)

---

## Overview

This document provides a comprehensive catalog of all validation error codes used in ebm-lib. Each error code follows a structured format and includes:

- **Code:** Unique identifier (e.g., "EPUB-CONTAINER-001")
- **Severity:** Critical, Error, Warning, or Info
- **Description:** What the error means
- **Common Causes:** Why this error occurs
- **Resolution:** How to fix the error
- **Auto-Repairable:** Whether automatic repair is possible
- **Safety Level:** If repairable, the safety level of the repair

### Error Code Naming Convention

```
<FORMAT>-<CATEGORY>-<NUMBER>
```

- **FORMAT:** EPUB, PDF, PDFA (PDF/A), or PDFUA (PDF/UA)
- **CATEGORY:** Container, Header, Trailer, Nav, etc.
- **NUMBER:** Three-digit unique identifier (001-999)

---

## Error Format Structure

All validation errors follow this unified structure:

### Go Structure

```go
type ValidationError struct {
    Code      string                 // Unique error code
    Message   string                 // Human-readable description
    Severity  Severity              // Error, Warning, or Info
    Location  *ErrorLocation        // Where the problem occurs
    Details   map[string]interface{} // Additional context
    Timestamp time.Time             // When error was detected
}

type ErrorLocation struct {
    File    string  // File path within ebook
    Line    int     // Line number (if applicable)
    Column  int     // Column number (if applicable)
    Path    string  // XPath or structural path
    Context string  // Surrounding context
}
```

### JSON Example

```json
{
  "code": "EPUB-CONTAINER-002",
  "message": "mimetype file must contain exactly 'application/epub+zip'",
  "severity": "error",
  "location": {
    "file": "mimetype",
    "line": 1,
    "column": 1,
    "path": "/mimetype",
    "context": "application/wrong"
  },
  "details": {
    "expected": "application/epub+zip",
    "found": "application/wrong",
    "repairable": true,
    "safety_level": "very_high"
  },
  "timestamp": "2025-12-28T10:30:45Z"
}
```

---

## EPUB Error Codes

### Container Errors (EPUB-CONTAINER-XXX)

These errors relate to the EPUB Open Container Format (OCF) structure, including ZIP archive format, mimetype file, and container.xml.

#### EPUB-CONTAINER-001: ZIP Invalid

**Severity:** Error  
**Category:** STRUCTURE  
**Auto-Repairable:** No

**Description:**  
The file is not a valid ZIP archive or the ZIP structure is corrupted.

**Common Causes:**
- File is not in ZIP format
- File has been truncated or incomplete download
- Corrupted ZIP structure (bad CRC, invalid headers)
- Wrong file type (not actually an EPUB)

**Specification Reference:**  
OCF 3.0, Section 3.1 - OCF ZIP Container

**Example:**
```json
{
  "code": "EPUB-CONTAINER-001",
  "message": "File is not a valid ZIP archive",
  "severity": "error",
  "details": {
    "error": "zip: not a valid zip file",
    "repairable": false
  }
}
```

**Resolution:**
- Verify file was completely downloaded
- Check file integrity (MD5/SHA checksums)
- Re-download from source
- For minor corruption, try specialized ZIP repair tools

---

#### EPUB-CONTAINER-002: Mimetype Invalid

**Severity:** Error  
**Category:** STRUCTURE  
**Auto-Repairable:** Yes  
**Safety Level:** Very High

**Description:**  
The mimetype file has incorrect content, compression, or format.

**Common Causes:**
- Mimetype contains wrong content (not "application/epub+zip")
- Mimetype file is compressed (must be stored uncompressed)
- Mimetype has extra whitespace or line breaks
- Mimetype file is missing or empty

**Specification Reference:**  
OCF 3.0, Section 3.3 - The mimetype File

**Examples:**

Wrong content:
```json
{
  "code": "EPUB-CONTAINER-002",
  "message": "mimetype file must contain exactly 'application/epub+zip'",
  "severity": "error",
  "details": {
    "expected": "application/epub+zip",
    "found": "application/wrong",
    "repairable": true,
    "safety_level": "very_high"
  }
}
```

Compressed mimetype:
```json
{
  "code": "EPUB-CONTAINER-002",
  "message": "mimetype file must be stored uncompressed",
  "severity": "error",
  "details": {
    "compression_method": 8,
    "expected_method": 0,
    "repairable": true,
    "safety_level": "high"
  }
}
```

**Resolution:**
- Auto-repair: Create/overwrite mimetype with exact content: `application/epub+zip`
- Ensure no whitespace, line breaks, or BOM
- Ensure stored uncompressed (ZIP Store method, not Deflate)
- Rebuild EPUB ZIP with correct mimetype

**Repair Strategy:**
```
Type: MIMETYPE_CONTENT_FIX
Action: Overwrite mimetype file with correct content
Safety: Very High (only affects single non-content file)
```

---

#### EPUB-CONTAINER-003: Mimetype Not First

**Severity:** Error  
**Category:** STRUCTURE  
**Auto-Repairable:** Yes  
**Safety Level:** High

**Description:**  
The mimetype file is not the first entry in the ZIP archive.

**Common Causes:**
- ZIP archive was created with files in wrong order
- EPUB was repacked incorrectly
- Archive was created with a tool that doesn't preserve file order
- Directory or other files added before mimetype

**Specification Reference:**  
OCF 3.0, Section 3.3 - The mimetype File (must be first file)

**Example:**
```json
{
  "code": "EPUB-CONTAINER-003",
  "message": "mimetype file must be first in ZIP archive, found 'META-INF/' instead",
  "severity": "error",
  "details": {
    "first_file": "META-INF/",
    "mimetype_position": 2,
    "repairable": true,
    "safety_level": "high"
  }
}
```

**Resolution:**
- Auto-repair: Rebuild ZIP archive with mimetype as first entry
- Use EPUB creation tools that maintain correct order
- Manually repack using command-line ZIP tools with proper ordering

**Repair Strategy:**
```
Type: ZIP_REBUILD
Action: Extract all files, rebuild ZIP with mimetype first
Safety: High (structure only, all content preserved)
Estimated Time: 2-10 seconds depending on EPUB size
```

---

#### EPUB-CONTAINER-004: Container XML Missing

**Severity:** Error  
**Category:** STRUCTURE  
**Auto-Repairable:** Conditional  
**Safety Level:** High

**Description:**  
The required META-INF/container.xml file is missing.

**Common Causes:**
- File was not included when creating the EPUB
- Incorrect directory structure
- File was deleted or corrupted
- Incomplete EPUB extraction/creation

**Specification Reference:**  
OCF 3.0, Section 3.5 - The container.xml File

**Example:**
```json
{
  "code": "EPUB-CONTAINER-004",
  "message": "Required file 'META-INF/container.xml' is missing",
  "severity": "error",
  "details": {
    "expected_path": "META-INF/container.xml",
    "repairable": true,
    "requires_path_guess": true,
    "safety_level": "high"
  }
}
```

**Resolution:**
- Auto-repair (conditional): Create minimal container.xml pointing to default OEBPS/content.opf
- Only auto-repairable if package document path can be guessed
- Manual: Create container.xml with correct rootfile path

**Repair Strategy:**
```
Type: CONTAINER_XML_CREATE
Action: Create minimal valid container.xml
Condition: If package document (*.opf) can be found in standard locations
Standard Locations: OEBPS/content.opf, content.opf, package.opf
Safety: High (if path is correct)
```

**Minimal container.xml template:**
```xml
<?xml version="1.0" encoding="UTF-8"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>
```

---

#### EPUB-CONTAINER-005: Container XML Invalid

**Severity:** Error  
**Category:** STRUCTURE  
**Auto-Repairable:** Conditional  
**Safety Level:** Medium-High

**Description:**  
The META-INF/container.xml file is malformed or contains invalid data.

**Common Causes:**
- XML syntax errors (unclosed tags, invalid characters)
- Missing required elements (rootfiles, rootfile)
- Empty or invalid rootfile entries
- No rootfile declarations
- Wrong namespace or XML structure

**Specification Reference:**  
OCF 3.0, Section 3.5 - The container.xml File

**Examples:**

Invalid XML:
```json
{
  "code": "EPUB-CONTAINER-005",
  "message": "META-INF/container.xml is not valid XML",
  "severity": "error",
  "details": {
    "error": "XML syntax error at line 3: unexpected EOF",
    "repairable": false
  }
}
```

No rootfiles:
```json
{
  "code": "EPUB-CONTAINER-005",
  "message": "META-INF/container.xml must contain at least one rootfile",
  "severity": "error",
  "details": {
    "rootfile_count": 0,
    "repairable": true,
    "safety_level": "high"
  }
}
```

Empty rootfile path:
```json
{
  "code": "EPUB-CONTAINER-005",
  "message": "Rootfile at index 0 has empty full-path attribute",
  "severity": "error",
  "details": {
    "rootfile_index": 0,
    "repairable": true,
    "safety_level": "high"
  }
}
```

**Resolution:**
- Auto-repair: Fix common XML issues, add missing rootfile if package can be found
- Manual: Fix XML syntax errors
- Ensure at least one rootfile element with non-empty full-path attribute

**Repair Strategy:**
```
Type: CONTAINER_XML_FIX
Actions:
  - Fix XML syntax if simple (missing closing tags)
  - Add missing rootfile if package document can be located
  - Normalize rootfile paths (remove leading slash, fix case)
Safety: Medium-High (depends on severity of XML errors)
```

---

### Navigation Errors (EPUB-NAV-XXX)

These errors relate to the EPUB 3 navigation document (nav.xhtml), which provides the table of contents and other navigation structures.

#### EPUB-NAV-001: Not Well-Formed

**Severity:** Error  
**Category:** CONTENT  
**Auto-Repairable:** No (generally requires manual intervention)  
**Safety Level:** N/A

**Description:**  
The navigation document is not well-formed XHTML.

**Common Causes:**
- XML/XHTML syntax errors (unclosed tags, missing quotes)
- Invalid character encoding
- Malformed HTML structure
- Invalid entities or special characters

**Specification Reference:**  
EPUB 3.3, Section 7 - EPUB Navigation Documents

**Example:**
```json
{
  "code": "EPUB-NAV-001",
  "message": "Navigation document is not well-formed XHTML",
  "severity": "error",
  "location": {
    "file": "OEBPS/nav.xhtml",
    "line": 15,
    "column": 8
  },
  "details": {
    "error": "XML syntax error: unclosed tag <li>",
    "repairable": false
  }
}
```

**Resolution:**
- Manual: Fix XHTML syntax errors
- Use XHTML validator to identify specific issues
- Use XML-aware editors (e.g., Sigil) to fix structure
- Check for invalid characters or encoding issues

---

#### EPUB-NAV-002: Missing TOC

**Severity:** Error  
**Category:** CONTENT  
**Auto-Repairable:** Conditional  
**Safety Level:** Medium

**Description:**  
The navigation document does not contain a required `<nav epub:type="toc">` element.

**Common Causes:**
- TOC navigation element not created
- Incorrect epub:type attribute value
- TOC element not marked with proper namespace
- Wrong element used (e.g., <div> instead of <nav>)

**Specification Reference:**  
EPUB 3.3, Section 7.3 - The nav Element (toc is required)

**Example:**
```json
{
  "code": "EPUB-NAV-002",
  "message": "Navigation document must contain <nav epub:type=\"toc\">",
  "severity": "error",
  "location": {
    "file": "OEBPS/nav.xhtml"
  },
  "details": {
    "nav_elements_found": 0,
    "repairable": true,
    "requires_manual_review": true,
    "safety_level": "medium"
  }
}
```

**Resolution:**
- Auto-repair (conditional): Generate minimal TOC from spine order
- May produce generic chapter names ("Chapter 1", "Chapter 2")
- Strongly recommend manual review and proper chapter titles

**Repair Strategy:**
```
Type: NAV_TOC_GENERATE
Action: Create basic TOC structure from spine items
Safety: Medium (heuristic-based, may need adjustment)
Output: Generic chapter structure requires manual refinement
```

---

#### EPUB-NAV-003: Invalid TOC Structure

**Severity:** Error  
**Category:** CONTENT  
**Auto-Repairable:** Yes  
**Safety Level:** High

**Description:**  
The TOC navigation element does not have the required `<ol>` structure.

**Common Causes:**
- Missing `<ol>` element within TOC nav
- TOC nav contains only text or other elements
- Incorrect nesting of navigation structure
- Wrong list type used (e.g., <ul> instead of <ol>)

**Specification Reference:**  
EPUB 3.3, Section 7.3 - The nav Element (must contain ol)

**Example:**
```json
{
  "code": "EPUB-NAV-003",
  "message": "TOC <nav> element must contain an <ol> element",
  "severity": "error",
  "location": {
    "file": "OEBPS/nav.xhtml",
    "path": "//nav[@epub:type='toc']"
  },
  "details": {
    "found_element": "div",
    "repairable": true,
    "safety_level": "high"
  }
}
```

**Resolution:**
- Auto-repair: Wrap content in proper `<ol>` structure
- Convert `<ul>` to `<ol>` if present
- Ensure proper nesting of `<li>` elements

**Repair Strategy:**
```
Type: NAV_STRUCTURE_FIX
Action: Add/convert to proper <ol> structure
Safety: High (structure correction only)
```

---

#### EPUB-NAV-004: Invalid Links

**Severity:** Error  
**Category:** CONTENT  
**Auto-Repairable:** Yes  
**Safety Level:** High

**Description:**  
The navigation document contains invalid or non-relative links.

**Common Causes:**
- Absolute URLs (http://, https://) used instead of relative paths
- Protocol-relative URLs (//) 
- Absolute paths starting with /
- Links pointing outside EPUB package (..)
- Empty href attributes

**Specification Reference:**  
EPUB 3.3, Section 7.3 - The nav Element (relative links required)

**Examples:**

Absolute URL:
```json
{
  "code": "EPUB-NAV-004",
  "message": "TOC contains invalid relative link: http://example.com/chapter.xhtml",
  "severity": "error",
  "location": {
    "file": "OEBPS/nav.xhtml",
    "line": 23,
    "path": "//nav[@epub:type='toc']//a[1]"
  },
  "details": {
    "href": "http://example.com/chapter.xhtml",
    "text": "Chapter 1",
    "link_type": "absolute_url",
    "repairable": false
  }
}
```

Empty link:
```json
{
  "code": "EPUB-NAV-004",
  "message": "TOC contains invalid relative link: ",
  "severity": "error",
  "details": {
    "href": "",
    "text": "Invalid Link",
    "link_type": "empty",
    "repairable": true,
    "safety_level": "high"
  }
}
```

**Resolution:**
- Auto-repair: Normalize relative paths, fix case mismatches
- Remove or fix absolute URLs (if target exists in EPUB)
- Validate all links point to existing resources

**Repair Strategy:**
```
Type: NAV_LINK_NORMALIZE
Actions:
  - Convert absolute paths to relative
  - Fix case sensitivity issues
  - Normalize path separators
  - Remove empty href attributes
Safety: High (verifiable corrections)
```

---

#### EPUB-NAV-005: Invalid Landmarks

**Severity:** Warning  
**Category:** CONTENT  
**Auto-Repairable:** Yes  
**Safety Level:** High

**Description:**  
The landmarks navigation element (if present) does not have the required `<ol>` structure.

**Common Causes:**
- Missing `<ol>` element within landmarks nav
- Landmarks nav contains only text or other elements
- Incorrect nesting of landmarks structure

**Specification Reference:**  
EPUB 3.3, Section 7.3 - The nav Element (landmarks strongly recommended)

**Example:**
```json
{
  "code": "EPUB-NAV-005",
  "message": "Landmarks <nav> element must contain an <ol> element",
  "severity": "warning",
  "location": {
    "file": "OEBPS/nav.xhtml",
    "path": "//nav[@epub:type='landmarks']"
  },
  "details": {
    "repairable": true,
    "safety_level": "high"
  }
}
```

**Resolution:**
- Auto-repair: Add proper `<ol>` structure to landmarks nav
- Ensure proper nesting of landmark items

---

#### EPUB-NAV-006: Missing Nav Element

**Severity:** Error  
**Category:** CONTENT  
**Auto-Repairable:** Conditional  
**Safety Level:** Medium

**Description:**  
The navigation document does not contain any `<nav>` elements.

**Common Causes:**
- Using `<div>` or other elements instead of `<nav>`
- Missing navigation structure entirely
- Navigation content not properly wrapped
- Legacy EPUB 2 NCX file mistakenly used as navigation document

**Specification Reference:**  
EPUB 3.3, Section 7.3 - The nav Element (at least one required)

**Example:**
```json
{
  "code": "EPUB-NAV-006",
  "message": "Navigation document must contain at least one <nav> element",
  "severity": "error",
  "location": {
    "file": "OEBPS/nav.xhtml"
  },
  "details": {
    "repairable": true,
    "requires_content_generation": true,
    "safety_level": "medium"
  }
}
```

**Resolution:**
- Auto-repair (conditional): Generate minimal `<nav>` structure
- Requires ability to determine document structure
- Manual review strongly recommended

---

## PDF Error Codes

### Header Errors (PDF-HEADER-XXX)

These errors relate to the PDF file header, which identifies the file as a PDF and specifies the PDF version.

#### PDF-HEADER-001: Invalid or Missing PDF Header

**Severity:** Critical  
**Category:** STRUCTURE  
**Auto-Repairable:** No  
**Safety Level:** N/A

**Description:**  
The file does not start with a valid PDF header signature.

**Common Causes:**
- File is not a PDF
- File is corrupted at the beginning
- File has been truncated or modified
- Wrong file format uploaded
- BOM or other characters before PDF header

**Specification Reference:**  
PDF 1.7, Section 7.5.2 - File Header

**Example:**
```json
{
  "code": "PDF-HEADER-001",
  "message": "Invalid or missing PDF header",
  "severity": "critical",
  "location": {
    "offset": 0
  },
  "details": {
    "expected": "%PDF-1.x where x=0-7",
    "found": "<!DOCTYPE",
    "repairable": false
  }
}
```

**Resolution:**
- Verify file is actually a PDF
- Check for corruption during transfer
- Re-download or re-export from source
- Cannot be auto-repaired (affects entire document structure)

---

#### PDF-HEADER-002: Invalid PDF Version Number

**Severity:** Critical  
**Category:** STRUCTURE  
**Auto-Repairable:** No  
**Safety Level:** N/A

**Description:**  
The PDF header contains an unsupported or invalid version number.

**Common Causes:**
- Version number outside the range 1.0-1.7
- Malformed version string
- Future PDF version not yet supported (e.g., PDF 2.0)
- Corrupted header

**Specification Reference:**  
PDF 1.7, Section 7.5.2 - File Header (version must be 1.0-1.7)

**Example:**
```json
{
  "code": "PDF-HEADER-002",
  "message": "Invalid PDF version number",
  "severity": "critical",
  "location": {
    "offset": 0,
    "line": 1
  },
  "details": {
    "expected": "1.0 through 1.7",
    "found": "%PDF-1.9",
    "repairable": false
  }
}
```

**Resolution:**
- Use PDF version 1.0-1.7 (PDF 2.0 requires different validation)
- Export from source application with compatible version
- Cannot be auto-repaired (affects feature availability)

---

### Trailer Errors (PDF-TRAILER-XXX)

These errors relate to the PDF file trailer, which provides information about the cross-reference table and document catalog.

#### PDF-TRAILER-001: Invalid or Missing startxref

**Severity:** Error  
**Category:** STRUCTURE  
**Auto-Repairable:** Yes  
**Safety Level:** High

**Description:**  
The startxref keyword or cross-reference offset is missing or malformed.

**Common Causes:**
- File truncation
- Missing startxref keyword
- Invalid offset value
- Corrupted trailer section

**Specification Reference:**  
PDF 1.7, Section 7.5.5 - File Trailer

**Example:**
```json
{
  "code": "PDF-TRAILER-001",
  "message": "Invalid or missing startxref",
  "severity": "error",
  "details": {
    "expected": "startxref <offset> before %%EOF",
    "repairable": true,
    "safety_level": "high"
  }
}
```

**Resolution:**
- Auto-repair: Recompute cross-reference table offset
- Scan file for xref location and update startxref value

**Repair Strategy:**
```
Type: STARTXREF_RECOMPUTE
Action: Calculate correct cross-reference offset
Safety: High (calculated value, no content changes)
Estimated Time: <1 second
```

---

#### PDF-TRAILER-002: Invalid Trailer Dictionary

**Severity:** Error  
**Category:** STRUCTURE  
**Auto-Repairable:** Conditional  
**Safety Level:** High (for simple fixes)

**Description:**  
The trailer dictionary is malformed or contains invalid entries.

**Common Causes:**
- Missing required trailer entries (Size, Root)
- Invalid dictionary syntax
- Corrupted trailer data
- Mismatched object references

**Specification Reference:**  
PDF 1.7, Section 7.5.5 - File Trailer

**Example:**
```json
{
  "code": "PDF-TRAILER-002",
  "message": "Invalid trailer dictionary",
  "severity": "error",
  "details": {
    "error": "Missing /Root entry",
    "repairable": false
  }
}
```

**Resolution:**
- Auto-repair (conditional): Fix common typos in trailer dictionary
- Correct /Size value mismatch
- Cannot repair missing /Root or severe corruption

**Repair Strategy:**
```
Type: TRAILER_DICT_FIX
Actions:
  - Fix /Size value to match object count
  - Correct common typos
  - Normalize dictionary formatting
Safety: High (for pattern-based corrections)
Limitations: Cannot fix missing required entries
```

---

#### PDF-TRAILER-003: Missing %%EOF Marker

**Severity:** Error  
**Category:** STRUCTURE  
**Auto-Repairable:** Yes  
**Safety Level:** Very High

**Description:**  
The required %%EOF end-of-file marker is missing.

**Common Causes:**
- File truncation during download or transfer
- Incomplete file write
- Storage corruption
- Network interruption during file transfer

**Specification Reference:**  
PDF 1.7, Section 7.5.5 - File Trailer (must end with %%EOF)

**Example:**
```json
{
  "code": "PDF-TRAILER-003",
  "message": "Missing %%EOF marker",
  "severity": "error",
  "details": {
    "expected": "%%EOF at end of file",
    "repairable": true,
    "safety_level": "very_high"
  }
}
```

**Resolution:**
- Auto-repair: Append `%%EOF` on a new line at end of file
- Simplest and safest PDF repair

**Repair Strategy:**
```
Type: EOF_MARKER_ADD
Action: Append "%%EOF\n" to end of file
Safety: Very High (purely additive, no content changes)
Estimated Time: <1 second
Success Rate: 99%+
```

---

### Cross-Reference Errors (PDF-XREF-XXX)

These errors relate to the PDF cross-reference table or stream, which provides byte offsets for all objects in the file.

#### PDF-XREF-001: Invalid or Damaged Cross-Reference Table

**Severity:** Error  
**Category:** STRUCTURE  
**Auto-Repairable:** No (requires specialized tools)  
**Safety Level:** N/A

**Description:**  
The cross-reference table or stream is malformed or cannot be parsed.

**Common Causes:**
- Corrupted xref table entries
- Invalid xref stream
- Missing xref table
- Incorrect offset values
- Damaged PDF structure

**Specification Reference:**  
PDF 1.7, Section 7.5.4 - Cross-Reference Table

**Example:**
```json
{
  "code": "PDF-XREF-001",
  "message": "Invalid or damaged cross-reference table",
  "severity": "error",
  "details": {
    "error": "Failed to parse cross-reference section",
    "repairable": false,
    "recommended_tool": "qpdf --check"
  }
}
```

**Resolution:**
- Cannot be auto-repaired (requires structural rebuild)
- Use specialized PDF repair tools: QPDF, Adobe Acrobat Pro
- May require rebuilding xref from scratch by scanning objects

---

#### PDF-XREF-002: Empty Cross-Reference Table

**Severity:** Error  
**Category:** STRUCTURE  
**Auto-Repairable:** No  
**Safety Level:** N/A

**Description:**  
The cross-reference table exists but contains no object entries.

**Common Causes:**
- Improperly generated PDF
- Corruption during PDF creation
- Invalid PDF writer implementation
- File truncation during xref generation

**Specification Reference:**  
PDF 1.7, Section 7.5.4 - Cross-Reference Table

**Example:**
```json
{
  "code": "PDF-XREF-002",
  "message": "Empty cross-reference table",
  "severity": "error",
  "details": {
    "object_count": 0,
    "repairable": false
  }
}
```

**Resolution:**
- Cannot be auto-repaired
- Regenerate PDF from source
- Rebuild xref table using PDF repair tools

---

#### PDF-XREF-003: Cross-Reference Table Has Overlapping Entries

**Severity:** Error  
**Category:** STRUCTURE  
**Auto-Repairable:** No  
**Safety Level:** N/A

**Description:**  
Multiple objects reference the same byte offset in the cross-reference table.

**Common Causes:**
- Incorrectly calculated offsets
- PDF writer bug
- Manual PDF editing errors
- Incremental update issues

**Specification Reference:**  
PDF 1.7, Section 7.5.4 - Cross-Reference Table

**Example:**
```json
{
  "code": "PDF-XREF-003",
  "message": "Cross-reference table has overlapping entries",
  "severity": "error",
  "details": {
    "offset": 1234,
    "objects": [5, 12],
    "repairable": false
  }
}
```

**Resolution:**
- Cannot be auto-repaired
- Rebuild xref table with correct byte offsets
- Use PDF repair tools to recalculate all offsets

---

### Catalog Errors (PDF-CATALOG-XXX)

These errors relate to the PDF catalog (document root), which is the entry point to the document structure.

#### PDF-CATALOG-001: Missing or Invalid Catalog Object

**Severity:** Critical  
**Category:** STRUCTURE  
**Auto-Repairable:** No  
**Safety Level:** N/A

**Description:**  
The document catalog (root object) is missing or cannot be accessed.

**Common Causes:**
- Invalid /Root entry in trailer
- Corrupted catalog object
- Missing catalog object definition
- Invalid object reference

**Specification Reference:**  
PDF 1.7, Section 7.7.2 - Document Catalog

**Example:**
```json
{
  "code": "PDF-CATALOG-001",
  "message": "Missing or invalid catalog object",
  "severity": "critical",
  "details": {
    "repairable": false
  }
}
```

**Resolution:**
- Cannot be auto-repaired (affects document root)
- Requires complete document structure rebuild
- Use PDF creation tools to regenerate from source

---

#### PDF-CATALOG-002: Catalog Missing /Type Entry or Invalid Type

**Severity:** Error  
**Category:** STRUCTURE  
**Auto-Repairable:** No  
**Safety Level:** N/A

**Description:**  
The catalog dictionary is missing the required /Type entry, or it's not set to /Catalog.

**Common Causes:**
- Missing /Type /Catalog entry
- Incorrect type value
- Malformed catalog dictionary
- PDF writer error

**Specification Reference:**  
PDF 1.7, Section 7.7.2 - Document Catalog (must have /Type /Catalog)

**Example:**
```json
{
  "code": "PDF-CATALOG-002",
  "message": "Catalog /Type must be /Catalog",
  "severity": "error",
  "details": {
    "found": "/NotCatalog",
    "expected": "/Catalog",
    "repairable": false
  }
}
```

**Resolution:**
- Cannot be auto-repaired (requires structural changes)
- Manually add/correct `/Type /Catalog` entry

---

#### PDF-CATALOG-003: Catalog Missing /Pages Entry

**Severity:** Critical  
**Category:** STRUCTURE  
**Auto-Repairable:** No  
**Safety Level:** N/A

**Description:**  
The catalog dictionary is missing the required /Pages entry.

**Common Causes:**
- Incomplete catalog dictionary
- Missing page tree root
- Corrupted catalog
- Invalid PDF generation

**Specification Reference:**  
PDF 1.7, Section 7.7.2 - Document Catalog (must have /Pages)

**Example:**
```json
{
  "code": "PDF-CATALOG-003",
  "message": "Catalog missing /Pages entry",
  "severity": "critical",
  "details": {
    "repairable": false
  }
}
```

**Resolution:**
- Cannot be auto-repaired (requires page tree rebuild)
- Must regenerate from source or use PDF repair tools

---

### General Structure Errors (PDF-STRUCTURE-XXX)

#### PDF-STRUCTURE-012: General Structure Error

**Severity:** Error  
**Category:** STRUCTURE  
**Auto-Repairable:** Varies  
**Safety Level:** Varies

**Description:**  
A general PDF structure error that doesn't fit into more specific categories.

**Common Causes:**
- Various structural issues
- Object numbering conflicts
- Duplicate object/generation pairs
- Parser failures
- Unexpected structure violations

**Specification Reference:**  
PDF 1.7, Various sections

**Example:**
```json
{
  "code": "PDF-STRUCTURE-012",
  "message": "Duplicate object number/generation pair",
  "severity": "error",
  "details": {
    "object_number": 5,
    "generation": 0,
    "repairable": false
  }
}
```

**Resolution:**
- Depends on specific error (see details field)
- May require manual intervention
- Use appropriate PDF repair tools based on issue type

---

## Severity Levels

### Critical

**Definition:** File cannot be opened, parsed, or is fundamentally invalid.

**Characteristics:**
- Prevents basic file operations
- Cannot proceed with validation
- Immediate rejection required

**Examples:**
- Invalid file format (not a ZIP for EPUB, not a PDF)
- Missing essential structure (no catalog, no header)
- Severe corruption preventing parsing

**Handling:**
```go
if err.Severity == ebmlib.SeverityError && isCritical(err.Code) {
    return fmt.Errorf("critical error: cannot proceed")
}
```

### Error

**Definition:** Fails conformance to EPUB or PDF specifications; file is invalid.

**Characteristics:**
- Violates required specifications
- May prevent proper rendering or functionality
- Must be fixed for compliance
- Some may be auto-repairable

**Examples:**
- Missing required elements (TOC, mimetype)
- Invalid structure (malformed XML, wrong catalog type)
- Incorrect formats (compressed mimetype, wrong links)

**Handling:**
```go
if report.ErrorCount() > 0 {
    // Attempt repair or reject
    attemptRepair()
}
```

### Warning

**Definition:** Non-critical issue that should be reviewed but doesn't invalidate the file.

**Characteristics:**
- Conformance issue but file may still work
- Best practice violation
- May cause problems in some readers
- Strongly recommended to fix

**Examples:**
- Missing optional elements (landmarks)
- Deprecated features
- Suboptimal structure

**Handling:**
```go
if report.WarningCount() > 0 {
    log.Printf("Warnings: %d", report.WarningCount())
    // May proceed with caution
}
```

### Info

**Definition:** Informational message about file structure or potential improvements.

**Characteristics:**
- No specification violation
- Best practice suggestion
- Optimization opportunity
- Optional improvement

**Examples:**
- File size optimization suggestions
- Performance improvements
- Accessibility enhancements

**Handling:**
```go
if verbose && report.InfoCount() > 0 {
    log.Printf("Info: %d", report.InfoCount())
}
```

---

## Error Handling Guidelines

### 1. Check Severity Before Taking Action

```go
func handleValidationReport(report *ebmlib.ValidationReport) error {
    // Critical/Error - must fix
    if report.ErrorCount() > 0 {
        return attemptRepair(report)
    }
    
    // Warning - review but may proceed
    if report.WarningCount() > 0 {
        logWarnings(report.Warnings)
    }
    
    // Info - optional
    if report.InfoCount() > 0 && verbose {
        logInfo(report.Info)
    }
    
    return nil
}
```

### 2. Filter Errors by Category

```go
func getStructureErrors(report *ebmlib.ValidationReport) []ebmlib.ValidationError {
    var structErrors []ebmlib.ValidationError
    for _, err := range report.Errors {
        if strings.HasPrefix(err.Code, "EPUB-CONTAINER-") ||
           strings.HasPrefix(err.Code, "PDF-HEADER-") ||
           strings.HasPrefix(err.Code, "PDF-TRAILER-") {
            structErrors = append(structErrors, err)
        }
    }
    return structErrors
}
```

### 3. Check Repairability

```go
func canAutoFix(errors []ebmlib.ValidationError) bool {
    for _, err := range errors {
        if repairable, ok := err.Details["repairable"].(bool); ok && repairable {
            continue
        }
        return false // Found non-repairable error
    }
    return true
}
```

### 4. Log Errors with Context

```go
func logError(err ebmlib.ValidationError) {
    log.Printf("[%s] %s", err.Severity, err.Code)
    log.Printf("  Message: %s", err.Message)
    
    if err.Location != nil {
        log.Printf("  Location: %s:%d:%d", 
            err.Location.File, 
            err.Location.Line, 
            err.Location.Column)
    }
    
    if len(err.Details) > 0 {
        log.Printf("  Details: %v", err.Details)
    }
}
```

### 5. Provide User-Friendly Messages

```go
func getUserMessage(err ebmlib.ValidationError) string {
    messages := map[string]string{
        "EPUB-CONTAINER-002": "The EPUB mimetype file is invalid. This can be automatically repaired.",
        "PDF-TRAILER-003": "The PDF is missing its end marker. This can be automatically repaired.",
        "EPUB-NAV-001": "The navigation document has syntax errors and requires manual correction.",
        // ... more mappings
    }
    
    if msg, ok := messages[err.Code]; ok {
        return msg
    }
    return err.Message
}
```

---

## Appendix: Quick Reference Tables

### Auto-Repairable Errors

| Code | Category | Safety Level | Estimated Time |
|------|----------|--------------|----------------|
| EPUB-CONTAINER-002 | Mimetype | Very High | <1s |
| EPUB-CONTAINER-003 | ZIP Order | High | 2-10s |
| EPUB-CONTAINER-004 | Container XML | High* | <1s |
| EPUB-NAV-003 | TOC Structure | High | <1s |
| EPUB-NAV-004 | Links | High | 1-5s |
| PDF-TRAILER-001 | Startxref | High | <1s |
| PDF-TRAILER-003 | EOF Marker | Very High | <1s |

\* Conditional on package document location

### Non-Repairable Errors (Require Manual Intervention)

| Code | Reason | Recommended Action |
|------|--------|-------------------|
| EPUB-CONTAINER-001 | ZIP corruption | Re-download or use ZIP repair tools |
| EPUB-NAV-001 | XHTML syntax | Fix manually in XML editor |
| PDF-HEADER-001 | Invalid format | Verify file type, re-export |
| PDF-HEADER-002 | Version incompatible | Export with PDF 1.7 or lower |
| PDF-XREF-001 | Structural damage | Use QPDF or Adobe Acrobat |
| PDF-CATALOG-XXX | Document root issues | Regenerate from source |

### Error Code Prefixes

| Prefix | Format | Category |
|--------|--------|----------|
| EPUB-CONTAINER- | EPUB | OCF structure |
| EPUB-NAV- | EPUB | Navigation documents |
| EPUB-OPF- | EPUB | Package documents |
| EPUB-CONTENT- | EPUB | Content documents |
| PDF-HEADER- | PDF | File header |
| PDF-TRAILER- | PDF | File trailer |
| PDF-XREF- | PDF | Cross-reference |
| PDF-CATALOG- | PDF | Document catalog |
| PDF-STRUCTURE- | PDF | General structure |

---

**For more information:**
- User Guide: See `docs/USER_GUIDE.md`
- Architecture: See `docs/ARCHITECTURE.md`
- EPUB Specification: `docs/specs/ebm-lib-EPUB-SPEC.md`
- PDF Specification: `docs/specs/ebm-lib-PDF-SPEC.md`
