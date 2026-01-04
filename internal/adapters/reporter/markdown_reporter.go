package reporter

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/petergi/ebook-mechanic-lib/internal/domain"
	"github.com/petergi/ebook-mechanic-lib/internal/ports"
)

// MarkdownReporter formats validation reports as Markdown.
type MarkdownReporter struct {
	filter *Filter
}

// NewMarkdownReporter returns a Markdown reporter without filters.
func NewMarkdownReporter() ports.Reporter {
	return &MarkdownReporter{}
}

// NewMarkdownReporterWithFilter returns a Markdown reporter with a filter applied.
func NewMarkdownReporterWithFilter(filter *Filter) ports.Reporter {
	return &MarkdownReporter{
		filter: filter,
	}
}

// Format renders a single report as Markdown.
func (r *MarkdownReporter) Format(_ context.Context, report *domain.ValidationReport, options *ports.ReportOptions) (string, error) {
	var sb strings.Builder

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

	sb.WriteString(fmt.Sprintf("# Validation Report: %s\n\n", report.FilePath))

	if report.IsValid {
		sb.WriteString("**Status:** ✅ Valid\n\n")
	} else {
		sb.WriteString("**Status:** ❌ Invalid\n\n")
	}

	sb.WriteString(fmt.Sprintf("- **File Type:** %s\n", report.FileType))
	sb.WriteString(fmt.Sprintf("- **Validation Time:** %s\n", report.ValidationTime.Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("- **Duration:** %s\n\n", report.Duration))

	sb.WriteString("## Summary\n\n")
	sb.WriteString(fmt.Sprintf("- **Total Issues:** %d\n", len(errors)+len(warnings)+len(info)))
	sb.WriteString(fmt.Sprintf("- **Errors:** %d\n", len(errors)))
	sb.WriteString(fmt.Sprintf("- **Warnings:** %d\n", len(warnings)))
	sb.WriteString(fmt.Sprintf("- **Info:** %d\n\n", len(info)))

	if len(errors) > 0 {
		sb.WriteString("## Errors\n\n")
		r.writeErrorsTable(&sb, errors, options)
		sb.WriteString("\n")
	}

	if len(warnings) > 0 {
		sb.WriteString("## Warnings\n\n")
		r.writeErrorsTable(&sb, warnings, options)
		sb.WriteString("\n")
	}

	if len(info) > 0 {
		sb.WriteString("## Information\n\n")
		r.writeErrorsTable(&sb, info, options)
		sb.WriteString("\n")
	}

	if options != nil && options.Verbose && len(report.Metadata) > 0 {
		sb.WriteString("## Metadata\n\n")
		for key, value := range report.Metadata {
			sb.WriteString(fmt.Sprintf("- **%s:** %v\n", key, value))
		}
		sb.WriteString("\n")
	}

	return sb.String(), nil
}

// Write writes a Markdown report to the provided writer.
func (r *MarkdownReporter) Write(ctx context.Context, report *domain.ValidationReport, writer io.Writer, options *ports.ReportOptions) error {
	formatted, err := r.Format(ctx, report, options)
	if err != nil {
		return err
	}

	_, err = writer.Write([]byte(formatted))
	return err
}

// WriteToFile writes a Markdown report to a file.
func (r *MarkdownReporter) WriteToFile(ctx context.Context, report *domain.ValidationReport, filePath string, options *ports.ReportOptions) error {
	file, err := os.Create(filePath) //nolint:gosec
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	return r.Write(ctx, report, file, options)
}

// FormatMultiple renders multiple reports as Markdown.
func (r *MarkdownReporter) FormatMultiple(ctx context.Context, reports []*domain.ValidationReport, options *ports.ReportOptions) (string, error) {
	var sb strings.Builder

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

	sb.WriteString("# Validation Report Summary\n\n")
	sb.WriteString(fmt.Sprintf("- **Total Files:** %d\n", len(reports)))
	sb.WriteString(fmt.Sprintf("- **Valid Files:** %d\n", validFiles))
	sb.WriteString(fmt.Sprintf("- **Invalid Files:** %d\n", invalidFiles))
	sb.WriteString(fmt.Sprintf("- **Total Issues:** %d\n", totalErrors+totalWarnings+totalInfo))
	sb.WriteString(fmt.Sprintf("- **Errors:** %d\n", totalErrors))
	sb.WriteString(fmt.Sprintf("- **Warnings:** %d\n", totalWarnings))
	sb.WriteString(fmt.Sprintf("- **Info:** %d\n\n", totalInfo))

	sb.WriteString("---\n\n")

	for i, report := range reports {
		reportStr, err := r.Format(ctx, report, options)
		if err != nil {
			return "", err
		}
		sb.WriteString(reportStr)

		if i < len(reports)-1 {
			sb.WriteString("---\n\n")
		}
	}

	return sb.String(), nil
}

// WriteMultiple writes multiple Markdown reports to the provided writer.
func (r *MarkdownReporter) WriteMultiple(ctx context.Context, reports []*domain.ValidationReport, writer io.Writer, options *ports.ReportOptions) error {
	formatted, err := r.FormatMultiple(ctx, reports, options)
	if err != nil {
		return err
	}

	_, err = writer.Write([]byte(formatted))
	return err
}

// WriteSummary writes a Markdown summary for multiple reports.
func (r *MarkdownReporter) WriteSummary(_ context.Context, reports []*domain.ValidationReport, writer io.Writer, _ *ports.ReportOptions) error {
	var sb strings.Builder

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

	sb.WriteString("# Validation Summary\n\n")
	sb.WriteString(fmt.Sprintf("- **Total Files:** %d\n", len(reports)))
	sb.WriteString(fmt.Sprintf("- **Valid Files:** %d\n", validFiles))
	sb.WriteString(fmt.Sprintf("- **Invalid Files:** %d\n", invalidFiles))
	sb.WriteString(fmt.Sprintf("- **Total Issues:** %d\n", totalErrors+totalWarnings+totalInfo))
	sb.WriteString(fmt.Sprintf("- **Errors:** %d\n", totalErrors))
	sb.WriteString(fmt.Sprintf("- **Warnings:** %d\n", totalWarnings))
	sb.WriteString(fmt.Sprintf("- **Info:** %d\n\n", totalInfo))

	sb.WriteString("## Files Overview\n\n")
	sb.WriteString("| File | Status | Errors | Warnings | Info |\n")
	sb.WriteString("|------|--------|--------|----------|------|\n")

	for _, report := range reports {
		status := "✅"
		if !report.IsValid {
			status = "❌"
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %d | %d | %d |\n",
			report.FilePath,
			status,
			report.ErrorCount(),
			report.WarningCount(),
			report.InfoCount()))
	}

	_, err := writer.Write([]byte(sb.String()))
	return err
}

func (r *MarkdownReporter) writeErrorsTable(sb *strings.Builder, errors []domain.ValidationError, options *ports.ReportOptions) {
	sb.WriteString("| Code | Message | Location |\n")
	sb.WriteString("|------|---------|----------|\n")

	for _, err := range errors {
		location := "N/A"
		if err.Location != nil {
			if err.Location.Line > 0 {
				location = fmt.Sprintf("%s (Line %d", err.Location.File, err.Location.Line)
				if err.Location.Column > 0 {
					location += fmt.Sprintf(", Col %d", err.Location.Column)
				}
				location += ")"
			} else {
				location = err.Location.File
			}
		}

		message := r.escapeMarkdown(err.Message)

		if options != nil && options.Verbose && len(err.Details) > 0 {
			message += "<br>"
			for key, value := range err.Details {
				message += fmt.Sprintf("<br>**%s:** %v", key, value)
			}
		}

		fmt.Fprintf(sb, "| `%s` | %s | %s |\n", err.Code, message, location)
	}
}

func (r *MarkdownReporter) escapeMarkdown(s string) string {
	replacer := strings.NewReplacer(
		"|", "\\|",
		"\n", "<br>",
	)
	return replacer.Replace(s)
}
