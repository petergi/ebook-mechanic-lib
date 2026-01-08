package epub

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/petergi/ebook-mechanic-lib/internal/domain"
	"github.com/petergi/ebook-mechanic-lib/internal/ports"
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

	if impl.containerValidator == nil {
		t.Error("containerValidator is nil")
	}
	if impl.opfValidator == nil {
		t.Error("opfValidator is nil")
	}
	if impl.navValidator == nil {
		t.Error("navValidator is nil")
	}
	if impl.contentValidator == nil {
		t.Error("contentValidator is nil")
	}
}

func TestPreview_EmptyReport(t *testing.T) {
	service := NewRepairService()
	ctx := context.Background()

	report := &domain.ValidationReport{
		FilePath:       "test.epub",
		FileType:       "EPUB",
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

func TestPreview_MimetypeInvalid(t *testing.T) {
	service := NewRepairService()
	ctx := context.Background()

	report := &domain.ValidationReport{
		FilePath: "test.epub",
		FileType: "EPUB",
		IsValid:  false,
		Errors: []domain.ValidationError{
			{
				Code:    ErrorCodeMimetypeInvalid,
				Message: "mimetype file must contain exactly 'application/epub+zip'",
				Location: &domain.ErrorLocation{
					Path: "mimetype",
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
	if action.Type != "fix_mimetype_content" {
		t.Errorf("Expected action type 'fix_mimetype_content', got '%s'", action.Type)
	}

	if !action.Automated {
		t.Error("Expected mimetype repair to be automated")
	}

	if !preview.CanAutoRepair {
		t.Error("Expected CanAutoRepair to be true for automated repairs")
	}

	if !preview.BackupRequired {
		t.Error("Expected BackupRequired to be true")
	}
}

func TestPreview_MimetypeNotFirst(t *testing.T) {
	service := NewRepairService()
	ctx := context.Background()

	report := &domain.ValidationReport{
		FilePath: "test.epub",
		FileType: "EPUB",
		IsValid:  false,
		Errors: []domain.ValidationError{
			{
				Code:    ErrorCodeMimetypeNotFirst,
				Message: "mimetype file must be first in ZIP archive",
				Location: &domain.ErrorLocation{
					Path: "mimetype",
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
	if action.Type != "fix_mimetype_order" {
		t.Errorf("Expected action type 'fix_mimetype_order', got '%s'", action.Type)
	}

	if !action.Automated {
		t.Error("Expected mimetype order repair to be automated")
	}
}

func TestPreview_OPFMissingNavDocument(t *testing.T) {
	service := NewRepairService()
	ctx := context.Background()

	report := &domain.ValidationReport{
		FilePath: "test.epub",
		FileType: "EPUB",
		IsValid:  false,
		Errors: []domain.ValidationError{
			{
				Code:    ErrorCodeOPFMissingNavDocument,
				Message: "OPF manifest must contain at least one item with properties='nav'",
				Location: &domain.ErrorLocation{
					Path: "OEBPS/content.opf",
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
	if action.Type != "add_nav_document" {
		t.Errorf("Expected action type 'add_nav_document', got '%s'", action.Type)
	}

	if !action.Automated {
		t.Error("Expected nav document repair to be automated")
	}
}

func TestPreview_OPFFileNotFound(t *testing.T) {
	service := NewRepairService()
	ctx := context.Background()

	report := &domain.ValidationReport{
		FilePath: "test.epub",
		FileType: "EPUB",
		IsValid:  false,
		Errors: []domain.ValidationError{
			{
				Code:    ErrorCodeOPFFileNotFound,
				Message: "Failed to read OPF file at OEBPS/content.opf: file not found",
				Location: &domain.ErrorLocation{
					Path: "OEBPS/content.opf",
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
	if action.Type != "create_opf" {
		t.Errorf("Expected action type 'create_opf', got '%s'", action.Type)
	}

	if !action.Automated {
		t.Error("Expected OPF creation to be automated")
	}
}

func TestPreview_NavMissing(t *testing.T) {
	service := NewRepairService()
	ctx := context.Background()

	report := &domain.ValidationReport{
		FilePath: "test.epub",
		FileType: "EPUB",
		IsValid:  false,
		Errors: []domain.ValidationError{
			{
				Code:    ErrorCodeNavMissingNavElement,
				Message: "Navigation document must contain at least one <nav> element",
				Location: &domain.ErrorLocation{
					Path: "OEBPS/nav.xhtml",
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
	if action.Type != "repair_nav_document" {
		t.Errorf("Expected action type 'repair_nav_document', got '%s'", action.Type)
	}

	if !action.Automated {
		t.Error("Expected nav repair to be automated")
	}
}

func TestPreview_ContentStructureRepair(t *testing.T) {
	service := NewRepairService()
	ctx := context.Background()

	report := &domain.ValidationReport{
		FilePath: "test.epub",
		FileType: "EPUB",
		IsValid:  false,
		Errors: []domain.ValidationError{
			{
				Code:    ErrorCodeContentInvalidNamespace,
				Message: "HTML element must have correct XHTML namespace",
				Location: &domain.ErrorLocation{
					Path: "OEBPS/content.xhtml",
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
	if action.Type != "repair_content_structure" {
		t.Errorf("Expected action type 'repair_content_structure', got '%s'", action.Type)
	}

	if !action.Automated {
		t.Error("Expected content structure repair to be automated")
	}
}

func TestPreview_OPFInvalidUniqueID(t *testing.T) {
	service := NewRepairService()
	ctx := context.Background()

	report := &domain.ValidationReport{
		FilePath: "test.epub",
		FileType: "EPUB",
		IsValid:  false,
		Errors: []domain.ValidationError{
			{
				Code:    ErrorCodeOPFInvalidUniqueID,
				Message: "unique-identifier 'uid' does not match any dc:identifier id",
				Location: &domain.ErrorLocation{
					Path: "OEBPS/content.opf",
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
	if action.Type != "fix_opf_unique_id" {
		t.Errorf("Expected action type 'fix_opf_unique_id', got '%s'", action.Type)
	}

	if !action.Automated {
		t.Error("Expected unique-identifier fix to be automated")
	}
}

func TestPreview_ContainerXMLMissing(t *testing.T) {
	service := NewRepairService()
	ctx := context.Background()

	report := &domain.ValidationReport{
		FilePath: "test.epub",
		FileType: "EPUB",
		IsValid:  false,
		Errors: []domain.ValidationError{
			{
				Code:    ErrorCodeContainerXMLMissing,
				Message: "Required file 'META-INF/container.xml' is missing",
				Location: &domain.ErrorLocation{
					Path: "META-INF/container.xml",
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
	if action.Type != "create_container_xml" {
		t.Errorf("Expected action type 'create_container_xml', got '%s'", action.Type)
	}

	if !action.Automated {
		t.Error("Expected container.xml creation to be automated")
	}
}

func TestPreview_ContentMissingDoctype(t *testing.T) {
	service := NewRepairService()
	ctx := context.Background()

	report := &domain.ValidationReport{
		FilePath: "test.epub",
		FileType: "EPUB",
		IsValid:  false,
		Errors: []domain.ValidationError{
			{
				Code:    ErrorCodeContentMissingDoctype,
				Message: "Content document is missing HTML5 DOCTYPE",
				Location: &domain.ErrorLocation{
					Path: "OEBPS/content.xhtml",
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
	if action.Type != "add_doctype" {
		t.Errorf("Expected action type 'add_doctype', got '%s'", action.Type)
	}

	if action.Target != "OEBPS/content.xhtml" {
		t.Errorf("Expected target 'OEBPS/content.xhtml', got '%s'", action.Target)
	}

	if !action.Automated {
		t.Error("Expected doctype addition to be automated")
	}
}

func TestPreview_OPFMetadataErrors(t *testing.T) {
	tests := []struct {
		name           string
		errorCode      string
		expectedType   string
		expectedTarget string
	}{
		{
			name:           "Missing Title",
			errorCode:      ErrorCodeOPFMissingTitle,
			expectedType:   "add_metadata_title",
			expectedTarget: "OEBPS/content.opf",
		},
		{
			name:           "Missing Identifier",
			errorCode:      ErrorCodeOPFMissingIdentifier,
			expectedType:   "add_metadata_identifier",
			expectedTarget: "OEBPS/content.opf",
		},
		{
			name:           "Missing Language",
			errorCode:      ErrorCodeOPFMissingLanguage,
			expectedType:   "add_metadata_language",
			expectedTarget: "OEBPS/content.opf",
		},
		{
			name:           "Missing Modified",
			errorCode:      ErrorCodeOPFMissingModified,
			expectedType:   "add_metadata_modified",
			expectedTarget: "OEBPS/content.opf",
		},
	}

	service := NewRepairService()
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := &domain.ValidationReport{
				FilePath: "test.epub",
				FileType: "EPUB",
				IsValid:  false,
				Errors: []domain.ValidationError{
					{
						Code:    tt.errorCode,
						Message: "Missing metadata",
						Location: &domain.ErrorLocation{
							Path: tt.expectedTarget,
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
			if action.Type != tt.expectedType {
				t.Errorf("Expected action type '%s', got '%s'", tt.expectedType, action.Type)
			}

			if action.Target != tt.expectedTarget {
				t.Errorf("Expected target '%s', got '%s'", tt.expectedTarget, action.Target)
			}

			if !action.Automated {
				t.Error("Expected metadata repair to be automated")
			}
		})
	}
}

func TestPreview_MultipleErrors(t *testing.T) {
	service := NewRepairService()
	ctx := context.Background()

	report := &domain.ValidationReport{
		FilePath: "test.epub",
		FileType: "EPUB",
		IsValid:  false,
		Errors: []domain.ValidationError{
			{
				Code:    ErrorCodeMimetypeInvalid,
				Message: "Invalid mimetype",
				Location: &domain.ErrorLocation{
					Path: "mimetype",
				},
			},
			{
				Code:    ErrorCodeContentMissingDoctype,
				Message: "Missing DOCTYPE",
				Location: &domain.ErrorLocation{
					Path: "OEBPS/content.xhtml",
				},
			},
			{
				Code:    ErrorCodeOPFMissingTitle,
				Message: "Missing title",
				Location: &domain.ErrorLocation{
					Path: "OEBPS/content.opf",
				},
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

	if !preview.CanAutoRepair {
		t.Error("Expected CanAutoRepair to be true for all automated repairs")
	}
}

func TestCanRepair(t *testing.T) {
	service := NewRepairService()
	ctx := context.Background()

	repairableCodes := []string{
		ErrorCodeMimetypeInvalid,
		ErrorCodeMimetypeNotFirst,
		ErrorCodeContainerXMLMissing,
		ErrorCodeContentMissingDoctype,
		ErrorCodeContentMissingHTML,
		ErrorCodeContentMissingHead,
		ErrorCodeContentMissingBody,
		ErrorCodeContentInvalidNamespace,
		ErrorCodeOPFMissingTitle,
		ErrorCodeOPFMissingIdentifier,
		ErrorCodeOPFMissingLanguage,
		ErrorCodeOPFMissingModified,
		ErrorCodeOPFMissingNavDocument,
		ErrorCodeOPFInvalidUniqueID,
		ErrorCodeNavMissingNavElement,
		ErrorCodeNavMissingTOC,
		ErrorCodeNavInvalidTOCStructure,
		ErrorCodeOPFFileNotFound,
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

	nonRepairableErr := &domain.ValidationError{
		Code:    "EPUB-UNKNOWN-999",
		Message: "Unknown error",
	}

	if service.CanRepair(ctx, nonRepairableErr) {
		t.Error("Expected unknown error code to not be repairable")
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

	result, err := service.Apply(ctx, "test.epub", preview)
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

func TestApply_MimetypeFix(t *testing.T) {
	service := NewRepairService()
	ctx := context.Background()

	tempDir := t.TempDir()
	testEPUB := filepath.Join(tempDir, "test.epub")

	if err := createTestEPUBWithBadMimetype(testEPUB); err != nil {
		t.Fatalf("Failed to create test EPUB: %v", err)
	}

	preview := &ports.RepairPreview{
		Actions: []ports.RepairAction{
			{
				Type:        "fix_mimetype_content",
				Description: "Fix mimetype content",
				Target:      "mimetype",
				Automated:   true,
			},
		},
		CanAutoRepair:  true,
		BackupRequired: true,
	}

	result, err := service.Apply(ctx, testEPUB, preview)
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

	if _, err := os.Stat(result.BackupPath); os.IsNotExist(err) {
		t.Error("Repaired file was not created")
	}

	if err := verifyMimetypeInEPUB(result.BackupPath); err != nil {
		t.Errorf("Mimetype verification failed: %v", err)
	}
}

func TestApply_DoctypeAddition(t *testing.T) {
	service := NewRepairService()
	ctx := context.Background()

	tempDir := t.TempDir()
	testEPUB := filepath.Join(tempDir, "test.epub")
	contentPath := "OEBPS/content.xhtml"

	if err := createTestEPUBWithoutDoctype(testEPUB, contentPath); err != nil {
		t.Fatalf("Failed to create test EPUB: %v", err)
	}

	preview := &ports.RepairPreview{
		Actions: []ports.RepairAction{
			{
				Type:        "add_doctype",
				Description: "Add DOCTYPE",
				Target:      contentPath,
				Automated:   true,
			},
		},
		CanAutoRepair:  true,
		BackupRequired: true,
	}

	result, err := service.Apply(ctx, testEPUB, preview)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected success, got error: %v", result.Error)
	}

	if err := verifyMimetypeInEPUB(result.BackupPath); err != nil {
		t.Errorf("Mimetype verification failed: %v", err)
	}

	content, err := readFileFromEPUB(result.BackupPath, contentPath)
	if err != nil {
		t.Fatalf("Failed to read content from repaired EPUB: %v", err)
	}

	if !strings.Contains(string(content), "<!DOCTYPE html>") {
		t.Error("DOCTYPE was not added to content document")
	}
}

func TestApply_OPFMetadataRepair(t *testing.T) {
	service := NewRepairService()
	ctx := context.Background()

	tempDir := t.TempDir()
	testEPUB := filepath.Join(tempDir, "test.epub")
	opfPath := "OEBPS/content.opf"

	if err := createTestEPUBWithIncompleteOPF(testEPUB, opfPath); err != nil {
		t.Fatalf("Failed to create test EPUB: %v", err)
	}

	preview := &ports.RepairPreview{
		Actions: []ports.RepairAction{
			{
				Type:        "add_metadata_title",
				Description: "Add title",
				Target:      opfPath,
				Automated:   true,
			},
			{
				Type:        "add_metadata_language",
				Description: "Add language",
				Target:      opfPath,
				Automated:   true,
			},
		},
		CanAutoRepair:  true,
		BackupRequired: true,
	}

	result, err := service.Apply(ctx, testEPUB, preview)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Expected success, got error: %v", result.Error)
	}

	opfData, err := readFileFromEPUB(result.BackupPath, opfPath)
	if err != nil {
		t.Fatalf("Failed to read OPF from repaired EPUB: %v", err)
	}

	var pkg Package
	if err := xml.Unmarshal(opfData, &pkg); err != nil {
		t.Fatalf("Failed to parse OPF: %v", err)
	}

	if len(pkg.Metadata.Titles) == 0 {
		t.Error("Title was not added")
	}

	if len(pkg.Metadata.Languages) == 0 {
		t.Error("Language was not added")
	}
}

func TestCreateBackup(t *testing.T) {
	service := NewRepairService()
	ctx := context.Background()

	tempDir := t.TempDir()
	sourcePath := filepath.Join(tempDir, "source.txt")
	backupPath := filepath.Join(tempDir, "backup.txt")

	content := []byte("test content")
	if err := os.WriteFile(sourcePath, content, 0600); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	if err := service.CreateBackup(ctx, sourcePath, backupPath); err != nil {
		t.Fatalf("CreateBackup failed: %v", err)
	}

	backupContent, err := os.ReadFile(backupPath) //nolint:gosec
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
	backupPath := filepath.Join(tempDir, "backup.txt")
	originalPath := filepath.Join(tempDir, "original.txt")

	backupContent := []byte("backup content")
	if err := os.WriteFile(backupPath, backupContent, 0600); err != nil {
		t.Fatalf("Failed to create backup file: %v", err)
	}

	if err := service.RestoreBackup(ctx, backupPath, originalPath); err != nil {
		t.Fatalf("RestoreBackup failed: %v", err)
	}

	restoredContent, err := os.ReadFile(originalPath) //nolint:gosec
	if err != nil {
		t.Fatalf("Failed to read restored file: %v", err)
	}

	if !bytes.Equal(backupContent, restoredContent) {
		t.Error("Restored content does not match backup")
	}
}

func TestAddDoctype(t *testing.T) {
	service := &RepairServiceImpl{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "No XML declaration",
			input:    "<html><head><title>Test</title></head></html>",
			expected: "<!DOCTYPE html>\n<html><head><title>Test</title></head></html>",
		},
		{
			name:     "With XML declaration",
			input:    "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<html><head><title>Test</title></head></html>",
			expected: "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<!DOCTYPE html>\n<html><head><title>Test</title></head></html>",
		},
		{
			name:     "Already has DOCTYPE",
			input:    "<!DOCTYPE html>\n<html><head><title>Test</title></head></html>",
			expected: "<!DOCTYPE html>\n<html><head><title>Test</title></head></html>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.addDoctype([]byte(tt.input))

			if string(result) != tt.expected {
				t.Errorf("Expected:\n%s\nGot:\n%s", tt.expected, string(result))
			}
		})
	}
}

func TestGenerateOutputPath(t *testing.T) {
	service := &RepairServiceImpl{}

	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "test.epub",
			expected: "test_repaired.epub",
		},
		{
			input:    "/path/to/book.epub",
			expected: "/path/to/book_repaired.epub",
		},
		{
			input:    "my.book.epub",
			expected: "my.book_repaired.epub",
		},
	}

	for _, tt := range tests {
		result := service.generateOutputPath(tt.input)
		if result != tt.expected {
			t.Errorf("For input %s, expected %s, got %s", tt.input, tt.expected, result)
		}
	}
}

func createTestEPUBWithBadMimetype(path string) error {
	f, err := os.Create(path) //nolint:gosec
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()

	zw := zip.NewWriter(f)
	defer func() {
		_ = zw.Close()
	}()

	w, err := zw.Create("mimetype")
	if err != nil {
		return err
	}
	if _, err := w.Write([]byte("wrong mimetype")); err != nil {
		return err
	}

	w, err = zw.Create("META-INF/container.xml")
	if err != nil {
		return err
	}
	containerXML := `<?xml version="1.0" encoding="UTF-8"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`
	if _, err := w.Write([]byte(containerXML)); err != nil {
		return err
	}

	return nil
}

func createTestEPUBWithContainer(path string, writeExtra func(*zip.Writer) error) error {
	f, err := os.Create(path) //nolint:gosec
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()

	zw := zip.NewWriter(f)
	defer func() {
		_ = zw.Close()
	}()

	header := &zip.FileHeader{
		Name:   "mimetype",
		Method: zip.Store,
	}
	w, err := zw.CreateHeader(header)
	if err != nil {
		return err
	}
	if _, err := w.Write([]byte(ExpectedMimetype)); err != nil {
		return err
	}

	w, err = zw.Create("META-INF/container.xml")
	if err != nil {
		return err
	}
	containerXML := `<?xml version="1.0" encoding="UTF-8"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`
	if _, err := w.Write([]byte(containerXML)); err != nil {
		return err
	}

	if writeExtra == nil {
		return nil
	}

	return writeExtra(zw)
}

func createTestEPUBWithoutDoctype(path, contentPath string) error {
	return createTestEPUBWithContainer(path, func(zw *zip.Writer) error {
		w, err := zw.Create(contentPath)
		if err != nil {
			return err
		}
		content := `<html xmlns="http://www.w3.org/1999/xhtml">
<head><title>Test</title></head>
<body><p>Test content</p></body>
</html>`
		_, err = w.Write([]byte(content))
		return err
	})
}

func createTestEPUBWithIncompleteOPF(path, opfPath string) error {
	return createTestEPUBWithContainer(path, func(zw *zip.Writer) error {
		w, err := zw.Create(opfPath)
		if err != nil {
			return err
		}
		opfXML := `<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" unique-identifier="bookid" version="3.3">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:identifier id="bookid">test-id</dc:identifier>
  </metadata>
  <manifest>
    <item id="nav" href="nav.xhtml" media-type="application/xhtml+xml" properties="nav"/>
  </manifest>
  <spine>
    <itemref idref="nav"/>
  </spine>
</package>`
		_, err = w.Write([]byte(opfXML))
		return err
	})
}

func verifyMimetypeInEPUB(path string) error {
	f, err := os.Open(path) //nolint:gosec
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()

	fi, err := f.Stat()
	if err != nil {
		return err
	}

	zr, err := zip.NewReader(f, fi.Size())
	if err != nil {
		return err
	}

	if len(zr.File) == 0 {
		return fmt.Errorf("ZIP archive is empty")
	}

	firstFile := zr.File[0]
	if firstFile.Name != "mimetype" {
		return fmt.Errorf("first file is not mimetype, got: %s", firstFile.Name)
	}

	if firstFile.Method != zip.Store {
		return fmt.Errorf("mimetype is compressed")
	}

	rc, err := firstFile.Open()
	if err != nil {
		return err
	}
	defer func() {
		_ = rc.Close()
	}()

	content, err := io.ReadAll(rc)
	if err != nil {
		return err
	}

	if string(content) != ExpectedMimetype {
		return fmt.Errorf("mimetype content is wrong: %s", string(content))
	}

	return nil
}

func readFileFromEPUB(epubPath, filePath string) ([]byte, error) {
	f, err := os.Open(epubPath) //nolint:gosec
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = f.Close()
	}()

	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	zr, err := zip.NewReader(f, fi.Size())
	if err != nil {
		return nil, err
	}

	for _, file := range zr.File {
		if file.Name == filePath {
			rc, err := file.Open()
			if err != nil {
				return nil, err
			}
			defer func() {
				_ = rc.Close()
			}()

			return io.ReadAll(rc)
		}
	}

	return nil, fmt.Errorf("file not found: %s", filePath)
}
