# PDF Repair Service - Quick Reference

## One-Minute Start

```go
import "github.com/example/project/internal/adapters/pdf"

// Create service
service := pdf.NewRepairService()

// Preview repairs
preview, _ := service.Preview(ctx, report)

// Apply if safe
if preview.CanAutoRepair {
    result, _ := service.Apply(ctx, "file.pdf", preview)
    fmt.Println(result.BackupPath) // _repaired.pdf
}
```

## What Can Be Repaired?

| ✅ Automated | ❌ Manual Intervention |
|-------------|----------------------|
| Missing %%EOF marker | Invalid header |
| Incorrect startxref | Damaged xref table |
| Trailer typos | Corrupt catalog |
| | Font embedding |
| | Compression changes |
| | Structure tree |

## Error Codes

```go
// Automatically repairable
pdf.ErrorCodePDFTrailer003  // Missing %%EOF
pdf.ErrorCodePDFTrailer001  // Invalid startxref
pdf.ErrorCodePDFTrailer002  // Trailer typos

// Manual intervention required
pdf.ErrorCodePDFHeader001/002   // Header issues
pdf.ErrorCodePDFXref001/002/003 // Xref issues
pdf.ErrorCodePDFCatalog001/002/003 // Catalog issues
```

## Check Repairability

```go
canRepair := service.CanRepair(ctx, validationError)
if canRepair {
    // Safe to auto-repair
}
```

## Backup Management

```go
// Create backup
service.CreateBackup(ctx, "file.pdf", "file.pdf.backup")

// Apply repairs
result, _ := service.Apply(ctx, "file.pdf", preview)

// Rollback if needed
if !result.Success {
    service.RestoreBackup(ctx, "file.pdf.backup", "file.pdf")
}
```

## Batch Repair

```go
for _, file := range files {
    preview, _ := service.Preview(ctx, report)
    
    if preview.CanAutoRepair {
        service.Apply(ctx, file, preview)
    } else {
        log.Printf("%s: Manual intervention required", file)
    }
}
```

## Preview Structure

```go
preview := &ports.RepairPreview{
    Actions:        []RepairAction  // What will be done
    CanAutoRepair:  bool            // Is it safe?
    EstimatedTime:  int64           // Milliseconds
    BackupRequired: bool            // Need backup?
    Warnings:       []string        // Why not safe?
}
```

## Result Structure

```go
result := &ports.RepairResult{
    Success:        bool            // Did it work?
    ActionsApplied: []RepairAction  // What was done
    BackupPath:     string          // Where is repaired file?
    Error:          error           // What went wrong?
}
```

## Action Types

```go
// Automated
"append_eof_marker"    // Append %%EOF
"recompute_startxref"  // Fix startxref offset
"fix_trailer_typos"    // Fix /Sise → /Size, etc.

// Manual
"manual_header_fix"    // Header needs fixing
"manual_xref_rebuild"  // Xref needs rebuild
"manual_catalog_fix"   // Catalog needs fixing
"manual_review"        // Unknown issue
```

## Common Patterns

### Pattern 1: Simple

```go
preview, _ := service.Preview(ctx, report)
result, _ := service.Apply(ctx, "file.pdf", preview)
```

### Pattern 2: Safe Guard

```go
preview, _ := service.Preview(ctx, report)

if !preview.CanAutoRepair {
    for _, warning := range preview.Warnings {
        fmt.Println(warning)
    }
    return
}

result, _ := service.Apply(ctx, "file.pdf", preview)
```

### Pattern 3: Custom Backup

```go
preview, _ := service.Preview(ctx, report)
result, _ := service.ApplyWithBackup(ctx, "file.pdf", preview, "custom_path.pdf")
```

## When to Use External Tools

| Issue | Recommended Tool |
|-------|-----------------|
| Font embedding | Adobe Acrobat Pro, Ghostscript |
| Compression | Ghostscript, qpdf |
| Xref rebuild | qpdf, MuPDF mutool |
| Structure tree | Adobe Acrobat Pro |
| PDF/A conversion | veraPDF, Adobe Acrobat |
| Accessibility | CommonLook PDF, PAC |

## Documentation Links

- **Full API**: [REPAIR_README.md](./REPAIR_README.md)
- **Limitations**: [REPAIR_LIMITATIONS.md](./REPAIR_LIMITATIONS.md)
- **Implementation**: [IMPLEMENTATION_SUMMARY.md](./IMPLEMENTATION_SUMMARY.md)
- **Error Codes**: [ERROR_CODES.md](./ERROR_CODES.md)
- **Examples**: [../../../examples/pdf_repair/main.go](../../../examples/pdf_repair/main.go)

## Key Principles

1. **Preview before apply** - Always inspect repairs first
2. **Safety first** - Only automate safe repairs
3. **Backup always** - Create backups before modifications
4. **Be explicit** - Clear error messages and warnings
5. **Know limits** - Use external tools for complex repairs

## Testing

```bash
# Run all tests
go test ./internal/adapters/pdf/...

# Run specific test
go test -run TestApply_AppendEOFMarker ./internal/adapters/pdf/

# With coverage
go test -cover ./internal/adapters/pdf/
```

## Performance

- Small repairs: < 100ms
- Memory: Entire file loaded
- Max file size: ~100MB
- Concurrency: Thread-safe

## Need Help?

1. Check error code in [ERROR_CODES.md](./ERROR_CODES.md)
2. Review limitations in [REPAIR_LIMITATIONS.md](./REPAIR_LIMITATIONS.md)
3. See examples in [examples/pdf_repair/main.go](../../../examples/pdf_repair/main.go)
4. Read full API in [REPAIR_README.md](./REPAIR_README.md)
