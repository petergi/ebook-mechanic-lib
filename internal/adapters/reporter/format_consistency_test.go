package reporter

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/example/project/internal/domain"
	"github.com/example/project/internal/ports"
)

func TestFormatConsistency_ErrorCodes(t *testing.T) {
	ctx := context.Background()

	report := createTestReport("consistency.epub", false)
	report.Errors = []domain.ValidationError{
		{Code: "CONSISTENCY-001", Message: "Test error 1", Severity: domain.SeverityError, Timestamp: time.Now()},
		{Code: "CONSISTENCY-002", Message: "Test error 2", Severity: domain.SeverityError, Timestamp: time.Now()},
	}

	jsonReporter := NewJSONReporter()
	mdReporter := NewMarkdownReporter()
	textReporter := NewTextReporter()

	options := &ports.ReportOptions{}

	jsonResult, _ := jsonReporter.Format(ctx, report, options)
	mdResult, _ := mdReporter.Format(ctx, report, options)
	textResult, _ := textReporter.Format(ctx, report, options)

	errorCodes := []string{"CONSISTENCY-001", "CONSISTENCY-002"}
	for _, code := range errorCodes {
		if !strings.Contains(jsonResult, code) {
			t.Errorf("JSON output missing error code: %s", code)
		}
		if !strings.Contains(mdResult, code) {
			t.Errorf("Markdown output missing error code: %s", code)
		}
		if !strings.Contains(textResult, code) {
			t.Errorf("Text output missing error code: %s", code)
		}
	}
}

func TestFormatConsistency_LocationInfo(t *testing.T) {
	ctx := context.Background()

	report := createTestReport("location.epub", false)
	report.Errors = []domain.ValidationError{
		{
			Code:      "LOC-001",
			Message:   "Location test",
			Severity:  domain.SeverityError,
			Timestamp: time.Now(),
			Location: &domain.ErrorLocation{
				File:   "test.opf",
				Line:   42,
				Column: 7,
				Path:   "OPS/test.opf",
			},
		},
	}

	jsonReporter := NewJSONReporter()
	mdReporter := NewMarkdownReporter()
	textReporter := NewTextReporter()

	options := &ports.ReportOptions{}

	jsonResult, _ := jsonReporter.Format(ctx, report, options)
	mdResult, _ := mdReporter.Format(ctx, report, options)
	textResult, _ := textReporter.Format(ctx, report, options)

	if !strings.Contains(jsonResult, "test.opf") {
		t.Error("JSON output missing file location")
	}
	if !strings.Contains(jsonResult, "42") {
		t.Error("JSON output missing line number")
	}

	if !strings.Contains(mdResult, "test.opf") {
		t.Error("Markdown output missing file location")
	}
	if !strings.Contains(mdResult, "42") {
		t.Error("Markdown output missing line number")
	}

	if !strings.Contains(textResult, "test.opf") {
		t.Error("Text output missing file location")
	}
	if !strings.Contains(textResult, "42") {
		t.Error("Text output missing line number")
	}
}

func TestFormatConsistency_SeverityCounts(t *testing.T) {
	ctx := context.Background()

	report := createTestReport("counts.epub", false)
	report.Errors = []domain.ValidationError{
		{Code: "E1", Severity: domain.SeverityError, Timestamp: time.Now()},
		{Code: "E2", Severity: domain.SeverityError, Timestamp: time.Now()},
		{Code: "E3", Severity: domain.SeverityError, Timestamp: time.Now()},
	}
	report.Warnings = []domain.ValidationError{
		{Code: "W1", Severity: domain.SeverityWarning, Timestamp: time.Now()},
		{Code: "W2", Severity: domain.SeverityWarning, Timestamp: time.Now()},
	}
	report.Info = []domain.ValidationError{
		{Code: "I1", Severity: domain.SeverityInfo, Timestamp: time.Now()},
	}

	jsonReporter := NewJSONReporter()
	mdReporter := NewMarkdownReporter()
	textReporter := NewTextReporter()

	options := &ports.ReportOptions{
		IncludeWarnings: true,
		IncludeInfo:     true,
	}

	jsonResult, _ := jsonReporter.Format(ctx, report, options)
	mdResult, _ := mdReporter.Format(ctx, report, options)
	textResult, _ := textReporter.Format(ctx, report, options)

	var jsonData map[string]interface{}
	_ = json.Unmarshal([]byte(jsonResult), &jsonData)

	summary := jsonData["summary"].(map[string]interface{})
	if summary["error_count"] != float64(3) {
		t.Errorf("JSON error count mismatch: got %v, want 3", summary["error_count"])
	}
	if summary["warning_count"] != float64(2) {
		t.Errorf("JSON warning count mismatch: got %v, want 2", summary["warning_count"])
	}
	if summary["info_count"] != float64(1) {
		t.Errorf("JSON info count mismatch: got %v, want 1", summary["info_count"])
	}

	if !strings.Contains(mdResult, "**Errors:** 3") {
		t.Error("Markdown missing correct error count")
	}
	if !strings.Contains(mdResult, "**Warnings:** 2") {
		t.Error("Markdown missing correct warning count")
	}
	if !strings.Contains(mdResult, "**Info:** 1") {
		t.Error("Markdown missing correct info count")
	}

	if !strings.Contains(textResult, "Errors:") || !strings.Contains(textResult, "3") {
		t.Error("Text missing correct error count")
	}
	if !strings.Contains(textResult, "Warnings:") || !strings.Contains(textResult, "2") {
		t.Error("Text missing correct warning count")
	}
	if !strings.Contains(textResult, "Info:") || !strings.Contains(textResult, "1") {
		t.Error("Text missing correct info count")
	}
}

func TestFormatConsistency_ValidStatus(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		isValid bool
	}{
		{"valid report", true},
		{"invalid report", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			report := createTestReport("status.epub", tc.isValid)
			if !tc.isValid {
				report.Errors = []domain.ValidationError{
					{Code: "E1", Severity: domain.SeverityError, Timestamp: time.Now()},
				}
			}

			jsonReporter := NewJSONReporter()
			mdReporter := NewMarkdownReporter()
			textReporter := NewTextReporter()

			options := &ports.ReportOptions{}

			jsonResult, _ := jsonReporter.Format(ctx, report, options)
			mdResult, _ := mdReporter.Format(ctx, report, options)
			textResult, _ := textReporter.Format(ctx, report, options)

			var jsonData map[string]interface{}
			_ = json.Unmarshal([]byte(jsonResult), &jsonData)

			if jsonData["is_valid"] != tc.isValid {
				t.Errorf("JSON is_valid mismatch: got %v, want %v", jsonData["is_valid"], tc.isValid)
			}

			if tc.isValid {
				if !strings.Contains(mdResult, "✅") && !strings.Contains(mdResult, "Valid") {
					t.Error("Markdown missing valid indicator")
				}
				if !strings.Contains(textResult, "VALID") {
					t.Error("Text missing valid status")
				}
			} else {
				if !strings.Contains(mdResult, "❌") && !strings.Contains(mdResult, "Invalid") {
					t.Error("Markdown missing invalid indicator")
				}
				if !strings.Contains(textResult, "INVALID") {
					t.Error("Text missing invalid status")
				}
			}
		})
	}
}

func TestFormatConsistency_FilePathPresence(t *testing.T) {
	ctx := context.Background()

	testFilePath := "test-book-with-special-chars_v2.epub"
	report := createTestReport(testFilePath, true)

	jsonReporter := NewJSONReporter()
	mdReporter := NewMarkdownReporter()
	textReporter := NewTextReporter()

	options := &ports.ReportOptions{}

	jsonResult, _ := jsonReporter.Format(ctx, report, options)
	mdResult, _ := mdReporter.Format(ctx, report, options)
	textResult, _ := textReporter.Format(ctx, report, options)

	if !strings.Contains(jsonResult, testFilePath) {
		t.Error("JSON output missing file path")
	}
	if !strings.Contains(mdResult, testFilePath) {
		t.Error("Markdown output missing file path")
	}
	if !strings.Contains(textResult, testFilePath) {
		t.Error("Text output missing file path")
	}
}

func TestFormatConsistency_MessageContent(t *testing.T) {
	ctx := context.Background()

	testMessage := "This is a detailed error message with special chars: <>&\""
	report := createTestReport("message.epub", false)
	report.Errors = []domain.ValidationError{
		{
			Code:      "MSG-001",
			Message:   testMessage,
			Severity:  domain.SeverityError,
			Timestamp: time.Now(),
		},
	}

	jsonReporter := NewJSONReporter()
	mdReporter := NewMarkdownReporter()
	textReporter := NewTextReporter()

	options := &ports.ReportOptions{}

	jsonResult, _ := jsonReporter.Format(ctx, report, options)
	mdResult, _ := mdReporter.Format(ctx, report, options)
	textResult, _ := textReporter.Format(ctx, report, options)

	var jsonData map[string]interface{}
	_ = json.Unmarshal([]byte(jsonResult), &jsonData)
	
	errors := jsonData["errors"].([]interface{})
	if len(errors) > 0 {
		firstError := errors[0].(map[string]interface{})
		if !strings.Contains(firstError["message"].(string), "detailed error message") {
			t.Error("JSON output missing/corrupted message content")
		}
	}

	if !strings.Contains(mdResult, "detailed error message") {
		t.Error("Markdown output missing message content")
	}

	if !strings.Contains(textResult, "detailed error message") {
		t.Error("Text output missing message content")
	}
}

func TestFormatConsistency_MultipleReports(t *testing.T) {
	ctx := context.Background()

	reports := []*domain.ValidationReport{
		createTestReport("file1.epub", true),
		createTestReport("file2.epub", false),
		createTestReport("file3.epub", false),
	}

	reports[1].Errors = []domain.ValidationError{
		{Code: "E1", Severity: domain.SeverityError, Timestamp: time.Now()},
	}
	reports[2].Errors = []domain.ValidationError{
		{Code: "E2", Severity: domain.SeverityError, Timestamp: time.Now()},
		{Code: "E3", Severity: domain.SeverityError, Timestamp: time.Now()},
	}

	jsonReporter := NewJSONReporter().(*JSONReporter)
	mdReporter := NewMarkdownReporter().(*MarkdownReporter)
	textReporter := NewTextReporter().(*TextReporter)

	options := &ports.ReportOptions{}

	jsonResult, _ := jsonReporter.FormatMultiple(ctx, reports, options)
	mdResult, _ := mdReporter.FormatMultiple(ctx, reports, options)
	textResult, _ := textReporter.FormatMultiple(ctx, reports, options)

	var jsonData map[string]interface{}
	_ = json.Unmarshal([]byte(jsonResult), &jsonData)

	if jsonData["total_files"] != float64(3) {
		t.Error("JSON multiple reports: incorrect total files")
	}
	if jsonData["valid_files"] != float64(1) {
		t.Error("JSON multiple reports: incorrect valid files count")
	}
	if jsonData["invalid_files"] != float64(2) {
		t.Error("JSON multiple reports: incorrect invalid files count")
	}

	if !strings.Contains(mdResult, "**Total Files:** 3") {
		t.Error("Markdown multiple reports: missing total files")
	}
	if !strings.Contains(textResult, "Total Files:     3") {
		t.Error("Text multiple reports: missing total files")
	}

	allFiles := []string{"file1.epub", "file2.epub", "file3.epub"}
	for _, file := range allFiles {
		if !strings.Contains(jsonResult, file) {
			t.Errorf("JSON missing file: %s", file)
		}
		if !strings.Contains(mdResult, file) {
			t.Errorf("Markdown missing file: %s", file)
		}
		if !strings.Contains(textResult, file) {
			t.Errorf("Text missing file: %s", file)
		}
	}
}

func TestFormatConsistency_EmptyDetailsHandling(t *testing.T) {
	ctx := context.Background()

	report := createTestReport("empty_details.epub", false)
	report.Errors = []domain.ValidationError{
		{
			Code:      "EMPTY-001",
			Message:   "Error without details",
			Severity:  domain.SeverityError,
			Timestamp: time.Now(),
			Details:   nil,
		},
		{
			Code:      "EMPTY-002",
			Message:   "Error with empty details",
			Severity:  domain.SeverityError,
			Timestamp: time.Now(),
			Details:   make(map[string]interface{}),
		},
	}

	jsonReporter := NewJSONReporter()
	mdReporter := NewMarkdownReporter()
	textReporter := NewTextReporter()

	options := &ports.ReportOptions{Verbose: true}

	jsonResult, err1 := jsonReporter.Format(ctx, report, options)
	mdResult, err2 := mdReporter.Format(ctx, report, options)
	textResult, err3 := textReporter.Format(ctx, report, options)

	if err1 != nil || err2 != nil || err3 != nil {
		t.Fatalf("Errors handling empty details: json=%v, md=%v, text=%v", err1, err2, err3)
	}

	if jsonResult == "" || mdResult == "" || textResult == "" {
		t.Error("Expected non-empty results even with nil/empty details")
	}
}

func TestFormatConsistency_TimestampFormatting(t *testing.T) {
	ctx := context.Background()

	now := time.Now()
	report := createTestReport("timestamp.epub", false)
	report.ValidationTime = now
	report.Errors = []domain.ValidationError{
		{
			Code:      "TS-001",
			Message:   "Timestamp test",
			Severity:  domain.SeverityError,
			Timestamp: now,
		},
	}

	jsonReporter := NewJSONReporter()
	options := &ports.ReportOptions{}

	jsonResult, _ := jsonReporter.Format(ctx, report, options)

	var jsonData map[string]interface{}
	_ = json.Unmarshal([]byte(jsonResult), &jsonData)

	if jsonData["validation_time"] == "" {
		t.Error("JSON missing validation time")
	}

	errors := jsonData["errors"].([]interface{})
	if len(errors) > 0 {
		firstError := errors[0].(map[string]interface{})
		if firstError["timestamp"] == "" {
			t.Error("JSON error missing timestamp")
		}
	}
}
