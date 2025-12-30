# PDF Repair Service - Limitations and Safety Guidelines

## Overview

The PDF Repair Service in ebm-lib provides **safe, automated repairs** for basic PDF 1.7 structural issues. This document outlines what can be safely repaired, what requires manual intervention, and why certain repairs are considered unsafe.

## Safe Repairs (Automated)

The following repairs are considered **safe** and can be performed automatically without user confirmation:

### 1. Append Missing %%EOF Marker (PDF-TRAILER-003)

**What it does:** Appends the required `%%EOF` end-of-file marker if missing.

**Safety Level:** Very High  
**Repairable:** Yes (Automated)  
**Risk:** Minimal - only adds missing marker at end of file

```go
// Example: Missing EOF repair
Before: ...startxref\n123\n
After:  ...startxref\n123\n%%EOF\n
```

**Limitations:**
- Does not verify if file was genuinely truncated
- Cannot recover lost content if file was cut short
- Only adds marker, does not validate file completeness

### 2. Recompute startxref Offset (PDF-TRAILER-001)

**What it does:** Recalculates the byte offset to the cross-reference table by scanning for the last `xref` keyword.

**Safety Level:** High  
**Repairable:** Yes (Automated)  
**Risk:** Low - only updates numerical offset value

```go
// Example: Incorrect startxref repair
Before: startxref\n999999\n%%EOF
After:  startxref\n107\n%%EOF
```

**Limitations:**
- Assumes the xref table itself is valid and present
- Cannot fix if multiple xref tables have structural issues
- Will use the last xref found, which may not always be correct for incremental updates
- Does not validate xref table contents

### 3. Fix Minor Trailer Dictionary Typos (PDF-TRAILER-002)

**What it does:** Corrects common typos in trailer dictionary entries (e.g., `/Sise` → `/Size`, `/root` → `/Root`).

**Safety Level:** High  
**Repairable:** Yes (Automated)  
**Risk:** Low - only fixes known typos in dictionary keys

```go
// Example: Trailer typo repair
Before: trailer << /Sise 10 /root 1 0 R >>
After:  trailer << /Size 10 /Root 1 0 R >>
```

**Limitations:**
- Only fixes known, common typos
- Does not validate dictionary values (only keys)
- Cannot fix structural dictionary errors
- Cannot rebuild missing trailer dictionaries
- Does not fix arbitrary syntax errors

## Unsafe Repairs (Manual Intervention Required)

The following repairs are considered **unsafe** and require manual intervention using specialized PDF repair tools:

### 1. Font Embedding and Subsetting

**Error Codes:** PDFA-FONT-* (PDF/A specific)

**Why Unsafe:**
- Requires access to original font files
- Font subsetting involves complex glyph extraction
- Unicode mapping generation requires font metrics
- Incorrect embedding can make text unreadable or unsearchable
- Legal issues with font licensing

**Recommended Tools:**
- Adobe Acrobat Preflight
- veraPDF (for PDF/A)
- Font embedding utilities

**Manual Steps:**
1. Identify which fonts need embedding
2. Obtain valid font files (licensed)
3. Use professional tool to embed/subset fonts
4. Validate ToUnicode CMap correctness

### 2. Compression Schemes

**Error Codes:** PDFA-COMPRESS-*

**Why Unsafe:**
- Requires decompression and re-encoding of content streams
- May alter image quality or introduce artifacts
- Complex filters (LZW, JBIG2, old JPEG) need specialized handling
- Risk of data loss or corruption
- Can significantly change file size

**Recommended Tools:**
- Ghostscript (with appropriate settings)
- Adobe Acrobat (Save As PDF/A)
- qpdf (for filter replacement)

**Manual Steps:**
1. Identify problematic compression filters
2. Extract and decompress content streams
3. Re-encode using allowed compression (Flate, JPEG2000)
4. Validate visual output matches original

### 3. Structure Tree and Tagged PDF

**Error Codes:** PDFUA-TAG-*, PDFUA-STRUCT-*

**Why Unsafe:**
- Requires semantic understanding of content
- Tag hierarchy affects reading order and accessibility
- Incorrect tagging can break screen readers
- Manual review needed for meaningful Alt text
- Automatic tag generation is heuristic and error-prone

**Recommended Tools:**
- Adobe Acrobat Pro (Accessibility tools)
- CommonLook PDF (tagging and remediation)
- PAC (PDF Accessibility Checker)

**Manual Steps:**
1. Analyze document structure and reading order
2. Create logical tag hierarchy (headings, lists, tables)
3. Add meaningful alternative text for images
4. Set appropriate tag roles and attributes
5. Validate with screen reader testing

### 4. Cross-Reference Table Rebuild

**Error Codes:** PDF-XREF-001, PDF-XREF-002, PDF-XREF-003

**Why Unsafe:**
- Requires complete file scan and object inventory
- May need to resolve object number conflicts
- Incremental updates complicate rebuild process
- Risk of missing or duplicating objects
- Can alter object generation numbers

**Recommended Tools:**
- qpdf (--linearize or repair mode)
- Ghostscript (re-writing PDF)
- pdftk (with repair flag)
- MuPDF mutool clean

**Manual Steps:**
1. Use specialized tool to scan for all objects
2. Rebuild xref table with correct offsets
3. Validate all object references resolve
4. Check for duplicate object numbers
5. Re-validate entire document structure

### 5. Document Catalog Repairs

**Error Codes:** PDF-CATALOG-001, PDF-CATALOG-002, PDF-CATALOG-003

**Why Unsafe:**
- Catalog is the root of document structure
- Missing or corrupt catalog requires structural rebuild
- Page tree reconstruction is complex
- Metadata and navigation may be lost
- Can affect document behavior and interactivity

**Recommended Tools:**
- pdftk (repair with template)
- qpdf (--linearize)
- Ghostscript (full rewrite)

**Manual Steps:**
1. Extract valid objects from corrupted file
2. Create new minimal catalog structure
3. Rebuild page tree from page objects
4. Restore metadata from damaged catalog
5. Re-establish interactive features (forms, links)

### 6. PDF Header Modifications

**Error Codes:** PDF-HEADER-001, PDF-HEADER-002

**Why Unsafe:**
- Changing header can affect file interpretation
- Binary data may immediately follow header
- Version change affects feature availability
- File format detection relies on header
- Can break compatibility with some readers

**Recommended Tools:**
- Manual hex editor (with caution)
- qpdf (for version upgrade/downgrade)
- Ghostscript (with explicit version setting)

**Manual Steps:**
1. Analyze file content for version-specific features
2. Use hex editor to modify header (risky)
3. Or use qpdf to re-write with correct version
4. Validate with multiple PDF readers
5. Check for rendering differences

## General Guidelines

### When to Repair Automatically

✅ **Safe to auto-repair:**
- Missing EOF marker
- Incorrect startxref offset (if xref table valid)
- Common trailer dictionary typos
- Missing newlines or whitespace issues

### When to Require Manual Intervention

⚠️ **Requires manual review:**
- Any change affecting visible content
- Font embedding or text encoding
- Image compression or color space changes
- Structure tree or accessibility tags
- Cross-reference rebuild
- Catalog reconstruction
- Header version changes

### Repair Workflow

1. **Validation Phase**
   - Run structure validation
   - Generate validation report
   - Identify error codes

2. **Preview Phase** (Dry-Run)
   - Call `Preview()` to see proposed repairs
   - Review automated vs. manual actions
   - Check warnings for unsafe operations

3. **Backup Phase**
   - Always create backup before repairs
   - Use `CreateBackup()` or allow automatic backup
   - Store backup path for potential restore

4. **Repair Phase**
   - Apply automated repairs with `Apply()`
   - Document manual repairs separately
   - Log all actions taken

5. **Validation Phase** (Post-Repair)
   - Re-validate repaired file
   - Compare before/after reports
   - Verify with PDF readers

6. **Rollback if Needed**
   - Use `RestoreBackup()` if repairs fail
   - Investigate why repairs didn't work
   - Consider manual intervention

## Code Example

```go
// Safe repair workflow
validator := pdf.NewStructureValidator()
repairService := pdf.NewRepairService()
ctx := context.Background()

// 1. Validate
result, err := validator.ValidateFile("document.pdf")
if err != nil {
    log.Fatal(err)
}

// Convert to domain report
report := convertToDomainReport(result)

// 2. Preview repairs
preview, err := repairService.Preview(ctx, report)
if err != nil {
    log.Fatal(err)
}

// 3. Check if safe to auto-repair
if !preview.CanAutoRepair {
    fmt.Println("Manual intervention required:")
    for _, warning := range preview.Warnings {
        fmt.Println("-", warning)
    }
    return
}

// 4. Apply automated repairs
result, err := repairService.Apply(ctx, "document.pdf", preview)
if err != nil {
    log.Fatal(err)
}

if result.Success {
    fmt.Printf("Repaired file saved to: %s\n", result.BackupPath)
    fmt.Printf("Actions applied: %d\n", len(result.ActionsApplied))
} else {
    fmt.Printf("Repair failed: %v\n", result.Error)
}
```

## External Tool Recommendations

### For Font Issues
- **Adobe Acrobat Pro DC** - Professional font embedding and subsetting
- **Ghostscript** - Command-line font embedding with `pdf2ps` and `ps2pdf`

### For Compression Issues
- **Ghostscript** - Recompress with `gs -sDEVICE=pdfwrite`
- **qpdf** - Linearize and optimize with `qpdf --linearize`

### For Structure Issues
- **qpdf** - Structural repairs with `qpdf --check` and repair modes
- **pdftk** - Simple structural operations
- **MuPDF mutool** - Clean and repair PDFs

### For Accessibility
- **Adobe Acrobat Pro** - Complete accessibility toolset
- **CommonLook PDF** - Professional PDF remediation
- **PAC (PDF Accessibility Checker)** - Free validation and guidance

### For PDF/A Conversion
- **veraPDF** - PDF/A validation and repair guidance
- **Adobe Acrobat Pro** - Save As PDF/A with preflight
- **Ghostscript** - Convert to PDF/A (limited)

## Implementation Notes

### Error Handling
- All repair operations return detailed error information
- Partial failures are logged but don't stop other repairs
- Original file is never modified directly
- Backup files use `_repaired.pdf` suffix

### Performance
- Small repairs (EOF, startxref) are near-instantaneous
- Large file repairs may take time for I/O
- No in-memory decompression of streams
- Minimal PDF parsing for safe repairs

### Testing
- Unit tests cover each repair scenario
- Integration tests validate end-to-end workflow
- Test fixtures include various corruption types
- Negative tests ensure unsafe repairs are blocked

## Conclusion

The PDF Repair Service focuses on **safe, automated repairs** that:
- Don't modify visible content
- Don't alter document structure significantly
- Don't require external resources (fonts, images)
- Don't change semantic meaning
- Can be applied without expert knowledge

For complex repairs involving fonts, compression, structure trees, or significant corruption, **always use professional PDF tools** and manual intervention.
