// Package main provides an example program for ebm-lib.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/example/project/pkg/ebmlib"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: basic_validation <file.epub|file.pdf>")
		os.Exit(1)
	}

	filePath := os.Args[1]
	fileExt := filePath[len(filePath)-4:]

	var report *ebmlib.ValidationReport
	var err error

	switch fileExt {
	case "epub":
		fmt.Printf("Validating EPUB: %s\n", filePath)
		report, err = ebmlib.ValidateEPUB(filePath)
	case ".pdf":
		fmt.Printf("Validating PDF: %s\n", filePath)
		report, err = ebmlib.ValidatePDF(filePath)
	default:
		log.Fatalf("Unsupported file type: %s (expected .epub or .pdf)", fileExt)
	}

	if err != nil {
		log.Fatalf("Validation error: %v", err)
	}

	fmt.Println("\n=== Validation Results ===")
	fmt.Printf("File: %s\n", report.FilePath)
	fmt.Printf("Type: %s\n", report.FileType)
	fmt.Printf("Valid: %v\n", report.IsValid)
	fmt.Printf("Duration: %v\n", report.Duration)

	if report.ErrorCount() > 0 {
		fmt.Printf("\n%d Errors:\n", report.ErrorCount())
		for i, err := range report.Errors {
			fmt.Printf("  %d. [%s] %s\n", i+1, err.Code, err.Message)
			if err.Location != nil {
				fmt.Printf("     Location: %s", err.Location.Path)
				if err.Location.Line > 0 {
					fmt.Printf(":%d", err.Location.Line)
				}
				fmt.Println()
			}
		}
	}

	if report.WarningCount() > 0 {
		fmt.Printf("\n%d Warnings:\n", report.WarningCount())
		for i, warn := range report.Warnings {
			fmt.Printf("  %d. [%s] %s\n", i+1, warn.Code, warn.Message)
		}
	}

	if report.InfoCount() > 0 {
		fmt.Printf("\n%d Info messages:\n", report.InfoCount())
		for i, info := range report.Info {
			fmt.Printf("  %d. [%s] %s\n", i+1, info.Code, info.Message)
		}
	}

	if report.IsValid {
		fmt.Println("\n✓ File is valid!")
	} else {
		fmt.Println("\n✗ File has validation errors")
		os.Exit(1)
	}
}
