# EPUB Accessibility Validator

The EPUB Accessibility Validator provides comprehensive accessibility compliance checking for EPUB content documents, supporting WCAG 2.1 (Level A and AA) and EPUB Accessibility 1.1 standards.

## Features

### Validation Checks

1. **Language Declarations**
   - Validates presence of `lang` and `xml:lang` attributes
   - Checks language code validity (ISO 639)

2. **Semantic HTML5 Structure**
   - Detects use of semantic elements (`<article>`, `<section>`, `<nav>`, `<header>`, `<footer>`, `<aside>`, `<main>`)
   - Encourages meaningful document structure

3. **ARIA Attributes and Roles**
   - Validates ARIA role values against WAI-ARIA specification
   - Checks for proper ARIA attribute usage
   - Ensures required labels are present for specific roles
   - Detects invalid or unknown ARIA attributes

4. **Image Alternative Text**
   - Ensures all images have `alt` attributes
   - Identifies decorative images with proper markup
   - Tracks completeness statistics

5. **Heading Hierarchy**
   - Validates heading structure (h1-h6)
   - Detects skipped heading levels
   - Identifies empty headings
   - Ensures documents start with h1

6. **Reading Order**
   - Detects disruptive positive `tabindex` values
   - Validates natural reading order preservation

7. **Table Accessibility**
   - Ensures data tables have proper headers (`<th>`)
   - Validates `headers` attribute usage
   - Handles presentation tables appropriately

8. **Form Accessibility**
   - Validates form control labels
   - Checks for `aria-label` or `<label>` associations
   - Ensures all interactive elements are properly labeled

9. **Media Elements**
   - Checks for captions/subtitles on video elements
   - Validates audio descriptions
   - Ensures alternative content for embedded media

10. **Landmark Regions**
    - Validates proper landmark usage
    - Ensures single main landmark per document
    - Encourages use of ARIA landmarks

## Accessibility Scoring System

The validator calculates a comprehensive score from 0-100:

| Component | Weight | Description |
|-----------|--------|-------------|
| Language Declaration | 5% | Presence and validity of lang attributes |
| Semantic Structure | 25% | Use of HTML5 semantic elements |
| ARIA Compliance | 20% | Proper ARIA roles and attributes |
| Alt Text Completeness | 25% | Images with appropriate alternative text |
| Heading Hierarchy | 15% | Proper heading structure without gaps |
| Reading Order | 10% | Natural reading order without disruption |

### Compliance Levels

- **90-100 (WCAG 2.1 AA):** Fully accessible, minimal to no errors
- **80-89 (WCAG 2.1 A):** Accessible with minor improvements needed
- **60-79 (Partial):** Some accessibility features, but significant issues remain
- **0-59 (Non-compliant):** Major accessibility barriers present

## Metadata Generation

The validator automatically generates EPUB Accessibility 1.1 metadata:

### Schema.org Properties

- `schema:accessMode`: textual, visual, auditory
- `schema:accessModeSufficient`: textual
- `schema:accessibilityFeature`: alternativeText, structuralNavigation, ARIA, tableOfContents
- `schema:accessibilityHazard`: none, flashing, sound, motionSimulation
- `schema:accessibilitySummary`: Human-readable summary with score and details

### Conformance Claims

- WCAG 2.1 Level A
- WCAG 2.1 Level AA
- EPUB Accessibility 1.1

## Usage

### Basic Validation

```go
import "github.com/petergi/ebook-mechanic-lib/internal/adapters/epub"

validator := epub.NewAccessibilityValidator()
result, err := validator.ValidateFile("content/chapter01.xhtml")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Accessibility Score: %d/100\n", result.Score.Total)
fmt.Printf("Compliance Level: %s\n", result.ComplianceLevel)
fmt.Printf("Valid: %v\n", result.Valid)
```

### Detailed Score Breakdown

```go
result, err := validator.ValidateBytes(contentData)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Score Breakdown:\n")
fmt.Printf("  Language: %d/%d\n", result.Score.LanguageDeclaration, 5)
fmt.Printf("  Semantic: %d/%d\n", result.Score.SemanticStructure, 25)
fmt.Printf("  ARIA: %d/%d\n", result.Score.ARIACompliance, 20)
fmt.Printf("  Alt Text: %d/%d\n", result.Score.AltTextCompleteness, 25)
fmt.Printf("  Headings: %d/%d\n", result.Score.HeadingHierarchy, 15)
fmt.Printf("  Reading Order: %d/%d\n", result.Score.ReadingOrder, 10)
```

### Error and Warning Reporting

```go
result, _ := validator.ValidateFile("content/chapter.xhtml")

// Process errors
for _, err := range result.Errors {
    fmt.Printf("ERROR [%s]: %s\n", err.Code, err.Message)
    if err.Details != nil {
        fmt.Printf("  Details: %+v\n", err.Details)
    }
}

// Process warnings
for _, warn := range result.Warnings {
    fmt.Printf("WARNING [%s]: %s\n", warn.Code, warn.Message)
}
```

### Metadata for Package Document

```go
result, _ := validator.ValidateFile("content/chapter.xhtml")

fmt.Println("Accessibility Metadata for OPF:")
fmt.Printf("Conformance: %v\n", result.Metadata.ConformanceClaims)
fmt.Printf("Features: %v\n", result.Metadata.AccessibilityFeatures)
fmt.Printf("Access Modes: %v\n", result.Metadata.AccessModes)
fmt.Printf("Summary: %s\n", result.Metadata.AccessibilitySummary)
```

### Validation with Reading Order Context

```go
spineOrder := []string{"chapter01.xhtml", "chapter02.xhtml", "chapter03.xhtml"}
result, err := validator.ValidateWithContext(reader, spineOrder)
```

## Error Codes

All accessibility error codes follow the pattern `EPUB-A11Y-XXX`:

| Code | Severity | Description |
|------|----------|-------------|
| EPUB-A11Y-001 | Error | Missing language declaration |
| EPUB-A11Y-002 | Warning | Invalid language code |
| EPUB-A11Y-003 | Warning | Missing semantic structure |
| EPUB-A11Y-004 | Warning | Invalid heading hierarchy |
| EPUB-A11Y-005 | Error | Missing alt text |
| EPUB-A11Y-006 | Warning | Empty alt text |
| EPUB-A11Y-007 | Error | Invalid ARIA role |
| EPUB-A11Y-008 | Warning | Invalid ARIA attribute |
| EPUB-A11Y-009 | Error | Missing ARIA label |
| EPUB-A11Y-010 | Warning | Invalid reading order |
| EPUB-A11Y-011 | Error | Missing table headers |
| EPUB-A11Y-012 | Error | Invalid table structure |
| EPUB-A11Y-013 | Error | Missing form labels |
| EPUB-A11Y-014 | Warning | Insufficient contrast |
| EPUB-A11Y-015 | Warning | Media missing alternative |
| EPUB-A11Y-016 | Error | Media overlay sync issue |
| EPUB-A11Y-017 | Warning | Missing skip links |
| EPUB-A11Y-018 | Error/Warning | Invalid landmarks |
| EPUB-A11Y-019 | Error | Empty heading |
| EPUB-A11Y-020 | Error | Skipped heading level |

## Standards Compliance

### WCAG 2.1

The validator checks compliance with Web Content Accessibility Guidelines 2.1:

- **Level A** (minimum): Basic accessibility features
- **Level AA** (recommended): Enhanced accessibility for broader audience
- **Level AAA** (optional): Highest level of accessibility

### EPUB Accessibility 1.1

Aligns with EPUB Accessibility 1.1 specification:

- Discovery metadata in package document
- Conformance to WCAG 2.1
- Distribution of accessibility metadata
- Page navigation
- MathML and other specialized content

## Best Practices

### For Content Creators

1. **Always declare language**: Add `lang` and `xml:lang` to `<html>`
2. **Use semantic HTML**: Prefer `<nav>`, `<article>`, `<section>` over generic `<div>`
3. **Provide alt text**: Describe images meaningfully, use `alt=""` for decorative images
4. **Maintain heading hierarchy**: Don't skip levels (h1 → h2 → h3)
5. **Label forms**: Associate all inputs with labels
6. **Use ARIA wisely**: Enhance, don't replace native HTML semantics
7. **Avoid positive tabindex**: Let natural DOM order determine focus

### For Publishers

1. **Validate early**: Check accessibility during content creation
2. **Document accessibility features**: Use generated metadata in OPF
3. **Test with assistive technology**: Validate with screen readers
4. **Provide accessibility statements**: Document conformance claims
5. **Support incremental improvements**: Use partial compliance reporting

## Partial Compliance Reporting

The validator supports incremental accessibility improvements by:

- Providing detailed score breakdowns by category
- Separating errors from warnings
- Generating metadata reflecting current state
- Allowing tracking of improvements over time

## Integration with EPUB Validation

The accessibility validator integrates with the main EPUB validator:

```go
epubValidator := epub.NewEPUBValidator()
accessibilityValidator := epub.NewAccessibilityValidator()

// Validate entire EPUB
epubResult, _ := epubValidator.ValidateFile(ctx, "book.epub")

// Validate specific content document for accessibility
a11yResult, _ := accessibilityValidator.ValidateFile("OEBPS/chapter01.xhtml")
```

## Performance Considerations

- Parses HTML once per document
- Efficient tree traversal algorithms
- Minimal memory footprint
- Suitable for batch processing of large EPUBs

## Future Enhancements

- Color contrast ratio calculation
- CSS accessibility checks
- MathML accessibility validation
- SVG accessibility validation
- Extended media overlay synchronization checks
- Automated repair suggestions
