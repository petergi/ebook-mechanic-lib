package reporter

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/example/project/internal/domain"
	"github.com/example/project/internal/ports"
)

func TestJSONReporter_Format(t *testing.T) {
	ctx := context.Background()
	reporter := NewJSONReporter()

	report := createTestReport("test.epub", true)
	options := &ports.ReportOptions{
		Format:          ports.FormatJSON,
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

	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(result), &jsonData); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if jsonData["file_path"] != "test.epub" {
		t.Errorf("Expected file_path to be 'test.epub', got %v", jsonData["file_path"])
	}

	if jsonData["is_valid"] != true {
		t.Errorf("Expected is_valid to be true, got %v", jsonData["is_valid"])
	}
}

func TestJSONReporter_FormatWithErrors(t *testing.T) {
	ctx := context.Background()
	reporter := NewJSONReporter()

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
		Format:          ports.FormatJSON,
		IncludeWarnings: true,
		IncludeInfo:     true,
	}

	result, err := reporter.Format(ctx, report, options)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(result), &jsonData); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	errors, ok := jsonData["errors"].([]interface{})
	if !ok || len(errors) == 0 {
		t.Fatal("Expected errors array")
	}

	firstError := errors[0].(map[string]interface{})
	if firstError["code"] != "EPUB-001" {
		t.Errorf("Expected code 'EPUB-001', got %v", firstError["code"])
	}

	location := firstError["location"].(map[string]interface{})
	if location["file"] != "content.opf" {
		t.Errorf("Expected file 'content.opf', got %v", location["file"])
	}
	if location["line"] != float64(10) {
		t.Errorf("Expected line 10, got %v", location["line"])
	}
}

func TestJSONReporter_MaxErrors(t *testing.T) {
	ctx := context.Background()
	reporter := NewJSONReporter()

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
		MaxErrors: 5,
	}

	result, err := reporter.Format(ctx, report, options)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(result), &jsonData); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	errors := jsonData["errors"].([]interface{})
	if len(errors) != 5 {
		t.Errorf("Expected 5 errors, got %d", len(errors))
	}
}

func TestJSONReporter_ExcludeWarnings(t *testing.T) {
	ctx := context.Background()
	reporter := NewJSONReporter()

	report := createTestReport("test.epub", false)
	report.Warnings = []domain.ValidationError{
		{
			Code:      "EPUB-W01",
			Message:   "Warning message",
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

	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(result), &jsonData); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	warnings := jsonData["warnings"].([]interface{})
	if len(warnings) != 0 {
		t.Errorf("Expected 0 warnings, got %d", len(warnings))
	}
}

func TestJSONReporter_WriteToFile(t *testing.T) {
	ctx := context.Background()
	reporter := NewJSONReporter()

	report := createTestReport("test.epub", true)
	options := &ports.ReportOptions{
		Format:          ports.FormatJSON,
		IncludeWarnings: true,
		IncludeInfo:     true,
	}

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "report.json")

	err := reporter.WriteToFile(ctx, report, outputPath, options)
	if err != nil {
		t.Fatalf("WriteToFile failed: %v", err)
	}

	data, err := os.ReadFile(outputPath) //nolint:gosec
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	var jsonData map[string]interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if jsonData["file_path"] != "test.epub" {
		t.Errorf("Expected file_path to be 'test.epub', got %v", jsonData["file_path"])
	}
}

func TestJSONReporter_FormatMultiple(t *testing.T) {
	ctx := context.Background()
	reporter := NewJSONReporter().(*JSONReporter)

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

	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(result), &jsonData); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	reportsArray := jsonData["reports"].([]interface{})
	if len(reportsArray) != 2 {
		t.Errorf("Expected 2 reports, got %d", len(reportsArray))
	}

	if jsonData["total_files"] != float64(2) {
		t.Errorf("Expected total_files to be 2, got %v", jsonData["total_files"])
	}

	if jsonData["valid_files"] != float64(1) {
		t.Errorf("Expected valid_files to be 1, got %v", jsonData["valid_files"])
	}

	if jsonData["invalid_files"] != float64(1) {
		t.Errorf("Expected invalid_files to be 1, got %v", jsonData["invalid_files"])
	}
}

func TestJSONReporter_WriteSummary(t *testing.T) {
	ctx := context.Background()
	reporter := NewJSONReporter().(*JSONReporter)

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

	var jsonData map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &jsonData); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if jsonData["total_files"] != float64(2) {
		t.Errorf("Expected total_files to be 2, got %v", jsonData["total_files"])
	}

	summary := jsonData["summary"].(map[string]interface{})
	if summary["error_count"] != float64(2) {
		t.Errorf("Expected error_count to be 2, got %v", summary["error_count"])
	}
	if summary["warning_count"] != float64(1) {
		t.Errorf("Expected warning_count to be 1, got %v", summary["warning_count"])
	}
}

func TestJSONReporter_WithFilter(t *testing.T) {
	ctx := context.Background()

	filter := &Filter{
		Severities: []domain.Severity{domain.SeverityError},
	}
	reporter := NewJSONReporterWithFilter(filter)

	report := createTestReport("test.epub", false)
	report.Errors = []domain.ValidationError{
		{Code: "E1", Severity: domain.SeverityError, Timestamp: time.Now()},
	}
	report.Warnings = []domain.ValidationError{
		{Code: "W1", Severity: domain.SeverityWarning, Timestamp: time.Now()},
	}

	options := &ports.ReportOptions{
		IncludeWarnings: true,
	}

	result, err := reporter.Format(ctx, report, options)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(result), &jsonData); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	errors := jsonData["errors"].([]interface{})
	warnings := jsonData["warnings"].([]interface{})

	if len(errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(errors))
	}
	if len(warnings) != 0 {
		t.Errorf("Expected 0 warnings (filtered), got %d", len(warnings))
	}
}

func createTestReport(filePath string, isValid bool) *domain.ValidationReport {
	return &domain.ValidationReport{
		FilePath:       filePath,
		FileType:       "EPUB",
		IsValid:        isValid,
		Errors:         make([]domain.ValidationError, 0),
		Warnings:       make([]domain.ValidationError, 0),
		Info:           make([]domain.ValidationError, 0),
		ValidationTime: time.Now(),
		Duration:       100 * time.Millisecond,
		Metadata:       make(map[string]interface{}),
	}
}
