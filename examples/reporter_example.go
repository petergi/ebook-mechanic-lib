package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/example/project/internal/adapters/reporter"
	"github.com/example/project/internal/domain"
	"github.com/example/project/internal/ports"
)

func main() {
	ctx := context.Background()

	report := createSampleValidationReport()

	fmt.Println("=== Reporter Examples ===\n")

	demonstrateJSONReporter(ctx, report)
	demonstrateMarkdownReporter(ctx, report)
	demonstrateTextReporter(ctx, report)
	demonstrateFiltering(ctx, report)
	demonstrateMultipleReports(ctx)

	fmt.Println("\nDone! Check the generated files in the current directory.")
}

func demonstrateJSONReporter(ctx context.Context, report *domain.ValidationReport) {
	fmt.Println("1. JSON Reporter Example")
	fmt.Println("------------------------")

	jsonReporter := reporter.NewJSONReporter()

	options := &ports.ReportOptions{
		Format:          ports.FormatJSON,
		IncludeWarnings: true,
		IncludeInfo:     true,
		Verbose:         true,
	}

	result, err := jsonReporter.Format(ctx, report, options)
	if err != nil {
		log.Printf("Error formatting JSON: %v\n", err)
		return
	}

	fmt.Println("JSON output (first 200 chars):")
	if len(result) > 200 {
		fmt.Printf("%s...\n\n", result[:200])
	} else {
		fmt.Printf("%s\n\n", result)
	}

	err = jsonReporter.WriteToFile(ctx, report, "report.json", options)
	if err != nil {
		log.Printf("Error writing JSON file: %v\n", err)
	} else {
		fmt.Println("Full report written to: report.json\n")
	}
}

func demonstrateMarkdownReporter(ctx context.Context, report *domain.ValidationReport) {
	fmt.Println("2. Markdown Reporter Example")
	fmt.Println("----------------------------")

	mdReporter := reporter.NewMarkdownReporter()

	options := &ports.ReportOptions{
		Format:          ports.FormatMarkdown,
		IncludeWarnings: true,
		IncludeInfo:     true,
		Verbose:         true,
	}

	result, err := mdReporter.Format(ctx, report, options)
	if err != nil {
		log.Printf("Error formatting Markdown: %v\n", err)
		return
	}

	fmt.Println("Markdown output (first 300 chars):")
	if len(result) > 300 {
		fmt.Printf("%s...\n\n", result[:300])
	} else {
		fmt.Printf("%s\n\n", result)
	}

	err = mdReporter.WriteToFile(ctx, report, "report.md", options)
	if err != nil {
		log.Printf("Error writing Markdown file: %v\n", err)
	} else {
		fmt.Println("Full report written to: report.md\n")
	}
}

func demonstrateTextReporter(ctx context.Context, report *domain.ValidationReport) {
	fmt.Println("3. Text Reporter Example (with colors)")
	fmt.Println("--------------------------------------")

	textReporter := reporter.NewTextReporter()

	options := &ports.ReportOptions{
		Format:          ports.FormatText,
		IncludeWarnings: true,
		IncludeInfo:     true,
		Verbose:         true,
		ColorEnabled:    true,
	}

	result, err := textReporter.Format(ctx, report, options)
	if err != nil {
		log.Printf("Error formatting text: %v\n", err)
		return
	}

	fmt.Println("Colored text output:")
	fmt.Printf("%s\n", result)

	optionsNoColor := &ports.ReportOptions{
		Format:          ports.FormatText,
		IncludeWarnings: true,
		IncludeInfo:     true,
		Verbose:         false,
		ColorEnabled:    false,
	}

	err = textReporter.WriteToFile(ctx, report, "report.txt", optionsNoColor)
	if err != nil {
		log.Printf("Error writing text file: %v\n", err)
	} else {
		fmt.Println("\nReport (without colors) written to: report.txt\n")
	}
}

func demonstrateFiltering(ctx context.Context, report *domain.ValidationReport) {
	fmt.Println("4. Filtering Examples")
	fmt.Println("--------------------")

	fmt.Println("a) Filter by severity (errors only):")
	filter := &reporter.Filter{
		Severities: []domain.Severity{domain.SeverityError},
	}
	jsonReporter := reporter.NewJSONReporterWithFilter(filter)
	options := &ports.ReportOptions{
		IncludeWarnings: true,
		IncludeInfo:     true,
	}

	result, err := jsonReporter.Format(ctx, report, options)
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("   Found %d errors (warnings/info filtered out)\n", countOccurrences(result, `"severity": "error"`))
	}

	fmt.Println("\nb) Filter by minimum severity (warnings and above):")
	filter2 := &reporter.Filter{
		MinSeverity: domain.SeverityWarning,
	}
	textReporter := reporter.NewTextReporterWithFilter(filter2)
	result2, err := textReporter.Format(ctx, report, &ports.ReportOptions{
		IncludeWarnings: true,
		IncludeInfo:     true,
		ColorEnabled:    false,
	})
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("   Output length: %d chars (info messages filtered)\n", len(result2))
	}

	fmt.Println("\nc) Filter by category (structure issues only):")
	filter3 := &reporter.Filter{
		Categories: []string{"structure"},
	}
	mdReporter := reporter.NewMarkdownReporterWithFilter(filter3)
	result3, err := mdReporter.Format(ctx, report, options)
	if err != nil {
		log.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("   Output length: %d chars (only structure category)\n\n", len(result3))
	}
}

func demonstrateMultipleReports(ctx context.Context) {
	fmt.Println("5. Multiple Reports Summary")
	fmt.Println("--------------------------")

	reports := []*domain.ValidationReport{
		createReport("book1.epub", true, 0, 0, 0),
		createReport("book2.epub", false, 3, 2, 1),
		createReport("book3.epub", false, 1, 0, 0),
		createReport("book4.epub", true, 0, 1, 2),
	}

	textReporter := reporter.NewTextReporter()
	options := &ports.ReportOptions{
		ColorEnabled: true,
	}

	var buf bytes.Buffer
	err := textReporter.WriteSummary(ctx, reports, &buf, options)
	if err != nil {
		log.Printf("Error generating summary: %v\n", err)
		return
	}

	fmt.Println(buf.String())

	jsonReporter := reporter.NewJSONReporter()
	err = jsonReporter.WriteToFile(ctx, &domain.ValidationReport{}, "summary.json", options)
	if err == nil {
		err = os.Remove("summary.json")
	}
}

func createSampleValidationReport() *domain.ValidationReport {
	return &domain.ValidationReport{
		FilePath: "sample-book.epub",
		FileType: "EPUB",
		IsValid:  false,
		Errors: []domain.ValidationError{
			{
				Code:      "EPUB-001",
				Message:   "Missing required metadata element: dc:title",
				Severity:  domain.SeverityError,
				Timestamp: time.Now(),
				Location: &domain.ErrorLocation{
					File:    "content.opf",
					Line:    15,
					Column:  8,
					Path:    "OEBPS/content.opf",
					Context: "<metadata>",
				},
				Details: map[string]interface{}{
					"category": "metadata",
					"standard": "EPUB3",
					"element":  "dc:title",
				},
			},
			{
				Code:      "EPUB-002",
				Message:   "Invalid spine item reference: chapter4.xhtml not found in manifest",
				Severity:  domain.SeverityError,
				Timestamp: time.Now(),
				Location: &domain.ErrorLocation{
					File:   "content.opf",
					Line:   45,
					Column: 12,
					Path:   "OEBPS/content.opf",
				},
				Details: map[string]interface{}{
					"category":      "structure",
					"standard":      "EPUB3",
					"missing_item":  "chapter4.xhtml",
					"spine_idref":   "ch4",
				},
			},
		},
		Warnings: []domain.ValidationError{
			{
				Code:      "EPUB-W01",
				Message:   "Deprecated HTML element used: <font>",
				Severity:  domain.SeverityWarning,
				Timestamp: time.Now(),
				Location: &domain.ErrorLocation{
					File:   "chapter1.xhtml",
					Line:   127,
					Column: 5,
					Path:   "OEBPS/Text/chapter1.xhtml",
				},
				Details: map[string]interface{}{
					"category": "content",
					"element":  "font",
					"recommendation": "Use CSS for styling instead",
				},
			},
			{
				Code:      "EPUB-W02",
				Message:   "Image file size exceeds recommended limit",
				Severity:  domain.SeverityWarning,
				Timestamp: time.Now(),
				Location: &domain.ErrorLocation{
					File: "cover.jpg",
					Path: "OEBPS/Images/cover.jpg",
				},
				Details: map[string]interface{}{
					"category":      "optimization",
					"file_size":     "5.2MB",
					"recommended":   "2MB",
				},
			},
		},
		Info: []domain.ValidationError{
			{
				Code:      "EPUB-I01",
				Message:   "Optional metadata element missing: dc:description",
				Severity:  domain.SeverityInfo,
				Timestamp: time.Now(),
				Details: map[string]interface{}{
					"category": "metadata",
					"element":  "dc:description",
					"benefit":  "Improves discoverability",
				},
			},
		},
		ValidationTime: time.Now(),
		Duration:       234 * time.Millisecond,
		Metadata: map[string]interface{}{
			"epub_version": "3.0",
			"file_size":    "12.4MB",
			"chapters":     10,
			"images":       25,
		},
	}
}

func createReport(filename string, isValid bool, errors, warnings, info int) *domain.ValidationReport {
	report := &domain.ValidationReport{
		FilePath:       filename,
		FileType:       "EPUB",
		IsValid:        isValid,
		Errors:         make([]domain.ValidationError, 0),
		Warnings:       make([]domain.ValidationError, 0),
		Info:           make([]domain.ValidationError, 0),
		ValidationTime: time.Now(),
		Duration:       100 * time.Millisecond,
		Metadata:       make(map[string]interface{}),
	}

	for i := 0; i < errors; i++ {
		report.Errors = append(report.Errors, domain.ValidationError{
			Code:      fmt.Sprintf("E%d", i+1),
			Message:   fmt.Sprintf("Error %d", i+1),
			Severity:  domain.SeverityError,
			Timestamp: time.Now(),
		})
	}

	for i := 0; i < warnings; i++ {
		report.Warnings = append(report.Warnings, domain.ValidationError{
			Code:      fmt.Sprintf("W%d", i+1),
			Message:   fmt.Sprintf("Warning %d", i+1),
			Severity:  domain.SeverityWarning,
			Timestamp: time.Now(),
		})
	}

	for i := 0; i < info; i++ {
		report.Info = append(report.Info, domain.ValidationError{
			Code:      fmt.Sprintf("I%d", i+1),
			Message:   fmt.Sprintf("Info %d", i+1),
			Severity:  domain.SeverityInfo,
			Timestamp: time.Now(),
		})
	}

	return report
}

func countOccurrences(s, substr string) int {
	count := 0
	for i := 0; i < len(s); i++ {
		if i+len(substr) <= len(s) && s[i:i+len(substr)] == substr {
			count++
			i += len(substr) - 1
		}
	}
	return count
}
