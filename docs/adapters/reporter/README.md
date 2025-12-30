# Reporter Adapters

This package provides multiple reporter implementations for formatting and outputting validation reports.

## Available Reporters

### JSON Reporter (`json_reporter.go`)

Produces structured JSON output suitable for programmatic consumption and integration with other tools.

**Features:**
- Structured JSON format with nested objects
- ISO 8601 timestamps
- Complete error details and metadata
- Support for single and multiple report formatting
- Summary generation

**Usage:**
```go
reporter := reporter.NewJSONReporter()
result, err := reporter.Format(ctx, report, &ports.ReportOptions{
    IncludeWarnings: true,
    IncludeInfo:     true,
})
```

### Markdown Reporter (`markdown_reporter.go`)

Generates Markdown-formatted reports suitable for documentation, GitHub issues, and static site generators.

**Features:**
- Markdown tables for structured data
- Headers and sections for easy navigation
- Status indicators (✅/❌)
- Suitable for version control and documentation
- Support for verbose output with metadata

**Usage:**
```go
reporter := reporter.NewMarkdownReporter()
err := reporter.WriteToFile(ctx, report, "report.md", &ports.ReportOptions{
    IncludeWarnings: true,
    IncludeInfo:     true,
    Verbose:         true,
})
```

### Text Reporter (`text_reporter.go`)

Produces human-readable text output with optional colorized terminal display.

**Features:**
- Structured text with clear sections
- Colorized output for terminal display
- Severity symbols (✗ for errors, ⚠ for warnings, ℹ for info)
- Box-drawing characters for visual separation
- Support for verbose details

**Usage:**
```go
reporter := reporter.NewTextReporter()
result, err := reporter.Format(ctx, report, &ports.ReportOptions{
    ColorEnabled:    true,
    IncludeWarnings: true,
    IncludeInfo:     true,
    Verbose:         true,
})
```

## Filtering

All reporters support filtering by severity, category, and standard.

### Filter Types

- **Severity Filtering**: Include only specific severity levels
- **Category Filtering**: Filter by error category (e.g., "structure", "metadata", "content")
- **Standard Filtering**: Filter by compliance standard (e.g., "EPUB3", "PDF/A")
- **Minimum Severity**: Filter out issues below a certain severity threshold

### Creating Filters

```go
filter := &reporter.Filter{
    Severities: []domain.Severity{domain.SeverityError},
    Categories: []string{"structure", "metadata"},
    Standards:  []string{"EPUB3"},
    MinSeverity: domain.SeverityWarning,
}

reporter := reporter.NewJSONReporterWithFilter(filter)
```

### Filter Examples

**Show only errors:**
```go
filter := &reporter.Filter{
    Severities: []domain.Severity{domain.SeverityError},
}
```

**Show errors and warnings (exclude info):**
```go
filter := &reporter.Filter{
    MinSeverity: domain.SeverityWarning,
}
```

**Show only structure-related issues:**
```go
filter := &reporter.Filter{
    Categories: []string{"structure"},
}
```

**Show only EPUB3 compliance issues:**
```go
filter := &reporter.Filter{
    Standards: []string{"EPUB3"},
}
```

**Combine multiple filters:**
```go
filter := &reporter.Filter{
    Severities: []domain.Severity{domain.SeverityError, domain.SeverityWarning},
    Categories: []string{"structure"},
    Standards:  []string{"EPUB3"},
}
```

## Report Options

The `ReportOptions` struct controls report formatting:

```go
type ReportOptions struct {
    Format         OutputFormat // Output format (json, text, markdown, etc.)
    IncludeWarnings bool         // Include warning-level issues
    IncludeInfo     bool         // Include info-level issues
    Verbose         bool         // Include detailed information
    ColorEnabled    bool         // Enable color output (text reporter)
    MaxErrors       int          // Maximum number of errors to display
}
```

## Color Support

The text reporter supports colorized output for terminal display:

- **Red**: Errors
- **Yellow**: Warnings
- **Blue**: Info
- **Green**: Success/Valid status
- **Cyan**: File paths
- **Bold**: Headers
- **Dim**: Timestamps and codes

Colors can be enabled or disabled using the `ColorEnabled` option.

## Multi-Report Support

All reporters implement the `MultiReporter` interface for handling multiple validation reports:

```go
reports := []*domain.ValidationReport{report1, report2, report3}

// Format all reports together
result, err := reporter.FormatMultiple(ctx, reports, options)

// Write summary only
err := reporter.WriteSummary(ctx, reports, writer, options)
```

## Output Methods

Each reporter supports three output methods:

1. **Format**: Returns formatted string
2. **Write**: Writes to an io.Writer
3. **WriteToFile**: Writes directly to a file

## Testing

Comprehensive test suites are provided:

- `filter_test.go`: Filter functionality tests
- `json_reporter_test.go`: JSON reporter tests
- `markdown_reporter_test.go`: Markdown reporter tests
- `text_reporter_test.go`: Text reporter tests
- `colors_test.go`: Color scheme tests
- `integration_test.go`: Cross-reporter integration tests

Run tests:
```bash
go test ./internal/adapters/reporter/...
```

## Examples

### Basic Usage

```go
ctx := context.Background()

// Create a reporter
reporter := reporter.NewJSONReporter()

// Format a report
options := &ports.ReportOptions{
    IncludeWarnings: true,
    IncludeInfo:     true,
}

result, err := reporter.Format(ctx, validationReport, options)
if err != nil {
    log.Fatal(err)
}

fmt.Println(result)
```

### With Filtering

```go
// Create a filter for errors only
filter := &reporter.Filter{
    Severities: []domain.Severity{domain.SeverityError},
}

// Create reporter with filter
reporter := reporter.NewTextReporterWithFilter(filter)

// Format with colors enabled
options := &ports.ReportOptions{
    ColorEnabled: true,
}

result, err := reporter.Format(ctx, validationReport, options)
```

### Multiple Reports

```go
reports := []*domain.ValidationReport{report1, report2, report3}

// Create markdown reporter
reporter := reporter.NewMarkdownReporter()

// Generate summary
var buf bytes.Buffer
err := reporter.WriteSummary(ctx, reports, &buf, &ports.ReportOptions{})

fmt.Println(buf.String())
```

## Format Consistency

All reporters maintain consistency in:
- Error code display
- Location information formatting
- Severity level representation
- Summary statistics
- Timestamp formatting

This ensures that switching between formats doesn't lose information or context.
