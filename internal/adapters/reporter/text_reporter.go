package reporter

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/example/project/internal/domain"
	"github.com/example/project/internal/ports"
)

// TextReporter formats validation reports as styled plain text.
type TextReporter struct {
	filter *Filter
}

// NewTextReporter returns a text reporter without filters.
func NewTextReporter() ports.Reporter {
	return &TextReporter{}
}

// NewTextReporterWithFilter returns a text reporter with a filter applied.
func NewTextReporterWithFilter(filter *Filter) ports.Reporter {
	return &TextReporter{
		filter: filter,
	}
}

// Format renders a single report as text.
func (r *TextReporter) Format(_ context.Context, report *domain.ValidationReport, options *ports.ReportOptions) (string, error) {
	var sb strings.Builder

	colors := NewColorScheme(options != nil && options.ColorEnabled)

	errors := report.Errors
	warnings := report.Warnings
	info := report.Info

	if r.filter != nil {
		errors = r.filter.FilterErrors(errors)
		warnings = r.filter.FilterErrors(warnings)
		info = r.filter.FilterErrors(info)
	}

	if options != nil {
		if !options.IncludeWarnings {
			warnings = []domain.ValidationError{}
		}
		if !options.IncludeInfo {
			info = []domain.ValidationError{}
		}
		if options.MaxErrors > 0 && len(errors) > options.MaxErrors {
			errors = errors[:options.MaxErrors]
		}
	}

	sb.WriteString(colors.ColorizeHeader("═══════════════════════════════════════════════════════════════\n"))
	sb.WriteString(colors.ColorizeHeader(fmt.Sprintf("  VALIDATION REPORT: %s\n", report.FilePath)))
	sb.WriteString(colors.ColorizeHeader("═══════════════════════════════════════════════════════════════\n\n"))

	statusText := "VALID"
	statusColor := colors.Success
	if !report.IsValid {
		statusText = "INVALID"
		statusColor = colors.Error
	}
	sb.WriteString(fmt.Sprintf("Status:          %s\n", colors.Colorize(statusText, statusColor)))
	sb.WriteString(fmt.Sprintf("File Type:       %s\n", report.FileType))
	sb.WriteString(fmt.Sprintf("Validation Time: %s\n", report.ValidationTime.Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("Duration:        %s\n\n", report.Duration))

	sb.WriteString(colors.ColorizeHeader("SUMMARY\n"))
	sb.WriteString(strings.Repeat("─", 63) + "\n")
	totalIssues := len(errors) + len(warnings) + len(info)
	sb.WriteString(fmt.Sprintf("Total Issues:    %d\n", totalIssues))
	sb.WriteString(fmt.Sprintf("Errors:          %s\n", colors.ColorizeError(fmt.Sprintf("%d", len(errors)))))
	sb.WriteString(fmt.Sprintf("Warnings:        %s\n", colors.ColorizeWarning(fmt.Sprintf("%d", len(warnings)))))
	sb.WriteString(fmt.Sprintf("Info:            %s\n\n", colors.ColorizeInfo(fmt.Sprintf("%d", len(info)))))

	if len(errors) > 0 {
		sb.WriteString(colors.ColorizeError("ERRORS\n"))
		sb.WriteString(strings.Repeat("─", 63) + "\n")
		r.writeErrors(&sb, errors, colors, options)
		sb.WriteString("\n")
	}

	if len(warnings) > 0 {
		sb.WriteString(colors.ColorizeWarning("WARNINGS\n"))
		sb.WriteString(strings.Repeat("─", 63) + "\n")
		r.writeErrors(&sb, warnings, colors, options)
		sb.WriteString("\n")
	}

	if len(info) > 0 {
		sb.WriteString(colors.ColorizeInfo("INFORMATION\n"))
		sb.WriteString(strings.Repeat("─", 63) + "\n")
		r.writeErrors(&sb, info, colors, options)
		sb.WriteString("\n")
	}

	if options != nil && options.Verbose && len(report.Metadata) > 0 {
		sb.WriteString(colors.ColorizeHeader("METADATA\n"))
		sb.WriteString(strings.Repeat("─", 63) + "\n")
		for key, value := range report.Metadata {
			sb.WriteString(fmt.Sprintf("%s: %v\n", key, value))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(colors.ColorizeHeader("═══════════════════════════════════════════════════════════════\n"))

	return sb.String(), nil
}

// Write writes a text report to the provided writer.
func (r *TextReporter) Write(ctx context.Context, report *domain.ValidationReport, writer io.Writer, options *ports.ReportOptions) error {
	formatted, err := r.Format(ctx, report, options)
	if err != nil {
		return err
	}

	_, err = writer.Write([]byte(formatted))
	return err
}

// WriteToFile writes a text report to a file.
func (r *TextReporter) WriteToFile(ctx context.Context, report *domain.ValidationReport, filePath string, options *ports.ReportOptions) error {
	file, err := os.Create(filePath) //nolint:gosec
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	return r.Write(ctx, report, file, options)
}

// FormatMultiple renders multiple reports as text.
func (r *TextReporter) FormatMultiple(ctx context.Context, reports []*domain.ValidationReport, options *ports.ReportOptions) (string, error) {
	var sb strings.Builder

	colors := NewColorScheme(options != nil && options.ColorEnabled)

	totalErrors := 0
	totalWarnings := 0
	totalInfo := 0
	validFiles := 0
	invalidFiles := 0

	for _, report := range reports {
		if report.IsValid {
			validFiles++
		} else {
			invalidFiles++
		}
		totalErrors += report.ErrorCount()
		totalWarnings += report.WarningCount()
		totalInfo += report.InfoCount()
	}

	sb.WriteString(colors.ColorizeHeader("═══════════════════════════════════════════════════════════════\n"))
	sb.WriteString(colors.ColorizeHeader("  VALIDATION SUMMARY\n"))
	sb.WriteString(colors.ColorizeHeader("═══════════════════════════════════════════════════════════════\n\n"))

	sb.WriteString(fmt.Sprintf("Total Files:     %d\n", len(reports)))
	sb.WriteString(fmt.Sprintf("Valid Files:     %s\n", colors.ColorizeSuccess(fmt.Sprintf("%d", validFiles))))
	sb.WriteString(fmt.Sprintf("Invalid Files:   %s\n", colors.ColorizeError(fmt.Sprintf("%d", invalidFiles))))
	sb.WriteString(fmt.Sprintf("Total Issues:    %d\n", totalErrors+totalWarnings+totalInfo))
	sb.WriteString(fmt.Sprintf("Errors:          %s\n", colors.ColorizeError(fmt.Sprintf("%d", totalErrors))))
	sb.WriteString(fmt.Sprintf("Warnings:        %s\n", colors.ColorizeWarning(fmt.Sprintf("%d", totalWarnings))))
	sb.WriteString(fmt.Sprintf("Info:            %s\n\n", colors.ColorizeInfo(fmt.Sprintf("%d", totalInfo))))

	sb.WriteString(colors.ColorizeHeader("═══════════════════════════════════════════════════════════════\n\n"))

	for i, report := range reports {
		reportStr, err := r.Format(ctx, report, options)
		if err != nil {
			return "", err
		}
		sb.WriteString(reportStr)

		if i < len(reports)-1 {
			sb.WriteString("\n")
		}
	}

	return sb.String(), nil
}

// WriteMultiple writes multiple text reports to the provided writer.
func (r *TextReporter) WriteMultiple(ctx context.Context, reports []*domain.ValidationReport, writer io.Writer, options *ports.ReportOptions) error {
	formatted, err := r.FormatMultiple(ctx, reports, options)
	if err != nil {
		return err
	}

	_, err = writer.Write([]byte(formatted))
	return err
}

// WriteSummary writes a text summary for multiple reports.
func (r *TextReporter) WriteSummary(_ context.Context, reports []*domain.ValidationReport, writer io.Writer, options *ports.ReportOptions) error {
	var sb strings.Builder

	colors := NewColorScheme(options != nil && options.ColorEnabled)

	totalErrors := 0
	totalWarnings := 0
	totalInfo := 0
	validFiles := 0
	invalidFiles := 0

	for _, report := range reports {
		if report.IsValid {
			validFiles++
		} else {
			invalidFiles++
		}
		totalErrors += report.ErrorCount()
		totalWarnings += report.WarningCount()
		totalInfo += report.InfoCount()
	}

	sb.WriteString(colors.ColorizeHeader("═══════════════════════════════════════════════════════════════\n"))
	sb.WriteString(colors.ColorizeHeader("  VALIDATION SUMMARY\n"))
	sb.WriteString(colors.ColorizeHeader("═══════════════════════════════════════════════════════════════\n\n"))

	sb.WriteString(fmt.Sprintf("Total Files:     %d\n", len(reports)))
	sb.WriteString(fmt.Sprintf("Valid Files:     %s\n", colors.ColorizeSuccess(fmt.Sprintf("%d", validFiles))))
	sb.WriteString(fmt.Sprintf("Invalid Files:   %s\n", colors.ColorizeError(fmt.Sprintf("%d", invalidFiles))))
	sb.WriteString(fmt.Sprintf("Total Issues:    %d\n", totalErrors+totalWarnings+totalInfo))
	sb.WriteString(fmt.Sprintf("Errors:          %s\n", colors.ColorizeError(fmt.Sprintf("%d", totalErrors))))
	sb.WriteString(fmt.Sprintf("Warnings:        %s\n", colors.ColorizeWarning(fmt.Sprintf("%d", totalWarnings))))
	sb.WriteString(fmt.Sprintf("Info:            %s\n\n", colors.ColorizeInfo(fmt.Sprintf("%d", totalInfo))))

	sb.WriteString(colors.ColorizeHeader("FILES OVERVIEW\n"))
	sb.WriteString(strings.Repeat("─", 63) + "\n")

	for _, report := range reports {
		status := colors.ColorizeSuccess("✓")
		if !report.IsValid {
			status = colors.ColorizeError("✗")
		}

		sb.WriteString(fmt.Sprintf("%s %s\n", status, colors.ColorizePath(report.FilePath)))
		sb.WriteString(fmt.Sprintf("    Errors: %s, Warnings: %s, Info: %s\n",
			colors.ColorizeError(fmt.Sprintf("%d", report.ErrorCount())),
			colors.ColorizeWarning(fmt.Sprintf("%d", report.WarningCount())),
			colors.ColorizeInfo(fmt.Sprintf("%d", report.InfoCount()))))
	}

	sb.WriteString("\n" + colors.ColorizeHeader("═══════════════════════════════════════════════════════════════\n"))

	_, err := writer.Write([]byte(sb.String()))
	return err
}

func (r *TextReporter) writeErrors(sb *strings.Builder, errors []domain.ValidationError, colors *ColorScheme, options *ports.ReportOptions) {
	for i, err := range errors {
		severitySymbol := r.getSeveritySymbol(err.Severity)
		fmt.Fprintf(sb, "\n[%d] %s %s %s\n",
			i+1,
			colors.ColorizeForSeverity(severitySymbol, err.Severity),
			colors.ColorizeCode(fmt.Sprintf("[%s]", err.Code)),
			err.Message)

		if err.Location != nil {
			locationStr := r.formatLocation(err.Location)
			fmt.Fprintf(sb, "    Location: %s\n", colors.ColorizePath(locationStr))
		}

		if options != nil && options.Verbose {
			if len(err.Details) > 0 {
				sb.WriteString("    Details:\n")
				for key, value := range err.Details {
					fmt.Fprintf(sb, "      %s: %v\n", key, value)
				}
			}
			fmt.Fprintf(sb, "    Timestamp: %s\n", colors.ColorizeDim(err.Timestamp.Format("2006-01-02 15:04:05")))
		}
	}
}

func (r *TextReporter) getSeveritySymbol(severity domain.Severity) string {
	switch severity {
	case domain.SeverityError:
		return "✗"
	case domain.SeverityWarning:
		return "⚠"
	case domain.SeverityInfo:
		return "ℹ"
	default:
		return "•"
	}
}

func (r *TextReporter) formatLocation(loc *domain.ErrorLocation) string {
	if loc.Line > 0 {
		if loc.Column > 0 {
			return fmt.Sprintf("%s (Line %d, Col %d)", loc.File, loc.Line, loc.Column)
		}
		return fmt.Sprintf("%s (Line %d)", loc.File, loc.Line)
	}
	if loc.Path != "" {
		return loc.Path
	}
	return loc.File
}
