// Package ebmlib provides a simple, high-level API for validating and repairing EPUB and PDF ebooks.
//
// This package wraps the internal validation and repair services with a clean, easy-to-use interface
// suitable for library consumers. It handles the complexity of the hexagonal architecture internally
// and exposes only the essential types and functions needed for common ebook validation and repair tasks.
//
// # Quick Start
//
// Validate an EPUB file:
//
//	report, err := ebmlib.ValidateEPUB("book.epub")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	if report.IsValid {
//	    fmt.Println("✓ EPUB is valid!")
//	} else {
//	    fmt.Printf("Found %d errors\n", report.ErrorCount())
//	}
//
// Validate a PDF file:
//
//	report, err := ebmlib.ValidatePDF("document.pdf")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Repair an invalid EPUB:
//
//	result, err := ebmlib.RepairEPUB("broken.epub")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	if result.Success {
//	    fmt.Printf("Repaired: %s\n", result.BackupPath)
//	}
//
// # Validation
//
// The library supports validation of EPUB 3.0 and PDF files. Validation functions
// return a ValidationReport containing detailed information about any errors, warnings,
// or informational messages found in the file.
//
// EPUB validation checks:
//   - Container structure (mimetype, META-INF/container.xml)
//   - OPF package document (metadata, manifest, spine)
//   - Navigation documents
//   - Content documents (XHTML/HTML)
//
// PDF validation checks:
//   - File header and version
//   - Cross-reference table
//   - Trailer dictionary
//   - Catalog structure
//
// # Repair
//
// The repair functionality attempts to automatically fix common issues found during
// validation. You can preview repair actions before applying them:
//
//	preview, err := ebmlib.PreviewEPUBRepair("broken.epub")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	fmt.Printf("Can auto-repair: %v\n", preview.CanAutoRepair)
//	for _, action := range preview.Actions {
//	    fmt.Printf("  - %s\n", action.Description)
//	}
//
// Then apply the repairs:
//
//	result, err := ebmlib.RepairEPUB("broken.epub")
//
// Or apply with a custom output path:
//
//	result, err := ebmlib.RepairEPUBWithPreview("broken.epub", preview, "fixed.epub")
//
// # Reporting
//
// Validation reports can be formatted in multiple ways:
//
//	// JSON format
//	jsonOutput, err := ebmlib.FormatReport(report, ebmlib.FormatJSON)
//
//	// Plain text format
//	textOutput, err := ebmlib.FormatReport(report, ebmlib.FormatText)
//
//	// Markdown format
//	mdOutput, err := ebmlib.FormatReport(report, ebmlib.FormatMarkdown)
//
// You can also customize report options:
//
//	options := &ebmlib.ReportOptions{
//	    Format:          ebmlib.FormatText,
//	    IncludeWarnings: true,
//	    IncludeInfo:     false,
//	    Verbose:         true,
//	    ColorEnabled:    true,
//	    MaxErrors:       50,
//	}
//	output, err := ebmlib.FormatReportWithOptions(ctx, report, options)
//
// # Working with Readers
//
// The library supports validation from io.Reader for both file and stream processing:
//
//	file, _ := os.Open("book.epub")
//	defer file.Close()
//
//	info, _ := file.Stat()
//	report, err := ebmlib.ValidateEPUBReader(file, info.Size())
//
// This is useful for validating uploads, streams, or files from non-filesystem sources.
//
// # Context Support
//
// All validation and repair functions have context-aware variants for cancellation
// and timeout support:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//
//	report, err := ebmlib.ValidateEPUBWithContext(ctx, filePath)
//
// # Error Handling
//
// The library distinguishes between validation errors (issues with the ebook content)
// and operational errors (file system, parsing, etc.):
//
//   - Operational errors are returned as Go errors
//   - Validation errors are contained in the ValidationReport
//
// Example:
//
//	report, err := ebmlib.ValidateEPUB(filePath)
//	if err != nil {
//	    // Operational error (file not found, corrupted, etc.)
//	    log.Fatal(err)
//	}
//
//	if !report.IsValid {
//	    // Validation errors (invalid EPUB content)
//	    for _, validationErr := range report.Errors {
//	        fmt.Printf("[%s] %s\n", validationErr.Code, validationErr.Message)
//	    }
//	}
//
// # Type Aliases
//
// This package re-exports key types from the internal domain:
//
//   - ValidationReport: Contains validation results
//   - ValidationError: Represents a validation issue
//   - RepairResult: Contains repair operation results
//   - RepairPreview: Preview of repair actions
//   - ReportOptions: Configuration for report formatting
//
// # Thread Safety
//
// All functions in this package are safe for concurrent use. Each operation creates
// its own validator/repairer instance, so multiple goroutines can validate/repair
// different files simultaneously.
//
// # Performance
//
// Validation performance depends on file size and complexity:
//   - Small EPUBs (<1MB): typically <100ms
//   - Medium EPUBs (1-10MB): typically 100ms-1s
//   - Large EPUBs (>10MB): may take several seconds
//   - PDFs: typically faster than equivalent-size EPUBs
//
// For batch processing, consider validating files in parallel using goroutines.
package ebmlib
