package ebmlib

import (
	"context"
	"fmt"
	"io"

	"github.com/petergi/ebook-mechanic-lib/internal/adapters/epub"
	"github.com/petergi/ebook-mechanic-lib/internal/adapters/pdf"
	"github.com/petergi/ebook-mechanic-lib/internal/adapters/reporter"
	"github.com/petergi/ebook-mechanic-lib/internal/domain"
	"github.com/petergi/ebook-mechanic-lib/internal/ports"
)

// ValidationReport contains the results of ebook validation including errors, warnings, and metadata.
type ValidationReport = domain.ValidationReport

// ValidationError represents a single validation issue with code, message, severity, and location.
type ValidationError = domain.ValidationError

// Severity indicates the importance level of a validation error.
type Severity = domain.Severity

// ErrorLocation specifies where an error occurred within the ebook file.
type ErrorLocation = domain.ErrorLocation

const (
	// SeverityError indicates a critical validation failure that prevents ebook from being valid.
	SeverityError = domain.SeverityError

	// SeverityWarning indicates a non-critical issue that should be reviewed but doesn't invalidate the ebook.
	SeverityWarning = domain.SeverityWarning

	// SeverityInfo provides informational messages about the ebook structure or content.
	SeverityInfo = domain.SeverityInfo
)

// RepairResult contains the outcome of a repair operation including success status and applied actions.
type RepairResult = ports.RepairResult

// RepairPreview describes the repair actions that would be performed without actually modifying the file.
type RepairPreview = ports.RepairPreview

// RepairAction describes a single repair operation that can be performed on an ebook.
type RepairAction = ports.RepairAction

// ReportOptions configures how validation reports are formatted and displayed.
type ReportOptions = ports.ReportOptions

// OutputFormat specifies the format for validation report output.
type OutputFormat = ports.OutputFormat

const (
	// FormatJSON outputs validation reports in JSON format.
	FormatJSON = ports.FormatJSON

	// FormatText outputs validation reports in plain text format.
	FormatText = ports.FormatText

	// FormatHTML outputs validation reports in HTML format.
	FormatHTML = ports.FormatHTML

	// FormatXML outputs validation reports in XML format.
	FormatXML = ports.FormatXML

	// FormatMarkdown outputs validation reports in Markdown format.
	FormatMarkdown = ports.FormatMarkdown
)

// ValidateEPUB validates an EPUB file at the given path.
// It performs comprehensive validation including container structure, package document,
// content documents, and navigation document checks.
//
// Returns a ValidationReport containing all errors, warnings, and metadata found,
// or an error if the file cannot be read or parsed.
//
// Example:
//
//	report, err := ebmlib.ValidateEPUB("book.epub")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if !report.IsValid {
//	    fmt.Printf("Found %d errors\n", report.ErrorCount())
//	}
func ValidateEPUB(filePath string) (*ValidationReport, error) {
	return ValidateEPUBWithContext(context.Background(), filePath)
}

// ValidateEPUBWithContext validates an EPUB file with a context for cancellation and timeout support.
// The context can be used to cancel long-running validations or enforce timeouts.
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//	report, err := ebmlib.ValidateEPUBWithContext(ctx, "book.epub")
func ValidateEPUBWithContext(ctx context.Context, filePath string) (*ValidationReport, error) {
	validator := epub.NewEPUBValidator()
	return validator.ValidateFile(ctx, filePath)
}

// ValidateEPUBReader validates an EPUB from an io.Reader.
// This is useful for validating uploads, streams, or files from non-filesystem sources.
// The size parameter must be the total size of the EPUB file in bytes.
//
// Example:
//
//	file, _ := os.Open("book.epub")
//	defer file.Close()
//	info, _ := file.Stat()
//	report, err := ebmlib.ValidateEPUBReader(file, info.Size())
func ValidateEPUBReader(reader io.Reader, size int64) (*ValidationReport, error) {
	return ValidateEPUBReaderWithContext(context.Background(), reader, size)
}

// ValidateEPUBReaderWithContext validates an EPUB from an io.Reader with context support.
// Combines the benefits of ValidateEPUBReader and context-aware operations.
func ValidateEPUBReaderWithContext(ctx context.Context, reader io.Reader, size int64) (*ValidationReport, error) {
	validator := epub.NewEPUBValidator()
	return validator.ValidateReader(ctx, reader, size)
}

// ValidatePDF validates a PDF file at the given path.
// Performs structural validation including header, trailer, cross-reference table,
// catalog object, and document structure checks according to PDF 1.7 specification.
//
// Returns a ValidationReport containing all structural errors found,
// or an error if the file cannot be read or parsed.
//
// Example:
//
//	report, err := ebmlib.ValidatePDF("document.pdf")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if report.IsValid {
//	    fmt.Println("PDF structure is valid")
//	}
func ValidatePDF(filePath string) (*ValidationReport, error) {
	return ValidatePDFWithContext(context.Background(), filePath)
}

// ValidatePDFWithContext validates a PDF file with a context for cancellation and timeout support.
// The context can be used to cancel long-running validations or enforce timeouts.
//
// Example:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
//	defer cancel()
//	report, err := ebmlib.ValidatePDFWithContext(ctx, "document.pdf")
func ValidatePDFWithContext(_ context.Context, filePath string) (*ValidationReport, error) {
	validator := pdf.NewStructureValidator()
	result, err := validator.ValidateFile(filePath)
	if err != nil {
		return nil, err
	}
	return convertPDFValidationResult(filePath, result), nil
}

// ValidatePDFReader validates a PDF from an io.Reader.
// This is useful for validating uploads, streams, or files from non-filesystem sources.
//
// Example:
//
//	file, _ := os.Open("document.pdf")
//	defer file.Close()
//	report, err := ebmlib.ValidatePDFReader(file)
func ValidatePDFReader(reader io.Reader) (*ValidationReport, error) {
	return ValidatePDFReaderWithContext(context.Background(), reader)
}

// ValidatePDFReaderWithContext validates a PDF from an io.Reader with context support.
// Combines the benefits of ValidatePDFReader and context-aware operations.
func ValidatePDFReaderWithContext(_ context.Context, reader io.Reader) (*ValidationReport, error) {
	validator := pdf.NewStructureValidator()
	result, err := validator.ValidateReader(reader)
	if err != nil {
		return nil, err
	}
	return convertPDFValidationResult("", result), nil
}

// RepairEPUB attempts to automatically repair an EPUB file.
// First validates the file, then applies safe automatic repairs for common issues.
// Creates a backup of the original file before modifying.
//
// Returns a RepairResult indicating success and the actions taken,
// or an error if validation or repair fails.
//
// Example:
//
//	result, err := ebmlib.RepairEPUB("broken.epub")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if result.Success {
//	    fmt.Printf("Repaired! Backup at: %s\n", result.BackupPath)
//	}
func RepairEPUB(filePath string) (*RepairResult, error) {
	return RepairEPUBWithContext(context.Background(), filePath)
}

// RepairEPUBWithContext repairs an EPUB file with context support for cancellation and timeouts.
func RepairEPUBWithContext(ctx context.Context, filePath string) (*RepairResult, error) {
	report, err := ValidateEPUBWithContext(ctx, filePath)
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	if report.IsValid {
		return &RepairResult{
			Success:        true,
			ActionsApplied: []RepairAction{},
		}, nil
	}

	repairService := epub.NewRepairService()
	preview, err := repairService.Preview(ctx, report)
	if err != nil {
		return nil, fmt.Errorf("repair preview failed: %w", err)
	}

	return repairService.Apply(ctx, filePath, preview)
}

// RepairEPUBWithPreview applies a pre-generated repair preview to an EPUB file.
// This allows you to review and potentially modify the repair actions before applying them.
// The repaired file is saved to the specified outputPath.
//
// Example:
//
//	preview, _ := ebmlib.PreviewEPUBRepair("broken.epub")
//	// Review preview.Actions...
//	result, err := ebmlib.RepairEPUBWithPreview("broken.epub", preview, "fixed.epub")
func RepairEPUBWithPreview(filePath string, preview *RepairPreview, outputPath string) (*RepairResult, error) {
	return RepairEPUBWithPreviewContext(context.Background(), filePath, preview, outputPath)
}

// RepairEPUBWithPreviewContext applies a repair preview with context support.
func RepairEPUBWithPreviewContext(ctx context.Context, filePath string, preview *RepairPreview, outputPath string) (*RepairResult, error) {
	repairService := epub.NewRepairService()
	return repairService.ApplyWithBackup(ctx, filePath, preview, outputPath)
}

// PreviewEPUBRepair generates a preview of repair actions without modifying the file.
// This allows you to see what repairs would be performed before actually applying them.
//
// Returns a RepairPreview containing all proposed actions and whether they can be
// automatically applied, or an error if validation fails.
//
// Example:
//
//	preview, err := ebmlib.PreviewEPUBRepair("broken.epub")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Can auto-repair: %v\n", preview.CanAutoRepair)
//	for _, action := range preview.Actions {
//	    fmt.Printf("  - %s\n", action.Description)
//	}
func PreviewEPUBRepair(filePath string) (*RepairPreview, error) {
	return PreviewEPUBRepairWithContext(context.Background(), filePath)
}

// PreviewEPUBRepairWithContext generates a repair preview with context support.
func PreviewEPUBRepairWithContext(ctx context.Context, filePath string) (*RepairPreview, error) {
	report, err := ValidateEPUBWithContext(ctx, filePath)
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	repairService := epub.NewRepairService()
	return repairService.Preview(ctx, report)
}

// RepairPDF attempts to automatically repair a PDF file.
// First validates the file, then applies safe automatic repairs for common structural issues.
// Creates a backup of the original file before modifying.
//
// Returns a RepairResult indicating success and the actions taken,
// or an error if validation or repair fails.
//
// Example:
//
//	result, err := ebmlib.RepairPDF("broken.pdf")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if result.Success {
//	    fmt.Printf("Repaired! Backup at: %s\n", result.BackupPath)
//	}
func RepairPDF(filePath string) (*RepairResult, error) {
	return RepairPDFWithContext(context.Background(), filePath)
}

// RepairPDFWithContext repairs a PDF file with context support for cancellation and timeouts.
func RepairPDFWithContext(ctx context.Context, filePath string) (*RepairResult, error) {
	report, err := ValidatePDFWithContext(ctx, filePath)
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	if report.IsValid {
		return &RepairResult{
			Success:        true,
			ActionsApplied: []RepairAction{},
		}, nil
	}

	repairService := pdf.NewRepairService()
	preview, err := repairService.Preview(ctx, report)
	if err != nil {
		return nil, fmt.Errorf("repair preview failed: %w", err)
	}

	return repairService.Apply(ctx, filePath, preview)
}

// RepairPDFWithPreview applies a pre-generated repair preview to a PDF file.
// This allows you to review and potentially modify the repair actions before applying them.
// The repaired file is saved to the specified outputPath.
//
// Example:
//
//	preview, _ := ebmlib.PreviewPDFRepair("broken.pdf")
//	// Review preview.Actions...
//	result, err := ebmlib.RepairPDFWithPreview("broken.pdf", preview, "fixed.pdf")
func RepairPDFWithPreview(filePath string, preview *RepairPreview, outputPath string) (*RepairResult, error) {
	return RepairPDFWithPreviewContext(context.Background(), filePath, preview, outputPath)
}

// RepairPDFWithPreviewContext applies a repair preview with context support.
func RepairPDFWithPreviewContext(ctx context.Context, filePath string, preview *RepairPreview, outputPath string) (*RepairResult, error) {
	repairService := pdf.NewRepairService()
	return repairService.ApplyWithBackup(ctx, filePath, preview, outputPath)
}

// PreviewPDFRepair generates a preview of repair actions without modifying the file.
// This allows you to see what repairs would be performed before actually applying them.
//
// Returns a RepairPreview containing all proposed actions and whether they can be
// automatically applied, or an error if validation fails.
//
// Example:
//
//	preview, err := ebmlib.PreviewPDFRepair("broken.pdf")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Can auto-repair: %v\n", preview.CanAutoRepair)
//	for _, action := range preview.Actions {
//	    fmt.Printf("  - %s\n", action.Description)
//	}
func PreviewPDFRepair(filePath string) (*RepairPreview, error) {
	return PreviewPDFRepairWithContext(context.Background(), filePath)
}

// PreviewPDFRepairWithContext generates a repair preview with context support.
func PreviewPDFRepairWithContext(ctx context.Context, filePath string) (*RepairPreview, error) {
	report, err := ValidatePDFWithContext(ctx, filePath)
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	repairService := pdf.NewRepairService()
	return repairService.Preview(ctx, report)
}

// FormatReport converts a validation report to the specified output format.
// Supported formats include JSON, Text, and Markdown.
//
// Returns the formatted report as a string, or an error if formatting fails.
//
// Example:
//
//	report, _ := ebmlib.ValidateEPUB("book.epub")
//	jsonOutput, _ := ebmlib.FormatReport(report, ebmlib.FormatJSON)
//	textOutput, _ := ebmlib.FormatReport(report, ebmlib.FormatText)
func FormatReport(report *ValidationReport, format OutputFormat) (string, error) {
	return FormatReportWithContext(context.Background(), report, format)
}

// FormatReportWithContext formats a validation report with context support.
func FormatReportWithContext(ctx context.Context, report *ValidationReport, format OutputFormat) (string, error) {
	options := &ReportOptions{
		Format:          format,
		IncludeWarnings: true,
		IncludeInfo:     true,
		Verbose:         true,
	}
	return FormatReportWithOptions(ctx, report, options)
}

// FormatReportWithOptions formats a validation report with custom options.
// Allows fine-grained control over report content, verbosity, and formatting.
//
// Example:
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
func FormatReportWithOptions(ctx context.Context, report *ValidationReport, options *ReportOptions) (string, error) {
	var rep ports.Reporter

	switch options.Format {
	case FormatJSON:
		rep = reporter.NewJSONReporter()
	case FormatText:
		rep = reporter.NewTextReporter()
	case FormatMarkdown:
		rep = reporter.NewMarkdownReporter()
	case FormatHTML, FormatXML:
		return "", fmt.Errorf("unsupported format: %s", options.Format)
	default:
		return "", fmt.Errorf("unsupported format: %s", options.Format)
	}

	return rep.Format(ctx, report, options)
}

// WriteReport writes a validation report to an io.Writer with the specified options.
// This is useful for writing reports to files, network streams, or custom outputs.
//
// Example:
//
//	options := &ebmlib.ReportOptions{Format: ebmlib.FormatJSON}
//	file, _ := os.Create("report.json")
//	defer file.Close()
//	err := ebmlib.WriteReport(report, file, options)
func WriteReport(report *ValidationReport, writer io.Writer, options *ReportOptions) error {
	return WriteReportWithContext(context.Background(), report, writer, options)
}

// WriteReportWithContext writes a validation report to an io.Writer with context support.
func WriteReportWithContext(ctx context.Context, report *ValidationReport, writer io.Writer, options *ReportOptions) error {
	var rep ports.Reporter

	switch options.Format {
	case FormatJSON:
		rep = reporter.NewJSONReporter()
	case FormatText:
		rep = reporter.NewTextReporter()
	case FormatMarkdown:
		rep = reporter.NewMarkdownReporter()
	case FormatHTML, FormatXML:
		return fmt.Errorf("unsupported format: %s", options.Format)
	default:
		return fmt.Errorf("unsupported format: %s", options.Format)
	}

	return rep.Write(ctx, report, writer, options)
}

// WriteReportToFile writes a validation report to a file with the specified options.
// Automatically creates the file if it doesn't exist, or overwrites it if it does.
//
// Example:
//
//	options := &ebmlib.ReportOptions{
//	    Format:          ebmlib.FormatMarkdown,
//	    IncludeWarnings: true,
//	    Verbose:         true,
//	}
//	err := ebmlib.WriteReportToFile(report, "report.md", options)
func WriteReportToFile(report *ValidationReport, filePath string, options *ReportOptions) error {
	return WriteReportToFileWithContext(context.Background(), report, filePath, options)
}

// WriteReportToFileWithContext writes a validation report to a file with context support.
func WriteReportToFileWithContext(ctx context.Context, report *ValidationReport, filePath string, options *ReportOptions) error {
	var rep ports.Reporter

	switch options.Format {
	case FormatJSON:
		rep = reporter.NewJSONReporter()
	case FormatText:
		rep = reporter.NewTextReporter()
	case FormatMarkdown:
		rep = reporter.NewMarkdownReporter()
	case FormatHTML, FormatXML:
		return fmt.Errorf("unsupported format: %s", options.Format)
	default:
		return fmt.Errorf("unsupported format: %s", options.Format)
	}

	return rep.WriteToFile(ctx, report, filePath, options)
}

func convertPDFValidationResult(filePath string, result *pdf.StructureValidationResult) *ValidationReport {
	report := &ValidationReport{
		FilePath: filePath,
		FileType: "PDF",
		IsValid:  result.Valid,
		Errors:   make([]ValidationError, 0),
		Warnings: make([]ValidationError, 0),
		Info:     make([]ValidationError, 0),
		Metadata: make(map[string]interface{}),
	}

	for _, err := range result.Errors {
		report.Errors = append(report.Errors, ValidationError{
			Code:     err.Code,
			Message:  err.Message,
			Severity: SeverityError,
			Details:  err.Details,
		})
	}

	return report
}
