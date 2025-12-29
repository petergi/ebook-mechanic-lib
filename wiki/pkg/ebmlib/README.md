# ebm-lib - Public API

This package provides a simple, high-level API for validating and repairing EPUB and PDF ebooks.

## Installation

```bash
go get github.com/example/project/pkg/ebmlib
```

## Quick Start

```go
import "github.com/example/project/pkg/ebmlib"

// Validate
report, err := ebmlib.ValidateEPUB("book.epub")
if err != nil {
    log.Fatal(err)
}

// Repair if needed
if !report.IsValid {
    result, err := ebmlib.RepairEPUB("book.epub")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Repaired: %s\n", result.BackupPath)
}
```

## API Overview

### Validation Functions

#### EPUB
```go
ValidateEPUB(filePath string) (*ValidationReport, error)
ValidateEPUBWithContext(ctx context.Context, filePath string) (*ValidationReport, error)
ValidateEPUBReader(reader io.Reader, size int64) (*ValidationReport, error)
ValidateEPUBReaderWithContext(ctx context.Context, reader io.Reader, size int64) (*ValidationReport, error)
```

#### PDF
```go
ValidatePDF(filePath string) (*ValidationReport, error)
ValidatePDFWithContext(ctx context.Context, filePath string) (*ValidationReport, error)
ValidatePDFReader(reader io.Reader) (*ValidationReport, error)
ValidatePDFReaderWithContext(ctx context.Context, reader io.Reader) (*ValidationReport, error)
```

### Repair Functions

#### EPUB
```go
RepairEPUB(filePath string) (*RepairResult, error)
RepairEPUBWithContext(ctx context.Context, filePath string) (*RepairResult, error)
PreviewEPUBRepair(filePath string) (*RepairPreview, error)
PreviewEPUBRepairWithContext(ctx context.Context, filePath string) (*RepairPreview, error)
RepairEPUBWithPreview(filePath string, preview *RepairPreview, outputPath string) (*RepairResult, error)
RepairEPUBWithPreviewContext(ctx context.Context, filePath string, preview *RepairPreview, outputPath string) (*RepairResult, error)
```

#### PDF
```go
RepairPDF(filePath string) (*RepairResult, error)
RepairPDFWithContext(ctx context.Context, filePath string) (*RepairResult, error)
PreviewPDFRepair(filePath string) (*RepairPreview, error)
PreviewPDFRepairWithContext(ctx context.Context, filePath string) (*RepairPreview, error)
RepairPDFWithPreview(filePath string, preview *RepairPreview, outputPath string) (*RepairResult, error)
RepairPDFWithPreviewContext(ctx context.Context, filePath string, preview *RepairPreview, outputPath string) (*RepairResult, error)
```

### Reporting Functions

```go
FormatReport(report *ValidationReport, format OutputFormat) (string, error)
FormatReportWithContext(ctx context.Context, report *ValidationReport, format OutputFormat) (string, error)
FormatReportWithOptions(ctx context.Context, report *ValidationReport, options *ReportOptions) (string, error)
WriteReport(report *ValidationReport, writer io.Writer, options *ReportOptions) error
WriteReportWithContext(ctx context.Context, report *ValidationReport, writer io.Writer, options *ReportOptions) error
WriteReportToFile(report *ValidationReport, filePath string, options *ReportOptions) error
WriteReportToFileWithContext(ctx context.Context, report *ValidationReport, filePath string, options *ReportOptions) error
```

## Main Types

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

### RepairAction

```go
type RepairAction struct {
    Type        string
    Description string
    Target      string
    Details     map[string]interface{}
    Automated   bool
}
```

### ReportOptions

```go
type ReportOptions struct {
    Format         OutputFormat
    IncludeWarnings bool
    IncludeInfo     bool
    Verbose         bool
    ColorEnabled    bool
    MaxErrors       int
}
```

## Constants

### Output Formats

```go
const (
    FormatJSON     OutputFormat = "json"
    FormatText     OutputFormat = "text"
    FormatHTML     OutputFormat = "html"
    FormatXML      OutputFormat = "xml"
    FormatMarkdown OutputFormat = "markdown"
)
```

### Severity Levels

```go
const (
    SeverityError   Severity = "error"
    SeverityWarning Severity = "warning"
    SeverityInfo    Severity = "info"
)
```

## Usage Examples

See the [examples directory](../../../examples/) for complete, runnable examples:

- **[basic_validation](../../../examples/basic_validation/main.go)** - Simple validation
- **[repair_example](../../../examples/repair_example/main.go)** - Repair workflow
- **[custom_reporting](../../../examples/custom_reporting/main.go)** - Report formatting
- **[advanced_validation](../../../examples/advanced_validation/main.go)** - Batch processing
- **[complete_workflow](../../../examples/complete_workflow/main.go)** - End-to-end workflow

## Design Principles

1. **Simple by default**: Most functions have simple variants without context
2. **Context support**: All operations have context-aware variants for cancellation/timeout
3. **Stream support**: Can validate from files or io.Reader
4. **Type safety**: Strong typing with exported domain types
5. **Error clarity**: Operational errors vs. validation errors are distinct
6. **Flexibility**: Multiple output formats and customization options

## Thread Safety

All functions are safe for concurrent use. Each operation creates its own internal instances.

## Performance Considerations

- **EPUB validation**: O(n) where n is the number of content files
- **PDF validation**: Generally faster than EPUB for same file size
- **Batch processing**: Use goroutines for parallel validation
- **Memory**: Validation loads files into memory; consider streaming for very large files

## Error Handling

The library uses two types of errors:

1. **Operational errors** (returned as Go errors):
   - File not found
   - Permission denied
   - File corruption
   - Parse failures

2. **Validation errors** (in ValidationReport):
   - EPUB/PDF specification violations
   - Structure issues
   - Content problems

## Maintenance

This is a stable public API. Breaking changes will be avoided, and deprecated functions will be maintained for at least one major version with clear migration paths.

## See Also

- [Main README](../../../README.md) - Complete documentation
- [Package documentation](../../../pkg/ebmlib/doc.go) - GoDoc comments
- [AGENTS.md](../../../AGENTS.md) - Development guide
