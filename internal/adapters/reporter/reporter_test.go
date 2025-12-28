package reporter

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/example/project/internal/domain"
	"github.com/example/project/internal/ports"
)

func TestReporterPackage_AllFormats(t *testing.T) {
	ctx := context.Background()

	report := createTestReport("package_test.epub", false)
	report.Errors = []domain.ValidationError{
		{
			Code:      "TEST-001",
			Message:   "Package test error",
			Severity:  domain.SeverityError,
			Timestamp: time.Now(),
			Location: &domain.ErrorLocation{
				File:   "test.opf",
				Line:   10,
				Column: 5,
			},
		},
	}

	options := &ports.ReportOptions{
		IncludeWarnings: true,
		IncludeInfo:     true,
		Verbose:         true,
		ColorEnabled:    false,
	}

	reporters := []struct {
		name     string
		reporter ports.Reporter
	}{
		{"JSON", NewJSONReporter()},
		{"Markdown", NewMarkdownReporter()},
		{"Text", NewTextReporter()},
	}

	for _, tc := range reporters {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tc.reporter.Format(ctx, report, options)
			if err != nil {
				t.Fatalf("Format failed: %v", err)
			}

			if result == "" {
				t.Error("Expected non-empty result")
			}
		})
	}
}

func TestReporterPackage_FileIO(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	report := createTestReport("io_test.epub", true)
	options := &ports.ReportOptions{}

	tests := []struct {
		name     string
		reporter ports.Reporter
		filename string
	}{
		{"JSON", NewJSONReporter(), "test.json"},
		{"Markdown", NewMarkdownReporter(), "test.md"},
		{"Text", NewTextReporter(), "test.txt"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(tmpDir, tc.filename)
			
			err := tc.reporter.WriteToFile(ctx, report, path, options)
			if err != nil {
				t.Fatalf("WriteToFile failed: %v", err)
			}

			info, err := os.Stat(path)
			if err != nil {
				t.Fatalf("File not created: %v", err)
			}

			if info.Size() == 0 {
				t.Error("Expected non-empty file")
			}
		})
	}
}

func TestReporterPackage_EdgeCases(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name   string
		report *domain.ValidationReport
	}{
		{
			name:   "empty report",
			report: createTestReport("empty.epub", true),
		},
		{
			name: "report with nil location",
			report: &domain.ValidationReport{
				FilePath: "nil_location.epub",
				FileType: "EPUB",
				IsValid:  false,
				Errors: []domain.ValidationError{
					{
						Code:      "TEST-001",
						Message:   "Error without location",
						Severity:  domain.SeverityError,
						Location:  nil,
						Timestamp: time.Now(),
					},
				},
				Warnings:       make([]domain.ValidationError, 0),
				Info:           make([]domain.ValidationError, 0),
				ValidationTime: time.Now(),
				Duration:       100 * time.Millisecond,
				Metadata:       make(map[string]interface{}),
			},
		},
		{
			name: "report with nil details",
			report: &domain.ValidationReport{
				FilePath: "nil_details.epub",
				FileType: "EPUB",
				IsValid:  false,
				Errors: []domain.ValidationError{
					{
						Code:      "TEST-001",
						Message:   "Error without details",
						Severity:  domain.SeverityError,
						Details:   nil,
						Timestamp: time.Now(),
					},
				},
				Warnings:       make([]domain.ValidationError, 0),
				Info:           make([]domain.ValidationError, 0),
				ValidationTime: time.Now(),
				Duration:       100 * time.Millisecond,
				Metadata:       make(map[string]interface{}),
			},
		},
		{
			name: "report with nil metadata",
			report: &domain.ValidationReport{
				FilePath:       "nil_metadata.epub",
				FileType:       "EPUB",
				IsValid:        true,
				Errors:         make([]domain.ValidationError, 0),
				Warnings:       make([]domain.ValidationError, 0),
				Info:           make([]domain.ValidationError, 0),
				ValidationTime: time.Now(),
				Duration:       100 * time.Millisecond,
				Metadata:       nil,
			},
		},
	}

	reporters := []ports.Reporter{
		NewJSONReporter(),
		NewMarkdownReporter(),
		NewTextReporter(),
	}

	options := &ports.ReportOptions{}

	for _, tc := range tests {
		for _, reporter := range reporters {
			t.Run(tc.name, func(t *testing.T) {
				_, err := reporter.Format(ctx, tc.report, options)
				if err != nil {
					t.Errorf("Format should not fail on edge case: %v", err)
				}
			})
		}
	}
}

func TestReporterPackage_OptionsHandling(t *testing.T) {
	ctx := context.Background()

	report := createTestReport("options_test.epub", false)
	report.Errors = []domain.ValidationError{
		{Code: "E1", Severity: domain.SeverityError, Timestamp: time.Now()},
		{Code: "E2", Severity: domain.SeverityError, Timestamp: time.Now()},
		{Code: "E3", Severity: domain.SeverityError, Timestamp: time.Now()},
	}
	report.Warnings = []domain.ValidationError{
		{Code: "W1", Severity: domain.SeverityWarning, Timestamp: time.Now()},
	}
	report.Info = []domain.ValidationError{
		{Code: "I1", Severity: domain.SeverityInfo, Timestamp: time.Now()},
	}

	tests := []struct {
		name    string
		options *ports.ReportOptions
	}{
		{
			name:    "nil options",
			options: nil,
		},
		{
			name: "exclude warnings",
			options: &ports.ReportOptions{
				IncludeWarnings: false,
				IncludeInfo:     true,
			},
		},
		{
			name: "exclude info",
			options: &ports.ReportOptions{
				IncludeWarnings: true,
				IncludeInfo:     false,
			},
		},
		{
			name: "max errors limit",
			options: &ports.ReportOptions{
				MaxErrors: 2,
			},
		},
		{
			name: "verbose",
			options: &ports.ReportOptions{
				Verbose: true,
			},
		},
		{
			name: "colors enabled",
			options: &ports.ReportOptions{
				ColorEnabled: true,
			},
		},
	}

	reporters := []ports.Reporter{
		NewJSONReporter(),
		NewMarkdownReporter(),
		NewTextReporter(),
	}

	for _, tc := range tests {
		for _, reporter := range reporters {
			t.Run(tc.name, func(t *testing.T) {
				_, err := reporter.Format(ctx, report, tc.options)
				if err != nil {
					t.Errorf("Format failed with options: %v", err)
				}
			})
		}
	}
}

func TestReporterPackage_FilterIntegration(t *testing.T) {
	ctx := context.Background()

	report := createTestReport("filter_integration.epub", false)
	report.Errors = []domain.ValidationError{
		{
			Code:      "E1",
			Severity:  domain.SeverityError,
			Timestamp: time.Now(),
			Details: map[string]interface{}{
				"category": "structure",
				"standard": "EPUB3",
			},
		},
		{
			Code:      "E2",
			Severity:  domain.SeverityError,
			Timestamp: time.Now(),
			Details: map[string]interface{}{
				"category": "metadata",
				"standard": "PDF/A",
			},
		},
	}

	filters := []struct {
		name   string
		filter *Filter
	}{
		{
			name: "by category",
			filter: &Filter{
				Categories: []string{"structure"},
			},
		},
		{
			name: "by standard",
			filter: &Filter{
				Standards: []string{"EPUB3"},
			},
		},
		{
			name: "by severity",
			filter: &Filter{
				Severities: []domain.Severity{domain.SeverityError},
			},
		},
		{
			name: "combined",
			filter: &Filter{
				Categories: []string{"structure"},
				Standards:  []string{"EPUB3"},
			},
		},
	}

	for _, fc := range filters {
		t.Run(fc.name, func(t *testing.T) {
			jsonReporter := NewJSONReporterWithFilter(fc.filter)
			mdReporter := NewMarkdownReporterWithFilter(fc.filter)
			textReporter := NewTextReporterWithFilter(fc.filter)

			_, err1 := jsonReporter.Format(ctx, report, &ports.ReportOptions{})
			_, err2 := mdReporter.Format(ctx, report, &ports.ReportOptions{})
			_, err3 := textReporter.Format(ctx, report, &ports.ReportOptions{})

			if err1 != nil || err2 != nil || err3 != nil {
				t.Errorf("Filtering failed: json=%v, md=%v, text=%v", err1, err2, err3)
			}
		})
	}
}

func BenchmarkJSONReporter_Format(b *testing.B) {
	ctx := context.Background()
	reporter := NewJSONReporter()
	report := createTestReport("bench.epub", false)
	report.Errors = make([]domain.ValidationError, 50)
	for i := range report.Errors {
		report.Errors[i] = domain.ValidationError{
			Code:      "BENCH-001",
			Message:   "Benchmark error",
			Severity:  domain.SeverityError,
			Timestamp: time.Now(),
		}
	}
	options := &ports.ReportOptions{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = reporter.Format(ctx, report, options)
	}
}

func BenchmarkMarkdownReporter_Format(b *testing.B) {
	ctx := context.Background()
	reporter := NewMarkdownReporter()
	report := createTestReport("bench.epub", false)
	report.Errors = make([]domain.ValidationError, 50)
	for i := range report.Errors {
		report.Errors[i] = domain.ValidationError{
			Code:      "BENCH-001",
			Message:   "Benchmark error",
			Severity:  domain.SeverityError,
			Timestamp: time.Now(),
		}
	}
	options := &ports.ReportOptions{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = reporter.Format(ctx, report, options)
	}
}

func BenchmarkTextReporter_Format(b *testing.B) {
	ctx := context.Background()
	reporter := NewTextReporter()
	report := createTestReport("bench.epub", false)
	report.Errors = make([]domain.ValidationError, 50)
	for i := range report.Errors {
		report.Errors[i] = domain.ValidationError{
			Code:      "BENCH-001",
			Message:   "Benchmark error",
			Severity:  domain.SeverityError,
			Timestamp: time.Now(),
		}
	}
	options := &ports.ReportOptions{ColorEnabled: false}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = reporter.Format(ctx, report, options)
	}
}

func BenchmarkFilter_FilterErrors(b *testing.B) {
	errors := make([]domain.ValidationError, 100)
	for i := range errors {
		errors[i] = domain.ValidationError{
			Code:     "BENCH-001",
			Severity: domain.Severity([]string{"error", "warning", "info"}[i%3]),
			Details: map[string]interface{}{
				"category": []string{"structure", "metadata", "content"}[i%3],
			},
		}
	}

	filter := &Filter{
		Categories: []string{"structure"},
		MinSeverity: domain.SeverityWarning,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = filter.FilterErrors(errors)
	}
}
