package reporter

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/petergi/ebook-mechanic-lib/internal/domain"
	"github.com/petergi/ebook-mechanic-lib/internal/ports"
)

func TestTextReporter_Format(t *testing.T) {
	ctx := context.Background()
	reporter := NewTextReporter()

	report := createTestReport("test.epub", true)
	options := &ports.ReportOptions{
		Format:          ports.FormatText,
		IncludeWarnings: true,
		IncludeInfo:     true,
		Verbose:         false,
		ColorEnabled:    false,
	}

	result, err := reporter.Format(ctx, report, options)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	if result == "" {
		t.Error("Expected non-empty result")
	}

	if !strings.Contains(result, "VALIDATION REPORT: test.epub") {
		t.Error("Expected report title in output")
	}

	if !strings.Contains(result, "VALID") {
		t.Error("Expected valid status in output")
	}

	if !strings.Contains(result, "SUMMARY") {
		t.Error("Expected summary section in output")
	}
}

func TestTextReporter_FormatWithColors(t *testing.T) {
	ctx := context.Background()
	reporter := NewTextReporter()

	report := createTestReport("test.epub", false)
	report.Errors = []domain.ValidationError{
		{
			Code:      "EPUB-001",
			Message:   "Invalid structure",
			Severity:  domain.SeverityError,
			Timestamp: time.Now(),
		},
	}

	options := &ports.ReportOptions{
		ColorEnabled: true,
	}

	result, err := reporter.Format(ctx, report, options)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	if !strings.Contains(result, "\033[") {
		t.Error("Expected ANSI color codes in output when colors enabled")
	}

	if !strings.Contains(result, colorReset) {
		t.Error("Expected color reset codes in output")
	}
}

func TestTextReporter_FormatWithoutColors(t *testing.T) {
	ctx := context.Background()
	reporter := NewTextReporter()

	report := createTestReport("test.epub", false)
	report.Errors = []domain.ValidationError{
		{
			Code:      "EPUB-001",
			Message:   "Invalid structure",
			Severity:  domain.SeverityError,
			Timestamp: time.Now(),
		},
	}

	options := &ports.ReportOptions{
		ColorEnabled: false,
	}

	result, err := reporter.Format(ctx, report, options)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	if strings.Contains(result, "\033[") {
		t.Error("Did not expect ANSI color codes in output when colors disabled")
	}
}

func TestTextReporter_FormatWithErrors(t *testing.T) {
	ctx := context.Background()
	reporter := NewTextReporter()

	report := createTestReport("test.epub", false)
	report.Errors = []domain.ValidationError{
		{
			Code:      "EPUB-001",
			Message:   "Invalid structure",
			Severity:  domain.SeverityError,
			Timestamp: time.Now(),
			Location: &domain.ErrorLocation{
				File:   "content.opf",
				Line:   10,
				Column: 5,
				Path:   "OPS/content.opf",
			},
			Details: map[string]interface{}{
				"category": "structure",
			},
		},
	}

	options := &ports.ReportOptions{
		IncludeWarnings: true,
		IncludeInfo:     true,
		Verbose:         true,
		ColorEnabled:    false,
	}

	result, err := reporter.Format(ctx, report, options)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	if !strings.Contains(result, "INVALID") {
		t.Error("Expected invalid status in output")
	}

	if !strings.Contains(result, "ERRORS") {
		t.Error("Expected errors section in output")
	}

	if !strings.Contains(result, "EPUB-001") {
		t.Error("Expected error code in output")
	}

	if !strings.Contains(result, "Invalid structure") {
		t.Error("Expected error message in output")
	}

	if !strings.Contains(result, "content.opf") {
		t.Error("Expected file location in output")
	}

	if !strings.Contains(result, "Line 10") {
		t.Error("Expected line number in output")
	}

	if !strings.Contains(result, "Col 5") {
		t.Error("Expected column number in output")
	}

	if options.Verbose && !strings.Contains(result, "Details:") {
		t.Error("Expected details in verbose output")
	}

	if options.Verbose && !strings.Contains(result, "category") {
		t.Error("Expected category in verbose details")
	}
}

func TestTextReporter_SeveritySymbols(t *testing.T) {
	ctx := context.Background()
	reporter := NewTextReporter()

	report := createTestReport("test.epub", false)
	report.Errors = []domain.ValidationError{
		{Code: "E1", Message: "Error", Severity: domain.SeverityError, Timestamp: time.Now()},
	}
	report.Warnings = []domain.ValidationError{
		{Code: "W1", Message: "Warning", Severity: domain.SeverityWarning, Timestamp: time.Now()},
	}
	report.Info = []domain.ValidationError{
		{Code: "I1", Message: "Info", Severity: domain.SeverityInfo, Timestamp: time.Now()},
	}

	options := &ports.ReportOptions{
		IncludeWarnings: true,
		IncludeInfo:     true,
		ColorEnabled:    false,
	}

	result, err := reporter.Format(ctx, report, options)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	if !strings.Contains(result, "✗") {
		t.Error("Expected error symbol (✗) in output")
	}

	if !strings.Contains(result, "⚠") {
		t.Error("Expected warning symbol (⚠) in output")
	}

	if !strings.Contains(result, "ℹ") {
		t.Error("Expected info symbol (ℹ) in output")
	}
}

func TestTextReporter_ExcludeWarnings(t *testing.T) {
	ctx := context.Background()
	reporter := NewTextReporter()

	report := createTestReport("test.epub", false)
	report.Warnings = []domain.ValidationError{
		{
			Code:      "EPUB-W01",
			Message:   "Warning",
			Severity:  domain.SeverityWarning,
			Timestamp: time.Now(),
		},
	}

	options := &ports.ReportOptions{
		IncludeWarnings: false,
		IncludeInfo:     true,
		ColorEnabled:    false,
	}

	result, err := reporter.Format(ctx, report, options)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	if strings.Contains(result, "WARNINGS") {
		t.Error("Did not expect warnings section in output")
	}

	if !strings.Contains(result, "Warnings:        0") {
		t.Error("Expected warnings count to be 0 in summary")
	}
}

func TestTextReporter_MaxErrors(t *testing.T) {
	ctx := context.Background()
	reporter := NewTextReporter()

	report := createTestReport("test.epub", false)
	report.Errors = make([]domain.ValidationError, 10)
	for i := range report.Errors {
		report.Errors[i] = domain.ValidationError{
			Code:      "TEST-001",
			Message:   "Test error",
			Severity:  domain.SeverityError,
			Timestamp: time.Now(),
		}
	}

	options := &ports.ReportOptions{
		MaxErrors:    3,
		ColorEnabled: false,
	}

	result, err := reporter.Format(ctx, report, options)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	errorCount := strings.Count(result, "[TEST-001]")
	if errorCount != 3 {
		t.Errorf("Expected 3 errors in output (max limit), got %d", errorCount)
	}
}

func TestTextReporter_WriteToFile(t *testing.T) {
	ctx := context.Background()
	reporter := NewTextReporter()

	report := createTestReport("test.epub", true)
	options := &ports.ReportOptions{
		Format:          ports.FormatText,
		IncludeWarnings: true,
		IncludeInfo:     true,
		ColorEnabled:    false,
	}

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "report.txt")

	err := reporter.WriteToFile(ctx, report, outputPath, options)
	if err != nil {
		t.Fatalf("WriteToFile failed: %v", err)
	}

	data, err := os.ReadFile(outputPath) //nolint:gosec
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "VALIDATION REPORT: test.epub") {
		t.Error("Expected report title in file content")
	}
}

func TestTextReporter_FormatMultiple(t *testing.T) {
	ctx := context.Background()
	reporter := NewTextReporter().(*TextReporter)

	reports := []*domain.ValidationReport{
		createTestReport("test1.epub", true),
		createTestReport("test2.epub", false),
	}

	reports[1].Errors = []domain.ValidationError{
		{
			Code:      "TEST-001",
			Message:   "Error",
			Severity:  domain.SeverityError,
			Timestamp: time.Now(),
		},
	}

	options := &ports.ReportOptions{
		IncludeWarnings: true,
		IncludeInfo:     true,
		ColorEnabled:    false,
	}

	result, err := reporter.FormatMultiple(ctx, reports, options)
	if err != nil {
		t.Fatalf("FormatMultiple failed: %v", err)
	}

	if !strings.Contains(result, "VALIDATION SUMMARY") {
		t.Error("Expected summary header in output")
	}

	if !strings.Contains(result, "test1.epub") {
		t.Error("Expected first file in output")
	}

	if !strings.Contains(result, "test2.epub") {
		t.Error("Expected second file in output")
	}

	if !strings.Contains(result, "Total Files:     2") {
		t.Error("Expected total files count in output")
	}

	if !strings.Contains(result, "Valid Files:") {
		t.Error("Expected valid files count in output")
	}

	if !strings.Contains(result, "Invalid Files:") {
		t.Error("Expected invalid files count in output")
	}
}

func TestTextReporter_WriteSummary(t *testing.T) {
	ctx := context.Background()
	reporter := NewTextReporter().(*TextReporter)

	reports := []*domain.ValidationReport{
		createTestReport("test1.epub", true),
		createTestReport("test2.epub", false),
	}

	reports[1].Errors = []domain.ValidationError{
		{Code: "E1", Severity: domain.SeverityError, Timestamp: time.Now()},
		{Code: "E2", Severity: domain.SeverityError, Timestamp: time.Now()},
	}
	reports[1].Warnings = []domain.ValidationError{
		{Code: "W1", Severity: domain.SeverityWarning, Timestamp: time.Now()},
	}

	options := &ports.ReportOptions{
		ColorEnabled: false,
	}

	var buf bytes.Buffer
	err := reporter.WriteSummary(ctx, reports, &buf, options)
	if err != nil {
		t.Fatalf("WriteSummary failed: %v", err)
	}

	result := buf.String()

	if !strings.Contains(result, "VALIDATION SUMMARY") {
		t.Error("Expected summary header in output")
	}

	if !strings.Contains(result, "FILES OVERVIEW") {
		t.Error("Expected files overview section in output")
	}

	if !strings.Contains(result, "test1.epub") {
		t.Error("Expected first file in output")
	}

	if !strings.Contains(result, "test2.epub") {
		t.Error("Expected second file in output")
	}

	if !strings.Contains(result, "✓") {
		t.Error("Expected valid status symbol")
	}

	if !strings.Contains(result, "✗") {
		t.Error("Expected invalid status symbol")
	}
}

func TestTextReporter_WithFilter(t *testing.T) {
	ctx := context.Background()

	filter := &Filter{
		MinSeverity: domain.SeverityWarning,
	}
	reporter := NewTextReporterWithFilter(filter)

	report := createTestReport("test.epub", false)
	report.Errors = []domain.ValidationError{
		{Code: "E1", Severity: domain.SeverityError, Timestamp: time.Now()},
	}
	report.Warnings = []domain.ValidationError{
		{Code: "W1", Severity: domain.SeverityWarning, Timestamp: time.Now()},
	}
	report.Info = []domain.ValidationError{
		{Code: "I1", Severity: domain.SeverityInfo, Timestamp: time.Now()},
	}

	options := &ports.ReportOptions{
		IncludeWarnings: true,
		IncludeInfo:     true,
		ColorEnabled:    false,
	}

	result, err := reporter.Format(ctx, report, options)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	if !strings.Contains(result, "E1") {
		t.Error("Expected error E1 in output")
	}

	if !strings.Contains(result, "W1") {
		t.Error("Expected warning W1 in output")
	}

	if strings.Contains(result, "I1") {
		t.Error("Did not expect info I1 in output (filtered by min severity)")
	}
}

func TestTextReporter_VerboseOutput(t *testing.T) {
	ctx := context.Background()
	reporter := NewTextReporter()

	report := createTestReport("test.epub", true)
	report.Metadata = map[string]interface{}{
		"version": "3.0",
		"title":   "Test Book",
	}

	options := &ports.ReportOptions{
		Verbose:      true,
		ColorEnabled: false,
	}

	result, err := reporter.Format(ctx, report, options)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	if !strings.Contains(result, "METADATA") {
		t.Error("Expected metadata section in verbose output")
	}

	if !strings.Contains(result, "version") {
		t.Error("Expected metadata key in output")
	}

	if !strings.Contains(result, "3.0") {
		t.Error("Expected metadata value in output")
	}
}

func TestTextReporter_LocationFormatting(t *testing.T) {
	ctx := context.Background()
	reporter := NewTextReporter()

	tests := []struct {
		name     string
		location *domain.ErrorLocation
		expected []string
	}{
		{
			name: "file with line and column",
			location: &domain.ErrorLocation{
				File:   "test.xhtml",
				Line:   10,
				Column: 5,
			},
			expected: []string{"test.xhtml", "Line 10", "Col 5"},
		},
		{
			name: "file with line only",
			location: &domain.ErrorLocation{
				File: "test.xhtml",
				Line: 10,
			},
			expected: []string{"test.xhtml", "Line 10"},
		},
		{
			name: "file only",
			location: &domain.ErrorLocation{
				File: "test.xhtml",
			},
			expected: []string{"test.xhtml"},
		},
		{
			name: "path only",
			location: &domain.ErrorLocation{
				Path: "OPS/content.opf",
			},
			expected: []string{"OPS/content.opf"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := createTestReport("test.epub", false)
			report.Errors = []domain.ValidationError{
				{
					Code:      "TEST-001",
					Message:   "Test error",
					Severity:  domain.SeverityError,
					Location:  tt.location,
					Timestamp: time.Now(),
				},
			}

			options := &ports.ReportOptions{
				ColorEnabled: false,
			}

			result, err := reporter.Format(ctx, report, options)
			if err != nil {
				t.Fatalf("Format failed: %v", err)
			}

			for _, exp := range tt.expected {
				if !strings.Contains(result, exp) {
					t.Errorf("Expected '%s' in location output", exp)
				}
			}
		})
	}
}
