// Package main provides an example program for ebm-lib.
package main

import (
	"fmt"
	"os"

	"github.com/example/project/internal/adapters/epub"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: epub_validation_example <path-to-epub>")
		os.Exit(1)
	}

	filePath := os.Args[1]

	validator := epub.NewContainerValidator()

	fmt.Printf("Validating EPUB: %s\n", filePath)
	fmt.Println("---")

	result, err := validator.ValidateFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error validating file: %v\n", err)
		os.Exit(1)
	}

	if result.Valid {
		fmt.Println("✓ EPUB container is valid!")
		fmt.Println()

		if len(result.Rootfiles) > 0 {
			fmt.Println("Rootfiles:")
			for i, rootfile := range result.Rootfiles {
				fmt.Printf("  %d. %s (%s)\n", i+1, rootfile.FullPath, rootfile.MediaType)
			}
		}
	} else {
		fmt.Println("✗ EPUB container is invalid")
		fmt.Println()

		if len(result.Errors) > 0 {
			fmt.Println("Errors:")
			for i, validationError := range result.Errors {
				fmt.Printf("  %d. [%s] %s\n", i+1, validationError.Code, validationError.Message)

				if len(validationError.Details) > 0 {
					fmt.Println("     Details:")
					for key, value := range validationError.Details {
						fmt.Printf("       - %s: %v\n", key, value)
					}
				}
			}
		}
	}

	os.Exit(0)
}
