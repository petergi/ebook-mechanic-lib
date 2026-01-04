package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/petergi/ebook-mechanic-lib/internal/adapters/epub"
	"github.com/petergi/ebook-mechanic-lib/internal/domain"
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

func TestEPUBValidatorIntegration_TableDriven_AllErrorCodes(t *testing.T) {
	testCases := []struct {
		name         string
		file         string
		expectedCode string
		shouldFail   bool
		description  string
	}{
		{
			name:         "Container_NotZip",
			file:         "invalid/not_zip.epub",
			expectedCode: "EPUB-CONTAINER-001",
			shouldFail:   true,
			description:  "File is not a valid ZIP archive",
		},
		{
			name:         "Container_WrongMimetype",
			file:         "invalid/wrong_mimetype.epub",
			expectedCode: "EPUB-CONTAINER-002",
			shouldFail:   true,
			description:  "mimetype file contains wrong content",
		},
		{
			name:         "Container_MimetypeNotFirst",
			file:         "invalid/mimetype_not_first.epub",
			expectedCode: "EPUB-CONTAINER-003",
			shouldFail:   true,
			description:  "mimetype file is not the first file in the ZIP",
		},
		{
			name:         "Container_NoContainer",
			file:         "invalid/no_container.epub",
			expectedCode: "EPUB-CONTAINER-004",
			shouldFail:   true,
			description:  "META-INF/container.xml is missing",
		},
		{
			name:         "Container_InvalidContainerXML",
			file:         "invalid/invalid_container_xml.epub",
			expectedCode: "EPUB-CONTAINER-005",
			shouldFail:   true,
			description:  "META-INF/container.xml is not valid XML",
		},
		{
			name:         "OPF_InvalidXML",
			file:         "invalid/invalid_opf.epub",
			expectedCode: "EPUB-OPF-001",
			shouldFail:   true,
			description:  "OPF file is not valid XML",
		},
		{
			name:         "OPF_MissingTitle",
			file:         "invalid/missing_title.epub",
			expectedCode: "EPUB-OPF-002",
			shouldFail:   true,
			description:  "OPF metadata missing dc:title",
		},
		{
			name:         "OPF_MissingIdentifier",
			file:         "invalid/missing_identifier.epub",
			expectedCode: "EPUB-OPF-003",
			shouldFail:   true,
			description:  "OPF metadata missing dc:identifier",
		},
		{
			name:         "OPF_MissingLanguage",
			file:         "invalid/missing_language.epub",
			expectedCode: "EPUB-OPF-004",
			shouldFail:   true,
			description:  "OPF metadata missing dc:language",
		},
		{
			name:         "OPF_MissingModified",
			file:         "invalid/missing_modified.epub",
			expectedCode: "EPUB-OPF-005",
			shouldFail:   true,
			description:  "OPF metadata missing dcterms:modified",
		},
		{
			name:         "OPF_MissingNavDocument",
			file:         "invalid/missing_nav_document.epub",
			expectedCode: "EPUB-OPF-009",
			shouldFail:   true,
			description:  "OPF manifest missing nav document",
		},
		{
			name:         "Nav_InvalidNavDocument",
			file:         "invalid/invalid_nav_document.epub",
			expectedCode: "EPUB-NAV-006",
			shouldFail:   true,
			description:  "Navigation document missing nav element",
		},
		{
			name:         "Content_InvalidContentDocument",
			file:         "invalid/invalid_content_document.epub",
			expectedCode: "EPUB-CONTENT-002",
			shouldFail:   true,
			description:  "Content document missing DOCTYPE or namespace",
		},
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

			if tc.shouldFail && report.IsValid {
				t.Errorf("Expected invalid EPUB, got valid")
			}

			if !tc.shouldFail && !report.IsValid {
				t.Errorf("Expected valid EPUB, got invalid. Errors: %v", report.Errors)
			}

			if tc.shouldFail && tc.expectedCode != "" {
				foundExpectedError := false
				for _, e := range report.Errors {
					if e.Code == tc.expectedCode {
						foundExpectedError = true
						t.Logf("Found expected error: [%s] %s", e.Code, e.Message)
						break
					}
				}

				if !foundExpectedError {
					t.Errorf("Expected error code %s (%s)", tc.expectedCode, tc.description)
					t.Logf("Actual errors:")
					for _, e := range report.Errors {
						t.Logf("  [%s] %s", e.Code, e.Message)
					}
				}
			}
		})
	}
}

func TestEPUBValidatorIntegration_ValidFiles(t *testing.T) {
	testCases := []struct {
		name        string
		file        string
		description string
	}{
		{
			name:        "MinimalValid",
			file:        "valid/minimal.epub",
			description: "Minimal valid EPUB 3.0",
		},
		{
			name:        "MultipleRootfiles",
			file:        "valid/multiple_rootfiles.epub",
			description: "EPUB with multiple rootfiles in container.xml",
		},
		{
			name:        "ComplexNested",
			file:        "valid/complex_nested.epub",
			description: "EPUB with complex nested directory structure",
		},
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

			if !report.IsValid {
				t.Errorf("Expected valid EPUB (%s), got invalid", tc.description)
				for _, e := range report.Errors {
					t.Logf("  Error: [%s] %s", e.Code, e.Message)
				}
			}
		})
	}
}

func TestEPUBValidatorIntegration_PerformanceLargeFiles(t *testing.T) {
	testCases := []struct {
		name          string
		file          string
		maxDurationMS int64
	}{
		{
			name:          "Large100Chapters",
			file:          "valid/large_100_chapters.epub",
			maxDurationMS: 5000,
		},
		{
			name:          "Large500Chapters",
			file:          "valid/large_500_chapters.epub",
			maxDurationMS: 15000,
		},
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

			if !report.IsValid {
				t.Errorf("Expected valid EPUB, got invalid")
				for _, e := range report.Errors {
					t.Logf("  Error: [%s] %s", e.Code, e.Message)
				}
			}

			durationMS := report.Duration.Milliseconds()
			t.Logf("Validation took %d ms", durationMS)

			if durationMS > tc.maxDurationMS {
				t.Logf("Warning: validation took %d ms, expected < %d ms", durationMS, tc.maxDurationMS)
			}
		})
	}
}

func TestEPUBValidatorIntegration_EdgeCases(t *testing.T) {
	testCases := []struct {
		name        string
		file        string
		shouldFail  bool
		description string
	}{
		{
			name:        "VeryLarge10MBPlus",
			file:        "edge_cases/large_10mb_plus.epub",
			shouldFail:  false,
			description: "Very large EPUB > 10MB",
		},
		{
			name:        "MimetypeCompressed",
			file:        "invalid/mimetype_compressed.epub",
			shouldFail:  true,
			description: "mimetype file is compressed instead of stored",
		},
		{
			name:        "CorruptZip",
			file:        "invalid/corrupt_zip.epub",
			shouldFail:  true,
			description: "Truncated/corrupt ZIP file",
		},
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

			if tc.shouldFail && report.IsValid {
				t.Errorf("Expected invalid EPUB (%s), got valid", tc.description)
			}

			if !tc.shouldFail && !report.IsValid {
				t.Errorf("Expected valid EPUB (%s), got invalid", tc.description)
				for _, e := range report.Errors {
					t.Logf("  Error: [%s] %s", e.Code, e.Message)
				}
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

func TestEPUBValidatorIntegration_ContainerValidation(t *testing.T) {
	testCases := []struct {
		name         string
		file         string
		expectedCode string
	}{
		{"NoRootfile", "invalid/no_rootfile.epub", "EPUB-CONTAINER-005"},
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
				for _, e := range report.Errors {
					t.Logf("  Got: [%s] %s", e.Code, e.Message)
				}
			}
		})
	}
}
