# ebm-lib

A Go library for validating and repairing EPUB and PDF ebooks, using hexagonal architecture with clean separation of concerns.

## Features

- **EPUB Validation**: Comprehensive validation of EPUB 3.0 files including structure, metadata, and content
- **PDF Validation**: PDF structure validation including header, trailer, cross-reference tables, and catalog
- **Automated Repair**: Automatic repair capabilities for common EPUB and PDF issues
- **Flexible Reporting**: Multiple output formats (JSON, Text, Markdown) with customizable options
- **Simple API**: Clean, intuitive public API in `pkg/ebmlib`

## Documentation

- `docs/README.md` - Documentation index
- `docs/USER_GUIDE.md` - CLI and library usage
- `docs/ARCHITECTURE.md` - System architecture
- `docs/adr/` - Architecture Decision Records

## Wiki

The GitHub wiki is a mirror of `docs/` in this repository. Edit those files and run:

```bash
make wiki-update
```

Other wiki operations:

```bash
make wiki-clone
make wiki-sync
make wiki-push
make wiki-pull
make wiki-status
make wiki-clean
```

## Quick Start

### Installation

```bash
go get github.com/example/project/pkg/ebmlib
```

### Basic Usage

```go
package main

import (
    "fmt"
    "log"
    "github.com/example/project/pkg/ebmlib"
)

func main() {
    // Validate an EPUB
    report, err := ebmlib.ValidateEPUB("book.epub")
    if err != nil {
        log.Fatal(err)
    }
    
    if report.IsValid {
        fmt.Println("✓ EPUB is valid!")
    } else {
        fmt.Printf("✗ Found %d errors\n", report.ErrorCount())
    }
    
    // Repair if needed
    if !report.IsValid {
        result, err := ebmlib.RepairEPUB("book.epub")
        if err != nil {
            log.Fatal(err)
        }
        fmt.Printf("Repaired: %s\n", result.BackupPath)
    }
}
```

## API Reference

### Validation Functions

#### EPUB Validation

```go
// Validate EPUB file
report, err := ebmlib.ValidateEPUB(filePath string) (*ValidationReport, error)

// Validate with context
report, err := ebmlib.ValidateEPUBWithContext(ctx context.Context, filePath string) (*ValidationReport, error)

// Validate from io.Reader
report, err := ebmlib.ValidateEPUBReader(reader io.Reader, size int64) (*ValidationReport, error)
```

#### PDF Validation

```go
// Validate PDF file
report, err := ebmlib.ValidatePDF(filePath string) (*ValidationReport, error)

// Validate with context
report, err := ebmlib.ValidatePDFWithContext(ctx context.Context, filePath string) (*ValidationReport, error)

// Validate from io.Reader
report, err := ebmlib.ValidatePDFReader(reader io.Reader) (*ValidationReport, error)
```

### Repair Functions

#### EPUB Repair

```go
// Repair EPUB (validate + auto-repair)
result, err := ebmlib.RepairEPUB(filePath string) (*RepairResult, error)

// Preview repair actions before applying
preview, err := ebmlib.PreviewEPUBRepair(filePath string) (*RepairPreview, error)

// Apply repair with custom output path
result, err := ebmlib.RepairEPUBWithPreview(filePath string, preview *RepairPreview, outputPath string) (*RepairResult, error)
```

#### PDF Repair

```go
// Repair PDF (validate + auto-repair)
result, err := ebmlib.RepairPDF(filePath string) (*RepairResult, error)

// Preview repair actions before applying
preview, err := ebmlib.PreviewPDFRepair(filePath string) (*RepairPreview, error)

// Apply repair with custom output path
result, err := ebmlib.RepairPDFWithPreview(filePath string, preview *RepairPreview, outputPath string) (*RepairResult, error)
```

### Reporting Functions

```go
// Format report in different formats
output, err := ebmlib.FormatReport(report *ValidationReport, format OutputFormat) (string, error)

// Available formats: FormatJSON, FormatText, FormatMarkdown

// Write report to io.Writer
err := ebmlib.WriteReport(report *ValidationReport, writer io.Writer, options *ReportOptions) error

// Write report to file
err := ebmlib.WriteReportToFile(report *ValidationReport, filePath string, options *ReportOptions) error

// Custom options
options := &ebmlib.ReportOptions{
    Format:          ebmlib.FormatJSON,
    IncludeWarnings: true,
    IncludeInfo:     true,
    Verbose:         true,
    ColorEnabled:    false,
    MaxErrors:       100,
}
```

## Usage Examples

### Example 1: Basic Validation

```go
report, err := ebmlib.ValidateEPUB("book.epub")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Valid: %v\n", report.IsValid)
fmt.Printf("Errors: %d\n", report.ErrorCount())
fmt.Printf("Warnings: %d\n", report.WarningCount())

for _, err := range report.Errors {
    fmt.Printf("  [%s] %s\n", err.Code, err.Message)
}
```

### Example 2: Repair with Preview

```go
// Preview what repairs would be done
preview, err := ebmlib.PreviewEPUBRepair("book.epub")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Can auto-repair: %v\n", preview.CanAutoRepair)
fmt.Printf("Actions: %d\n", len(preview.Actions))

for _, action := range preview.Actions {
    fmt.Printf("  - %s (automated: %v)\n", action.Description, action.Automated)
}

// Apply repairs
result, err := ebmlib.RepairEPUB("book.epub")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Success: %v\n", result.Success)
fmt.Printf("Output: %s\n", result.BackupPath)
```

### Example 3: Custom Reporting

```go
report, _ := ebmlib.ValidateEPUB("book.epub")

// Generate JSON report
jsonOutput, _ := ebmlib.FormatReport(report, ebmlib.FormatJSON)
fmt.Println(jsonOutput)

// Generate colored text report
textOptions := &ebmlib.ReportOptions{
    Format:          ebmlib.FormatText,
    IncludeWarnings: true,
    IncludeInfo:     false,
    ColorEnabled:    true,
}
ebmlib.WriteReportToFile(report, "report.txt", textOptions)

// Generate markdown report
mdOptions := &ebmlib.ReportOptions{
    Format:  ebmlib.FormatMarkdown,
    Verbose: true,
}
ebmlib.WriteReportToFile(report, "report.md", mdOptions)
```

### Example 4: Working with Readers

```go
file, _ := os.Open("book.epub")
defer file.Close()

info, _ := file.Stat()
report, err := ebmlib.ValidateEPUBReader(file, info.Size())
if err != nil {
    log.Fatal(err)
}
```

## Running Examples

The `examples/` directory contains complete working examples:

```bash
# Basic validation
go run examples/basic_validation.go testdata/sample.epub

# Repair example
go run examples/repair_example.go testdata/broken.epub

# Custom reporting
go run examples/custom_reporting.go testdata/sample.epub
```

## Project Structure

```
.
├── cmd/                    # Application entrypoints
├── internal/
│   ├── domain/            # Domain entities and business logic
│   ├── ports/             # Interface definitions (ports)
│   └── adapters/          # Implementation of ports (adapters)
├── pkg/
│   └── ebmlib/            # Public API library
├── testdata/              # Test fixtures and sample data
└── examples/              # Example usage code
```

## Development

### Prerequisites

- Go 1.21 or higher

### Installation

```bash
make install
```

### Building

```bash
make build
```

### Running

```bash
make run
```

### Testing

```bash
# Run all tests
make test

# Run unit tests only
make test-unit

# Run integration tests only
make test-integration

# Generate coverage report
make coverage

# Run performance benchmarks
make test-bench

# Create performance baseline
make bench-baseline

# Compare with baseline
make bench-compare
```

### Code Quality

```bash
# Format code
make fmt

# Run vet
make vet

# Run linter
make lint
```

## Performance Benchmarking

The library includes comprehensive performance benchmarks for validation throughput, reporter formatting, and repair operations. See [docs/BENCHMARKING.md](docs/BENCHMARKING.md) for detailed information.

### Quick Start

```bash
# Run all benchmarks
make test-bench

# Create baseline for regression detection
make bench-baseline

# Compare current performance with baseline
make bench-compare
```

### Benchmark Categories

- **EPUB Validation**: Small (<1MB), Medium (1-10MB), Large (>10MB) files
- **PDF Validation**: Various file sizes and validation modes
- **Reporter Formatting**: Different error counts (10, 100, 1K, 10K errors)
- **Repair Service**: Preview and apply operations

### Performance Targets

| Operation | Target Time | Memory Target |
|-----------|-------------|---------------|
| EPUB Small | < 2ms | < 500 KB |
| EPUB Medium | < 20ms | < 5 MB |
| PDF Small | < 1ms | < 200 KB |
| Reporter (100 errors) | < 1ms | < 500 KB |

See [docs/tests/integration/BENCHMARKS.md](docs/tests/integration/BENCHMARKS.md) for complete baseline metrics.

## Make Targets

Run `make help` to see all available targets with descriptions.

## API Usage Patterns

### Pattern 1: Simple Validation Pipeline

```go
func validateAndReport(filePath string) error {
    // Validate
    report, err := ebmlib.ValidateEPUB(filePath)
    if err != nil {
        return err
    }
    
    // Format as text
    output, err := ebmlib.FormatReport(report, ebmlib.FormatText)
    if err != nil {
        return err
    }
    
    fmt.Println(output)
    return nil
}
```

### Pattern 2: Batch Processing

```go
func validateMultipleFiles(files []string) {
    for _, file := range files {
        report, err := ebmlib.ValidateEPUB(file)
        if err != nil {
            log.Printf("Error validating %s: %v", file, err)
            continue
        }
        
        if !report.IsValid {
            log.Printf("%s: %d errors", file, report.ErrorCount())
        }
    }
}
```

### Pattern 3: Conditional Repair

```go
func repairIfNeeded(filePath string) error {
    // Validate first
    report, err := ebmlib.ValidateEPUB(filePath)
    if err != nil {
        return err
    }
    
    if report.IsValid {
        fmt.Println("File is valid, no repair needed")
        return nil
    }
    
    // Preview repairs
    preview, err := ebmlib.PreviewEPUBRepair(filePath)
    if err != nil {
        return err
    }
    
    if !preview.CanAutoRepair {
        return fmt.Errorf("manual intervention required")
    }
    
    // Apply repairs
    result, err := ebmlib.RepairEPUB(filePath)
    if err != nil {
        return err
    }
    
    fmt.Printf("Repaired: %s\n", result.BackupPath)
    return nil
}
```

### Pattern 4: Custom Output Formats

```go
func generateReports(filePath string) error {
    report, err := ebmlib.ValidateEPUB(filePath)
    if err != nil {
        return err
    }
    
    // JSON for machine processing
    ebmlib.WriteReportToFile(report, "report.json", &ebmlib.ReportOptions{
        Format:          ebmlib.FormatJSON,
        IncludeWarnings: true,
        IncludeInfo:     true,
    })
    
    // Text for console
    ebmlib.WriteReportToFile(report, "report.txt", &ebmlib.ReportOptions{
        Format:       ebmlib.FormatText,
        ColorEnabled: false,
    })
    
    // Markdown for documentation
    ebmlib.WriteReportToFile(report, "report.md", &ebmlib.ReportOptions{
        Format:  ebmlib.FormatMarkdown,
        Verbose: true,
    })
    
    return nil
}
```

### Pattern 5: Stream Processing

```go
func validateFromStream(r io.Reader, size int64) (*ebmlib.ValidationReport, error) {
    report, err := ebmlib.ValidateEPUBReader(r, size)
    if err != nil {
        return nil, err
    }
    
    return report, nil
}

// Usage with HTTP upload
func handleUpload(w http.ResponseWriter, r *http.Request) {
    file, header, _ := r.FormFile("ebook")
    defer file.Close()
    
    report, err := validateFromStream(file, header.Size)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    jsonOutput, _ := ebmlib.FormatReport(report, ebmlib.FormatJSON)
    w.Header().Set("Content-Type", "application/json")
    w.Write([]byte(jsonOutput))
}
```

### Pattern 6: Context-Aware Processing

```go
func validateWithTimeout(filePath string, timeout time.Duration) error {
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()
    
    report, err := ebmlib.ValidateEPUBWithContext(ctx, filePath)
    if err != nil {
        return err
    }
    
    if !report.IsValid {
        return fmt.Errorf("validation failed: %d errors", report.ErrorCount())
    }
    
    return nil
}
```

### Pattern 7: Error Analysis

```go
func analyzeErrors(report *ebmlib.ValidationReport) {
    errorsByCode := make(map[string]int)
    errorsByLocation := make(map[string]int)
    
    for _, err := range report.Errors {
        errorsByCode[err.Code]++
        if err.Location != nil {
            errorsByLocation[err.Location.File]++
        }
    }
    
    fmt.Println("Error distribution by code:")
    for code, count := range errorsByCode {
        fmt.Printf("  %s: %d\n", code, count)
    }
    
    fmt.Println("\nError distribution by file:")
    for file, count := range errorsByLocation {
        fmt.Printf("  %s: %d\n", file, count)
    }
}
```

### Pattern 8: Integration with CI/CD

```go
func ciValidation(filePath string) int {
    report, err := ebmlib.ValidateEPUB(filePath)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Validation failed: %v\n", err)
        return 2
    }
    
    // Write detailed report
    ebmlib.WriteReportToFile(report, "validation-report.json", &ebmlib.ReportOptions{
        Format: ebmlib.FormatJSON,
    })
    
    if !report.IsValid {
        fmt.Fprintf(os.Stderr, "Validation errors: %d\n", report.ErrorCount())
        return 1
    }
    
    if report.WarningCount() > 0 {
        fmt.Printf("Warnings: %d\n", report.WarningCount())
    }
    
    return 0
}
```

## Type Reference

### ValidationReport

```go
type ValidationReport struct {
    FilePath       string
    FileType       string
    IsValid        bool
    Errors         []ValidationError
    Warnings       []ValidationError
    Info           []ValidationError
    ValidationTime time.Time
    Duration       time.Duration
    Metadata       map[string]interface{}
}

// Methods
func (r *ValidationReport) HasErrors() bool
func (r *ValidationReport) HasWarnings() bool
func (r *ValidationReport) ErrorCount() int
func (r *ValidationReport) WarningCount() int
func (r *ValidationReport) InfoCount() int
func (r *ValidationReport) TotalIssues() int
```

### ValidationError

```go
type ValidationError struct {
    Code      string
    Message   string
    Severity  Severity
    Location  *ErrorLocation
    Details   map[string]interface{}
    Timestamp time.Time
}
```

### RepairResult

```go
type RepairResult struct {
    Success        bool
    ActionsApplied []RepairAction
    Report         *ValidationReport
    BackupPath     string
    Error          error
}
```

### RepairPreview

```go
type RepairPreview struct {
    Actions        []RepairAction
    CanAutoRepair  bool
    EstimatedTime  int64
    BackupRequired bool
    Warnings       []string
}
```

## CLI Usage

The CLI lives under `cmd/` and exposes validation, repair, and batch operations.

```bash
# Validate a single file
ebm-cli validate book.epub

# Validate with JSON output and severity filtering
ebm-cli validate document.pdf --format json --min-severity warning

# Repair in place with backup
ebm-cli repair book.epub --in-place --backup

# Batch validate a directory with 8 workers
ebm-cli batch validate ./library --jobs 8 --progress simple

# Batch repair with glob patterns
ebm-cli batch repair ./books/**/*.epub --in-place --backup
```

Run `ebm-cli --help`, `ebm-cli validate --help`, and `ebm-cli batch --help` for detailed flag and example references.

## Dependencies

- `archive/zip` - Standard library for ZIP archive handling
- `golang.org/x/net/html` - HTML parsing
- `github.com/unidoc/unipdf/v3` - PDF processing

## License

TBD
