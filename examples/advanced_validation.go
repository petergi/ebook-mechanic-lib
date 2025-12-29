package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/example/project/pkg/ebmlib"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: advanced_validation <directory>")
		fmt.Println("Validates all EPUB and PDF files in the specified directory")
		os.Exit(1)
	}

	directory := os.Args[1]

	fmt.Printf("=== Advanced Validation Example ===\n\n")
	fmt.Printf("Scanning directory: %s\n\n", directory)

	demonstrateBatchValidation(directory)
	fmt.Println()
	demonstrateContextValidation()
	fmt.Println()
	demonstrateErrorAnalysis()
}

func demonstrateBatchValidation(directory string) {
	fmt.Println("--- Batch Validation ---")

	files, err := findEbookFiles(directory)
	if err != nil {
		log.Printf("Error scanning directory: %v", err)
		return
	}

	if len(files) == 0 {
		fmt.Println("No EPUB or PDF files found")
		return
	}

	fmt.Printf("Found %d ebook files\n\n", len(files))

	validCount := 0
	invalidCount := 0
	errorCount := 0

	for i, file := range files {
		fmt.Printf("[%d/%d] Validating: %s\n", i+1, len(files), filepath.Base(file))

		report, err := validateFile(file)
		if err != nil {
			log.Printf("  ✗ Error: %v\n", err)
			errorCount++
			continue
		}

		if report.IsValid {
			fmt.Printf("  ✓ Valid (%v)\n", report.Duration)
			validCount++
		} else {
			fmt.Printf("  ✗ Invalid: %d errors, %d warnings\n",
				report.ErrorCount(), report.WarningCount())
			invalidCount++

			for _, err := range report.Errors {
				fmt.Printf("      [%s] %s\n", err.Code, err.Message)
			}
		}
	}

	fmt.Printf("\nSummary:\n")
	fmt.Printf("  Total files: %d\n", len(files))
	fmt.Printf("  Valid: %d\n", validCount)
	fmt.Printf("  Invalid: %d\n", invalidCount)
	fmt.Printf("  Errors: %d\n", errorCount)
}

func demonstrateContextValidation() {
	fmt.Println("--- Context-Aware Validation ---")

	fmt.Println("Validating with 5-second timeout...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if len(os.Args) > 2 {
		filePath := os.Args[2]

		start := time.Now()
		report, err := ebmlib.ValidateEPUBWithContext(ctx, filePath)
		duration := time.Since(start)

		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				fmt.Printf("  ✗ Validation timed out after %v\n", duration)
			} else {
				fmt.Printf("  ✗ Error: %v\n", err)
			}
			return
		}

		fmt.Printf("  Completed in %v (within timeout)\n", duration)
		if report.IsValid {
			fmt.Println("  ✓ Valid")
		} else {
			fmt.Printf("  ✗ Invalid: %d errors\n", report.ErrorCount())
		}
	} else {
		fmt.Println("  (Provide a second argument to test context validation)")
	}
}

func demonstrateErrorAnalysis() {
	fmt.Println("--- Error Analysis ---")

	if len(os.Args) < 3 {
		fmt.Println("  (Provide a file path to analyze errors)")
		return
	}

	filePath := os.Args[2]
	report, err := validateFile(filePath)
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	if report.IsValid {
		fmt.Println("File is valid, no errors to analyze")
		return
	}

	errorsByCode := make(map[string][]ebmlib.ValidationError)
	errorsByFile := make(map[string]int)
	severityCounts := make(map[ebmlib.Severity]int)

	for _, err := range report.Errors {
		errorsByCode[err.Code] = append(errorsByCode[err.Code], err)
		severityCounts[err.Severity]++

		if err.Location != nil && err.Location.File != "" {
			errorsByFile[err.Location.File]++
		}
	}

	fmt.Printf("\nError Distribution by Code:\n")
	for code, errors := range errorsByCode {
		fmt.Printf("  %s: %d occurrences\n", code, len(errors))
		if len(errors) > 0 && len(errors) <= 3 {
			for _, err := range errors {
				fmt.Printf("    - %s\n", err.Message)
			}
		}
	}

	if len(errorsByFile) > 0 {
		fmt.Printf("\nError Distribution by File:\n")
		for file, count := range errorsByFile {
			fmt.Printf("  %s: %d errors\n", file, count)
		}
	}

	fmt.Printf("\nError Distribution by Severity:\n")
	for severity, count := range severityCounts {
		fmt.Printf("  %s: %d\n", severity, count)
	}

	if len(report.Warnings) > 0 {
		fmt.Printf("\nWarnings: %d\n", len(report.Warnings))
		for i, warn := range report.Warnings {
			if i < 3 {
				fmt.Printf("  - [%s] %s\n", warn.Code, warn.Message)
			}
		}
		if len(report.Warnings) > 3 {
			fmt.Printf("  ... and %d more\n", len(report.Warnings)-3)
		}
	}

	suggestRepairs(report)
}

func suggestRepairs(report *ebmlib.ValidationReport) {
	fmt.Println("\nRepair Suggestions:")

	repairableCount := 0
	manualCount := 0

	for _, err := range report.Errors {
		switch err.Code {
		case "EPUB-001", "EPUB-002":
			repairableCount++
		default:
			if err.Severity == ebmlib.SeverityError {
				manualCount++
			}
		}
	}

	if repairableCount > 0 {
		fmt.Printf("  ✓ %d errors may be automatically repairable\n", repairableCount)
		fmt.Println("    Run with -repair flag to attempt automatic repair")
	}

	if manualCount > 0 {
		fmt.Printf("  ⚠ %d errors may require manual intervention\n", manualCount)
	}

	if repairableCount == 0 && manualCount == 0 {
		fmt.Println("  No specific repair suggestions available")
	}
}

func validateFile(filePath string) (*ebmlib.ValidationReport, error) {
	ext := filepath.Ext(filePath)

	switch ext {
	case ".epub":
		return ebmlib.ValidateEPUB(filePath)
	case ".pdf":
		return ebmlib.ValidatePDF(filePath)
	default:
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}
}

func findEbookFiles(directory string) ([]string, error) {
	var files []string

	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		if ext == ".epub" || ext == ".pdf" {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}
