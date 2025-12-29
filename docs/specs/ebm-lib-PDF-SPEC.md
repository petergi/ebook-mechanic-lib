# **ebm-lib Master PDF Specifications**

**Project Name**: ebm-lib  
**Focus**: PDF Validation, Archival (PDF/A), and Accessibility (PDF/UA) Modules  
**Description**: Comprehensive reference for PDF handling in ebm-lib, covering basic well-formed PDF 1.7, long-term archival (PDF/A family), and universal accessibility (PDF/UA family).  
**Status**: Phased Implementation (Basic PDF 1.7 → PDF/A → PDF/UA)  
**Last Updated**: December 28, 2025  
**Target Language**: Go (Golang) – Hexagonal Architecture  

## 1. Overall Product Vision for PDF Functionality

ebm-lib's PDF module provides reliable, standards-compliant validation and limited safe repair capabilities across three progressive layers:

- **Layer 1 (Phase 1)**: Basic well-formed PDF 1.7 – ensures the file is syntactically correct and opens in standard readers  
- **Layer 2 (Phase 2)**: PDF/A archival conformance – long-term preservation & self-containment  
- **Layer 3 (Phase 3)**: PDF/UA accessibility conformance – universal access for users with disabilities  

All layers build on the same parser foundation and share consistent error reporting (code, message, severity, location/page/object).

**Authoritative References** (use as single source of truth):
- Adobe PDF 1.7 Reference: https://www.adobe.com/go/pdfreference  
- ISO 32000-2:2020 (PDF 2.0): Basis for modern features  
- veraPDF: Behavioral oracle for PDF/A & PDF/UA validation  
- Local project file: `/Users/petergiannopoulos/Documents/Projects/Personal/Active/ebm-lib/docs/specs/ebm-lib-PDF-SPEC.md`

## 2. Target PDF Standards Summary

| Layer | Primary Target                  | Base PDF Version | ISO Standard          | Status (Dec 2025)         | Priority |
|-------|---------------------------------|------------------|-----------------------|---------------------------|----------|
| 1     | Well-formed PDF 1.7             | PDF 1.7          | ISO 32000-1:2008      | Universally supported     | Phase 1 – Must Have |
| 2     | PDF/A archival                  | PDF 2.0 (A-4)    | ISO 19005-4:2020      | Current & recommended     | Phase 2 |
| 2     | Legacy archival compatibility   | PDF 1.7          | PDF/A-2b / PDF/A-3b   | Still widely required     | Phase 2 |
| 3     | PDF/UA accessibility            | PDF 1.7          | ISO 14289-1:2014      | Dominant legacy           | Phase 3 |
| 3     | Modern accessibility            | PDF 2.0          | ISO 14289-2:2024      | Current gold standard (published March 15, 2024) | Phase 3 |

## 3. Layer 1 – Basic Well-Formed PDF 1.7 Validation

**Goal**: File is syntactically correct and opens/render correctly in major readers.

### Key Checks
- Header: `%PDF-1.x` (x=0–7)  
- Trailer: Valid `%%EOF`, `startxref` pointing to xref  
- Cross-reference: Valid table or stream, no overlaps  
- Objects: Correct numbering/generation, balanced delimiters  
- Catalog: `/Type /Catalog`, reachable, with `/Pages`  
- Fonts/Resources: Embedded/subset fonts recommended (warn if missing)  
- Common corruptions: Truncated file, invalid object streams, password-locked (report as locked)

### Repair Capabilities (Safe Only)
- Append missing `%%EOF`  
- Fix minor trailer typos  
- Recompute `startxref`  
- Basic linear scan for damaged xref

## 4. Layer 2 – PDF/A Archival Validation

**Goal**: Long-term preservation – self-contained, visually stable forever.

### Supported Profiles (Recommended Order)
| Profile     | Base      | Conformance Levels | Key Features                                      | Recommendation (2025+) |
|-------------|-----------|--------------------|---------------------------------------------------|------------------------|
| PDF/A-4     | PDF 2.0   | Default / 4f / 4e  | Modern, simplified, optional tagging              | New archives (primary) |
| PDF/A-2b    | PDF 1.7   | b (basic)          | Transparency, layers, JPEG2000                    | High compatibility     |
| PDF/A-3b    | PDF 1.7   | b                  | + Arbitrary attachments                           | Invoices/hybrids       |
| PDF/A-1b    | PDF 1.4   | b                  | Very strict, legacy                               | Only if required       |

### Core PDF/A Requirements (All Versions)
- All fonts fully embedded/subsetted  
- Device-independent color (embedded ICC or allowed Device*)  
- No prohibited compression (LZW forbidden)  
- Valid XMP PDF/A identifier (`pdfaid:Part`, `pdfaid:Conformance`)  
- No JavaScript, no audio/video/3D (except 4e), no encryption blocking viewing  

### Common Failures to Detect
- Missing/invalid XMP identifier  
- Non-embedded fonts  
- Prohibited filters  
- Incorrect color spaces  
- Attachments not PDF/A (in A-1/2)

## 5. Layer 3 – PDF/UA Accessibility Validation

**Goal**: Universal access via screen readers, keyboard, braille, etc. (complements WCAG 2.x).

### Supported Profiles
| Profile     | Base      | Key Enhancements (vs previous)                          | Recommendation (2025+) |
|-------------|-----------|----------------------------------------------------------|------------------------|
| PDF/UA-2    | PDF 2.0   | New tags (DocumentFragment, Em, Strong, FENote), MathML, namespaces | Modern / future-proof (primary) |
| PDF/UA-1    | PDF 1.7   | Tagged PDF, Alt text, Unicode mapping, logical structure | Legacy compatibility   |

### Core PDF/UA Requirements
- **Tagged PDF**: Full logical structure tree (mandatory)  
- **Semantic correctness**: Tags match meaning, proper nesting  
- **Alternative text**: Every image/figure has Alt/ActualText  
- **Unicode mapping**: All text via ToUnicode or embedded fonts  
- **Reading order**: Structure matches natural order  
- **Language & metadata**: Document + sections tagged with Lang, proper XMP (PDF/UA identifier + `pdfuaid:rev` 2024 for UA-2)  
- **Annotations/Forms**: Visible, tab order = structure order, labeled  

### Common Failures to Detect
- Missing Alt text  
- Incorrect tag hierarchy/nesting  
- Untagged content (real content marked Artifact)  
- Missing language tags  
- Non-embedded fonts / no Unicode maps  
- Invalid annotation handling  

## 6. Implementation Roadmap & Priorities

| Phase | Focus                              | Target Standards                  | Key Milestones                              | Dependencies |
|-------|------------------------------------|-----------------------------------|---------------------------------------------|--------------|
| 1     | Basic well-formedness              | PDF 1.7                           | Header/trailer/xref/catalog validation      | Parser choice (unipdf/pdfcpu) |
| 2     | Archival conformance               | PDF/A-4 + PDF/A-2b/3b             | XMP identifier, font embedding, compression | Phase 1       |
| 3     | Accessibility conformance          | PDF/UA-2 (primary) + PDF/UA-1     | Tagged structure, Alt text, language tags   | Phase 1+2     |

## 7. Recommended Tools & References

- **Parser Libraries**: `github.com/unidoc/unipdf` or `github.com/pdfcpu/pdfcpu`  
- **Validation Oracle**: veraPDF (gold standard for PDF/A & PDF/UA)  
- **Free Resources**: PDF Association (no-cost ISO 32000-2, 14289-2, 19005-4)  
- **Error Style**: Structured (code, severity, location, message) matching EPUB module  


## 8. Sample Error Formats

All validation errors **must** follow this unified, structured format.  
This ensures consistent output for CLI, API, logging, and UI consumers.

### 8.1 Core Error Structure (JSON-like / Go struct representation)

```go
type ValidationError struct {
    Code           string             // Unique error code (e.g. "PDF-HEADER-001")
    Severity       string             // "CRITICAL", "ERROR", "WARNING", "INFO"
    Message        string             // Human-readable description
    Details        string             // Optional extended explanation
    Location       ErrorLocation      // Where the problem occurs
    Category       string             // "STRUCTURE", "METADATA", "FONTS", "TAGS", "ACCESSIBILITY", etc.
    Standard       string             // e.g. "PDF 1.7", "PDF/A-4", "PDF/UA-2"
    Conformance    string             // e.g. "PDF/A-2b", "PDF/UA-1"
    Repairable     bool               // true if library can suggest/apply safe fix
    SuggestedFix   string             // Short description of possible repair (if Repairable=true)
}
```

```go
type ErrorLocation struct {
    PageNumber     int                // 1-based page number (0 if not page-specific)
    ObjectID       string             // e.g. "12 0 obj", "xref stream"
    Offset         int64              // Byte offset in file (optional)
    XPath          string             // For structure tree / XMP (e.g. "/StructTreeRoot/K[3]/K[1]")
    Element        string             // e.g. "Catalog", "Figure", "XMP packet"
}
```

### 8.2 Severity Levels

| Level      | Meaning                                      | Action Recommendation                     |
|------------|----------------------------------------------|-------------------------------------------|
| CRITICAL   | File cannot be opened/parsed at all          | Immediate rejection required              |
| ERROR      | Fails conformance – not archival/accessible  | Must be fixed for compliance              |
| WARNING    | Conformance issue but may still be usable    | Strongly recommended to fix               |
| INFO       | Best practice violation / potential issue    | Optional improvement                      |

### 8.3 Sample Errors – Layer 1 (Basic PDF 1.7)

```json
{
  "code": "PDF-HEADER-001",
  "severity": "CRITICAL",
  "message": "Invalid or missing PDF header",
  "details": "File does not start with '%PDF-1.' pattern",
  "location": {
    "pageNumber": 0,
    "offset": 0
  },
  "category": "STRUCTURE",
  "standard": "PDF 1.7",
  "repairable": false
}
```

```json
{
  "code": "PDF-TRAILER-003",
  "severity": "ERROR",
  "message": "Missing or invalid %%EOF marker",
  "details": "File ends without '%%EOF' on its own line",
  "location": {
    "offset": 1245678
  },
  "category": "STRUCTURE",
  "standard": "PDF 1.7",
  "repairable": true,
  "suggestedFix": "Append '%%EOF' at end of file"
}
```

### 8.4 Sample Errors – Layer 2 (PDF/A Archival)

```json
{
  "code": "PDFA-META-001",
  "severity": "ERROR",
  "message": "Missing PDF/A identifier in XMP metadata",
  "details": "No pdfaid:Part / pdfaid:Conformance in XMP packet",
  "location": {
    "element": "XMP packet"
  },
  "category": "METADATA",
  "standard": "PDF/A-4",
  "conformance": "PDF/A-4",
  "repairable": true,
  "suggestedFix": "Add proper PDF/A XMP schema"
}
```

```json
{
  "code": "PDFA-FONT-002",
  "severity": "ERROR",
  "message": "Non-embedded font used",
  "details": "Font 'Helvetica' is not embedded or subsetted",
  "location": {
    "pageNumber": 3,
    "objectID": "45 0 obj"
  },
  "category": "FONTS",
  "standard": "PDF/A-2b",
  "conformance": "PDF/A-2b",
  "repairable": false
}
```

### 8.5 Sample Errors – Layer 3 (PDF/UA Accessibility)

```json
{
  "code": "PDFUA-TAG-001",
  "severity": "ERROR",
  "message": "Image without alternative text",
  "details": "Figure tag missing Alt or ActualText entry",
  "location": {
    "pageNumber": 5,
    "xpath": "/StructTreeRoot/K[2]/K[1]"
  },
  "category": "ACCESSIBILITY",
  "standard": "PDF/UA-2",
  "conformance": "PDF/UA-2",
  "repairable": true,
  "suggestedFix": "Add placeholder Alt text and prompt user for real description"
}
```

```json
{
  "code": "PDFUA-ORDER-003",
  "severity": "WARNING",
  "message": "Reading order does not match logical structure",
  "details": "Content sequence differs from structure tree order",
  "location": {
    "pageNumber": 1
  },
  "category": "ACCESSIBILITY",
  "standard": "PDF/UA-1",
  "conformance": "PDF/UA-1",
  "repairable": false
}
```

### 8.6 Recommended Implementation Notes

- Use a central error registry (map[string]ErrorTemplate) for codes & default messages  

- Support JSON output for API/CLI, human-readable text for logs  

- Include `errorCode` in all panic/recover scenarios  

- Allow filtering by severity/category/standard in validation reports  

- Align codes with veraPDF error naming where possible (e.g. prefix "PDFA-", "PDFUA-")  

  

## 9. Repair Strategy Tables



### 9.1 Layer 1 – Basic Well-Formed PDF 1.7 Repairs

| Error Code Prefix | Issue Description                          | Repair Strategy                                      | Safety Level | User Confirmation Needed? | Repairable Automatically? |
|-------------------|--------------------------------------------|------------------------------------------------------|--------------|----------------------------|----------------------------|
| PDF-HEADER-       | Missing/invalid %PDF- header               | Prepend correct %PDF-1.7 header + version comment   | Very High    | Yes (changes file start)   | No                         |
| PDF-TRAILER-      | Missing %%EOF marker                       | Append %%EOF + final newline                         | Very High    | No                         | Yes                        |
| PDF-TRAILER-      | Incorrect startxref value                  | Recompute cross-reference offset                     | High         | No                         | Yes                        |
| PDF-XREF-         | Truncated or damaged xref table/stream     | Attempt linear scan + rebuild index (fallback mode) | Medium       | Yes                        | Conditional (warn)         |
| PDF-OBJECT-       | Minor syntax errors in trailer dictionary  | Correct known typos (e.g. /Size value mismatch)     | High         | No                         | Yes                        |
| PDF-ENCRYPT-      | Password-protected (viewer-locked)         | No automatic repair (remove encryption unsafe)      | —            | —                          | No                         |

### 9.2 Layer 2 – PDF/A Archival Repairs

| Error Code Prefix | Issue Description                              | Repair Strategy                                                  | Safety Level | User Confirmation Needed? | Repairable Automatically? |
|-------------------|------------------------------------------------|------------------------------------------------------------------|--------------|----------------------------|----------------------------|
| PDFA-META-        | Missing/invalid PDF/A XMP identifier           | Inject minimal valid PDF/A XMP packet (based on target level)   | High         | Yes                        | Conditional (template)     |
| PDFA-FONT-        | Non-embedded font                              | No safe auto-repair (requires font subsetting/re-embedding)     | —            | —                          | No                         |
| PDFA-COMPRESS-    | Prohibited compression (LZW, old JPEG)         | No safe repair (requires re-encoding content streams)           | —            | —                          | No                         |
| PDFA-COLOR-       | Device-dependent color without ICC profile     | Inject default sRGB ICC profile if safe to assume               | Medium       | Yes                        | Conditional                |
| PDFA-ATTACH-      | Non-PDF/A attachments (in PDF/A-1/2)           | Remove attachments (lossy) or convert to PDF/A if possible      | Low          | Yes                        | No                         |
| PDFA-JS-          | JavaScript present                             | Remove JavaScript actions (may affect forms/annotations)        | Medium       | Yes                        | Conditional                |

### 9.3 Layer 3 – PDF/UA Accessibility Repairs

| Error Code Prefix | Issue Description                              | Repair Strategy                                                  | Safety Level | User Confirmation Needed? | Repairable Automatically? |
|-------------------|------------------------------------------------|------------------------------------------------------------------|--------------|----------------------------|----------------------------|
| PDFUA-TAG-        | Image/figure missing Alt/ActualText            | Insert placeholder Alt text ("Image on page X") + log for review| High         | No (placeholder)           | Yes (with warning)         |
| PDFUA-TAG-        | Incorrect tag hierarchy / nesting              | Attempt minimal re-nesting (heading levels, lists)              | Medium       | Yes                        | Conditional                |
| PDFUA-ORDER-      | Reading order mismatches structure             | Set /Tabs = /S (structure order) on page if safe                 | High         | No                         | Yes                        |
| PDFUA-LANG-       | Missing document or section language tag       | Add /Lang entry in Catalog (default "en" or detected)           | High         | No                         | Yes                        |
| PDFUA-FONT-       | Text without Unicode mapping                   | No safe auto-repair (requires ToUnicode CMap generation)        | —            | —                          | No                         |
| PDFUA-ANNOT-      | Hidden or unlabeled annotations                | Make visible + add default Contents ("Annotation")              | Medium       | Yes                        | Conditional                |
| PDFUA-STRUCT-     | Untagged real content (marked as Artifact)     | Convert Artifact → Figure/P/Sect (heuristic-based)              | Low          | Yes                        | No                         |

### 9.4 General Repair Guidelines

- **Always** offer **dry-run mode** (`--dry-run` / `--preview`) that shows proposed changes without modifying the file  
- **Always** produce a **diff-like report** (before/after byte ranges or textual summary)  
- **Never** auto-repair anything that changes:  
  - Visible text content  
  - Image data  
  - Layout coordinates  
  - Form field values  
- **Log level**: Every repair attempt must be logged (even if skipped) with reason  
- **Output artifact**: Optional generation of repaired file with suffix (e.g. `_repaired.pdf`)  
- **Best practice**: For high-risk repairs (fonts, compression, structure tree), only offer guidance messages with external tool recommendations (veraPDF, Acrobat Preflight, CommonLook)

