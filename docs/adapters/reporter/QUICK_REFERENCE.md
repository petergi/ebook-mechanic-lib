# Reporter Adapters - Quick Reference

## Quick Start

### Basic Usage

```go
import (
    "context"
    "github.com/example/project/internal/adapters/reporter"
    "github.com/example/project/internal/ports"
)

ctx := context.Background()

// Create reporter
jsonReporter := reporter.NewJSONReporter()

// Format report
result, err := jsonReporter.Format(ctx, validationReport, &ports.ReportOptions{
    IncludeWarnings: true,
    IncludeInfo:     true,
})
```

## Available Reporters

| Reporter | Constructor | Best For |
|----------|-------------|----------|
| JSON | `NewJSONReporter()` | API responses, tool integration |
| Markdown | `NewMarkdownReporter()` | Documentation, GitHub issues |
| Text | `NewTextReporter()` | Terminal output, logs |

## Common Options

```go
options := &ports.ReportOptions{
    IncludeWarnings: true,    // Include warnings in output
    IncludeInfo:     true,    // Include info messages
    Verbose:         true,    // Show detailed information
    ColorEnabled:    true,    // Enable colored output (text only)
    MaxErrors:       50,      // Limit number of errors shown
}
```

## Filtering

### By Severity

```go
// Show only errors
filter := &reporter.Filter{
    Severities: []domain.Severity{domain.SeverityError},
}

// Show errors and warnings (no info)
filter := &reporter.Filter{
    MinSeverity: domain.SeverityWarning,
}

reporter := reporter.NewJSONReporterWithFilter(filter)
```

### By Category

```go
filter := &reporter.Filter{
    Categories: []string{"structure", "metadata"},
}
reporter := reporter.NewTextReporterWithFilter(filter)
```

### By Standard

```go
filter := &reporter.Filter{
    Standards: []string{"EPUB3"},
}
reporter := reporter.NewMarkdownReporterWithFilter(filter)
```

### Combined Filters

```go
filter := &reporter.Filter{
    Severities: []domain.Severity{domain.SeverityError, domain.SeverityWarning},
    Categories: []string{"structure"},
    Standards:  []string{"EPUB3"},
}
```

## Output Methods

### To String

```go
result, err := reporter.Format(ctx, report, options)
fmt.Println(result)
```

### To Writer

```go
var buf bytes.Buffer
err := reporter.Write(ctx, report, &buf, options)
```

### To File

```go
err := reporter.WriteToFile(ctx, report, "report.json", options)
```

## Multiple Reports

### Format All

```go
reports := []*domain.ValidationReport{report1, report2, report3}
result, err := reporter.FormatMultiple(ctx, reports, options)
```

### Summary Only

```go
err := reporter.WriteSummary(ctx, reports, os.Stdout, options)
```

## Color Output (Text Reporter Only)

```go
textReporter := reporter.NewTextReporter()

// Enable colors
options := &ports.ReportOptions{
    ColorEnabled: true,
}

result, err := textReporter.Format(ctx, report, options)
```

### Color Meanings

- 🔴 **Red**: Errors
- 🟡 **Yellow**: Warnings
- 🔵 **Blue**: Info messages
- 🟢 **Green**: Valid status
- 🔷 **Cyan**: File paths
- **Bold**: Headers
- **Dim**: Timestamps, codes

## Format Comparison

### JSON
- Machine-readable
- Structured data
- Best for APIs
- Easy to parse

### Markdown
- Human-readable
- Great for docs
- Version control friendly
- Tables for data

### Text
- Terminal friendly
- Color support
- Box drawing
- Quick scanning

## Common Patterns

### API Response

```go
jsonReporter := reporter.NewJSONReporter()
result, _ := jsonReporter.Format(ctx, report, &ports.ReportOptions{})
w.Header().Set("Content-Type", "application/json")
w.Write([]byte(result))
```

### Terminal Output

```go
textReporter := reporter.NewTextReporter()
result, _ := textReporter.Format(ctx, report, &ports.ReportOptions{
    ColorEnabled: true,
})
fmt.Print(result)
```

### GitHub Issue

```go
mdReporter := reporter.NewMarkdownReporter()
result, _ := mdReporter.Format(ctx, report, &ports.ReportOptions{
    Verbose: true,
})
// Post result as issue comment
```

### CI/CD Integration

```go
// Fail on errors, show all issues
filter := &reporter.Filter{
    MinSeverity: domain.SeverityInfo,  // Show everything
}
reporter := reporter.NewTextReporterWithFilter(filter)

result, _ := reporter.Format(ctx, report, &ports.ReportOptions{})
fmt.Print(result)

if report.ErrorCount() > 0 {
    os.Exit(1)
}
```

### Batch Processing

```go
var reports []*domain.ValidationReport

// Validate multiple files
for _, file := range files {
    report, _ := validator.ValidateFile(ctx, file)
    reports = append(reports, report)
}

// Generate summary
jsonReporter := reporter.NewJSONReporter()
err := jsonReporter.WriteToFile(ctx, nil, "batch-summary.json", options)
```

## Error Handling

```go
result, err := reporter.Format(ctx, report, options)
if err != nil {
    log.Printf("Reporter error: %v", err)
    // Handle error appropriately
}
```

## Testing

```go
import "testing"

func TestMyReporter(t *testing.T) {
    reporter := reporter.NewJSONReporter()
    report := createTestReport()
    
    result, err := reporter.Format(context.Background(), report, nil)
    if err != nil {
        t.Fatalf("Format failed: %v", err)
    }
    
    if result == "" {
        t.Error("Expected non-empty result")
    }
}
```

## Performance Tips

1. **Reuse reporter instances** - Reporters are stateless and thread-safe
2. **Use appropriate filters** - Reduce output size by filtering early
3. **Limit errors** - Use `MaxErrors` option for large reports
4. **Disable colors** - When piping to files or non-terminal outputs
5. **Batch writes** - Use `WriteMultiple` for multiple reports

## Troubleshooting

### Empty Output
- Check if report has data
- Verify options aren't filtering everything
- Ensure report is not nil

### Missing Colors
- Ensure `ColorEnabled: true`
- Only works with text reporter
- Check terminal supports ANSI codes

### Slow Performance
- Use filters to reduce data
- Set `MaxErrors` limit
- Disable `Verbose` mode

### Memory Issues
- Process reports in batches
- Use streaming writers
- Implement pagination for large datasets

## Best Practices

1. **Choose the right format** for your use case
2. **Apply filters early** to reduce processing
3. **Handle errors** appropriately
4. **Use consistent options** across your application
5. **Test with edge cases** (empty reports, nil values)
6. **Document assumptions** in your code
7. **Validate filter criteria** before use

## See Also

- [README.md](README.md) - Full documentation
- [examples/reporter_example/main.go](../../../examples/reporter_example/main.go) - Complete examples
- [internal/ports/reporter.go](../../../internal/ports/reporter.go) - Interface definitions
- [internal/domain/validation.go](../../../internal/domain/validation.go) - Data structures
