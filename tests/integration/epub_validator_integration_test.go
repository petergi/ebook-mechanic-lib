package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/example/project/internal/adapters/epub"
	"github.com/example/project/internal/domain"
)

func TestEPUBValidatorIntegration_ValidMinimal(t *testing.T) {
	validator := epub.NewEPUBValidator()
	ctx := context.Background()

	testFile := filepath.Join("..", "..", "testdata", "epub", "valid", "minimal.epub")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skipf("Test file not found: %s", testFile)
	}

	report, err := validator.ValidateFile(ctx, testFile)
	if err != nil {
		t.Fatalf("ValidateFile failed: %v", err)
	}

	if !report.IsValid {
		t.Errorf("Expected valid EPUB, got invalid. Errors: %d", len(report.Errors))
		for _, e := range report.Errors {
			t.Logf("  Error: [%s] %s", e.Code, e.Message)
		}
	}

	if report.FileType != "EPUB" {
		t.Errorf("Expected FileType='EPUB', got '%s'", report.FileType)
	}

	if len(report.Errors) != 0 {
		t.Errorf("Expected 0 errors, got %d", len(report.Errors))
	}
}

func TestEPUBValidatorIntegration_AllErrorCodes(t *testing.T) {
	testCases := []struct {
		name         string
		file         string
		expectedCode string
	}{
		{"NotZip", "invalid/not_zip.epub", "EPUB-CONTAINER-001"},
		{"WrongMimetype", "invalid/wrong_mimetype.epub", "EPUB-CONTAINER-002"},
		{"MimetypeNotFirst", "invalid/mimetype_not_first.epub", "EPUB-CONTAINER-003"},
		{"NoContainer", "invalid/no_container.epub", "EPUB-CONTAINER-004"},
		{"InvalidContainerXML", "invalid/invalid_container_xml.epub", "EPUB-CONTAINER-005"},
		{"InvalidOPF", "invalid/invalid_opf.epub", "EPUB-OPF-001"},
		{"MissingTitle", "invalid/missing_title.epub", "EPUB-OPF-002"},
		{"MissingIdentifier", "invalid/missing_identifier.epub", "EPUB-OPF-003"},
		{"MissingLanguage", "invalid/missing_language.epub", "EPUB-OPF-004"},
		{"MissingModified", "invalid/missing_modified.epub", "EPUB-OPF-005"},
		{"MissingNavDocument", "invalid/missing_nav_document.epub", "EPUB-OPF-009"},
	}

	validator := epub.NewEPUBValidator()
	ctx := context.Background()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testFile := filepath.Join("..", "..", "testdata", "epub", tc.file)
			if _, err := os.Stat(testFile); os.IsNotExist(err) {
				t.Skipf("Test file not found: %s", testFile)
			}

			report, err := validator.ValidateFile(ctx, testFile)
			if err != nil {
				t.Fatalf("ValidateFile failed: %v", err)
			}

			if report.IsValid {
				t.Error("Expected invalid EPUB, got valid")
			}

			foundExpectedError := false
			for _, e := range report.Errors {
				if e.Code == tc.expectedCode {
					foundExpectedError = true
					break
				}
			}

			if !foundExpectedError {
				t.Errorf("Expected error code %s", tc.expectedCode)
			}
		})
	}
}

func TestEPUBValidatorIntegration_ReportStructure(t *testing.T) {
	validator := epub.NewEPUBValidator()
	ctx := context.Background()

	testFile := filepath.Join("..", "..", "testdata", "epub", "valid", "minimal.epub")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skipf("Test file not found: %s", testFile)
	}

	report, err := validator.ValidateFile(ctx, testFile)
	if err != nil {
		t.Fatalf("ValidateFile failed: %v", err)
	}

	if report.FilePath != testFile {
		t.Errorf("Expected FilePath='%s', got '%s'", testFile, report.FilePath)
	}

	if report.FileType != "EPUB" {
		t.Errorf("Expected FileType='EPUB', got '%s'", report.FileType)
	}

	if report.ValidationTime.IsZero() {
		t.Error("Expected ValidationTime to be set")
	}

	if report.Duration == 0 {
		t.Error("Expected Duration to be set")
	}

	if report.Errors == nil {
		t.Error("Expected Errors slice to be initialized")
	}
}

func TestEPUBValidatorIntegration_ErrorStructure(t *testing.T) {
	validator := epub.NewEPUBValidator()
	ctx := context.Background()

	testFile := filepath.Join("..", "..", "testdata", "epub", "invalid", "wrong_mimetype.epub")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skipf("Test file not found: %s", testFile)
	}

	report, err := validator.ValidateFile(ctx, testFile)
	if err != nil {
		t.Fatalf("ValidateFile failed: %v", err)
	}

	if len(report.Errors) == 0 {
		t.Fatal("Expected at least one error")
	}

	for i, e := range report.Errors {
		if e.Code == "" {
			t.Errorf("Error %d: Code should not be empty", i)
		}

		if e.Message == "" {
			t.Errorf("Error %d: Message should not be empty", i)
		}

		if e.Severity != domain.SeverityError {
			t.Errorf("Error %d: Expected Severity='error', got '%s'", i, e.Severity)
		}

		if e.Timestamp.IsZero() {
			t.Errorf("Error %d: Timestamp should be set", i)
		}

		if e.Location == nil {
			t.Errorf("Error %d: Location should be set", i)
		}
	}
}
