# ebm-lib User Guide

**Version:** 1.0  
**Last Updated:** December 2025

## Table of Contents

1. [Introduction](#introduction)
2. [Quick Start](#quick-start)
3. [Validation Workflow](#validation-workflow)
4. [Error Code Reference](#error-code-reference)
5. [Repair Strategies](#repair-strategies)
6. [Best Practices](#best-practices)
7. [Advanced Usage](#advanced-usage)
8. [Troubleshooting](#troubleshooting)

---

## Introduction

ebm-lib is a comprehensive Go library for validating and repairing EPUB 3.x and PDF 1.7 ebook files. It provides:

- **Complete validation** of EPUB structure, metadata, content documents, and navigation
- **PDF structural validation** including header, trailer, cross-reference tables, and catalog
- **Automatic repair** capabilities for common issues
- **Multiple output formats** (JSON, Text, Markdown) for validation reports
- **Context support** for timeouts and cancellation
- **Stream processing** for validating uploads and non-filesystem sources

### Target Standards

- **EPUB:** EPUB 3.0+ with compatibility for EPUB 3.3 (W3C Recommendation)
- **PDF:** PDF 1.7 / ISO 32000-1:2008

---

## Quick Start

### Installation

```bash
go get github.com/example/project/pkg/ebmlib
```

### CLI Quick Start

```bash
# Validate a single file (default command)
ebm-cli book.epub

# Validate explicitly
ebm-cli validate document.pdf --format json --min-severity warning

# Repair in place with backup
ebm-cli repair broken.epub --in-place --backup

# Batch validate with progress
ebm-cli batch validate ./library --jobs 8 --progress simple
```

For local dev runs, you can pass arguments through the Makefile:

```bash
make run RUN_ARGS="book.epub"
```

### Basic EPUB Validation

```go
package main

import (
    "fmt"
    "log"
    "github.com/example/project/pkg/ebmlib"
)

func main() {
    // Validate an EPUB file
    report, err := ebmlib.ValidateEPUB("book.epub")
    if err != nil {
        log.Fatal(err)
    }
    
    // Check validation results
    if report.IsValid {
        fmt.Println("✓ EPUB is valid!")
    } else {
        fmt.Printf("✗ Found %d errors\n", report.ErrorCount())
        for _, e := range report.Errors {
            fmt.Printf("  [%s] %s\n", e.Code, e.Message)
        }
    }
}
```

### Basic PDF Validation

```go
// Validate a PDF file
report, err := ebmlib.ValidatePDF("document.pdf")
if err != nil {
    log.Fatal(err)
}

if report.IsValid {
    fmt.Println("✓ PDF structure is valid!")
}
```

### Simple Repair

```go
// Validate and repair an EPUB
result, err := ebmlib.RepairEPUB("broken.epub")
if err != nil {
    log.Fatal(err)
}

if result.Success {
    fmt.Printf("✓ Repaired! Backup saved to: %s\n", result.BackupPath)
    fmt.Printf("Applied %d repair actions\n", len(result.ActionsApplied))
}
```

---

## Validation Workflow

### Standard Validation Flow

```
┌─────────────────────┐
│  Select File        │
└──────┬──────────────┘
       │
       ▼
┌─────────────────────┐
│  Validate File      │
│  (ValidateEPUB or   │
│   ValidatePDF)      │
└──────┬──────────────┘
       │
       ▼
┌─────────────────────┐
│  Review Report      │
│  - Errors           │
│  - Warnings         │
│  - Info messages    │
└──────┬──────────────┘
       │
       ├─── Valid? ──► Done
       │
       ▼ Invalid
┌─────────────────────┐
│  Preview Repair     │
│  (PreviewRepair)    │
└──────┬──────────────┘
       │
       ▼
┌─────────────────────┐
│  Apply Repair       │
│  (RepairEPUB or     │
│   RepairPDF)        │
└──────┬──────────────┘
       │
       ▼
┌─────────────────────┐
│  Re-validate        │
└─────────────────────┘
```

### Step-by-Step Validation

#### Step 1: Validate the File

```go
import (
    "context"
    "time"
    "github.com/example/project/pkg/ebmlib"
)

// Basic validation
report, err := ebmlib.ValidateEPUB("book.epub")

// With timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
report, err := ebmlib.ValidateEPUBWithContext(ctx, "book.epub")

// From io.Reader (for uploads)
file, _ := os.Open("book.epub")
defer file.Close()
info, _ := file.Stat()
report, err := ebmlib.ValidateEPUBReader(file, info.Size())
```

#### Step 2: Analyze the Report

```go
fmt.Printf("File: %s\n", report.FilePath)
fmt.Printf("Type: %s\n", report.FileType)
fmt.Printf("Valid: %v\n", report.IsValid)
fmt.Printf("Duration: %v\n", report.Duration)

// Count issues by severity
fmt.Printf("Errors: %d\n", report.ErrorCount())
fmt.Printf("Warnings: %d\n", report.WarningCount())
fmt.Printf("Info: %d\n", report.InfoCount())

// Examine individual errors
for _, err := range report.Errors {
    fmt.Printf("[%s] %s\n", err.Code, err.Message)
    
    if err.Location != nil {
        fmt.Printf("  Location: %s", err.Location.File)
        if err.Location.Line > 0 {
            fmt.Printf(":%d:%d", err.Location.Line, err.Location.Column)
        }
        fmt.Println()
    }
    
    // Additional details
    if len(err.Details) > 0 {
        fmt.Printf("  Details: %v\n", err.Details)
    }
}
```

#### Step 3: Generate Reports

```go
// JSON format (for APIs)
jsonOutput, err := ebmlib.FormatReport(report, ebmlib.FormatJSON)

// Text format (for console)
textOutput, err := ebmlib.FormatReport(report, ebmlib.FormatText)

// Markdown format (for documentation)
mdOutput, err := ebmlib.FormatReport(report, ebmlib.FormatMarkdown)

// Save to file with custom options
options := &ebmlib.ReportOptions{
    Format:          ebmlib.FormatText,
    IncludeWarnings: true,
    IncludeInfo:     false,
    Verbose:         true,
    ColorEnabled:    false,
    MaxErrors:       100,
}
err = ebmlib.WriteReportToFile(report, "report.txt", options)
```

#### Step 4: Preview Repairs

```go
// Preview what repairs would be done
preview, err := ebmlib.PreviewEPUBRepair("broken.epub")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Can auto-repair: %v\n", preview.CanAutoRepair)
fmt.Printf("Estimated time: %d seconds\n", preview.EstimatedTime)
fmt.Printf("Backup required: %v\n", preview.BackupRequired)

// Review proposed actions
for i, action := range preview.Actions {
    fmt.Printf("%d. %s\n", i+1, action.Description)
    fmt.Printf("   Type: %s\n", action.Type)
    fmt.Printf("   Target: %s\n", action.Target)
    fmt.Printf("   Automated: %v\n", action.Automated)
}

// Review warnings
for _, warning := range preview.Warnings {
    fmt.Printf("⚠ %s\n", warning)
}
```

#### Step 5: Apply Repairs

```go
// Option 1: Direct repair (automatic preview + apply)
result, err := ebmlib.RepairEPUB("broken.epub")

// Option 2: Apply pre-reviewed preview
result, err := ebmlib.RepairEPUBWithPreview("broken.epub", preview, "fixed.epub")

// Check results
if result.Success {
    fmt.Printf("✓ Repair successful!\n")
    fmt.Printf("Backup: %s\n", result.BackupPath)
    fmt.Printf("Applied %d actions:\n", len(result.ActionsApplied))
    for _, action := range result.ActionsApplied {
        fmt.Printf("  - %s\n", action.Description)
    }
} else {
    fmt.Printf("✗ Repair failed: %v\n", result.Error)
}

// Re-validate after repair
newReport, _ := ebmlib.ValidateEPUB("broken.epub")
if newReport.IsValid {
    fmt.Println("✓ File is now valid!")
}
```

---

## Error Code Reference

### EPUB Error Codes

#### Container Errors (EPUB-CONTAINER-XXX)

| Code | Severity | Description | Auto-Repairable |
|------|----------|-------------|-----------------|
| **EPUB-CONTAINER-001** | Error | File is not a valid ZIP archive | No |
| **EPUB-CONTAINER-002** | Error | Mimetype file has incorrect content or compression | Yes |
| **EPUB-CONTAINER-003** | Error | Mimetype file is not the first entry in ZIP archive | Yes |
| **EPUB-CONTAINER-004** | Error | Required file META-INF/container.xml is missing | Yes* |
| **EPUB-CONTAINER-005** | Error | META-INF/container.xml is malformed or invalid | Yes* |

\* Auto-repairable if package document path can be guessed

**Example Error:**

```json
{
  "code": "EPUB-CONTAINER-002",
  "message": "mimetype file must contain exactly 'application/epub+zip'",
  "severity": "error",
  "details": {
    "expected": "application/epub+zip",
    "found": "application/wrong"
  }
}
```

**Resolution:** Ensure mimetype file:
- Contains exactly "application/epub+zip" with no extra whitespace
- Is stored uncompressed (ZIP Store method)
- Is the first file in the ZIP archive

#### Navigation Errors (EPUB-NAV-XXX)

| Code | Severity | Description | Auto-Repairable |
|------|----------|-------------|-----------------|
| **EPUB-NAV-001** | Error | Navigation document is not well-formed XHTML | No |
| **EPUB-NAV-002** | Error | Missing required `<nav epub:type="toc">` element | Yes* |
| **EPUB-NAV-003** | Error | TOC navigation element missing `<ol>` structure | Yes* |
| **EPUB-NAV-004** | Error | Navigation contains invalid or non-relative links | Yes |
| **EPUB-NAV-005** | Error | Landmarks navigation missing `<ol>` structure | Yes |
| **EPUB-NAV-006** | Error | Navigation document missing `<nav>` elements | Yes* |

\* May generate basic structure; manual review recommended

**Example Error:**

```json
{
  "code": "EPUB-NAV-004",
  "message": "TOC contains invalid relative link: http://example.com/chapter.xhtml",
  "severity": "error",
  "details": {
    "href": "http://example.com/chapter.xhtml",
    "text": "Chapter 1"
  }
}
```

**Resolution:** Use only relative links within the EPUB package (e.g., "chapter1.xhtml", "content/chapter2.xhtml#section1")

### PDF Error Codes

#### Header Errors (PDF-HEADER-XXX)

| Code | Severity | Description | Auto-Repairable |
|------|----------|-------------|-----------------|
| **PDF-HEADER-001** | Critical | Invalid or missing PDF header | No |
| **PDF-HEADER-002** | Critical | Invalid PDF version number (must be 1.0-1.7) | No |

**Example Error:**

```json
{
  "code": "PDF-HEADER-001",
  "message": "Invalid or missing PDF header",
  "severity": "critical",
  "details": {
    "expected": "%PDF-1.x where x=0-7"
  }
}
```

**Resolution:** Ensure file starts with `%PDF-1.` followed by version number 0-7

#### Trailer Errors (PDF-TRAILER-XXX)

| Code | Severity | Description | Auto-Repairable |
|------|----------|-------------|-----------------|
| **PDF-TRAILER-001** | Error | Invalid or missing startxref | Yes |
| **PDF-TRAILER-002** | Error | Invalid trailer dictionary | Yes* |
| **PDF-TRAILER-003** | Error | Missing %%EOF marker | Yes |

\* Safe for common typos; may require manual repair for severe corruption

**Example Error:**

```json
{
  "code": "PDF-TRAILER-003",
  "message": "Missing %%EOF marker",
  "severity": "error",
  "details": {
    "expected": "%%EOF at end of file"
  }
}
```

**Resolution:** Append `%%EOF` on a new line at the end of the file (auto-repairable)

#### Cross-Reference Errors (PDF-XREF-XXX)

| Code | Severity | Description | Auto-Repairable |
|------|----------|-------------|-----------------|
| **PDF-XREF-001** | Error | Invalid or damaged cross-reference table | No |
| **PDF-XREF-002** | Error | Empty cross-reference table | No |
| **PDF-XREF-003** | Error | Cross-reference table has overlapping entries | No |

**Resolution:** Requires structural rebuild; use specialized PDF repair tools

#### Catalog Errors (PDF-CATALOG-XXX)

| Code | Severity | Description | Auto-Repairable |
|------|----------|-------------|-----------------|
| **PDF-CATALOG-001** | Critical | Missing or invalid catalog object | No |
| **PDF-CATALOG-002** | Error | Catalog /Type must be /Catalog | No |
| **PDF-CATALOG-003** | Critical | Catalog missing /Pages entry | No |

**Resolution:** Requires structural rebuild; affects document root

#### General Structure Errors (PDF-STRUCTURE-XXX)

| Code | Severity | Description | Auto-Repairable |
|------|----------|-------------|-----------------|
| **PDF-STRUCTURE-012** | Error | General structure error (e.g., duplicate objects) | Varies |

**Resolution:** Depends on specific issue; see error details

---

## Repair Strategies

### EPUB Repair Strategies

#### 1. Container & ZIP-Level Repairs

**Mimetype Repairs** (Very High Safety)
- **Issue:** Invalid or missing mimetype file
- **Action:** Create/overwrite with exact content: `application/epub+zip`
- **Safety:** Very High - No existing content affected
- **Auto-repair:** Yes

```go
// Example: Auto-repair mimetype issue
result, err := ebmlib.RepairEPUB("broken.epub")
// Automatically fixes mimetype if needed
```

**ZIP Structure Repairs** (High Safety)
- **Issue:** Mimetype not first in archive or compressed
- **Action:** Rebuild ZIP with mimetype first, uncompressed
- **Safety:** High - Structure only, content preserved
- **Auto-repair:** Yes

**Container.xml Repairs** (High Safety)
- **Issue:** Missing or empty META-INF/container.xml
- **Action:** Create minimal valid container.xml pointing to default OEBPS/content.opf
- **Safety:** High - If path can be guessed
- **Auto-repair:** Yes (conditional)

#### 2. Content Document Repairs

**DOCTYPE Addition** (Very High Safety)
- **Issue:** Missing HTML5 DOCTYPE
- **Action:** Prepend `<!DOCTYPE html>`
- **Safety:** Very High - Purely additive
- **Auto-repair:** Yes

**Namespace Correction** (High Safety)
- **Issue:** Wrong namespace or missing lang attributes
- **Action:** Add/correct `xmlns="http://www.w3.org/1999/xhtml" lang="en"`
- **Safety:** High - Standard correction
- **Auto-repair:** Yes

**Well-Formedness Fixes** (Medium Safety)
- **Issue:** Not well-formed XML (unclosed tags, invalid entities)
- **Action:** Use XML parser with auto-recovery
- **Safety:** Medium - May alter structure slightly
- **Auto-repair:** Conditional (requires user confirmation)

#### 3. Navigation Document Repairs

**TOC Generation** (Medium Safety)
- **Issue:** Missing or invalid `<nav epub:type="toc">`
- **Action:** Generate minimal TOC from spine order
- **Safety:** Medium - Heuristic-based
- **Auto-repair:** Conditional

**Link Normalization** (High Safety)
- **Issue:** Broken internal links / href targets
- **Action:** Scan & fix relative paths to existing files
- **Safety:** High - Verifiable corrections
- **Auto-repair:** Yes

### PDF Repair Strategies

#### 1. Trailer Repairs (Very High Safety)

**EOF Marker Addition**
- **Issue:** Missing %%EOF marker
- **Action:** Append `%%EOF` at end of file
- **Safety:** Very High - Only adds missing marker
- **Auto-repair:** Yes
- **Estimated Time:** <1 second

```go
// Example: Auto-repair missing EOF
result, err := ebmlib.RepairPDF("broken.pdf")
if result.Success {
    for _, action := range result.ActionsApplied {
        if action.Type == "EOF_MARKER" {
            fmt.Println("✓ Added missing %%EOF marker")
        }
    }
}
```

**Startxref Recomputation** (High Safety)
- **Issue:** Incorrect startxref value
- **Action:** Recompute cross-reference offset
- **Safety:** High - Calculated value
- **Auto-repair:** Yes
- **Estimated Time:** <1 second

**Trailer Dictionary Fixes** (High Safety)
- **Issue:** Minor syntax errors in trailer dictionary
- **Action:** Correct known typos (e.g., /Size value mismatch)
- **Safety:** High - Pattern-based corrections
- **Auto-repair:** Yes
- **Estimated Time:** <1 second

#### 2. Non-Repairable Issues

The following require manual intervention or specialized tools:

- **Header corruption** - Missing or invalid `%PDF-` header cannot be reconstructed safely
- **Cross-reference table damage** - Requires full xref rebuild (not supported)
- **Catalog issues** - Affects document root
- **Encryption/password protection** - Cannot be removed safely

**Recommended External Tools:**
- Adobe Acrobat Pro Preflight
- PDFtk (PDF Toolkit)
- QPDF
- VeraPDF (for validation)

### Repair Output Paths

- **Default output**: writes `<file>.repaired.<ext>` (leaves original untouched)
- **In-place**: `--in-place` replaces the original file
- **Backup**: `--backup` keeps a copy of the original (optionally `--backup-dir`)

### Repair Safety Levels

| Level | Description | User Confirmation | Backup Required |
|-------|-------------|-------------------|-----------------|
| **Very High** | Only adds missing elements; no content changes | No | Optional |
| **High** | Corrects structure/metadata; content preserved | No | Yes |
| **Medium** | May alter structure; heuristic-based | Yes | Yes |
| **Low** | Potentially lossy; removes/replaces content | Yes | Required |

### Repair Preview Example

```go
preview, err := ebmlib.PreviewEPUBRepair("broken.epub")

// Check if safe auto-repair is possible
if preview.CanAutoRepair {
    fmt.Println("✓ All repairs can be automated")
} else {
    fmt.Println("⚠ Manual review required")
}

// Review each action
for _, action := range preview.Actions {
    fmt.Printf("Action: %s\n", action.Description)
    fmt.Printf("  Type: %s\n", action.Type)
    fmt.Printf("  Target: %s\n", action.Target)
    fmt.Printf("  Automated: %v\n", action.Automated)
    
    // Check safety level from details
    if safety, ok := action.Details["safety"]; ok {
        fmt.Printf("  Safety: %v\n", safety)
    }
}
```

---

## Best Practices

### 1. Always Validate Before Processing

```go
// Good: Validate first
report, err := ebmlib.ValidateEPUB("book.epub")
if err != nil {
    return fmt.Errorf("validation error: %w", err)
}

if report.IsValid {
    // Process the valid file
    processBook("book.epub")
} else {
    // Handle errors appropriately
    logErrors(report.Errors)
}
```

### 2. Use Context for Long-Running Operations

```go
import (
    "context"
    "time"
)

func validateWithTimeout(filePath string) error {
    // Set reasonable timeout
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
    defer cancel()
    
    report, err := ebmlib.ValidateEPUBWithContext(ctx, filePath)
    if err != nil {
        return err
    }
    
    // Process report...
    return nil
}
```

### 3. Preview Repairs Before Applying

```go
// Good: Preview first
preview, err := ebmlib.PreviewEPUBRepair("book.epub")
if err != nil {
    return err
}

// Review and confirm
if !preview.CanAutoRepair {
    return fmt.Errorf("manual intervention required")
}

for _, warning := range preview.Warnings {
    log.Printf("Warning: %s", warning)
}

// User confirms or auto-proceed based on policy
if userConfirms() {
    result, err := ebmlib.RepairEPUB("book.epub")
    // Handle result...
}
```

### 4. Always Keep Backups

```go
// Good: Backup before repair
import "os"

func safeRepair(filePath string) error {
    // Create manual backup
    backupPath := filePath + ".backup"
    input, _ := os.ReadFile(filePath)
    os.WriteFile(backupPath, input, 0644)
    
    // Now repair (library also creates backup)
    result, err := ebmlib.RepairEPUB(filePath)
    if err != nil {
        // Restore from backup if repair fails
        os.Rename(backupPath, filePath)
        return err
    }
    
    if result.Success {
        fmt.Printf("Original backup: %s\n", backupPath)
        fmt.Printf("Library backup: %s\n", result.BackupPath)
    }
    
    return nil
}
```

### 5. Handle Errors by Severity

```go
func analyzeReport(report *ebmlib.ValidationReport) {
    // Critical errors - must fix
    criticalCount := 0
    for _, err := range report.Errors {
        if err.Severity == ebmlib.SeverityError {
            criticalCount++
            log.Printf("ERROR [%s]: %s", err.Code, err.Message)
        }
    }
    
    // Warnings - should review
    for _, warn := range report.Warnings {
        log.Printf("WARNING [%s]: %s", warn.Code, warn.Message)
    }
    
    // Info - optional improvements
    if verbose {
        for _, info := range report.Info {
            log.Printf("INFO [%s]: %s", info.Code, info.Message)
        }
    }
    
    // Decision logic
    if criticalCount > 0 {
        // Must repair before use
        attemptRepair()
    } else if report.WarningCount() > 0 {
        // May proceed with caution
        proceedWithWarnings()
    } else {
        // All clear
        proceed()
    }
}
```

### 6. Use Appropriate Output Formats

```go
// JSON for APIs and machine processing
jsonReport, _ := ebmlib.FormatReport(report, ebmlib.FormatJSON)
apiResponse := map[string]interface{}{
    "status": "validated",
    "report": jsonReport,
}

// Text for console and logs
textOptions := &ebmlib.ReportOptions{
    Format:       ebmlib.FormatText,
    ColorEnabled: true, // If terminal supports colors
    Verbose:      false,
}
ebmlib.WriteReportToFile(report, "console.log", textOptions)

// Markdown for documentation and reports
mdOptions := &ebmlib.ReportOptions{
    Format:  ebmlib.FormatMarkdown,
    Verbose: true,
}
ebmlib.WriteReportToFile(report, "VALIDATION_REPORT.md", mdOptions)
```

### 7. Batch Processing

```go
func validateBatch(files []string) map[string]*ebmlib.ValidationReport {
    results := make(map[string]*ebmlib.ValidationReport)
    
    // Use goroutines for parallel processing
    type result struct {
        file   string
        report *ebmlib.ValidationReport
        err    error
    }
    
    ch := make(chan result, len(files))
    
    for _, file := range files {
        go func(f string) {
            report, err := ebmlib.ValidateEPUB(f)
            ch <- result{file: f, report: report, err: err}
        }(file)
    }
    
    // Collect results
    for range files {
        r := <-ch
        if r.err != nil {
            log.Printf("Error validating %s: %v", r.file, r.err)
            continue
        }
        results[r.file] = r.report
    }
    
    return results
}
```

### 8. Stream Processing for Uploads

```go
import "net/http"

func handleUpload(w http.ResponseWriter, r *http.Request) {
    // Parse multipart form
    file, header, err := r.FormFile("ebook")
    if err != nil {
        http.Error(w, "Invalid upload", http.StatusBadRequest)
        return
    }
    defer file.Close()
    
    // Validate from stream (no temporary file needed)
    var report *ebmlib.ValidationReport
    if strings.HasSuffix(header.Filename, ".epub") {
        report, err = ebmlib.ValidateEPUBReader(file, header.Size)
    } else if strings.HasSuffix(header.Filename, ".pdf") {
        report, err = ebmlib.ValidatePDFReader(file)
    } else {
        http.Error(w, "Unsupported format", http.StatusBadRequest)
        return
    }
    
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    // Return JSON response
    jsonOutput, _ := ebmlib.FormatReport(report, ebmlib.FormatJSON)
    w.Header().Set("Content-Type", "application/json")
    w.Write([]byte(jsonOutput))
}
```

---

## Advanced Usage

### Custom Error Filtering

```go
func filterErrorsByCode(report *ebmlib.ValidationReport, codes []string) []ebmlib.ValidationError {
    codeSet := make(map[string]bool)
    for _, code := range codes {
        codeSet[code] = true
    }
    
    filtered := make([]ebmlib.ValidationError, 0)
    for _, err := range report.Errors {
        if codeSet[err.Code] {
            filtered = append(filtered, err)
        }
    }
    
    return filtered
}

// Usage
report, _ := ebmlib.ValidateEPUB("book.epub")
containerErrors := filterErrorsByCode(report, []string{
    "EPUB-CONTAINER-001",
    "EPUB-CONTAINER-002",
    "EPUB-CONTAINER-003",
})
```

### Error Statistics

```go
func generateErrorStatistics(reports []*ebmlib.ValidationReport) {
    stats := make(map[string]int)
    filesByError := make(map[string][]string)
    
    for _, report := range reports {
        for _, err := range report.Errors {
            stats[err.Code]++
            filesByError[err.Code] = append(filesByError[err.Code], report.FilePath)
        }
    }
    
    // Print statistics
    fmt.Println("Error Statistics:")
    for code, count := range stats {
        fmt.Printf("  %s: %d occurrences\n", code, count)
        fmt.Printf("    Affected files: %v\n", filesByError[code])
    }
}
```

### Progressive Validation

```go
// Validate in stages for large files
func progressiveValidation(filePath string) error {
    ctx := context.Background()
    
    // Stage 1: Quick structure check
    fmt.Println("Stage 1: Structure validation...")
    report, err := ebmlib.ValidateEPUBWithContext(ctx, filePath)
    if err != nil {
        return fmt.Errorf("stage 1 failed: %w", err)
    }
    
    if !report.IsValid {
        fmt.Printf("Structure issues found: %d errors\n", report.ErrorCount())
        return fmt.Errorf("validation failed at stage 1")
    }
    
    // Stage 2: Content validation (if needed)
    fmt.Println("Stage 2: Content validation...")
    // Additional validation logic...
    
    return nil
}
```

---

## Troubleshooting

### Common Issues

#### Issue: "File is not a valid ZIP archive" (EPUB-CONTAINER-001)

**Symptoms:** Cannot open EPUB file, validation fails immediately

**Causes:**
- File is corrupted during download or transfer
- File is not actually an EPUB (wrong extension)
- ZIP structure is damaged

**Solutions:**
1. Re-download the file
2. Verify file integrity (checksum)
3. Try opening with a ZIP utility to confirm structure
4. Use specialized ZIP repair tools if corruption is minor

#### Issue: "Navigation document not well-formed" (EPUB-NAV-001)

**Symptoms:** EPUB validation fails on navigation document

**Causes:**
- XHTML syntax errors (unclosed tags)
- Invalid character encoding
- Malformed HTML structure

**Solutions:**
1. Preview repair to see if auto-fixable:
   ```go
   preview, _ := ebmlib.PreviewEPUBRepair("book.epub")
   ```
2. If not auto-repairable, manually fix XHTML in navigation document
3. Use XHTML validation tools to identify specific syntax errors

#### Issue: "Missing %%EOF marker" (PDF-TRAILER-003)

**Symptoms:** PDF validation fails, file may not open in some readers

**Causes:**
- File truncation during transfer
- Incomplete file write
- Storage corruption

**Solutions:**
1. Auto-repair (recommended):
   ```go
   result, _ := ebmlib.RepairPDF("document.pdf")
   ```
2. Manually append `%%EOF` to the end of the file

#### Issue: Repair fails with "manual intervention required"

**Symptoms:** Preview shows `CanAutoRepair: false`

**Causes:**
- Errors require structural changes
- Multiple conflicting issues
- Safety level too low for automatic repair

**Solutions:**
1. Review preview warnings:
   ```go
   for _, warning := range preview.Warnings {
       fmt.Println(warning)
   }
   ```
2. Fix issues manually using tools like:
   - **EPUB:** Sigil, Calibre
   - **PDF:** Adobe Acrobat Pro, QPDF
3. Re-validate after manual fixes

### Performance Issues

#### Large File Validation

```go
// Use timeout for very large files
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

report, err := ebmlib.ValidateEPUBWithContext(ctx, "large-book.epub")
```

#### Memory Issues

```go
// Use streaming validation for uploads
// Instead of loading entire file into memory
report, err := ebmlib.ValidateEPUBReader(request.Body, contentLength)
```

### Getting Help

1. **Check error details:** Most errors include detailed information in the `Details` field
2. **Review error code documentation:** See [Error Code Reference](#error-code-reference)
3. **Enable verbose reporting:** Use `Verbose: true` in ReportOptions
4. **Check logs:** Validation operations log detailed information
5. **Consult specifications:** Refer to EPUB 3.3 or PDF 1.7 specs for standard requirements

---

## Appendix: Quick Reference

### Common Validation Patterns

```go
// Pattern 1: Validate and log
report, _ := ebmlib.ValidateEPUB("book.epub")
textOutput, _ := ebmlib.FormatReport(report, ebmlib.FormatText)
log.Println(textOutput)

// Pattern 2: Validate and repair if needed
report, _ := ebmlib.ValidateEPUB("book.epub")
if !report.IsValid {
    ebmlib.RepairEPUB("book.epub")
}

// Pattern 3: Validate with timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
report, _ := ebmlib.ValidateEPUBWithContext(ctx, "book.epub")

// Pattern 4: Stream validation
file, _ := os.Open("book.epub")
defer file.Close()
info, _ := file.Stat()
report, _ := ebmlib.ValidateEPUBReader(file, info.Size())
```

### Report Format Examples

```go
// JSON (for APIs)
options := &ebmlib.ReportOptions{Format: ebmlib.FormatJSON}

// Text (for console)
options := &ebmlib.ReportOptions{
    Format: ebmlib.FormatText,
    ColorEnabled: true,
}

// Markdown (for docs)
options := &ebmlib.ReportOptions{
    Format: ebmlib.FormatMarkdown,
    Verbose: true,
}
```

---

**For more information:**
- API Documentation: See `pkg/ebmlib/doc.go`
- Architecture: See `docs/ARCHITECTURE.md`
- Error Codes: See `docs/ERROR_CODES.md`
- Examples: See `examples/` directory
