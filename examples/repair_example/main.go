// Package main provides an example program for EBMLib.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/example/project/pkg/ebmlib"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: repair_example <file.epub|file.pdf>")
		os.Exit(1)
	}

	filePath := os.Args[1]
	fileExt := filePath[len(filePath)-4:]

	fmt.Printf("=== Repairing %s ===\n\n", filePath)

	switch fileExt {
	case "epub":
		repairEPUB(filePath)
	case ".pdf":
		repairPDF(filePath)
	default:
		log.Fatalf("Unsupported file type: %s (expected .epub or .pdf)", fileExt)
	}
}

func repairEPUB(filePath string) {
	fmt.Println("Step 1: Validating EPUB...")
	report, err := ebmlib.ValidateEPUB(filePath)
	if err != nil {
		log.Fatalf("Validation failed: %v", err)
	}

	if report.IsValid {
		fmt.Println("✓ File is already valid, no repair needed")
		return
	}

	fmt.Printf("Found %d errors\n\n", report.ErrorCount())

	fmt.Println("Step 2: Previewing repair actions...")
	preview, err := ebmlib.PreviewEPUBRepair(filePath)
	if err != nil {
		log.Fatalf("Preview failed: %v", err)
	}

	fmt.Printf("Repair preview:\n")
	fmt.Printf("  Can auto-repair: %v\n", preview.CanAutoRepair)
	fmt.Printf("  Backup required: %v\n", preview.BackupRequired)
	fmt.Printf("  Actions: %d\n\n", len(preview.Actions))

	for i, action := range preview.Actions {
		fmt.Printf("  %d. %s\n", i+1, action.Description)
		fmt.Printf("     Type: %s\n", action.Type)
		fmt.Printf("     Target: %s\n", action.Target)
		fmt.Printf("     Automated: %v\n", action.Automated)
		if len(action.Details) > 0 {
			fmt.Printf("     Details: %+v\n", action.Details)
		}
		fmt.Println()
	}

	if len(preview.Warnings) > 0 {
		fmt.Println("Warnings:")
		for _, warning := range preview.Warnings {
			fmt.Printf("  ⚠ %s\n", warning)
		}
		fmt.Println()
	}

	fmt.Println("Step 3: Applying repairs...")
	result, err := ebmlib.RepairEPUB(filePath)
	if err != nil {
		log.Fatalf("Repair failed: %v", err)
	}

	if result.Success {
		fmt.Printf("✓ Repair successful!\n")
		fmt.Printf("  Output file: %s\n", result.BackupPath)
		fmt.Printf("  Actions applied: %d\n", len(result.ActionsApplied))

		for i, action := range result.ActionsApplied {
			fmt.Printf("    %d. %s\n", i+1, action.Description)
		}
	} else {
		fmt.Printf("✗ Repair failed: %v\n", result.Error)
		os.Exit(1)
	}
}

func repairPDF(filePath string) {
	fmt.Println("Step 1: Validating PDF...")
	report, err := ebmlib.ValidatePDF(filePath)
	if err != nil {
		log.Fatalf("Validation failed: %v", err)
	}

	if report.IsValid {
		fmt.Println("✓ File is already valid, no repair needed")
		return
	}

	fmt.Printf("Found %d errors\n\n", report.ErrorCount())

	fmt.Println("Step 2: Previewing repair actions...")
	preview, err := ebmlib.PreviewPDFRepair(filePath)
	if err != nil {
		log.Fatalf("Preview failed: %v", err)
	}

	fmt.Printf("Repair preview:\n")
	fmt.Printf("  Can auto-repair: %v\n", preview.CanAutoRepair)
	fmt.Printf("  Backup required: %v\n", preview.BackupRequired)
	fmt.Printf("  Actions: %d\n\n", len(preview.Actions))

	for i, action := range preview.Actions {
		fmt.Printf("  %d. %s\n", i+1, action.Description)
		fmt.Printf("     Type: %s\n", action.Type)
		fmt.Printf("     Target: %s\n", action.Target)
		fmt.Printf("     Automated: %v\n", action.Automated)
		fmt.Println()
	}

	if len(preview.Warnings) > 0 {
		fmt.Println("Warnings:")
		for _, warning := range preview.Warnings {
			fmt.Printf("  ⚠ %s\n", warning)
		}
		fmt.Println()
	}

	if !preview.CanAutoRepair {
		fmt.Println("⚠ Some repairs require manual intervention")
		fmt.Println("Only automated repairs will be applied")
	}

	fmt.Println("Step 3: Applying automated repairs...")
	result, err := ebmlib.RepairPDF(filePath)
	if err != nil {
		log.Fatalf("Repair failed: %v", err)
	}

	if result.Success {
		fmt.Printf("✓ Repair successful!\n")
		fmt.Printf("  Output file: %s\n", result.BackupPath)
		fmt.Printf("  Actions applied: %d\n", len(result.ActionsApplied))

		for i, action := range result.ActionsApplied {
			fmt.Printf("    %d. %s\n", i+1, action.Description)
		}
	} else {
		fmt.Printf("✗ Repair failed: %v\n", result.Error)
		os.Exit(1)
	}
}
