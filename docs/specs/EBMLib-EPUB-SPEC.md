# EBMLib System Requirements Specification – EPUB Validation Module

**Project Name**: EBMLib  
**Module Focus**: EPUB Validation and Repair  
**Description**: EBMLib is a multi-platform library (usable in CLIs, web apps, mobile apps, and desktop UIs) that provides comprehensive tools for working with EPUB files. The EPUB module emphasizes **validation** and, where possible, **automatic repair** of invalid EPUBs.  
**Status**: Draft / Implementation Phase  
**Last Updated**: December 28, 2025  

## 1. Product Vision for EPUB Functionality

The core library **must** deliver a full suite of EPUB validation features, with at least the coverage of **epubcheck** (current version 5.x for EPUB 3.3).  
Key goals:  
- Detect and report validation errors clearly  
- Attempt **non-destructive repairs** where safe and possible (e.g., fix well-formedness, add missing DOCTYPE, normalize paths)  
- Target **EPUB 3.0+** (maximum compatibility with EPUB 3.3 as of late 2025)  

## 2. Target EPUB Standard

- **Primary**: EPUB 3.3 (W3C Recommendation, published March 27, 2025)  
- **Backward Compatibility**: Must validate EPUB 3.x files; optional legacy support for EPUB 2.0.1 (via ADR decision)  
- **Authoritative Specification Source**:  
  - Official HTML: https://www.w3.org/TR/epub-33/  
  - Editor's Draft: https://w3c.github.io/epub-specs/epub33/core/  
  - GitHub Repo: https://github.com/w3c/epub-specs/tree/main/epub33  

> **Note**: Copy the full spec by opening the official HTML page, selecting all (Ctrl+A), and copying into a text editor for reference.

## 3. Required Validation Checks (Minimum epubcheck Parity)

The library **must** validate and report issues in these core areas:

### 3.1 Container Format (OCF – ZIP Archive)
- File extension: `.epub`  
- Must be a **valid ZIP archive** (not RAR, 7z, etc.)  
  - Use standard Go `archive/zip` (custom Swift `ZipArchive` only for non-core components)  
- **First file**: `mimetype`  
  - Exact content: `application/epub+zip` (no whitespace, no line breaks)  
  - Must be **uncompressed** and **first in archive**  
- Required file: `META-INF/container.xml`  
  - Must exist and be non-empty  
  - Must be valid XML  
  - Must contain at least one `<rootfile>` pointing to a `.opf` package document  

### 3.2 Package Document (content.opf)
- Must be valid XML  
- Required attributes: `version="3.3"`, `unique-identifier`  
- **Metadata** (minimum):  
  - `<dc:title>`  
  - `<dc:identifier id="...">` (matches `unique-identifier`)  
  - `<dc:language>`  
  - `<meta property="dcterms:modified">YYYY-MM-DD</meta>`  
- **Manifest**:  
  - At least one item for content document (`application/xhtml+xml`)  
  - Navigation document item with `properties="nav"`  
- **Spine**: At least one `<itemref>`  
- All `href` paths: **relative**, URL-encoded, correct case  

### 3.3 Content Documents (XHTML)
- Must be **well-formed XML**  
- Must include **HTML5 DOCTYPE**: `<!DOCTYPE html>`  
- Required elements: `<html>`, `<head>`, `<body>`  
- Valid namespaces: `xmlns="http://www.w3.org/1999/xhtml"`  
- Charset: `<meta charset="utf-8"/>` recommended  

### 3.4 Navigation Document (usually `nav.xhtml`)
- **Required** in EPUB 3+  
- Must be listed in manifest with `properties="nav"`  
- Must contain:  
  - `<nav epub:type="toc">` (Table of Contents with nested `<ol>`)  
  - `<nav epub:type="landmarks">` (strongly recommended)  
- Links must use relative paths and point to valid spine items  

### 3.5 General File & Structure Rules
- All XML files: well-formed, proper entities, UTF-8 encoding  
- File/folder names: ASCII preferred; avoid spaces/special chars  
- Images: ≤ 5.6 million pixels each (enforced by Apple Books and others)  
- No DRM in container (external only)  
- No prohibited features: LZW compression, external references (unless allowed with fallbacks)  

## 4. Minimal Valid EPUB 3.3 Skeleton (Reference)

```
mimetype
META-INF/
└── container.xml
OEBPS/  (or any folder)
 ├── content.opf
 └── xhtml/
     ├── document.xhtml
     └── nav.xhtml
```

### 4.1 `mimetype`
```
application/epub+zip
```

### 4.2 `META-INF/container.xml`
```xml
<?xml version="1.0" encoding="UTF-8"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>
```

### 4.3 `content.opf` (minimal)
```xml
<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" unique-identifier="bookid" version="3.3">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:dcterms="http://purl.org/dc/terms/">
    <dc:title>Minimal EPUB Example</dc:title>
    <dc:identifier id="bookid">urn:uuid:123e4567-e89b-12d3-a456-426614174000</dc:identifier>
    <dc:language>en</dc:language>
    <meta property="dcterms:modified">2025-12-28</meta>
  </metadata>

  <manifest>
    <item id="doc" href="xhtml/document.xhtml" media-type="application/xhtml+xml"/>
    <item id="nav" href="xhtml/nav.xhtml" media-type="application/xhtml+xml" properties="nav"/>
  </manifest>

  <spine>
    <itemref idref="doc"/>
  </spine>
</package>
```

### 4.4 `xhtml/document.xhtml` (minimal)
```html
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" lang="en" xml:lang="en">
  <head>
    <meta charset="utf-8"/>
    <title>Minimal EPUB</title>
  </head>
  <body>
    <h1>Hello, World!</h1>
    <p>This is the smallest possible content document.</p>
  </body>
</html>
```

## 5. Repair Capabilities (Where Safe)

The library **should** offer optional repair modes:  
- Add missing DOCTYPE  
- Insert `<html>`, `<head>`, `<body>` if absent  
- Normalize `mimetype` whitespace  
- Fix relative paths in OPF  
- Generate minimal `nav.xhtml` if missing  
- **Never** alter content without user confirmation or dry-run mode  

## 6. Validation Checklist (Quick Reference)

| Check                    | Required?     | Notes |
| ------------------------ | ------------- | ----- |
| `.epub` extension        | Yes           |       |
| Valid ZIP                | Yes           |       |
| `mimetype` first & exact | Yes           |       |
| `container.xml` valid    | Yes           |       |
| At least one `.opf`      | Yes           |       |
| Minimal metadata         | Yes           |       |
| At least one XHTML       | Yes           |       |
| Navigation document      | Yes (EPUB 3+) |       |
| All XML well-formed      | Yes           |       |
| Passes epubcheck         | Target        |       |

## 7. EPUB 3.3 Specification Summary (High-Level Structure)

1. **Introduction** – Terminology, conformance  
2. **Publication Conformance** – High-level rules  
3. **Publication Resources** – Media types, fallbacks  
4. **OCF Container** – ZIP, `mimetype`, `container.xml`  
5. **Package Document** – `.opf` structure  
6. **Content Documents** – XHTML, SVG, scripting  
7. **Navigation Document** – TOC, landmarks  
8. **Media Overlays** – Audio sync  
9. **Accessibility** – WCAG integration  
10. **Appendices** – Vocabulary, examples  

## 8. Additional Resources

| Resource            | Link                                | Purpose              |
| ------------------- | ----------------------------------- | -------------------- |
| EPUBCheck           | https://github.com/w3c/epubcheck    | Validation reference |
| Accessibility 1.1   | https://www.w3.org/TR/epub-a11y-11/ | Accessibility rules  |
| Reading Systems 3.3 | https://www.w3.org/TR/epub-rs-33/   | Reader requirements  |

---



### 9. EPUB Repair Strategy Tables

The strategies focus on the **most common real-world EPUB validation issues** (based on epubcheck reports, Calibre behavior, and community practices as of late 2025).

### 

**General Principles for EPUB Repairs** (apply to all strategies):

- **Always** offer **dry-run/preview mode** — show proposed changes without modifying the original file
- Generate **repair report** (diff-like summary or list of changes)
- Create repaired file with suffix (e.g. `_repaired.epub`) or in a temp directory
- **Never** auto-alter semantic content (text, images, chapter order) without explicit user consent
- Log every attempted/skipped repair with reason
- Use ZIP manipulation carefully (Go's `archive/zip` + temp files recommended)
- Validate before & after with internal epubcheck-like rules or external epubcheck call (if integrated)

#### 9.1. Container & ZIP-Level Repairs

| Error / Issue                                   | Typical epubcheck Code(s) | Repair Strategy                                              | Safety Level | User Confirmation Needed? | Auto-repair Recommended? |
| ----------------------------------------------- | ------------------------- | ------------------------------------------------------------ | ------------ | ------------------------- | ------------------------ |
| Invalid or missing `mimetype` file              | OPF-001, Mimetype-related | Create/overwrite with exact single-line content: `application/epub+zip` (no extra whitespace/newline) | Very High    | No                        | Yes                      |
| `mimetype` not first in archive or compressed   | Container-related         | Rebuild ZIP: ensure `mimetype` is first entry, stored uncompressed (Store method) | High         | No                        | Yes                      |
| Corrupted ZIP structure (bad CRC, truncated)    | — (ZIP-level failure)     | Attempt ZIP repair using library like `github.com/klauspost/compress/zip` or external `unzip -FF` style logic | Medium       | Yes                       | Conditional (warn)       |
| Missing or empty `META-INF/container.xml`       | RSC-001, Container-001    | Create minimal valid container.xml pointing to default `OEBPS/content.opf` | High         | Yes                       | Yes (if path guessable)  |
| Wrong case or path in `container.xml` full-path | Container-related         | Normalize path (case-insensitive fix, remove leading `/`)    | High         | No                        | Yes                      |

#### 9.2. Package Document (`content.opf`) Repairs

| Error / Issue                                           | Typical epubcheck Code(s) | Repair Strategy                                              | Safety Level | User Confirmation Needed? | Auto-repair Recommended?   |
| ------------------------------------------------------- | ------------------------- | ------------------------------------------------------------ | ------------ | ------------------------- | -------------------------- |
| Missing required metadata (title, identifier, language) | OPF-xxx                   | Inject minimal valid values (e.g. title="Untitled", lang="en", generate UUID) | Medium       | Yes                       | Conditional (placeholders) |
| Invalid or missing `dcterms:modified` date              | OPF-028, OPF-029          | Add current date in ISO format                               | High         | No                        | Yes                        |
| Broken relative paths in manifest/spine (`href`)        | RSC-007, RSC-008          | Resolve & normalize paths (fix `../`, double slashes, case mismatches) | High         | No                        | Yes                        |
| Duplicate or missing IDs in manifest                    | OPF-xxx                   | Auto-generate unique IDs, fix references                     | Medium       | Yes                       | Conditional                |
| Missing navigation document (`properties="nav"`)        | OPF-xxx, NAV-001          | Create minimal `nav.xhtml` with basic toc/landmarks if none exists | Medium       | Yes                       | Conditional (heuristic)    |

#### 9.3. Content Documents (XHTML) Repairs

| Error / Issue                                              | Typical epubcheck Code(s) | Repair Strategy                                              | Safety Level | User Confirmation Needed? | Auto-repair Recommended? |
| ---------------------------------------------------------- | ------------------------- | ------------------------------------------------------------ | ------------ | ------------------------- | ------------------------ |
| Not well-formed XML (unclosed tags, invalid entities)      | RSC-005, HTM-001          | Use XML parser with auto-recovery (e.g. `golang.org/x/net/html`) + fix basic nesting | Medium       | Yes                       | Conditional (risky)      |
| Missing HTML5 DOCTYPE                                      | HTM-002                   | Prepend `<!DOCTYPE html>`                                    | Very High    | No                        | Yes                      |
| Missing structural elements (`<html>`, `<head>`, `<body>`) | HTM-xxx                   | Wrap content in minimal valid structure if raw HTML detected | High         | Yes                       | Conditional              |
| Wrong namespace or missing `lang`/`xml:lang`               | HTM-xxx                   | Add/correct `xmlns="http://www.w3.org/1999/xhtml" lang="en"` | High         | No                        | Yes                      |
| Invalid/escaped characters (e.g. &quot; in text)           | RSC-xxx                   | Decode/encode entities properly (common in converted files)  | Medium       | No                        | Yes                      |

#### 9.4. Navigation & Other Common Repairs

| Error / Issue                                   | Typical epubcheck Code(s) | Repair Strategy                                       | Safety Level | User Confirmation Needed? | Auto-repair Recommended? |
| ----------------------------------------------- | ------------------------- | ----------------------------------------------------- | ------------ | ------------------------- | ------------------------ |
| Missing or invalid `<nav epub:type="toc">`      | NAV-001, NAV-002          | Generate minimal TOC from spine order if possible     | Medium       | Yes                       | Conditional (basic)      |
| Broken internal links / href targets            | RSC-007, RSC-008          | Scan & fix relative paths to existing files           | High         | Yes                       | Yes                      |
| Overly large images (>5.6M pixels – Apple rule) | — (warning)               | Warn only (downscale risky & lossy)                   | —            | —                         | No                       |
| Legacy EPUB 2 (NCX) in EPUB 3 file              | —                         | Add EPUB 3 nav if missing, keep NCX for compatibility | High         | No                        | Yes                      |

### Implementation Recommendations for EBMLib

- **Repair engine** → Separate port/adapter: `RepairService` interface with `Preview()` and `Apply()` methods
- **Prioritize** → Start with container/mimetype (highest success rate, lowest risk)
- **Safety guardrails** → Max repair depth (e.g. 3 attempts), always keep original backup
- **Fallbacks** → For complex cases (structure tree rebuild, font issues) → output detailed guidance + recommend external tools (Calibre, Sigil)
- **Testing** → Use corpus of known broken EPUBs from community (e.g. epubcheck test suite + real corrupted samples)

This table covers ~80-90% of real-world EPUB repair scenarios encountered in practice (mimetype, paths, metadata, basic well-formedness).  
More invasive repairs (full HTML parsing & semantic fixing) should remain optional/advanced features.

