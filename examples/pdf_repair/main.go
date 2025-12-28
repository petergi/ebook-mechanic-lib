package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/example/project/internal/adapters/pdf"
	"github.com/example/project/internal/domain"
	"github.com/example/project/internal/ports"
)

func main() {
	fmt.Println("PDF Repair Service Example")
	fmt.Println("===========================\n")

	// Example 1: Basic repair workflow
	example1_BasicRepairWorkflow()

	fmt.Println("\n==================================================\n")

	// Example 2: Handling unsafe repairs
	example2_UnsafeRepairs()

	fmt.Println("\n==================================================\n")

	// Example 3: Batch repair
	example3_BatchRepair()
}

func example1_BasicRepairWorkflow() {
	fmt.Println("Example 1: Basic Repair Workflow")
	fmt.Println("---------------------------------")

	ctx := context.Background()
	repairService := pdf.NewRepairService()

	// Simulate a validation report with missing EOF
	report := &domain.ValidationReport{
		FilePath: "document.pdf",
		FileType: "PDF",
		IsValid:  false,
		Errors: []domain.ValidationError{
			{
				Code:    pdf.ErrorCodePDFTrailer003,
				Message: "Missing %%EOF marker",
				Details: map[string]interface{}{
					"expected": "%%EOF at end of file",
				},
			},
		},
		ValidationTime: time.Now(),
	}

	fmt.Printf("Validating: %s\n", report.FilePath)
	fmt.Printf("Errors found: %d\n\n", len(report.Errors))

	// Preview repairs
	preview, err := repairService.Preview(ctx, report)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Repair Preview:")
	fmt.Printf("  Total Actions: %d\n", len(preview.Actions))
	fmt.Printf("  Can Auto-Repair: %v\n", preview.CanAutoRepair)
	fmt.Printf("  Backup Required: %v\n", preview.BackupRequired)
	fmt.Printf("  Estimated Time: %dms\n\n", preview.EstimatedTime)

	fmt.Println("Proposed Actions:")
	for i, action := range preview.Actions {
		fmt.Printf("  %d. %s\n", i+1, action.Description)
		fmt.Printf("     Type: %s\n", action.Type)
		fmt.Printf("     Target: %s\n", action.Target)
		fmt.Printf("     Automated: %v\n", action.Automated)
		if len(action.Details) > 0 {
			fmt.Printf("     Details: %v\n", action.Details)
		}
		fmt.Println()
	}

	// Note: In real usage, you would call Apply here
	// result, err := repairService.Apply(ctx, "document.pdf", preview)
	fmt.Println("✓ Preview completed successfully")
}

func example2_UnsafeRepairs() {
	fmt.Println("Example 2: Handling Unsafe Repairs")
	fmt.Println("-----------------------------------")

	ctx := context.Background()
	repairService := pdf.NewRepairService()

	// Simulate a validation report with unsafe errors
	report := &domain.ValidationReport{
		FilePath: "corrupted.pdf",
		FileType: "PDF",
		IsValid:  false,
		Errors: []domain.ValidationError{
			{
				Code:    pdf.ErrorCodePDFTrailer003,
				Message: "Missing %%EOF marker",
			},
			{
				Code:    pdf.ErrorCodePDFHeader001,
				Message: "Invalid or missing PDF header",
			},
			{
				Code:    pdf.ErrorCodePDFXref001,
				Message: "Invalid or damaged cross-reference table",
			},
		},
		ValidationTime: time.Now(),
	}

	fmt.Printf("Validating: %s\n", report.FilePath)
	fmt.Printf("Errors found: %d\n\n", len(report.Errors))

	preview, err := repairService.Preview(ctx, report)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Can Auto-Repair: %v\n\n", preview.CanAutoRepair)

	if !preview.CanAutoRepair {
		fmt.Println("⚠ Manual intervention required!\n")
		fmt.Println("Warnings:")
		for _, warning := range preview.Warnings {
			fmt.Printf("  - %s\n", warning)
		}
		fmt.Println()

		fmt.Println("Repair Actions:")
		automatedCount := 0
		manualCount := 0

		for i, action := range preview.Actions {
			fmt.Printf("  %d. %s\n", i+1, action.Description)
			fmt.Printf("     Automated: %v\n", action.Automated)

			if !action.Automated {
				manualCount++
				if reason, ok := action.Details["reason"].(string); ok {
					fmt.Printf("     Reason: %s\n", reason)
				}
			} else {
				automatedCount++
			}
			fmt.Println()
		}

		fmt.Printf("Summary: %d automated, %d manual\n\n", automatedCount, manualCount)

		fmt.Println("Recommended Tools for Manual Repairs:")
		fmt.Println("  - qpdf: For structural repairs and xref rebuild")
		fmt.Println("  - Ghostscript: For rewriting PDFs")
		fmt.Println("  - Adobe Acrobat: For complex repairs")
		fmt.Println("  - MuPDF mutool: For cleaning and repair")
	}
}

func example3_BatchRepair() {
	fmt.Println("Example 3: Batch Repair with Rollback")
	fmt.Println("--------------------------------------")

	ctx := context.Background()
	repairService := pdf.NewRepairService()

	// Simulate multiple files with various issues
	files := []struct {
		path   string
		report *domain.ValidationReport
	}{
		{
			path: "doc1.pdf",
			report: &domain.ValidationReport{
				FilePath: "doc1.pdf",
				IsValid:  false,
				Errors: []domain.ValidationError{
					{Code: pdf.ErrorCodePDFTrailer003, Message: "Missing EOF"},
				},
			},
		},
		{
			path: "doc2.pdf",
			report: &domain.ValidationReport{
				FilePath: "doc2.pdf",
				IsValid:  false,
				Errors: []domain.ValidationError{
					{Code: pdf.ErrorCodePDFTrailer001, Message: "Invalid startxref"},
				},
			},
		},
		{
			path: "doc3.pdf",
			report: &domain.ValidationReport{
				FilePath: "doc3.pdf",
				IsValid:  false,
				Errors: []domain.ValidationError{
					{Code: pdf.ErrorCodePDFHeader001, Message: "Invalid header"},
				},
			},
		},
	}

	fmt.Printf("Processing %d files...\n\n", len(files))

	repairedCount := 0
	skippedCount := 0
	failedCount := 0

	for _, file := range files {
		fmt.Printf("Processing: %s\n", file.path)

		if file.report.IsValid {
			fmt.Println("  ✓ Valid - skipping\n")
			skippedCount++
			continue
		}

		preview, err := repairService.Preview(ctx, file.report)
		if err != nil {
			fmt.Printf("  ✗ Preview failed: %v\n\n", err)
			failedCount++
			continue
		}

		if !preview.CanAutoRepair {
			fmt.Println("  ⚠ Manual intervention required - skipping")
			fmt.Printf("    Reason: %s\n\n", preview.Warnings[0])
			skippedCount++
			continue
		}

		fmt.Printf("  → Applying %d repair(s)...\n", len(preview.Actions))

		// Note: In real usage, you would:
		// 1. Create explicit backup
		// 2. Apply repairs
		// 3. Validate result
		// 4. Rollback if needed

		/*
		   backupPath := file.path + ".backup"
		   if err := repairService.CreateBackup(ctx, file.path, backupPath); err != nil {
		       fmt.Printf("  ✗ Backup failed: %v\n\n", err)
		       failedCount++
		       continue
		   }

		   result, err := repairService.Apply(ctx, file.path, preview)
		   if err != nil || !result.Success {
		       fmt.Printf("  ✗ Repair failed: %v\n", result.Error)
		       fmt.Println("  → Restoring backup...")
		       repairService.RestoreBackup(ctx, backupPath, file.path)
		       failedCount++
		       continue
		   }
		*/

		fmt.Println("  ✓ Repaired successfully\n")
		repairedCount++
	}

	fmt.Println("Batch Repair Summary:")
	fmt.Printf("  Repaired: %d\n", repairedCount)
	fmt.Printf("  Skipped: %d\n", skippedCount)
	fmt.Printf("  Failed: %d\n", failedCount)
	fmt.Printf("  Total: %d\n", len(files))
}

// Helper function to check if error is repairable
func checkRepairability(ctx context.Context, service ports.PDFRepairService, errors []domain.ValidationError) {
	fmt.Println("Checking repairability of errors:")

	for i, err := range errors {
		canRepair := service.CanRepair(ctx, &err)
		status := "✗ Not repairable"
		if canRepair {
			status = "✓ Repairable"
		}

		fmt.Printf("  %d. %s (%s) - %s\n", i+1, err.Code, err.Message, status)
	}
	fmt.Println()
}

// Helper function to display detailed action information
func displayActionDetails(action ports.RepairAction) {
	fmt.Printf("Action Details:\n")
	fmt.Printf("  Type: %s\n", action.Type)
	fmt.Printf("  Description: %s\n", action.Description)
	fmt.Printf("  Target: %s\n", action.Target)
	fmt.Printf("  Automated: %v\n", action.Automated)

	if len(action.Details) > 0 {
		fmt.Printf("  Details:\n")
		for k, v := range action.Details {
			fmt.Printf("    %s: %v\n", k, v)
		}
	}
}
