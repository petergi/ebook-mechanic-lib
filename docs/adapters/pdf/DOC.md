# PDF Structure Validator Documentation

## Overview

The PDF Structure Validator implements comprehensive validation of PDF basic well-formedness according to PDF 1.7 specifications (ISO 32000-1:2008). It uses the unipdf library for robust PDF parsing and validation.

## Implementation Details

### Validation Strategy

The validator uses a two-phase approach:

1. **Pre-Parse Phase**: Fast checks on raw bytes
   - Header format validation
   - EOF marker presence
   - startxref keyword validation
   
2. **Parse Phase**: Deep structure validation using unipdf
   - Cross-reference table parsing
   - Catalog object validation
   - Object numbering verification

### Error Handling

All validation errors are collected rather than failing fast. This provides comprehensive feedback about all issues in a single validation pass.

```go
type ValidationError struct {
    Code     string                 // Unique error code (PDF-XXX-NNN)
    Message  string                 // Human-readable message
    Details  map[string]interface{} // Additional context
}
```

### Design Decisions

#### Why Pre-Parse Validation?

Pre-parse validation catches critical issues before attempting full PDF parsing:

- **Performance**: Avoids expensive parsing for obviously invalid files
- **Error Quality**: Provides specific error messages for common issues
- **Resource Protection**: Prevents parser crashes on severely malformed files

#### Why unipdf?

Selected for:
- Robust cross-reference table handling (both table and stream formats)
- Good error recovery for damaged PDFs
- Access to low-level PDF structure
- Active maintenance and good documentation

See ADR-003 for detailed comparison with pdfcpu.

#### Error Code Design

Error codes follow a hierarchical pattern:

```
PDF-[COMPONENT]-[NUMBER]
    │      │        │
    │      │        └─ Sequential number (001-999)
    │      └─────────── Component identifier
    └────────────────── Format identifier
```

Components:
- HEADER: File header issues
- TRAILER: Trailer and EOF issues
- XREF: Cross-reference issues
- CATALOG: Document catalog issues
- STRUCTURE: General structural issues

### Validation Checks

#### 1. Header Validation

**Check**: File starts with `%PDF-1.x` where x is 0-7

**Implementation**:
```go
headerPattern := regexp.MustCompile(`^%PDF-1\.[0-7]`)
if !headerPattern.Match(data) {
    // Error: PDF-HEADER-001 or PDF-HEADER-002
}
```

**Edge Cases**:
- Empty files
- Files with whitespace before header
- Files with comment after version
- Binary marker after header (common but optional)

#### 2. Trailer Validation

**Checks**:
- `%%EOF` marker present at end of file
- `startxref` keyword followed by valid offset
- Offset value is numeric and within file bounds

**Implementation**:
```go
// Check last 1KB for EOF marker (handles linearized PDFs)
lastBytes := data[len(data)-1024:]
eofIndex := bytes.LastIndex(lastBytes, []byte("%%EOF"))

// Validate startxref pattern
startxrefPattern := regexp.MustCompile(`startxref\s+(\d+)\s+%%EOF`)
```

**Edge Cases**:
- Linearized PDFs (two trailers)
- Incremental updates (multiple trailers)
- Whitespace variations around keywords
- Comments after EOF

#### 3. Cross-Reference Validation

**Checks**:
- Cross-reference table/stream exists
- Not empty (has object entries)
- No overlapping byte offsets
- All offsets are within file bounds

**Implementation**:
```go
xrefTable := parser.GetXrefTable()
offsets := make(map[int64][]int)
for _, objNum := range xrefTable.GetObjectNums() {
    xrefObj, _ := xrefTable.Get(objNum)
    offsets[xrefObj.Offset] = append(offsets[xrefObj.Offset], objNum)
}
// Check for overlaps
```

**Edge Cases**:
- Cross-reference streams (PDF 1.5+)
- Hybrid-reference files (table + stream)
- Free object entries
- Compressed object streams

#### 4. Catalog Validation

**Checks**:
- Catalog object exists and is reachable
- `/Type` entry equals `/Catalog`
- `/Pages` entry exists and is valid reference

**Implementation**:
```go
catalog := pdfReader.GetCatalog()
dict, ok := core.GetDict(catalogDict)

typeObj := dict.Get("Type")
typeName, ok := core.GetName(typeObj)
// Verify typeName.String() == "Catalog"

pagesObj := dict.Get("Pages")
// Verify pages object exists
```

**Edge Cases**:
- Indirect catalog reference
- Missing optional entries
- Invalid object references
- Circular references

#### 5. Object Numbering Validation

**Check**: No duplicate object number/generation pairs

**Implementation**:
```go
seenObjects := make(map[string]bool)
for _, objNum := range xrefTable.GetObjectNums() {
    xrefObj, _ := xrefTable.Get(objNum)
    key := fmt.Sprintf("%d_%d", objNum, xrefObj.Generation)
    if seenObjects[key] {
        // Error: PDF-STRUCTURE-012
    }
}
```

**Edge Cases**:
- Object streams (objects within objects)
- Free list entries
- Generation number increments

## Testing Strategy

### Test Categories

1. **Positive Tests**: Valid PDFs should pass
   - All supported versions (1.0-1.7)
   - Minimal valid structure
   - Complex multi-page documents

2. **Negative Tests**: Invalid PDFs should fail with specific errors
   - Missing/invalid headers
   - Truncated files
   - Damaged structures
   - Missing required objects

3. **Edge Case Tests**:
   - Empty files
   - Single-byte files
   - Large files
   - Multiple simultaneous errors

### Test Data Generation

Tests use in-memory PDF generation rather than fixture files:

**Advantages**:
- No binary blobs in repository
- Easy to modify and extend
- Clear test intent in code
- Version control friendly

**Example**:
```go
func createPDFWithMissingEOF() []byte {
    return []byte(`%PDF-1.4
1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj
xref
0 2
0000000000 65535 f 
0000000009 00000 n 
trailer
<< /Size 2 /Root 1 0 R >>
startxref
58
`)  // Note: no %%EOF
}
```

### Coverage Goals

- All error codes triggered by at least one test
- All validation functions covered
- All edge cases documented and tested
- Integration with unipdf library validated

## Performance Considerations

### Optimization Strategies

1. **Fast Fail on Header**: Check header before parsing
2. **Buffered Reading**: Read last 1KB for EOF check instead of full file
3. **Lazy Evaluation**: Only parse objects when needed
4. **Error Accumulation**: Collect errors without throwing exceptions

### Memory Usage

- **Small Files** (<1MB): Load entire file into memory
- **Large Files** (>1MB): Stream validation where possible
- **Cross-Reference**: Keep xref table in memory (typically <1% of file size)

### Typical Performance

- **Valid PDF**: 1-5ms for files <100KB
- **Invalid Header**: <1ms (pre-parse rejection)
- **Damaged Xref**: 10-50ms (parser recovery attempts)

## Error Messages

### Message Guidelines

1. **Be Specific**: Indicate exact problem
2. **Be Actionable**: Suggest how to fix when possible
3. **Include Context**: Add relevant details (offsets, values)
4. **Be Consistent**: Use standard terminology

### Example Error Messages

**Good**:
```
"Missing %%EOF marker"
"Invalid PDF version number (found: 1.9, expected: 1.0-1.7)"
"Cross-reference table has overlapping entries at offset 1234"
```

**Bad**:
```
"Invalid file"
"Error parsing"
"Something is wrong with the PDF"
```

## Repair Service

The PDF Repair Service provides safe, automated repairs for basic PDF structural issues. It follows a preview/apply workflow to ensure safety and transparency.

### Key Features

- **Preview Before Apply**: Inspect all proposed repairs before applying them
- **Safe Repairs Only**: Only automates non-destructive repairs
- **Backup Management**: Automatic backup creation before modifications
- **Detailed Reporting**: Action-level repair information
- **Error Classification**: Distinguishes safe vs. unsafe repairs

### Supported Repairs

**Automated (Safe)**:
- Append missing `%%EOF` markers (PDF-TRAILER-003)
- Recompute incorrect `startxref` offsets (PDF-TRAILER-001)
- Fix minor trailer dictionary typos (PDF-TRAILER-002)

**Requires Manual Intervention**:
- Header modifications (PDF-HEADER-001/002)
- Cross-reference table rebuild (PDF-XREF-001/002/003)
- Catalog repairs (PDF-CATALOG-001/002/003)
- Font embedding and subsetting
- Compression scheme changes
- Structure tree modifications

### Usage Example

```go
repairService := pdf.NewRepairService()
validator := pdf.NewStructureValidator()

// Validate
result, _ := validator.ValidateFile("broken.pdf")
report := convertToReport("broken.pdf", result)

// Preview repairs
preview, _ := repairService.Preview(ctx, report)

// Apply if safe
if preview.CanAutoRepair {
    result, _ := repairService.Apply(ctx, "broken.pdf", preview)
    fmt.Printf("Repaired: %s\n", result.BackupPath)
}
```

### Documentation

- **[REPAIR_README.md](./REPAIR_README.md)**: Complete API documentation and usage examples
- **[REPAIR_LIMITATIONS.md](./REPAIR_LIMITATIONS.md)**: Safety guidelines and limitations for each repair type

### Design Principles

1. **Safety First**: Never modify original file directly
2. **Transparency**: Always show what will be changed
3. **Conservative Approach**: Only automate safe repairs
4. **Backup Always**: Create backups before any modification
5. **Clear Communication**: Explain why manual intervention is needed

## Future Enhancements

### Phase 2: PDF/A Validation
- XMP metadata validation
- Font embedding checks
- Color space validation
- Compression method checks

### Phase 3: PDF/UA Validation
- Tagged structure validation
- Alternative text presence
- Reading order validation
- Language tagging

### Additional Features
- ✅ Repair service (implemented)
- Performance profiling
- Validation reports (JSON/HTML)
- Batch validation support
- Stream-based repairs for large files
- Metadata repair implementation

## References

- **PDF 1.7 Specification**: ISO 32000-1:2008
- **unipdf Documentation**: https://github.com/unidoc/unipdf
- **Project Spec**: docs/specs/ebm-lib-PDF-SPEC.md
- **ADR-003**: docs/adr/ADR-003-unipdf-over-pdfcpu.md
- **Error Codes**: [ERROR_CODES.md](./ERROR_CODES.md)
