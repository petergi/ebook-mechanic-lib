package reporter

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/example/project/internal/domain"
	"github.com/example/project/internal/ports"
)

func TestMarkdownReporter_Format(t *testing.T) {
	ctx := context.Background()
	reporter := NewMarkdownReporter()

	report := createTestReport("test.epub", true)
	options := &ports.ReportOptions{
		Format:          ports.FormatMarkdown,
		IncludeWarnings: true,
		IncludeInfo:     true,
		Verbose:         false,
	}

	result, err := reporter.Format(ctx, report, options)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	if result == "" {
		t.Error("Expected non-empty result")
	}

	if !strings.Contains(result, "# Validation Report: test.epub") {
		t.Error("Expected report title in output")
	}

	if !strings.Contains(result, "✅ Valid") {
		t.Error("Expected valid status in output")
	}

	if !strings.Contains(result, "## Summary") {
		t.Error("Expected summary section in output")
	}
}

func TestMarkdownReporter_FormatWithErrors(t *testing.T) {
	ctx := context.Background()
	reporter := NewMarkdownReporter()

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
		Format:          ports.FormatMarkdown,
		IncludeWarnings: true,
		IncludeInfo:     true,
		Verbose:         true,
	}

	result, err := reporter.Format(ctx, report, options)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	if !strings.Contains(result, "❌ Invalid") {
		t.Error("Expected invalid status in output")
	}

	if !strings.Contains(result, "## Errors") {
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

	if options.Verbose && !strings.Contains(result, "category") {
		t.Error("Expected details in verbose output")
	}
}

func TestMarkdownReporter_FormatWithWarnings(t *testing.T) {
	ctx := context.Background()
	reporter := NewMarkdownReporter()

	report := createTestReport("test.epub", false)
	report.Warnings = []domain.ValidationError{
		{
			Code:      "EPUB-W01",
			Message:   "Deprecated element",
			Severity:  domain.SeverityWarning,
			Timestamp: time.Now(),
			Location: &domain.ErrorLocation{
				File: "chapter1.xhtml",
				Line: 25,
			},
		},
	}

	options := &ports.ReportOptions{
		IncludeWarnings: true,
		IncludeInfo:     false,
	}

	result, err := reporter.Format(ctx, report, options)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	if !strings.Contains(result, "## Warnings") {
		t.Error("Expected warnings section in output")
	}

	if !strings.Contains(result, "EPUB-W01") {
		t.Error("Expected warning code in output")
	}

	if !strings.Contains(result, "chapter1.xhtml") {
		t.Error("Expected file in output")
	}

	if !strings.Contains(result, "Line 25") {
		t.Error("Expected line number in output")
	}
}

func TestMarkdownReporter_ExcludeWarnings(t *testing.T) {
	ctx := context.Background()
	reporter := NewMarkdownReporter()

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
	}

	result, err := reporter.Format(ctx, report, options)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	if strings.Contains(result, "## Warnings") {
		t.Error("Did not expect warnings section in output")
	}

	if !strings.Contains(result, "**Warnings:** 0") {
		t.Error("Expected warnings count to be 0 in summary")
	}
}

func TestMarkdownReporter_WriteToFile(t *testing.T) {
	ctx := context.Background()
	reporter := NewMarkdownReporter()

	report := createTestReport("test.epub", true)
	options := &ports.ReportOptions{
		Format:          ports.FormatMarkdown,
		IncludeWarnings: true,
		IncludeInfo:     true,
	}

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "report.md")

	err := reporter.WriteToFile(ctx, report, outputPath, options)
	if err != nil {
		t.Fatalf("WriteToFile failed: %v", err)
	}

	data, err := os.ReadFile(outputPath) //nolint:gosec
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "# Validation Report: test.epub") {
		t.Error("Expected report title in file content")
	}
}

func TestMarkdownReporter_FormatMultiple(t *testing.T) {
	ctx := context.Background()
	reporter := NewMarkdownReporter().(*MarkdownReporter)

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
	}

	result, err := reporter.FormatMultiple(ctx, reports, options)
	if err != nil {
		t.Fatalf("FormatMultiple failed: %v", err)
	}

	if !strings.Contains(result, "# Validation Report Summary") {
		t.Error("Expected summary header in output")
	}

	if !strings.Contains(result, "test1.epub") {
		t.Error("Expected first file in output")
	}

	if !strings.Contains(result, "test2.epub") {
		t.Error("Expected second file in output")
	}

	if !strings.Contains(result, "**Total Files:** 2") {
		t.Error("Expected total files count in output")
	}

	if !strings.Contains(result, "**Valid Files:** 1") {
		t.Error("Expected valid files count in output")
	}

	if !strings.Contains(result, "**Invalid Files:** 1") {
		t.Error("Expected invalid files count in output")
	}
}

func TestMarkdownReporter_WriteSummary(t *testing.T) {
	ctx := context.Background()
	reporter := NewMarkdownReporter().(*MarkdownReporter)

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

	options := &ports.ReportOptions{}

	var buf bytes.Buffer
	err := reporter.WriteSummary(ctx, reports, &buf, options)
	if err != nil {
		t.Fatalf("WriteSummary failed: %v", err)
	}

	result := buf.String()

	if !strings.Contains(result, "# Validation Summary") {
		t.Error("Expected summary header in output")
	}

	if !strings.Contains(result, "## Files Overview") {
		t.Error("Expected files overview section in output")
	}

	if !strings.Contains(result, "test1.epub") {
		t.Error("Expected first file in table")
	}

	if !strings.Contains(result, "test2.epub") {
		t.Error("Expected second file in table")
	}

	if !strings.Contains(result, "✅") {
		t.Error("Expected valid status symbol")
	}

	if !strings.Contains(result, "❌") {
		t.Error("Expected invalid status symbol")
	}
}

func TestMarkdownReporter_EscapeMarkdown(t *testing.T) {
	ctx := context.Background()
	reporter := NewMarkdownReporter()

	report := createTestReport("test.epub", false)
	report.Errors = []domain.ValidationError{
		{
			Code:      "TEST-001",
			Message:   "Message with | pipe and\nnewline",
			Severity:  domain.SeverityError,
			Timestamp: time.Now(),
		},
	}

	options := &ports.ReportOptions{}

	result, err := reporter.Format(ctx, report, options)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	if strings.Contains(result, "| pipe") && !strings.Contains(result, "\\|") {
		t.Error("Expected escaped pipe in markdown table")
	}

	if strings.Contains(result, "\nnewline") && !strings.Contains(result, "<br>") {
		t.Error("Expected newline to be replaced with <br> in markdown table")
	}
}

func TestMarkdownReporter_WithFilter(t *testing.T) {
	ctx := context.Background()

	filter := &Filter{
		Categories: []string{"structure"},
	}
	reporter := NewMarkdownReporterWithFilter(filter)

	report := createTestReport("test.epub", false)
	report.Errors = []domain.ValidationError{
		{
			Code:      "E1",
			Severity:  domain.SeverityError,
			Timestamp: time.Now(),
			Details: map[string]interface{}{
				"category": "structure",
			},
		},
		{
			Code:      "E2",
			Severity:  domain.SeverityError,
			Timestamp: time.Now(),
			Details: map[string]interface{}{
				"category": "metadata",
			},
		},
	}

	options := &ports.ReportOptions{}

	result, err := reporter.Format(ctx, report, options)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	errorCount := strings.Count(result, "| `E")
	if errorCount != 1 {
		t.Errorf("Expected 1 error row in table (after filtering), got %d", errorCount)
	}
}

func TestMarkdownReporter_VerboseMetadata(t *testing.T) {
	ctx := context.Background()
	reporter := NewMarkdownReporter()

	report := createTestReport("test.epub", true)
	report.Metadata = map[string]interface{}{
		"version": "3.0",
		"title":   "Test Book",
	}

	options := &ports.ReportOptions{
		Verbose: true,
	}

	result, err := reporter.Format(ctx, report, options)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	if !strings.Contains(result, "## Metadata") {
		t.Error("Expected metadata section in verbose output")
	}

	if !strings.Contains(result, "version") {
		t.Error("Expected metadata key in output")
	}

	if !strings.Contains(result, "3.0") {
		t.Error("Expected metadata value in output")
	}
}
