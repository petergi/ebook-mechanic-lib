package reporter

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/example/project/internal/domain"
	"github.com/example/project/internal/ports"
)

// JSONReporter formats validation reports as JSON.
type JSONReporter struct {
	filter *Filter
}

// NewJSONReporter returns a JSON reporter without filters.
func NewJSONReporter() ports.Reporter {
	return &JSONReporter{}
}

// NewJSONReporterWithFilter returns a JSON reporter with a filter applied.
func NewJSONReporterWithFilter(filter *Filter) ports.Reporter {
	return &JSONReporter{
		filter: filter,
	}
}

type jsonReport struct {
	FilePath       string                 `json:"file_path"`
	FileType       string                 `json:"file_type"`
	IsValid        bool                   `json:"is_valid"`
	Errors         []jsonValidationError  `json:"errors"`
	Warnings       []jsonValidationError  `json:"warnings"`
	Info           []jsonValidationError  `json:"info"`
	ValidationTime string                 `json:"validation_time"`
	Duration       string                 `json:"duration"`
	Summary        jsonSummary            `json:"summary"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

type jsonValidationError struct {
	Code      string                 `json:"code"`
	Message   string                 `json:"message"`
	Severity  string                 `json:"severity"`
	Location  *jsonErrorLocation     `json:"location,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp string                 `json:"timestamp"`
}

type jsonErrorLocation struct {
	File    string `json:"file,omitempty"`
	Line    int    `json:"line,omitempty"`
	Column  int    `json:"column,omitempty"`
	Path    string `json:"path,omitempty"`
	Context string `json:"context,omitempty"`
}

type jsonSummary struct {
	TotalIssues  int `json:"total_issues"`
	ErrorCount   int `json:"error_count"`
	WarningCount int `json:"warning_count"`
	InfoCount    int `json:"info_count"`
}

type jsonMultiReport struct {
	Reports      []jsonReport `json:"reports"`
	Summary      jsonSummary  `json:"summary"`
	TotalFiles   int          `json:"total_files"`
	ValidFiles   int          `json:"valid_files"`
	InvalidFiles int          `json:"invalid_files"`
}

// Format renders a single report as JSON.
func (r *JSONReporter) Format(_ context.Context, report *domain.ValidationReport, options *ports.ReportOptions) (string, error) {
	jsonRep := r.convertReport(report, options)

	data, err := json.MarshalIndent(jsonRep, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return string(data), nil
}

// Write writes a JSON report to the provided writer.
func (r *JSONReporter) Write(ctx context.Context, report *domain.ValidationReport, writer io.Writer, options *ports.ReportOptions) error {
	formatted, err := r.Format(ctx, report, options)
	if err != nil {
		return err
	}

	_, err = writer.Write([]byte(formatted))
	return err
}

// WriteToFile writes a JSON report to a file.
func (r *JSONReporter) WriteToFile(ctx context.Context, report *domain.ValidationReport, filePath string, options *ports.ReportOptions) error {
	file, err := os.Create(filePath) //nolint:gosec
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	return r.Write(ctx, report, file, options)
}

// FormatMultiple renders multiple reports as JSON.
func (r *JSONReporter) FormatMultiple(_ context.Context, reports []*domain.ValidationReport, options *ports.ReportOptions) (string, error) {
	jsonReps := make([]jsonReport, 0, len(reports))
	totalErrors := 0
	totalWarnings := 0
	totalInfo := 0
	validFiles := 0
	invalidFiles := 0

	for _, report := range reports {
		jsonRep := r.convertReport(report, options)
		jsonReps = append(jsonReps, jsonRep)

		totalErrors += jsonRep.Summary.ErrorCount
		totalWarnings += jsonRep.Summary.WarningCount
		totalInfo += jsonRep.Summary.InfoCount

		if report.IsValid {
			validFiles++
		} else {
			invalidFiles++
		}
	}

	multiReport := jsonMultiReport{
		Reports: jsonReps,
		Summary: jsonSummary{
			TotalIssues:  totalErrors + totalWarnings + totalInfo,
			ErrorCount:   totalErrors,
			WarningCount: totalWarnings,
			InfoCount:    totalInfo,
		},
		TotalFiles:   len(reports),
		ValidFiles:   validFiles,
		InvalidFiles: invalidFiles,
	}

	data, err := json.MarshalIndent(multiReport, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return string(data), nil
}

// WriteMultiple writes multiple JSON reports to the provided writer.
func (r *JSONReporter) WriteMultiple(ctx context.Context, reports []*domain.ValidationReport, writer io.Writer, options *ports.ReportOptions) error {
	formatted, err := r.FormatMultiple(ctx, reports, options)
	if err != nil {
		return err
	}

	_, err = writer.Write([]byte(formatted))
	return err
}

// WriteSummary writes a JSON summary for multiple reports.
func (r *JSONReporter) WriteSummary(_ context.Context, reports []*domain.ValidationReport, writer io.Writer, _ *ports.ReportOptions) error {
	totalErrors := 0
	totalWarnings := 0
	totalInfo := 0
	validFiles := 0
	invalidFiles := 0

	for _, report := range reports {
		totalErrors += report.ErrorCount()
		totalWarnings += report.WarningCount()
		totalInfo += report.InfoCount()

		if report.IsValid {
			validFiles++
		} else {
			invalidFiles++
		}
	}

	summary := map[string]interface{}{
		"total_files":   len(reports),
		"valid_files":   validFiles,
		"invalid_files": invalidFiles,
		"summary": jsonSummary{
			TotalIssues:  totalErrors + totalWarnings + totalInfo,
			ErrorCount:   totalErrors,
			WarningCount: totalWarnings,
			InfoCount:    totalInfo,
		},
	}

	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	_, err = writer.Write(data)
	return err
}

func (r *JSONReporter) convertReport(report *domain.ValidationReport, options *ports.ReportOptions) jsonReport {
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

	return jsonReport{
		FilePath:       report.FilePath,
		FileType:       report.FileType,
		IsValid:        report.IsValid,
		Errors:         r.convertErrors(errors),
		Warnings:       r.convertErrors(warnings),
		Info:           r.convertErrors(info),
		ValidationTime: report.ValidationTime.Format("2006-01-02T15:04:05Z07:00"),
		Duration:       report.Duration.String(),
		Summary: jsonSummary{
			TotalIssues:  len(errors) + len(warnings) + len(info),
			ErrorCount:   len(errors),
			WarningCount: len(warnings),
			InfoCount:    len(info),
		},
		Metadata: report.Metadata,
	}
}

func (r *JSONReporter) convertErrors(errors []domain.ValidationError) []jsonValidationError {
	result := make([]jsonValidationError, 0, len(errors))

	for _, err := range errors {
		jsonErr := jsonValidationError{
			Code:      err.Code,
			Message:   err.Message,
			Severity:  string(err.Severity),
			Details:   err.Details,
			Timestamp: err.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
		}

		if err.Location != nil {
			jsonErr.Location = &jsonErrorLocation{
				File:    err.Location.File,
				Line:    err.Location.Line,
				Column:  err.Location.Column,
				Path:    err.Location.Path,
				Context: err.Location.Context,
			}
		}

		result = append(result, jsonErr)
	}

	return result
}
