package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/example/project/pkg/ebmlib"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: complete_workflow <file.epub|file.pdf>")
		fmt.Println("\nThis example demonstrates a complete validation and repair workflow:")
		fmt.Println("  1. Validate the file")
		fmt.Println("  2. Generate reports in multiple formats")
		fmt.Println("  3. Preview repair actions if needed")
		fmt.Println("  4. Apply repairs and validate again")
		os.Exit(1)
	}

	filePath := os.Args[1]

	fmt.Printf("=== Complete Workflow for %s ===\n\n", filepath.Base(filePath))

	step1Validate(filePath)
	report, err := step2GenerateReports(filePath)
	if err != nil {
		log.Fatal(err)
	}

	if !report.IsValid {
		step3PreviewRepairs(filePath)
		step4ApplyRepairs(filePath)
		step5ValidateRepaired(filePath)
	}

	fmt.Println("\n=== Workflow Complete ===")
}

func step1Validate(filePath string) {
	fmt.Println("STEP 1: Initial Validation")
	fmt.Println("---------------------------")

	var report *ebmlib.ValidationReport
	var err error

	ext := filepath.Ext(filePath)
	switch ext {
	case ".epub":
		report, err = ebmlib.ValidateEPUB(filePath)
	case ".pdf":
		report, err = ebmlib.ValidatePDF(filePath)
	default:
		log.Fatalf("Unsupported file type: %s", ext)
	}

	if err != nil {
		log.Fatalf("Validation failed: %v", err)
	}

	fmt.Printf("File Type: %s\n", report.FileType)
	fmt.Printf("Validation Time: %v\n", report.Duration)
	fmt.Printf("Status: ")

	if report.IsValid {
		fmt.Println("✓ VALID")
		fmt.Println("\nNo issues found. File is compliant.")
	} else {
		fmt.Println("✗ INVALID")
		fmt.Printf("\nIssues found:\n")
		fmt.Printf("  Errors:   %d\n", report.ErrorCount())
		fmt.Printf("  Warnings: %d\n", report.WarningCount())
		fmt.Printf("  Info:     %d\n", report.InfoCount())

		if report.ErrorCount() > 0 {
			fmt.Println("\nFirst 3 errors:")
			for i, err := range report.Errors {
				if i >= 3 {
					break
				}
				fmt.Printf("  %d. [%s] %s\n", i+1, err.Code, err.Message)
				if err.Location != nil && err.Location.Path != "" {
					fmt.Printf("     Location: %s", err.Location.Path)
					if err.Location.Line > 0 {
						fmt.Printf(":%d", err.Location.Line)
					}
					fmt.Println()
				}
			}
		}
	}

	fmt.Println()
}

func step2GenerateReports(filePath string) (*ebmlib.ValidationReport, error) {
	fmt.Println("STEP 2: Generate Reports")
	fmt.Println("-------------------------")

	var report *ebmlib.ValidationReport
	var err error

	ext := filepath.Ext(filePath)
	switch ext {
	case ".epub":
		report, err = ebmlib.ValidateEPUB(filePath)
	case ".pdf":
		report, err = ebmlib.ValidatePDF(filePath)
	}

	if err != nil {
		return nil, err
	}

	baseDir := filepath.Dir(filePath)
	baseName := filepath.Base(filePath)
	nameWithoutExt := baseName[:len(baseName)-len(filepath.Ext(baseName))]

	jsonPath := filepath.Join(baseDir, nameWithoutExt+"_validation.json")
	err = ebmlib.WriteReportToFile(report, jsonPath, &ebmlib.ReportOptions{
		Format:          ebmlib.FormatJSON,
		IncludeWarnings: true,
		IncludeInfo:     true,
		Verbose:         true,
	})
	if err != nil {
		log.Printf("Warning: Failed to write JSON report: %v", err)
	} else {
		fmt.Printf("✓ JSON report: %s\n", jsonPath)
	}

	textPath := filepath.Join(baseDir, nameWithoutExt+"_validation.txt")
	err = ebmlib.WriteReportToFile(report, textPath, &ebmlib.ReportOptions{
		Format:          ebmlib.FormatText,
		IncludeWarnings: true,
		IncludeInfo:     true,
		Verbose:         false,
		ColorEnabled:    false,
	})
	if err != nil {
		log.Printf("Warning: Failed to write text report: %v", err)
	} else {
		fmt.Printf("✓ Text report: %s\n", textPath)
	}

	mdPath := filepath.Join(baseDir, nameWithoutExt+"_validation.md")
	err = ebmlib.WriteReportToFile(report, mdPath, &ebmlib.ReportOptions{
		Format:          ebmlib.FormatMarkdown,
		IncludeWarnings: true,
		IncludeInfo:     true,
		Verbose:         true,
	})
	if err != nil {
		log.Printf("Warning: Failed to write Markdown report: %v", err)
	} else {
		fmt.Printf("✓ Markdown report: %s\n", mdPath)
	}

	fmt.Println()
	return report, nil
}

func step3PreviewRepairs(filePath string) {
	fmt.Println("STEP 3: Preview Repair Actions")
	fmt.Println("-------------------------------")

	var preview *ebmlib.RepairPreview
	var err error

	ext := filepath.Ext(filePath)
	switch ext {
	case ".epub":
		preview, err = ebmlib.PreviewEPUBRepair(filePath)
	case ".pdf":
		preview, err = ebmlib.PreviewPDFRepair(filePath)
	}

	if err != nil {
		log.Printf("Preview failed: %v\n", err)
		return
	}

	fmt.Printf("Can auto-repair: %v\n", preview.CanAutoRepair)
	fmt.Printf("Backup required: %v\n", preview.BackupRequired)
	fmt.Printf("Estimated time: %dms\n", preview.EstimatedTime)
	fmt.Printf("Total actions: %d\n\n", len(preview.Actions))

	if len(preview.Actions) == 0 {
		fmt.Println("No repair actions available")
		return
	}

	automatedCount := 0
	manualCount := 0

	fmt.Println("Repair actions:")
	for i, action := range preview.Actions {
		fmt.Printf("%d. %s\n", i+1, action.Description)
		fmt.Printf("   Type: %s\n", action.Type)
		fmt.Printf("   Target: %s\n", action.Target)
		fmt.Printf("   Automated: %v\n", action.Automated)

		if action.Automated {
			automatedCount++
		} else {
			manualCount++
		}

		if len(action.Details) > 0 {
			fmt.Printf("   Details: %+v\n", action.Details)
		}
		fmt.Println()
	}

	fmt.Printf("Summary: %d automated, %d manual\n", automatedCount, manualCount)

	if len(preview.Warnings) > 0 {
		fmt.Println("\nWarnings:")
		for _, warning := range preview.Warnings {
			fmt.Printf("  ⚠ %s\n", warning)
		}
	}

	fmt.Println()
}

func step4ApplyRepairs(filePath string) {
	fmt.Println("STEP 4: Apply Repairs")
	fmt.Println("----------------------")

	var result *ebmlib.RepairResult
	var err error

	ext := filepath.Ext(filePath)
	switch ext {
	case ".epub":
		result, err = ebmlib.RepairEPUB(filePath)
	case ".pdf":
		result, err = ebmlib.RepairPDF(filePath)
	}

	if err != nil {
		log.Printf("Repair failed: %v\n", err)
		return
	}

	if result.Success {
		fmt.Println("✓ Repair successful!")
		fmt.Printf("  Output file: %s\n", result.BackupPath)
		fmt.Printf("  Actions applied: %d\n", len(result.ActionsApplied))

		if len(result.ActionsApplied) > 0 {
			fmt.Println("\nActions applied:")
			for i, action := range result.ActionsApplied {
				fmt.Printf("  %d. %s\n", i+1, action.Description)
			}
		}
	} else {
		fmt.Println("✗ Repair failed")
		if result.Error != nil {
			fmt.Printf("  Error: %v\n", result.Error)
		}
	}

	fmt.Println()
}

func step5ValidateRepaired(filePath string) {
	fmt.Println("STEP 5: Validate Repaired File")
	fmt.Println("-------------------------------")

	repairedPath := getRepairedPath(filePath)

	if _, err := os.Stat(repairedPath); os.IsNotExist(err) {
		fmt.Println("Repaired file not found, skipping validation")
		return
	}

	var report *ebmlib.ValidationReport
	var err error

	ext := filepath.Ext(repairedPath)
	switch ext {
	case ".epub":
		report, err = ebmlib.ValidateEPUB(repairedPath)
	case ".pdf":
		report, err = ebmlib.ValidatePDF(repairedPath)
	}

	if err != nil {
		log.Printf("Validation of repaired file failed: %v\n", err)
		return
	}

	fmt.Printf("Validation result: ")
	if report.IsValid {
		fmt.Println("✓ VALID")
		fmt.Println("The repaired file is now compliant!")
	} else {
		fmt.Println("✗ INVALID")
		fmt.Printf("The repaired file still has %d errors\n", report.ErrorCount())

		if report.ErrorCount() > 0 {
			fmt.Println("\nRemaining errors:")
			for i, err := range report.Errors {
				if i >= 3 {
					fmt.Printf("  ... and %d more\n", report.ErrorCount()-3)
					break
				}
				fmt.Printf("  %d. [%s] %s\n", i+1, err.Code, err.Message)
			}
		}
	}

	fmt.Println()
}

func getRepairedPath(originalPath string) string {
	ext := filepath.Ext(originalPath)
	base := originalPath[:len(originalPath)-len(ext)]
	return base + "_repaired" + ext
}
