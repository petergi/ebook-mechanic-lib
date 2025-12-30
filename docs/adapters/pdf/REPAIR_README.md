# PDF Repair Service

A safe, automated repair service for basic PDF 1.7 structural issues, following the hexagonal architecture pattern.

## Features

- **Preview/Apply Workflow**: Inspect repairs before applying them
- **Safe Repairs Only**: Automatically repairs non-destructive issues
- **Backup Management**: Creates backups before modifications
- **Detailed Reporting**: Provides action-level repair information
- **Error Classification**: Distinguishes safe vs. unsafe repairs

## Supported Repairs

### Automated (Safe) Repairs

| Error Code | Issue | Repair Action | Safety Level |
|------------|-------|---------------|--------------|
| PDF-TRAILER-003 | Missing %%EOF marker | Append %%EOF to end of file | Very High |
| PDF-TRAILER-001 | Invalid startxref offset | Recompute xref table offset | High |
| PDF-TRAILER-002 | Trailer dictionary typos | Fix common typos (/Sise→/Size, /root→/Root) | High |

### Manual Intervention Required

| Error Code | Issue | Reason |
|------------|-------|--------|
| PDF-HEADER-001/002 | Invalid header | Header changes can corrupt structure |
| PDF-XREF-001/002/003 | Cross-reference issues | Requires complete structural rebuild |
| PDF-CATALOG-001/002/003 | Catalog problems | Affects document root structure |

See [REPAIR_LIMITATIONS.md](./REPAIR_LIMITATIONS.md) for detailed safety analysis.

## Installation

The PDF repair service is part of the ebm-lib adapters package:

```go
import "github.com/example/project/internal/adapters/pdf"
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/example/project/internal/adapters/pdf"
)

func main() {
    ctx := context.Background()
    repairService := pdf.NewRepairService()
    validator := pdf.NewStructureValidator()

    // 1. Validate the PDF
    result, err := validator.ValidateFile("broken.pdf")
    if err != nil {
        log.Fatal(err)
    }

    if result.Valid {
        fmt.Println("PDF is valid, no repairs needed")
        return
    }

    // 2. Convert to domain report
    report := convertValidationResult("broken.pdf", result)

    // 3. Preview repairs
    preview, err := repairService.Preview(ctx, report)
    if err != nil {
        log.Fatal(err)
    }

    // 4. Check if auto-repair is possible
    if !preview.CanAutoRepair {
        fmt.Println("Manual intervention required:")
        for _, warning := range preview.Warnings {
            fmt.Printf("  - %s\n", warning)
        }
        return
    }

    // 5. Show proposed actions
    fmt.Printf("Proposed repairs (%d actions):\n", len(preview.Actions))
    for i, action := range preview.Actions {
        fmt.Printf("%d. %s - %s\n", i+1, action.Type, action.Description)
    }

    // 6. Apply repairs (write repaired output to a new file)
    result, err := repairService.ApplyWithBackup(ctx, "broken.pdf", preview, "broken.repaired.pdf")
    if err != nil {
        log.Fatal(err)
    }

    if result.Success {
        fmt.Printf("✓ Repaired file saved to: %s\n", result.BackupPath)
        fmt.Printf("✓ Applied %d repairs\n", len(result.ActionsApplied))
    } else {
        fmt.Printf("✗ Repair failed: %v\n", result.Error)
    }
}
```

## API Reference

### Creating a Repair Service

```go
repairService := pdf.NewRepairService()
```

Returns an implementation of `ports.PDFRepairService` interface.

### Preview Repairs

```go
preview, err := repairService.Preview(ctx, report)
```

**Parameters:**
- `ctx`: Context for cancellation
- `report`: Validation report from structure validator

**Returns:**
- `preview`: Preview of repair actions
- `err`: Error if preview generation fails

**Preview Structure:**
```go
type RepairPreview struct {
    Actions        []RepairAction  // List of repair actions
    CanAutoRepair  bool           // True if all repairs are automated
    EstimatedTime  int64          // Estimated time in milliseconds
    BackupRequired bool           // True if backup should be created
    Warnings       []string       // Warnings about manual repairs needed
}
```

### Apply Repairs

```go
result, err := repairService.Apply(ctx, filePath, preview)
```

**Parameters:**
- `ctx`: Context for cancellation
- `filePath`: Path to PDF file to repair
- `preview`: Preview from Preview() call

**Returns:**
- `result`: Result of repair operation
- `err`: Error if operation fails

**Result Structure:**
```go
type RepairResult struct {
    Success        bool            // True if all repairs succeeded
    ActionsApplied []RepairAction  // Actions that were applied
    Report         *ValidationReport // Post-repair validation (optional)
    BackupPath     string          // Path passed to ApplyWithBackup (repaired output)
    Error          error           // Error if repair failed
}
```

### Apply with Custom Backup Path

```go
result, err := repairService.ApplyWithBackup(ctx, filePath, preview, backupPath)
```

Same as `Apply()` but allows specifying the output path.

### Check if Error is Repairable

```go
canRepair := repairService.CanRepair(ctx, validationError)
```

**Parameters:**
- `ctx`: Context for cancellation
- `validationError`: Single validation error

**Returns:**
- `bool`: True if error can be automatically repaired

### Backup Management

#### Create Backup

```go
err := repairService.CreateBackup(ctx, sourcePath, backupPath)
```

Creates a copy of the file before repairs.

#### Restore Backup

```go
err := repairService.RestoreBackup(ctx, backupPath, originalPath)
```

Restores a file from backup.

**Note:** `BackupPath` is the output path when using `ApplyWithBackup`. When using the CLI with `--in-place --backup`, the CLI reports the repaired output path separately from the backup of the original.

### High-Level Repair Functions

#### Repair Structure Issues

```go
result, err := repairService.RepairStructure(ctx, filePath)
```

Validates and repairs structural issues in one call.

#### Repair Metadata Issues

```go
result, err := repairService.RepairMetadata(ctx, filePath)
```

Reserved for future metadata repairs (currently returns success with no actions).

## Repair Actions

Each repair action has the following structure:

```go
type RepairAction struct {
    Type        string                 // Action type identifier
    Description string                 // Human-readable description
    Target      string                 // Target component (trailer, xref, etc.)
    Details     map[string]interface{} // Additional details
    Automated   bool                   // True if can be applied automatically
}
```

### Action Types

**Automated Actions:**
- `append_eof_marker`: Appends missing %%EOF marker
- `recompute_startxref`: Recalculates startxref offset
- `fix_trailer_typos`: Fixes common trailer dictionary typos

**Manual Actions:**
- `manual_header_fix`: Header requires manual intervention
- `manual_xref_rebuild`: Cross-reference rebuild needed
- `manual_catalog_fix`: Catalog repair needed
- `manual_review`: Generic manual review required

## Examples

### Example 1: Repair Missing EOF Marker

```go
ctx := context.Background()
service := pdf.NewRepairService()

// Create a validation report with missing EOF error
report := &domain.ValidationReport{
    FilePath: "test.pdf",
    IsValid:  false,
    Errors: []domain.ValidationError{
        {
            Code:    pdf.ErrorCodePDFTrailer003,
            Message: "Missing %%EOF marker",
        },
    },
}

preview, _ := service.Preview(ctx, report)
result, _ := service.Apply(ctx, "test.pdf", preview)

if result.Success {
    fmt.Println("EOF marker added successfully")
}
```

### Example 2: Handle Unsafe Repairs

```go
preview, _ := service.Preview(ctx, report)

if !preview.CanAutoRepair {
    fmt.Println("Cannot auto-repair. Manual intervention required:")
    
    for _, action := range preview.Actions {
        if !action.Automated {
            fmt.Printf("- %s: %s\n", action.Type, action.Description)
            if reason, ok := action.Details["reason"]; ok {
                fmt.Printf("  Reason: %s\n", reason)
            }
        }
    }
    
    // Guide user to manual tools
    fmt.Println("\nRecommended tools:")
    fmt.Println("- qpdf: For structural repairs")
    fmt.Println("- Ghostscript: For rewriting PDFs")
    fmt.Println("- Adobe Acrobat: For complex repairs")
}
```

### Example 3: Batch Repair with Rollback

```go
files := []string{"doc1.pdf", "doc2.pdf", "doc3.pdf"}

for _, file := range files {
    result, err := validator.ValidateFile(file)
    if err != nil {
        continue
    }
    
    if result.Valid {
        continue
    }
    
    report := convertToReport(file, result)
    preview, _ := service.Preview(ctx, report)
    
    if !preview.CanAutoRepair {
        log.Printf("%s: Requires manual intervention\n", file)
        continue
    }
    
    // Create explicit backup
    backupPath := file + ".backup"
    service.CreateBackup(ctx, file, backupPath)
    
    repairResult, _ := service.Apply(ctx, file, preview)
    
    if !repairResult.Success {
        log.Printf("%s: Repair failed, restoring backup\n", file)
        service.RestoreBackup(ctx, backupPath, file)
        continue
    }
    
    log.Printf("%s: Successfully repaired\n", file)
}
```

### Example 4: Detailed Action Reporting

```go
preview, _ := service.Preview(ctx, report)

fmt.Printf("Repair Preview for %s\n", report.FilePath)
fmt.Printf("==================%s\n", strings.Repeat("=", len(report.FilePath)))
fmt.Printf("Total Actions: %d\n", len(preview.Actions))
fmt.Printf("Can Auto-Repair: %v\n", preview.CanAutoRepair)
fmt.Printf("Backup Required: %v\n", preview.BackupRequired)
fmt.Printf("Estimated Time: %dms\n\n", preview.EstimatedTime)

for i, action := range preview.Actions {
    fmt.Printf("Action %d:\n", i+1)
    fmt.Printf("  Type: %s\n", action.Type)
    fmt.Printf("  Description: %s\n", action.Description)
    fmt.Printf("  Target: %s\n", action.Target)
    fmt.Printf("  Automated: %v\n", action.Automated)
    
    if len(action.Details) > 0 {
        fmt.Printf("  Details:\n")
        for k, v := range action.Details {
            fmt.Printf("    %s: %v\n", k, v)
        }
    }
    fmt.Println()
}

if len(preview.Warnings) > 0 {
    fmt.Println("Warnings:")
    for _, warning := range preview.Warnings {
        fmt.Printf("  ⚠ %s\n", warning)
    }
}
```

## Testing

Run the test suite:

```bash
go test ./internal/adapters/pdf/...
```

Run specific test:

```bash
go test -run TestApply_AppendEOFMarker ./internal/adapters/pdf/
```

Run with coverage:

```bash
go test -cover ./internal/adapters/pdf/
```

## Architecture

The repair service follows the hexagonal architecture pattern:

```
┌─────────────────────────────────────────┐
│         ports.PDFRepairService          │  Port (Interface)
│  - Preview(report) → RepairPreview      │
│  - Apply(path, preview) → RepairResult  │
│  - CanRepair(error) → bool              │
└─────────────────────────────────────────┘
                    ↑
                    │ implements
                    │
┌─────────────────────────────────────────┐
│        pdf.RepairServiceImpl            │  Adapter (Implementation)
│  - validator: StructureValidator        │
│  - generateRepairActions()              │
│  - applyRepairs()                       │
│  - appendEOFMarker()                    │
│  - recomputeStartxref()                 │
│  - fixTrailerTypos()                    │
└─────────────────────────────────────────┘
```

### Design Principles

1. **Preview Before Apply**: Always show what will be changed
2. **Safety First**: Only automate safe repairs
3. **Backup Always**: Create backups before modifications
4. **Explicit Actions**: Clear action types and descriptions
5. **No Side Effects**: Original file never modified directly

## Limitations

See [REPAIR_LIMITATIONS.md](./REPAIR_LIMITATIONS.md) for:
- Detailed safety analysis of each repair type
- Why certain repairs require manual intervention
- Recommended external tools for complex repairs
- Guidelines for fonts, compression, and structure tree repairs

## Integration with Validation

The repair service is designed to work with the structure validator:

```go
// Validation → Repair workflow
validator := pdf.NewStructureValidator()
repairService := pdf.NewRepairService()

// 1. Validate
validationResult, _ := validator.ValidateFile("document.pdf")

// 2. Convert errors to repair actions
report := convertToReport("document.pdf", validationResult)

// 3. Preview repairs
preview, _ := repairService.Preview(ctx, report)

// 4. Apply if safe
if preview.CanAutoRepair {
    result, _ := repairService.Apply(ctx, "document.pdf", preview)
}
```

## Error Handling

The service uses Go's standard error handling:

```go
result, err := repairService.Apply(ctx, filePath, preview)
if err != nil {
    // I/O error or operation error
    log.Fatal(err)
}

if !result.Success {
    // Repair logic error
    log.Printf("Repair failed: %v", result.Error)
}
```

**Error Types:**
- `err != nil`: I/O errors, invalid parameters, system errors
- `result.Error != nil`: Repair logic errors (e.g., cannot find xref table)

## Performance Considerations

- **Memory**: Loads entire PDF into memory for repairs
- **Speed**: Safe repairs are fast (< 100ms for typical files)
- **File Size**: Suitable for files up to 100MB in memory
- **Concurrency**: Service is stateless, safe for concurrent use

## Future Enhancements

Planned features (not yet implemented):

- [ ] Stream-based repairs for large files
- [ ] Incremental update handling
- [ ] Metadata repair implementation
- [ ] PDF/A conversion assistance
- [ ] Accessibility tag generation (with manual review)
- [ ] Compression optimization
- [ ] Interactive form repair

## Contributing

When adding new repair types:

1. Add error code constant in `structure_validator.go`
2. Implement repair logic in `repair_service.go`
3. Add repair action generation in `generateRepairActions()`
4. Add to `applyRepairs()` switch statement
5. Write unit tests in `repair_service_test.go`
6. Document safety level in `REPAIR_LIMITATIONS.md`
7. Update this README with examples

## License

Part of ebm-lib project. See LICENSE file for details.

## See Also

- [ERROR_CODES.md](./ERROR_CODES.md) - Complete error code reference
- [REPAIR_LIMITATIONS.md](./REPAIR_LIMITATIONS.md) - Safety guidelines and limitations
- [docs/specs/ebm-lib-PDF-SPEC.md](../../../docs/specs/ebm-lib-PDF-SPEC.md) - PDF validation spec
- [internal/ports/repair.go](../../../internal/ports/repair.go) - Repair service interface definition
