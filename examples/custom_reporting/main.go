// Package main provides an example program for EBMLib.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/example/project/pkg/ebmlib"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: custom_reporting <file.epub|file.pdf>")
		os.Exit(1)
	}

	filePath := os.Args[1]
	fileExt := filePath[len(filePath)-4:]

	var report *ebmlib.ValidationReport
	var err error

	switch fileExt {
	case "epub":
		fmt.Printf("Validating EPUB: %s\n\n", filePath)
		report, err = ebmlib.ValidateEPUB(filePath)
	case ".pdf":
		fmt.Printf("Validating PDF: %s\n\n", filePath)
		report, err = ebmlib.ValidatePDF(filePath)
	default:
		log.Fatalf("Unsupported file type: %s (expected .epub or .pdf)", fileExt)
	}

	if err != nil {
		log.Fatalf("Validation error: %v", err)
	}

	demonstrateJSONReport(report)
	demonstrateTextReport(report)
	demonstrateMarkdownReport(report)
	demonstrateCustomOptions(report)
}

func demonstrateJSONReport(report *ebmlib.ValidationReport) {
	fmt.Println("=== JSON Report ===")

	jsonOutput, err := ebmlib.FormatReport(report, ebmlib.FormatJSON)
	if err != nil {
		log.Printf("Error formatting JSON: %v", err)
		return
	}

	if len(jsonOutput) > 500 {
		fmt.Printf("%s...\n\n", jsonOutput[:500])
	} else {
		fmt.Printf("%s\n\n", jsonOutput)
	}

	err = ebmlib.WriteReportToFile(report, "validation_report.json", &ebmlib.ReportOptions{
		Format:          ebmlib.FormatJSON,
		IncludeWarnings: true,
		IncludeInfo:     true,
		Verbose:         true,
	})
	if err != nil {
		log.Printf("Error writing JSON file: %v", err)
	} else {
		fmt.Println("Full JSON report written to: validation_report.json")
		fmt.Println()
	}
}

func demonstrateTextReport(report *ebmlib.ValidationReport) {
	fmt.Println("=== Text Report (with colors) ===")

	ctx := context.Background()

	options := &ebmlib.ReportOptions{
		Format:          ebmlib.FormatText,
		IncludeWarnings: true,
		IncludeInfo:     true,
		Verbose:         true,
		ColorEnabled:    true,
	}

	textOutput, err := ebmlib.FormatReportWithOptions(ctx, report, options)
	if err != nil {
		log.Printf("Error formatting text: %v", err)
		return
	}

	fmt.Println(textOutput)

	optionsNoColor := &ebmlib.ReportOptions{
		Format:          ebmlib.FormatText,
		IncludeWarnings: true,
		IncludeInfo:     true,
		Verbose:         false,
		ColorEnabled:    false,
	}

	err = ebmlib.WriteReportToFile(report, "validation_report.txt", optionsNoColor)
	if err != nil {
		log.Printf("Error writing text file: %v", err)
	} else {
		fmt.Println()
		fmt.Println("Text report (without colors) written to: validation_report.txt")
		fmt.Println()
	}
}

func demonstrateMarkdownReport(report *ebmlib.ValidationReport) {
	fmt.Println("=== Markdown Report ===")

	ctx := context.Background()

	options := &ebmlib.ReportOptions{
		Format:          ebmlib.FormatMarkdown,
		IncludeWarnings: true,
		IncludeInfo:     true,
		Verbose:         true,
	}

	mdOutput, err := ebmlib.FormatReportWithOptions(ctx, report, options)
	if err != nil {
		log.Printf("Error formatting Markdown: %v", err)
		return
	}

	if len(mdOutput) > 500 {
		fmt.Printf("%s...\n\n", mdOutput[:500])
	} else {
		fmt.Printf("%s\n\n", mdOutput)
	}

	err = ebmlib.WriteReportToFile(report, "validation_report.md", options)
	if err != nil {
		log.Printf("Error writing Markdown file: %v", err)
	} else {
		fmt.Println("Markdown report written to: validation_report.md")
		fmt.Println()
	}
}

func demonstrateCustomOptions(report *ebmlib.ValidationReport) {
	fmt.Println("=== Custom Options Examples ===")
	fmt.Println()

	ctx := context.Background()

	fmt.Println("1. Errors only (no warnings/info):")
	options1 := &ebmlib.ReportOptions{
		Format:          ebmlib.FormatText,
		IncludeWarnings: false,
		IncludeInfo:     false,
		Verbose:         false,
		ColorEnabled:    false,
	}
	output1, err := ebmlib.FormatReportWithOptions(ctx, report, options1)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		lines := countLines(output1)
		fmt.Printf("   Output: %d lines (errors only)\n", lines)
	}

	fmt.Println()
	fmt.Println("2. Limited errors (max 5):")
	options2 := &ebmlib.ReportOptions{
		Format:          ebmlib.FormatText,
		IncludeWarnings: true,
		IncludeInfo:     true,
		MaxErrors:       5,
		Verbose:         false,
		ColorEnabled:    false,
	}
	output2, err := ebmlib.FormatReportWithOptions(ctx, report, options2)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		lines := countLines(output2)
		fmt.Printf("   Output: %d lines (max 5 errors)\n", lines)
	}

	fmt.Println()
	fmt.Println("3. Compact format (non-verbose):")
	options3 := &ebmlib.ReportOptions{
		Format:          ebmlib.FormatText,
		IncludeWarnings: true,
		IncludeInfo:     true,
		Verbose:         false,
		ColorEnabled:    false,
	}
	output3, err := ebmlib.FormatReportWithOptions(ctx, report, options3)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		lines := countLines(output3)
		fmt.Printf("   Output: %d lines (compact)\n", lines)
	}

	fmt.Println()
	fmt.Println("4. Verbose format with all details:")
	options4 := &ebmlib.ReportOptions{
		Format:          ebmlib.FormatJSON,
		IncludeWarnings: true,
		IncludeInfo:     true,
		Verbose:         true,
	}
	output4, err := ebmlib.FormatReportWithOptions(ctx, report, options4)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("   Output: %d bytes (verbose JSON)\n", len(output4))
	}
}

func countLines(s string) int {
	count := 0
	for _, c := range s {
		if c == '\n' {
			count++
		}
	}
	return count
}
