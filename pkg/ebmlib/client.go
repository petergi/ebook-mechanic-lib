package ebmlib

import (
	"context"
	"fmt"
	"io"

	"github.com/example/project/internal/adapters/epub"
	"github.com/example/project/internal/adapters/pdf"
	"github.com/example/project/internal/adapters/reporter"
	"github.com/example/project/internal/domain"
	"github.com/example/project/internal/ports"
)

type ValidationReport = domain.ValidationReport
type ValidationError = domain.ValidationError
type Severity = domain.Severity
type ErrorLocation = domain.ErrorLocation

const (
	SeverityError   = domain.SeverityError
	SeverityWarning = domain.SeverityWarning
	SeverityInfo    = domain.SeverityInfo
)

type RepairResult = ports.RepairResult
type RepairPreview = ports.RepairPreview
type RepairAction = ports.RepairAction

type ReportOptions = ports.ReportOptions
type OutputFormat = ports.OutputFormat

const (
	FormatJSON     = ports.FormatJSON
	FormatText     = ports.FormatText
	FormatHTML     = ports.FormatHTML
	FormatXML      = ports.FormatXML
	FormatMarkdown = ports.FormatMarkdown
)

func ValidateEPUB(filePath string) (*ValidationReport, error) {
	return ValidateEPUBWithContext(context.Background(), filePath)
}

func ValidateEPUBWithContext(ctx context.Context, filePath string) (*ValidationReport, error) {
	validator := epub.NewEPUBValidator()
	return validator.ValidateFile(ctx, filePath)
}

func ValidateEPUBReader(reader io.Reader, size int64) (*ValidationReport, error) {
	return ValidateEPUBReaderWithContext(context.Background(), reader, size)
}

func ValidateEPUBReaderWithContext(ctx context.Context, reader io.Reader, size int64) (*ValidationReport, error) {
	validator := epub.NewEPUBValidator()
	return validator.ValidateReader(ctx, reader, size)
}

func ValidatePDF(filePath string) (*ValidationReport, error) {
	return ValidatePDFWithContext(context.Background(), filePath)
}

func ValidatePDFWithContext(ctx context.Context, filePath string) (*ValidationReport, error) {
	validator := pdf.NewStructureValidator()
	result, err := validator.ValidateFile(filePath)
	if err != nil {
		return nil, err
	}
	return convertPDFValidationResult(filePath, result), nil
}

func ValidatePDFReader(reader io.Reader) (*ValidationReport, error) {
	return ValidatePDFReaderWithContext(context.Background(), reader)
}

func ValidatePDFReaderWithContext(ctx context.Context, reader io.Reader) (*ValidationReport, error) {
	validator := pdf.NewStructureValidator()
	result, err := validator.ValidateReader(reader)
	if err != nil {
		return nil, err
	}
	return convertPDFValidationResult("", result), nil
}

func RepairEPUB(filePath string) (*RepairResult, error) {
	return RepairEPUBWithContext(context.Background(), filePath)
}

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

func RepairEPUBWithPreview(filePath string, preview *RepairPreview, outputPath string) (*RepairResult, error) {
	return RepairEPUBWithPreviewContext(context.Background(), filePath, preview, outputPath)
}

func RepairEPUBWithPreviewContext(ctx context.Context, filePath string, preview *RepairPreview, outputPath string) (*RepairResult, error) {
	repairService := epub.NewRepairService()
	return repairService.ApplyWithBackup(ctx, filePath, preview, outputPath)
}

func PreviewEPUBRepair(filePath string) (*RepairPreview, error) {
	return PreviewEPUBRepairWithContext(context.Background(), filePath)
}

func PreviewEPUBRepairWithContext(ctx context.Context, filePath string) (*RepairPreview, error) {
	report, err := ValidateEPUBWithContext(ctx, filePath)
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	repairService := epub.NewRepairService()
	return repairService.Preview(ctx, report)
}

func RepairPDF(filePath string) (*RepairResult, error) {
	return RepairPDFWithContext(context.Background(), filePath)
}

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

func RepairPDFWithPreview(filePath string, preview *RepairPreview, outputPath string) (*RepairResult, error) {
	return RepairPDFWithPreviewContext(context.Background(), filePath, preview, outputPath)
}

func RepairPDFWithPreviewContext(ctx context.Context, filePath string, preview *RepairPreview, outputPath string) (*RepairResult, error) {
	repairService := pdf.NewRepairService()
	return repairService.ApplyWithBackup(ctx, filePath, preview, outputPath)
}

func PreviewPDFRepair(filePath string) (*RepairPreview, error) {
	return PreviewPDFRepairWithContext(context.Background(), filePath)
}

func PreviewPDFRepairWithContext(ctx context.Context, filePath string) (*RepairPreview, error) {
	report, err := ValidatePDFWithContext(ctx, filePath)
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	repairService := pdf.NewRepairService()
	return repairService.Preview(ctx, report)
}

func FormatReport(report *ValidationReport, format OutputFormat) (string, error) {
	return FormatReportWithContext(context.Background(), report, format)
}

func FormatReportWithContext(ctx context.Context, report *ValidationReport, format OutputFormat) (string, error) {
	options := &ReportOptions{
		Format:          format,
		IncludeWarnings: true,
		IncludeInfo:     true,
		Verbose:         true,
	}
	return FormatReportWithOptions(ctx, report, options)
}

func FormatReportWithOptions(ctx context.Context, report *ValidationReport, options *ReportOptions) (string, error) {
	var rep ports.Reporter

	switch options.Format {
	case FormatJSON:
		rep = reporter.NewJSONReporter()
	case FormatText:
		rep = reporter.NewTextReporter()
	case FormatMarkdown:
		rep = reporter.NewMarkdownReporter()
	default:
		return "", fmt.Errorf("unsupported format: %s", options.Format)
	}

	return rep.Format(ctx, report, options)
}

func WriteReport(report *ValidationReport, writer io.Writer, options *ReportOptions) error {
	return WriteReportWithContext(context.Background(), report, writer, options)
}

func WriteReportWithContext(ctx context.Context, report *ValidationReport, writer io.Writer, options *ReportOptions) error {
	var rep ports.Reporter

	switch options.Format {
	case FormatJSON:
		rep = reporter.NewJSONReporter()
	case FormatText:
		rep = reporter.NewTextReporter()
	case FormatMarkdown:
		rep = reporter.NewMarkdownReporter()
	default:
		return fmt.Errorf("unsupported format: %s", options.Format)
	}

	return rep.Write(ctx, report, writer, options)
}

func WriteReportToFile(report *ValidationReport, filePath string, options *ReportOptions) error {
	return WriteReportToFileWithContext(context.Background(), report, filePath, options)
}

func WriteReportToFileWithContext(ctx context.Context, report *ValidationReport, filePath string, options *ReportOptions) error {
	var rep ports.Reporter

	switch options.Format {
	case FormatJSON:
		rep = reporter.NewJSONReporter()
	case FormatText:
		rep = reporter.NewTextReporter()
	case FormatMarkdown:
		rep = reporter.NewMarkdownReporter()
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
