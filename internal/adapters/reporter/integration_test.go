package reporter

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/example/project/internal/domain"
	"github.com/example/project/internal/ports"
)

func TestIntegration_AllReportersWithComplexReport(t *testing.T) {
	ctx := context.Background()

	report := createComplexValidationReport()

	reporters := map[string]ports.Reporter{
		"JSON":     NewJSONReporter(),
		"Markdown": NewMarkdownReporter(),
		"Text":     NewTextReporter(),
	}

	options := &ports.ReportOptions{
		IncludeWarnings: true,
		IncludeInfo:     true,
		Verbose:         true,
		ColorEnabled:    false,
	}

	for name, reporter := range reporters {
		t.Run(name, func(t *testing.T) {
			result, err := reporter.Format(ctx, report, options)
			if err != nil {
				t.Fatalf("Format failed for %s: %v", name, err)
			}

			if result == "" {
				t.Errorf("Expected non-empty result for %s", name)
			}

			if !strings.Contains(result, "complex.epub") {
				t.Errorf("Expected file path in %s output", name)
			}

			if !strings.Contains(result, "EPUB-001") {
				t.Errorf("Expected error code in %s output", name)
			}

			if !strings.Contains(result, "EPUB-W01") {
				t.Errorf("Expected warning code in %s output", name)
			}

			if !strings.Contains(result, "EPUB-I01") {
				t.Errorf("Expected info code in %s output", name)
			}
		})
	}
}

func TestIntegration_FilteringAcrossReporters(t *testing.T) {
	ctx := context.Background()

	report := createComplexValidationReport()

	filter := &Filter{
		Severities: []domain.Severity{domain.SeverityError},
	}

	reporters := map[string]ports.Reporter{
		"JSON":     NewJSONReporterWithFilter(filter),
		"Markdown": NewMarkdownReporterWithFilter(filter),
		"Text":     NewTextReporterWithFilter(filter),
	}

	options := &ports.ReportOptions{
		IncludeWarnings: true,
		IncludeInfo:     true,
	}

	for name, reporter := range reporters {
		t.Run(name, func(t *testing.T) {
			result, err := reporter.Format(ctx, report, options)
			if err != nil {
				t.Fatalf("Format failed for %s: %v", name, err)
			}

			if !strings.Contains(result, "EPUB-001") {
				t.Errorf("Expected error code in filtered %s output", name)
			}

			if strings.Contains(result, "EPUB-W01") {
				t.Errorf("Did not expect warning code in filtered %s output", name)
			}

			if strings.Contains(result, "EPUB-I01") {
				t.Errorf("Did not expect info code in filtered %s output", name)
			}
		})
	}
}

func TestIntegration_CategoryFiltering(t *testing.T) {
	ctx := context.Background()

	report := createReportWithCategories()

	filter := &Filter{
		Categories: []string{"structure"},
	}

	reporters := map[string]ports.Reporter{
		"JSON":     NewJSONReporterWithFilter(filter),
		"Markdown": NewMarkdownReporterWithFilter(filter),
		"Text":     NewTextReporterWithFilter(filter),
	}

	options := &ports.ReportOptions{
		IncludeWarnings: true,
		IncludeInfo:     true,
	}

	for name, reporter := range reporters {
		t.Run(name, func(t *testing.T) {
			result, err := reporter.Format(ctx, report, options)
			if err != nil {
				t.Fatalf("Format failed for %s: %v", name, err)
			}

			if !strings.Contains(result, "Structure error") {
				t.Errorf("Expected structure error in filtered %s output", name)
			}

			if strings.Contains(result, "Metadata error") {
				t.Errorf("Did not expect metadata error in filtered %s output", name)
			}
		})
	}
}

func TestIntegration_StandardFiltering(t *testing.T) {
	ctx := context.Background()

	report := createReportWithStandards()

	filter := &Filter{
		Standards: []string{"EPUB3"},
	}

	reporters := map[string]ports.Reporter{
		"JSON":     NewJSONReporterWithFilter(filter),
		"Markdown": NewMarkdownReporterWithFilter(filter),
		"Text":     NewTextReporterWithFilter(filter),
	}

	options := &ports.ReportOptions{}

	for name, reporter := range reporters {
		t.Run(name, func(t *testing.T) {
			result, err := reporter.Format(ctx, report, options)
			if err != nil {
				t.Fatalf("Format failed for %s: %v", name, err)
			}

			if !strings.Contains(result, "EPUB3 compliance") {
				t.Errorf("Expected EPUB3 error in filtered %s output", name)
			}

			if strings.Contains(result, "PDF/A compliance") {
				t.Errorf("Did not expect PDF/A error in filtered %s output", name)
			}
		})
	}
}

func TestIntegration_MinSeverityFiltering(t *testing.T) {
	ctx := context.Background()

	report := createComplexValidationReport()

	filter := &Filter{
		MinSeverity: domain.SeverityWarning,
	}

	reporters := map[string]ports.Reporter{
		"JSON":     NewJSONReporterWithFilter(filter),
		"Markdown": NewMarkdownReporterWithFilter(filter),
		"Text":     NewTextReporterWithFilter(filter),
	}

	options := &ports.ReportOptions{
		IncludeWarnings: true,
		IncludeInfo:     true,
	}

	for name, reporter := range reporters {
		t.Run(name, func(t *testing.T) {
			result, err := reporter.Format(ctx, report, options)
			if err != nil {
				t.Fatalf("Format failed for %s: %v", name, err)
			}

			if !strings.Contains(result, "EPUB-001") {
				t.Errorf("Expected error in filtered %s output", name)
			}

			if !strings.Contains(result, "EPUB-W01") {
				t.Errorf("Expected warning in filtered %s output", name)
			}

			if strings.Contains(result, "EPUB-I01") {
				t.Errorf("Did not expect info in filtered %s output (below min severity)", name)
			}
		})
	}
}

func TestIntegration_MultipleReportsFormatting(t *testing.T) {
	ctx := context.Background()

	reports := []*domain.ValidationReport{
		createTestReport("file1.epub", true),
		createTestReport("file2.epub", false),
		createTestReport("file3.pdf", false),
	}

	reports[1].Errors = []domain.ValidationError{
		{Code: "E1", Message: "Error 1", Severity: domain.SeverityError, Timestamp: time.Now()},
	}

	reports[2].Errors = []domain.ValidationError{
		{Code: "E2", Message: "Error 2", Severity: domain.SeverityError, Timestamp: time.Now()},
		{Code: "E3", Message: "Error 3", Severity: domain.SeverityError, Timestamp: time.Now()},
	}

	reporters := map[string]interface{}{
		"JSON":     NewJSONReporter().(*JSONReporter),
		"Markdown": NewMarkdownReporter().(*MarkdownReporter),
		"Text":     NewTextReporter().(*TextReporter),
	}

	options := &ports.ReportOptions{}

	for name, rep := range reporters {
		t.Run(name, func(t *testing.T) {
			var result string
			var err error
			switch reporter := rep.(type) {
			case *JSONReporter:
				result, err = reporter.FormatMultiple(ctx, reports, options)
			case *MarkdownReporter:
				result, err = reporter.FormatMultiple(ctx, reports, options)
			case *TextReporter:
				result, err = reporter.FormatMultiple(ctx, reports, options)
			}
			if err != nil {
				t.Fatalf("FormatMultiple failed for %s: %v", name, err)
			}

			if !strings.Contains(result, "file1.epub") {
				t.Errorf("Expected file1.epub in %s output", name)
			}

			if !strings.Contains(result, "file2.epub") {
				t.Errorf("Expected file2.epub in %s output", name)
			}

			if !strings.Contains(result, "file3.pdf") {
				t.Errorf("Expected file3.pdf in %s output", name)
			}
		})
	}
}

func TestIntegration_EmptyReport(t *testing.T) {
	ctx := context.Background()

	report := createTestReport("empty.epub", true)

	reporters := map[string]ports.Reporter{
		"JSON":     NewJSONReporter(),
		"Markdown": NewMarkdownReporter(),
		"Text":     NewTextReporter(),
	}

	options := &ports.ReportOptions{}

	for name, reporter := range reporters {
		t.Run(name, func(t *testing.T) {
			result, err := reporter.Format(ctx, report, options)
			if err != nil {
				t.Fatalf("Format failed for %s: %v", name, err)
			}

			if result == "" {
				t.Errorf("Expected non-empty result for %s even with empty report", name)
			}
		})
	}
}

func TestIntegration_LargeReport(t *testing.T) {
	ctx := context.Background()

	report := createTestReport("large.epub", false)
	report.Errors = make([]domain.ValidationError, 100)
	for i := range report.Errors {
		report.Errors[i] = domain.ValidationError{
			Code:      "LARGE-001",
			Message:   "Large report error",
			Severity:  domain.SeverityError,
			Timestamp: time.Now(),
		}
	}

	reporters := map[string]ports.Reporter{
		"JSON":     NewJSONReporter(),
		"Markdown": NewMarkdownReporter(),
		"Text":     NewTextReporter(),
	}

	options := &ports.ReportOptions{}

	for name, reporter := range reporters {
		t.Run(name, func(t *testing.T) {
			result, err := reporter.Format(ctx, report, options)
			if err != nil {
				t.Fatalf("Format failed for %s: %v", name, err)
			}

			if result == "" {
				t.Errorf("Expected non-empty result for %s with large report", name)
			}
		})
	}
}

func TestIntegration_ConsistencyAcrossFormats(t *testing.T) {
	ctx := context.Background()

	report := createComplexValidationReport()

	jsonReporter := NewJSONReporter()
	mdReporter := NewMarkdownReporter()
	textReporter := NewTextReporter()

	options := &ports.ReportOptions{
		IncludeWarnings: true,
		IncludeInfo:     true,
	}

	jsonResult, err := jsonReporter.Format(ctx, report, options)
	if err != nil {
		t.Fatalf("JSON format failed: %v", err)
	}

	mdResult, err := mdReporter.Format(ctx, report, options)
	if err != nil {
		t.Fatalf("Markdown format failed: %v", err)
	}

	textResult, err := textReporter.Format(ctx, report, options)
	if err != nil {
		t.Fatalf("Text format failed: %v", err)
	}

	requiredElements := []string{
		"complex.epub",
		"EPUB-001",
		"EPUB-W01",
		"EPUB-I01",
		"Critical error",
		"Deprecation warning",
		"Information message",
	}

	for _, elem := range requiredElements {
		if !strings.Contains(jsonResult, elem) {
			t.Errorf("JSON output missing required element: %s", elem)
		}
		if !strings.Contains(mdResult, elem) {
			t.Errorf("Markdown output missing required element: %s", elem)
		}
		if !strings.Contains(textResult, elem) {
			t.Errorf("Text output missing required element: %s", elem)
		}
	}
}

func createComplexValidationReport() *domain.ValidationReport {
	return &domain.ValidationReport{
		FilePath: "complex.epub",
		FileType: "EPUB",
		IsValid:  false,
		Errors: []domain.ValidationError{
			{
				Code:      "EPUB-001",
				Message:   "Critical error",
				Severity:  domain.SeverityError,
				Timestamp: time.Now(),
				Location: &domain.ErrorLocation{
					File:    "content.opf",
					Line:    10,
					Column:  5,
					Path:    "OPS/content.opf",
					Context: "<metadata>",
				},
				Details: map[string]interface{}{
					"category": "structure",
					"standard": "EPUB3",
				},
			},
			{
				Code:      "EPUB-002",
				Message:   "Missing required element",
				Severity:  domain.SeverityError,
				Timestamp: time.Now(),
				Location: &domain.ErrorLocation{
					File: "content.opf",
					Path: "OPS/content.opf",
				},
				Details: map[string]interface{}{
					"category": "structure",
				},
			},
		},
		Warnings: []domain.ValidationError{
			{
				Code:      "EPUB-W01",
				Message:   "Deprecation warning",
				Severity:  domain.SeverityWarning,
				Timestamp: time.Now(),
				Location: &domain.ErrorLocation{
					File: "chapter1.xhtml",
					Line: 25,
					Path: "OPS/chapter1.xhtml",
				},
				Details: map[string]interface{}{
					"category": "content",
				},
			},
		},
		Info: []domain.ValidationError{
			{
				Code:      "EPUB-I01",
				Message:   "Information message",
				Severity:  domain.SeverityInfo,
				Timestamp: time.Now(),
				Details: map[string]interface{}{
					"category": "metadata",
				},
			},
		},
		ValidationTime: time.Now(),
		Duration:       250 * time.Millisecond,
		Metadata: map[string]interface{}{
			"version": "3.0",
			"title":   "Complex Book",
		},
	}
}

func createReportWithCategories() *domain.ValidationReport {
	return &domain.ValidationReport{
		FilePath: "categorized.epub",
		FileType: "EPUB",
		IsValid:  false,
		Errors: []domain.ValidationError{
			{
				Code:      "CAT-001",
				Message:   "Structure error",
				Severity:  domain.SeverityError,
				Timestamp: time.Now(),
				Details: map[string]interface{}{
					"category": "structure",
				},
			},
			{
				Code:      "CAT-002",
				Message:   "Metadata error",
				Severity:  domain.SeverityError,
				Timestamp: time.Now(),
				Details: map[string]interface{}{
					"category": "metadata",
				},
			},
			{
				Code:      "CAT-003",
				Message:   "Content error",
				Severity:  domain.SeverityError,
				Timestamp: time.Now(),
				Details: map[string]interface{}{
					"category": "content",
				},
			},
		},
		Warnings:       make([]domain.ValidationError, 0),
		Info:           make([]domain.ValidationError, 0),
		ValidationTime: time.Now(),
		Duration:       100 * time.Millisecond,
		Metadata:       make(map[string]interface{}),
	}
}

func createReportWithStandards() *domain.ValidationReport {
	return &domain.ValidationReport{
		FilePath: "standards.epub",
		FileType: "EPUB",
		IsValid:  false,
		Errors: []domain.ValidationError{
			{
				Code:      "STD-001",
				Message:   "EPUB3 compliance issue",
				Severity:  domain.SeverityError,
				Timestamp: time.Now(),
				Details: map[string]interface{}{
					"standard": "EPUB3",
				},
			},
			{
				Code:      "STD-002",
				Message:   "PDF/A compliance issue",
				Severity:  domain.SeverityError,
				Timestamp: time.Now(),
				Details: map[string]interface{}{
					"standard": "PDF/A",
				},
			},
		},
		Warnings:       make([]domain.ValidationError, 0),
		Info:           make([]domain.ValidationError, 0),
		ValidationTime: time.Now(),
		Duration:       100 * time.Millisecond,
		Metadata:       make(map[string]interface{}),
	}
}
