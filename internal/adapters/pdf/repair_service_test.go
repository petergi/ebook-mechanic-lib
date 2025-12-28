package pdf

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/example/project/internal/domain"
	"github.com/example/project/internal/ports"
)

func TestNewRepairService(t *testing.T) {
	service := NewRepairService()
	if service == nil {
		t.Fatal("NewRepairService returned nil")
	}

	impl, ok := service.(*RepairServiceImpl)
	if !ok {
		t.Fatal("NewRepairService did not return *RepairServiceImpl")
	}

	if impl.validator == nil {
		t.Error("validator is nil")
	}
}

func TestPreview_EmptyReport(t *testing.T) {
	service := NewRepairService()
	ctx := context.Background()

	report := &domain.ValidationReport{
		FilePath:       "test.pdf",
		FileType:       "PDF",
		IsValid:        true,
		Errors:         make([]domain.ValidationError, 0),
		ValidationTime: time.Now(),
	}

	preview, err := service.Preview(ctx, report)
	if err != nil {
		t.Fatalf("Preview failed: %v", err)
	}

	if preview == nil {
		t.Fatal("Preview returned nil")
	}

	if len(preview.Actions) != 0 {
		t.Errorf("Expected 0 actions, got %d", len(preview.Actions))
	}

	if preview.BackupRequired {
		t.Error("Expected BackupRequired to be false for empty report")
	}

	if !preview.CanAutoRepair {
		t.Error("Expected CanAutoRepair to be true for empty report")
	}
}

func TestPreview_NilReport(t *testing.T) {
	service := NewRepairService()
	ctx := context.Background()

	_, err := service.Preview(ctx, nil)
	if err == nil {
		t.Error("Expected error for nil report")
	}
}

func TestPreview_MissingEOF(t *testing.T) {
	service := NewRepairService()
	ctx := context.Background()

	report := &domain.ValidationReport{
		FilePath: "test.pdf",
		FileType: "PDF",
		IsValid:  false,
		Errors: []domain.ValidationError{
			{
				Code:    ErrorCodePDFTrailer003,
				Message: "Missing %%EOF marker",
				Location: &domain.ErrorLocation{
					Path: "trailer",
				},
				Details: map[string]interface{}{
					"expected": "%%EOF at end of file",
				},
			},
		},
		ValidationTime: time.Now(),
	}

	preview, err := service.Preview(ctx, report)
	if err != nil {
		t.Fatalf("Preview failed: %v", err)
	}

	if len(preview.Actions) != 1 {
		t.Fatalf("Expected 1 action, got %d", len(preview.Actions))
	}

	action := preview.Actions[0]
	if action.Type != "append_eof_marker" {
		t.Errorf("Expected action type 'append_eof_marker', got '%s'", action.Type)
	}

	if !action.Automated {
		t.Error("Expected EOF marker append to be automated")
	}

	if !preview.CanAutoRepair {
		t.Error("Expected CanAutoRepair to be true for automated repairs")
	}

	if !preview.BackupRequired {
		t.Error("Expected BackupRequired to be true")
	}
}

func TestPreview_InvalidStartxref(t *testing.T) {
	service := NewRepairService()
	ctx := context.Background()

	report := &domain.ValidationReport{
		FilePath: "test.pdf",
		FileType: "PDF",
		IsValid:  false,
		Errors: []domain.ValidationError{
			{
				Code:    ErrorCodePDFTrailer001,
				Message: "Invalid or missing startxref",
				Location: &domain.ErrorLocation{
					Path: "trailer",
				},
				Details: map[string]interface{}{
					"expected": "startxref <offset> before %%EOF",
				},
			},
		},
		ValidationTime: time.Now(),
	}

	preview, err := service.Preview(ctx, report)
	if err != nil {
		t.Fatalf("Preview failed: %v", err)
	}

	if len(preview.Actions) != 1 {
		t.Fatalf("Expected 1 action, got %d", len(preview.Actions))
	}

	action := preview.Actions[0]
	if action.Type != "recompute_startxref" {
		t.Errorf("Expected action type 'recompute_startxref', got '%s'", action.Type)
	}

	if !action.Automated {
		t.Error("Expected startxref recomputation to be automated")
	}
}

func TestPreview_TrailerTypos(t *testing.T) {
	service := NewRepairService()
	ctx := context.Background()

	report := &domain.ValidationReport{
		FilePath: "test.pdf",
		FileType: "PDF",
		IsValid:  false,
		Errors: []domain.ValidationError{
			{
				Code:    ErrorCodePDFTrailer002,
				Message: "Invalid trailer dictionary",
				Location: &domain.ErrorLocation{
					Path: "trailer",
				},
			},
		},
		ValidationTime: time.Now(),
	}

	preview, err := service.Preview(ctx, report)
	if err != nil {
		t.Fatalf("Preview failed: %v", err)
	}

	if len(preview.Actions) != 1 {
		t.Fatalf("Expected 1 action, got %d", len(preview.Actions))
	}

	action := preview.Actions[0]
	if action.Type != "fix_trailer_typos" {
		t.Errorf("Expected action type 'fix_trailer_typos', got '%s'", action.Type)
	}

	if !action.Automated {
		t.Error("Expected trailer typo fix to be automated")
	}
}

func TestPreview_UnsafeRepairs(t *testing.T) {
	tests := []struct {
		name          string
		errorCode     string
		expectedType  string
		shouldAutomate bool
	}{
		{
			name:          "Invalid Header",
			errorCode:     ErrorCodePDFHeader001,
			expectedType:  "manual_header_fix",
			shouldAutomate: false,
		},
		{
			name:          "Invalid Version",
			errorCode:     ErrorCodePDFHeader002,
			expectedType:  "manual_header_fix",
			shouldAutomate: false,
		},
		{
			name:          "Damaged Xref",
			errorCode:     ErrorCodePDFXref001,
			expectedType:  "manual_xref_rebuild",
			shouldAutomate: false,
		},
		{
			name:          "Missing Catalog",
			errorCode:     ErrorCodePDFCatalog001,
			expectedType:  "manual_catalog_fix",
			shouldAutomate: false,
		},
	}

	service := NewRepairService()
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := &domain.ValidationReport{
				FilePath: "test.pdf",
				FileType: "PDF",
				IsValid:  false,
				Errors: []domain.ValidationError{
					{
						Code:    tt.errorCode,
						Message: "Test error",
						Details: map[string]interface{}{},
					},
				},
				ValidationTime: time.Now(),
			}

			preview, err := service.Preview(ctx, report)
			if err != nil {
				t.Fatalf("Preview failed: %v", err)
			}

			if len(preview.Actions) != 1 {
				t.Fatalf("Expected 1 action, got %d", len(preview.Actions))
			}

			action := preview.Actions[0]
			if action.Type != tt.expectedType {
				t.Errorf("Expected action type '%s', got '%s'", tt.expectedType, action.Type)
			}

			if action.Automated != tt.shouldAutomate {
				t.Errorf("Expected automated=%v, got %v", tt.shouldAutomate, action.Automated)
			}

			if preview.CanAutoRepair {
				t.Error("Expected CanAutoRepair to be false for unsafe repairs")
			}

			if len(preview.Warnings) == 0 {
				t.Error("Expected warnings for unsafe repairs")
			}
		})
	}
}

func TestPreview_MultipleErrors(t *testing.T) {
	service := NewRepairService()
	ctx := context.Background()

	report := &domain.ValidationReport{
		FilePath: "test.pdf",
		FileType: "PDF",
		IsValid:  false,
		Errors: []domain.ValidationError{
			{
				Code:    ErrorCodePDFTrailer003,
				Message: "Missing EOF",
			},
			{
				Code:    ErrorCodePDFTrailer001,
				Message: "Invalid startxref",
			},
			{
				Code:    ErrorCodePDFHeader001,
				Message: "Invalid header",
			},
		},
		ValidationTime: time.Now(),
	}

	preview, err := service.Preview(ctx, report)
	if err != nil {
		t.Fatalf("Preview failed: %v", err)
	}

	if len(preview.Actions) != 3 {
		t.Fatalf("Expected 3 actions, got %d", len(preview.Actions))
	}

	if preview.CanAutoRepair {
		t.Error("Expected CanAutoRepair to be false when unsafe repairs present")
	}
}

func TestCanRepair(t *testing.T) {
	service := NewRepairService()
	ctx := context.Background()

	repairableCodes := []string{
		ErrorCodePDFTrailer003,
		ErrorCodePDFTrailer001,
	}

	for _, code := range repairableCodes {
		err := &domain.ValidationError{
			Code:    code,
			Message: "Test error",
		}

		if !service.CanRepair(ctx, err) {
			t.Errorf("Expected code %s to be repairable", code)
		}
	}

	nonRepairableCodes := []string{
		ErrorCodePDFHeader001,
		ErrorCodePDFHeader002,
		ErrorCodePDFXref001,
		ErrorCodePDFCatalog001,
	}

	for _, code := range nonRepairableCodes {
		err := &domain.ValidationError{
			Code:    code,
			Message: "Test error",
		}

		if service.CanRepair(ctx, err) {
			t.Errorf("Expected code %s to not be automatically repairable", code)
		}
	}

	if service.CanRepair(ctx, nil) {
		t.Error("Expected nil error to not be repairable")
	}
}

func TestApply_NoActions(t *testing.T) {
	service := NewRepairService()
	ctx := context.Background()

	preview := &ports.RepairPreview{
		Actions:        make([]ports.RepairAction, 0),
		CanAutoRepair:  true,
		BackupRequired: false,
	}

	result, err := service.Apply(ctx, "test.pdf", preview)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	if result.Success {
		t.Error("Expected success to be false when no actions to apply")
	}

	if result.Error == nil {
		t.Error("Expected error when no actions to apply")
	}
}

func TestApply_AppendEOFMarker(t *testing.T) {
	service := NewRepairService()
	ctx := context.Background()

	tempDir := t.TempDir()
	testPDF := filepath.Join(tempDir, "test.pdf")

	pdfContent := createMinimalPDFWithoutEOF()
	if err := os.WriteFile(testPDF, pdfContent, 0644); err != nil {
		t.Fatalf("Failed to create test PDF: %v", err)
	}

	preview := &ports.RepairPreview{
		Actions: []ports.RepairAction{
			{
				Type:        "append_eof_marker",
				Description: "Append EOF marker",
				Target:      "trailer",
				Automated:   true,
			},
		},
		CanAutoRepair:  true,
		BackupRequired: true,
	}

	result, err := service.Apply(ctx, testPDF, preview)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected success, got error: %v", result.Error)
	}

	if len(result.ActionsApplied) != 1 {
		t.Errorf("Expected 1 action applied, got %d", len(result.ActionsApplied))
	}

	if result.BackupPath == "" {
		t.Error("Expected backup path to be set")
	}

	repairedData, err := os.ReadFile(result.BackupPath)
	if err != nil {
		t.Fatalf("Failed to read repaired file: %v", err)
	}

	if !bytes.Contains(repairedData, []byte("%%EOF")) {
		t.Error("EOF marker was not appended")
	}

	lastBytes := repairedData[len(repairedData)-10:]
	if !bytes.Contains(lastBytes, []byte("%%EOF")) {
		t.Error("EOF marker not at end of file")
	}
}

func TestApply_RecomputeStartxref(t *testing.T) {
	service := NewRepairService()
	ctx := context.Background()

	tempDir := t.TempDir()
	testPDF := filepath.Join(tempDir, "test.pdf")

	pdfContent := createMinimalPDFWithBadStartxref()
	if err := os.WriteFile(testPDF, pdfContent, 0644); err != nil {
		t.Fatalf("Failed to create test PDF: %v", err)
	}

	preview := &ports.RepairPreview{
		Actions: []ports.RepairAction{
			{
				Type:        "recompute_startxref",
				Description: "Recompute startxref",
				Target:      "trailer",
				Automated:   true,
			},
		},
		CanAutoRepair:  true,
		BackupRequired: true,
	}

	result, err := service.Apply(ctx, testPDF, preview)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected success, got error: %v", result.Error)
	}

	repairedData, err := os.ReadFile(result.BackupPath)
	if err != nil {
		t.Fatalf("Failed to read repaired file: %v", err)
	}

	if !strings.Contains(string(repairedData), "startxref") {
		t.Error("startxref keyword not found in repaired file")
	}
}

func TestApply_FixTrailerTypos(t *testing.T) {
	service := NewRepairService()
	ctx := context.Background()

	tempDir := t.TempDir()
	testPDF := filepath.Join(tempDir, "test.pdf")

	pdfContent := createMinimalPDFWithTrailerTypos()
	if err := os.WriteFile(testPDF, pdfContent, 0644); err != nil {
		t.Fatalf("Failed to create test PDF: %v", err)
	}

	preview := &ports.RepairPreview{
		Actions: []ports.RepairAction{
			{
				Type:        "fix_trailer_typos",
				Description: "Fix trailer typos",
				Target:      "trailer",
				Automated:   true,
			},
		},
		CanAutoRepair:  true,
		BackupRequired: true,
	}

	result, err := service.Apply(ctx, testPDF, preview)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected success, got error: %v", result.Error)
	}

	repairedData, err := os.ReadFile(result.BackupPath)
	if err != nil {
		t.Fatalf("Failed to read repaired file: %v", err)
	}

	repairedStr := string(repairedData)
	if strings.Contains(repairedStr, "/Sise") {
		t.Error("/Sise typo was not fixed to /Size")
	}

	if strings.Contains(repairedStr, "/root") && !strings.Contains(repairedStr, "/Root") {
		t.Error("/root typo was not fixed to /Root")
	}
}

func TestApply_MultipleRepairs(t *testing.T) {
	service := NewRepairService()
	ctx := context.Background()

	tempDir := t.TempDir()
	testPDF := filepath.Join(tempDir, "test.pdf")

	pdfContent := createMinimalPDFWithMultipleIssues()
	if err := os.WriteFile(testPDF, pdfContent, 0644); err != nil {
		t.Fatalf("Failed to create test PDF: %v", err)
	}

	preview := &ports.RepairPreview{
		Actions: []ports.RepairAction{
			{
				Type:        "append_eof_marker",
				Description: "Append EOF marker",
				Target:      "trailer",
				Automated:   true,
			},
			{
				Type:        "recompute_startxref",
				Description: "Recompute startxref",
				Target:      "trailer",
				Automated:   true,
			},
		},
		CanAutoRepair:  true,
		BackupRequired: true,
	}

	result, err := service.Apply(ctx, testPDF, preview)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected success, got error: %v", result.Error)
	}

	if len(result.ActionsApplied) != 2 {
		t.Errorf("Expected 2 actions applied, got %d", len(result.ActionsApplied))
	}
}

func TestCreateBackup(t *testing.T) {
	service := NewRepairService()
	ctx := context.Background()

	tempDir := t.TempDir()
	sourcePath := filepath.Join(tempDir, "source.pdf")
	backupPath := filepath.Join(tempDir, "backup.pdf")

	content := []byte("%PDF-1.7\ntest content\n%%EOF")
	if err := os.WriteFile(sourcePath, content, 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	if err := service.CreateBackup(ctx, sourcePath, backupPath); err != nil {
		t.Fatalf("CreateBackup failed: %v", err)
	}

	backupContent, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("Failed to read backup file: %v", err)
	}

	if !bytes.Equal(content, backupContent) {
		t.Error("Backup content does not match source")
	}
}

func TestRestoreBackup(t *testing.T) {
	service := NewRepairService()
	ctx := context.Background()

	tempDir := t.TempDir()
	backupPath := filepath.Join(tempDir, "backup.pdf")
	originalPath := filepath.Join(tempDir, "original.pdf")

	backupContent := []byte("%PDF-1.7\nbackup content\n%%EOF")
	if err := os.WriteFile(backupPath, backupContent, 0644); err != nil {
		t.Fatalf("Failed to create backup file: %v", err)
	}

	if err := service.RestoreBackup(ctx, backupPath, originalPath); err != nil {
		t.Fatalf("RestoreBackup failed: %v", err)
	}

	restoredContent, err := os.ReadFile(originalPath)
	if err != nil {
		t.Fatalf("Failed to read restored file: %v", err)
	}

	if !bytes.Equal(backupContent, restoredContent) {
		t.Error("Restored content does not match backup")
	}
}

func TestGenerateOutputPath(t *testing.T) {
	service := &RepairServiceImpl{}

	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "test.pdf",
			expected: "test_repaired.pdf",
		},
		{
			input:    "/path/to/document.pdf",
			expected: "/path/to/document_repaired.pdf",
		},
		{
			input:    "my.file.pdf",
			expected: "my.file_repaired.pdf",
		},
	}

	for _, tt := range tests {
		result := service.generateOutputPath(tt.input)
		if result != tt.expected {
			t.Errorf("For input %s, expected %s, got %s", tt.input, tt.expected, result)
		}
	}
}

func TestAppendEOFMarker(t *testing.T) {
	service := &RepairServiceImpl{}

	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{
			name:     "No EOF marker",
			input:    []byte("%PDF-1.7\nstartxref\n123"),
			expected: "%%EOF",
		},
		{
			name:     "Already has EOF",
			input:    []byte("%PDF-1.7\nstartxref\n123\n%%EOF\n"),
			expected: "%%EOF",
		},
		{
			name:     "No trailing newline",
			input:    []byte("%PDF-1.7\nstartxref\n123"),
			expected: "\n%%EOF\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &repairContext{
				data: tt.input,
			}

			err := service.appendEOFMarker(ctx)
			if err != nil {
				t.Fatalf("appendEOFMarker failed: %v", err)
			}

			if !strings.Contains(string(ctx.data), tt.expected) {
				t.Errorf("Expected output to contain %q, got %q", tt.expected, string(ctx.data))
			}
		})
	}
}

func TestRecomputeStartxref(t *testing.T) {
	service := &RepairServiceImpl{}

	pdfContent := []byte(`%PDF-1.7
1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj
2 0 obj
<< /Type /Pages /Count 0 /Kids [] >>
endobj
xref
0 3
0000000000 65535 f 
0000000009 00000 n 
0000000058 00000 n 
trailer
<< /Size 3 /Root 1 0 R >>
startxref
9999999
%%EOF`)

	ctx := &repairContext{
		data: pdfContent,
	}

	err := service.recomputeStartxref(ctx)
	if err != nil {
		t.Fatalf("recomputeStartxref failed: %v", err)
	}

	if strings.Contains(string(ctx.data), "9999999") {
		t.Error("Old incorrect startxref value still present")
	}

	if !strings.Contains(string(ctx.data), "startxref") {
		t.Error("startxref keyword not found")
	}
}

func TestFixTrailerTypos(t *testing.T) {
	service := &RepairServiceImpl{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Fix /Sise typo",
			input:    "trailer << /Sise 10 /Root 1 0 R >>",
			expected: "/Size 10",
		},
		{
			name:     "Fix /root typo",
			input:    "trailer << /Size 10 /root 1 0 R >>",
			expected: "/Root 1 0 R",
		},
		{
			name:     "No typos",
			input:    "trailer << /Size 10 /Root 1 0 R >>",
			expected: "trailer <<",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &repairContext{
				data: []byte(tt.input),
			}

			err := service.fixTrailerTypos(ctx)
			if err != nil {
				t.Fatalf("fixTrailerTypos failed: %v", err)
			}

			if !strings.Contains(string(ctx.data), tt.expected) {
				t.Errorf("Expected output to contain %q, got %q", tt.expected, string(ctx.data))
			}
		})
	}
}

func createMinimalPDFWithoutEOF() []byte {
	return []byte(`%PDF-1.7
1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj
2 0 obj
<< /Type /Pages /Count 0 /Kids [] >>
endobj
xref
0 3
0000000000 65535 f 
0000000009 00000 n 
0000000058 00000 n 
trailer
<< /Size 3 /Root 1 0 R >>
startxref
107
`)
}

func createMinimalPDFWithBadStartxref() []byte {
	return []byte(`%PDF-1.7
1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj
2 0 obj
<< /Type /Pages /Count 0 /Kids [] >>
endobj
xref
0 3
0000000000 65535 f 
0000000009 00000 n 
0000000058 00000 n 
trailer
<< /Size 3 /Root 1 0 R >>
startxref
999999
%%EOF
`)
}

func createMinimalPDFWithTrailerTypos() []byte {
	return []byte(`%PDF-1.7
1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj
2 0 obj
<< /Type /Pages /Count 0 /Kids [] >>
endobj
xref
0 3
0000000000 65535 f 
0000000009 00000 n 
0000000058 00000 n 
trailer
<< /Sise 3 /root 1 0 R >>
startxref
107
%%EOF
`)
}

func createMinimalPDFWithMultipleIssues() []byte {
	return []byte(`%PDF-1.7
1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj
2 0 obj
<< /Type /Pages /Count 0 /Kids [] >>
endobj
xref
0 3
0000000000 65535 f 
0000000009 00000 n 
0000000058 00000 n 
trailer
<< /Size 3 /Root 1 0 R >>
startxref
999999
`)
}
