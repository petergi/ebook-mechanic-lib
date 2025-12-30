# EPUB Test Fixtures

This directory contains comprehensive EPUB test fixtures covering all validation error codes and edge cases.

## Quick Start

Generate all fixtures:

```bash
go run generate_fixtures.go
```

This creates the following directory structure:

```
epub/
├── valid/
│   ├── minimal.epub
│   ├── multiple_rootfiles.epub
│   ├── complex_nested.epub
│   ├── large_100_chapters.epub
│   └── large_500_chapters.epub
├── invalid/
│   ├── not_zip.epub
│   ├── corrupt_zip.epub
│   ├── wrong_mimetype.epub
│   ├── mimetype_not_first.epub
│   ├── mimetype_compressed.epub
│   ├── no_container.epub
│   ├── invalid_container_xml.epub
│   ├── no_rootfile.epub
│   ├── invalid_opf.epub
│   ├── missing_title.epub
│   ├── missing_identifier.epub
│   ├── missing_language.epub
│   ├── missing_modified.epub
│   ├── missing_nav_document.epub
│   ├── invalid_nav_document.epub
│   └── invalid_content_document.epub
└── edge_cases/
    └── large_10mb_plus.epub
```

## Fixture Details

### Valid EPUBs

#### minimal.epub
- **Purpose**: Baseline valid EPUB 3.0
- **Structure**: Single chapter with navigation
- **Size**: ~1 KB
- **Contents**:
  - mimetype (uncompressed, first)
  - META-INF/container.xml
  - OEBPS/content.opf (with all required metadata)
  - OEBPS/nav.xhtml (navigation document)
  - OEBPS/content.xhtml (single content document)

#### multiple_rootfiles.epub
- **Purpose**: Test handling of multiple rootfiles in container.xml
- **Structure**: Two OPF files referenced
- **Size**: ~2 KB
- **Note**: Valid per EPUB spec; validators should handle gracefully

#### complex_nested.epub
- **Purpose**: Test complex directory structures
- **Structure**: 
  - Nested directories: `content/xhtml/chapters/`
  - Separate directories for styles and images
  - Relative path references
- **Size**: ~3 KB
- **Files**: OPF, nav, 2 chapters, CSS, image (JPEG header)

#### large_100_chapters.epub
- **Purpose**: Performance testing - medium size
- **Structure**: 100 chapters with navigation
- **Size**: ~500 KB
- **Use**: Benchmark validation speed

#### large_500_chapters.epub
- **Purpose**: Performance testing - large size
- **Structure**: 500 chapters with navigation
- **Size**: ~2.5 MB
- **Use**: Test scalability

### Invalid EPUBs

Each invalid fixture targets a specific error code:

#### Container Errors (EPUB-CONTAINER-XXX)

**not_zip.epub**
- **Error**: EPUB-CONTAINER-001
- **Issue**: Plain text file, not a ZIP archive
- **Detection**: ZIP header validation fails

**corrupt_zip.epub**
- **Error**: EPUB-CONTAINER-001
- **Issue**: Truncated ZIP file (last 50 bytes removed)
- **Detection**: ZIP structure parsing fails

**wrong_mimetype.epub**
- **Error**: EPUB-CONTAINER-002
- **Issue**: mimetype contains "application/zip" instead of "application/epub+zip"
- **Detection**: mimetype content validation

**mimetype_not_first.epub**
- **Error**: EPUB-CONTAINER-003
- **Issue**: dummy.txt appears before mimetype in ZIP
- **Detection**: First file name check

**mimetype_compressed.epub**
- **Error**: EPUB-CONTAINER-002
- **Issue**: mimetype file stored with compression
- **Detection**: ZIP file method check (should be STORE, not DEFLATE)

**no_container.epub**
- **Error**: EPUB-CONTAINER-004
- **Issue**: Missing META-INF/container.xml
- **Detection**: Required file existence check

**invalid_container_xml.epub**
- **Error**: EPUB-CONTAINER-005
- **Issue**: container.xml is malformed XML ("<invalid xml")
- **Detection**: XML parsing error

**no_rootfile.epub**
- **Error**: EPUB-CONTAINER-005
- **Issue**: container.xml has empty <rootfiles> element
- **Detection**: Rootfile count validation

#### OPF Errors (EPUB-OPF-XXX)

**invalid_opf.epub**
- **Error**: EPUB-OPF-001
- **Issue**: content.opf is malformed XML
- **Detection**: XML parsing error

**missing_title.epub**
- **Error**: EPUB-OPF-002
- **Issue**: OPF metadata lacks dc:title
- **Detection**: Required metadata validation

**missing_identifier.epub**
- **Error**: EPUB-OPF-003
- **Issue**: OPF metadata lacks dc:identifier
- **Detection**: Required metadata validation

**missing_language.epub**
- **Error**: EPUB-OPF-004
- **Issue**: OPF metadata lacks dc:language
- **Detection**: Required metadata validation

**missing_modified.epub**
- **Error**: EPUB-OPF-005
- **Issue**: OPF metadata lacks meta property="dcterms:modified"
- **Detection**: Required metadata validation (EPUB 3.0 requirement)

**missing_nav_document.epub**
- **Error**: EPUB-OPF-009
- **Issue**: OPF manifest has no item with properties="nav"
- **Detection**: Manifest navigation item validation

#### Navigation Errors (EPUB-NAV-XXX)

**invalid_nav_document.epub**
- **Error**: EPUB-NAV-006
- **Issue**: nav.xhtml exists but contains no <nav> element
- **Detection**: Navigation element presence check

#### Content Document Errors (EPUB-CONTENT-XXX)

**invalid_content_document.epub**
- **Error**: EPUB-CONTENT-002
- **Issue**: content.xhtml missing DOCTYPE and xmlns
- **Detection**: XHTML structure validation

### Edge Cases

**edge_cases/large_10mb_plus.epub**
- **Purpose**: Stress testing with very large file
- **Structure**: 2000+ chapters
- **Size**: ~10+ MB
- **Use**: Memory and performance limits testing

## Test Coverage Matrix

| Error Code | Fixture | Test Status |
|------------|---------|-------------|
| EPUB-CONTAINER-001 | not_zip.epub, corrupt_zip.epub | ✅ |
| EPUB-CONTAINER-002 | wrong_mimetype.epub, mimetype_compressed.epub | ✅ |
| EPUB-CONTAINER-003 | mimetype_not_first.epub | ✅ |
| EPUB-CONTAINER-004 | no_container.epub | ✅ |
| EPUB-CONTAINER-005 | invalid_container_xml.epub, no_rootfile.epub | ✅ |
| EPUB-OPF-001 | invalid_opf.epub | ✅ |
| EPUB-OPF-002 | missing_title.epub | ✅ |
| EPUB-OPF-003 | missing_identifier.epub | ✅ |
| EPUB-OPF-004 | missing_language.epub | ✅ |
| EPUB-OPF-005 | missing_modified.epub | ✅ |
| EPUB-OPF-009 | missing_nav_document.epub | ✅ |
| EPUB-NAV-006 | invalid_nav_document.epub | ✅ |
| EPUB-CONTENT-002 | invalid_content_document.epub | ✅ |

## Integration Testing

Tests are located in `../../tests/integration/epub_validator_integration_test.go`:

```go
// Table-driven test covering all error codes
func TestEPUBValidatorIntegration_TableDriven_AllErrorCodes(t *testing.T)

// Valid file tests
func TestEPUBValidatorIntegration_ValidFiles(t *testing.T)

// Performance tests
func TestEPUBValidatorIntegration_PerformanceLargeFiles(t *testing.T)

// Edge case tests
func TestEPUBValidatorIntegration_EdgeCases(t *testing.T)
```

## EPUB Specification Compliance

Fixtures are designed to comply with or intentionally violate:

- **EPUB 3.3 Specification** (latest)
- **OCF (Open Container Format) 3.0**
- **OPF (Open Packaging Format) 3.0**

Key validation points:
- Container structure (ZIP format)
- mimetype file requirements (uncompressed, first, exact content)
- META-INF/container.xml structure
- OPF metadata requirements (title, identifier, language, modified)
- Navigation document requirements
- XHTML content document structure

## Extending Fixtures

To add a new fixture:

1. Create a function in `generate_fixtures.go`:
   ```go
   func createEPUBNewCase() []byte {
       buf := new(bytes.Buffer)
       zipWriter := zip.NewWriter(buf)
       // ... build EPUB structure
       zipWriter.Close()
       return buf.Bytes()
   }
   ```

2. Add to the fixtures map:
   ```go
   fixtures := map[string][]byte{
       // ...
       "invalid/new_case.epub": createEPUBNewCase(),
   }
   ```

3. Regenerate: `go run generate_fixtures.go`

4. Add test case in `../../tests/integration/epub_validator_integration_test.go`

## Validation with epubcheck

To cross-validate fixtures with the reference validator:

```bash
# Install epubcheck (Java required)
# Download from https://github.com/w3c/epubcheck/releases

# Validate a fixture
java -jar epubcheck.jar valid/minimal.epub

# Should output: No errors or warnings detected
```

Compare results with our validator to ensure alignment with the standard.
