package integration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/example/project/internal/adapters/pdf"
)

func TestPDFValidatorIntegration_ValidMinimal(t *testing.T) {
	validator := pdf.NewStructureValidator()

	testFile := filepath.Join("..", "..", "testdata", "pdf", "valid", "minimal.pdf")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skipf("Test file not found: %s", testFile)
	}

	result, err := validator.ValidateFile(testFile)
	if err != nil {
		t.Fatalf("ValidateFile failed: %v", err)
	}

	if !result.Valid {
		t.Errorf("Expected valid PDF, got invalid. Errors: %d", len(result.Errors))
		for _, e := range result.Errors {
			t.Logf("  Error: [%s] %s", e.Code, e.Message)
		}
	}

	if len(result.Errors) != 0 {
		t.Errorf("Expected 0 errors, got %d", len(result.Errors))
	}
}

func TestPDFValidatorIntegration_TableDriven_AllErrorCodes(t *testing.T) {
	testCases := []struct {
		name         string
		file         string
		expectedCode string
		shouldFail   bool
		description  string
	}{
		{
			name:         "Header_NotPDF",
			file:         "invalid/not_pdf.pdf",
			expectedCode: "PDF-HEADER-001",
			shouldFail:   true,
			description:  "File is not a PDF (missing header)",
		},
		{
			name:         "Header_NoHeader",
			file:         "invalid/no_header.pdf",
			expectedCode: "PDF-HEADER-001",
			shouldFail:   true,
			description:  "File has no PDF header",
		},
		{
			name:         "Header_InvalidVersion",
			file:         "invalid/invalid_version.pdf",
			expectedCode: "PDF-HEADER-002",
			shouldFail:   true,
			description:  "PDF version number is invalid",
		},
		{
			name:         "Trailer_NoEOF",
			file:         "invalid/no_eof.pdf",
			expectedCode: "PDF-TRAILER-003",
			shouldFail:   true,
			description:  "Missing %%EOF marker",
		},
		{
			name:         "Trailer_NoStartxref",
			file:         "invalid/no_startxref.pdf",
			expectedCode: "PDF-TRAILER-001",
			shouldFail:   true,
			description:  "Missing or invalid startxref",
		},
		{
			name:         "Xref_CorruptXref",
			file:         "invalid/corrupt_xref.pdf",
			expectedCode: "PDF-XREF-001",
			shouldFail:   true,
			description:  "Corrupt cross-reference table",
		},
		{
			name:         "Catalog_NoCatalog",
			file:         "invalid/no_catalog.pdf",
			expectedCode: "PDF-CATALOG-001",
			shouldFail:   true,
			description:  "Missing catalog object",
		},
		{
			name:         "Catalog_InvalidCatalog",
			file:         "invalid/invalid_catalog.pdf",
			expectedCode: "PDF-CATALOG-003",
			shouldFail:   true,
			description:  "Catalog missing /Pages entry",
		},
	}

	validator := pdf.NewStructureValidator()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testFile := filepath.Join("..", "..", "testdata", "pdf", tc.file)
			if _, err := os.Stat(testFile); os.IsNotExist(err) {
				t.Skipf("Test file not found: %s", testFile)
			}

			result, err := validator.ValidateFile(testFile)
			if err != nil {
				t.Fatalf("ValidateFile failed: %v", err)
			}

			if tc.shouldFail && result.Valid {
				t.Errorf("Expected invalid PDF, got valid")
			}

			if !tc.shouldFail && !result.Valid {
				t.Errorf("Expected valid PDF, got invalid. Errors: %v", result.Errors)
			}

			if tc.shouldFail && tc.expectedCode != "" {
				foundExpectedError := false
				for _, e := range result.Errors {
					if e.Code == tc.expectedCode {
						foundExpectedError = true
						t.Logf("Found expected error: [%s] %s", e.Code, e.Message)
						break
					}
				}

				if !foundExpectedError {
					t.Errorf("Expected error code %s (%s)", tc.expectedCode, tc.description)
					t.Logf("Actual errors:")
					for _, e := range result.Errors {
						t.Logf("  [%s] %s", e.Code, e.Message)
					}
				}
			}
		})
	}
}

func TestPDFValidatorIntegration_ValidFiles(t *testing.T) {
	testCases := []struct {
		name        string
		file        string
		description string
	}{
		{
			name:        "MinimalValid",
			file:        "valid/minimal.pdf",
			description: "Minimal valid PDF 1.4",
		},
		{
			name:        "WithImages",
			file:        "valid/with_images.pdf",
			description: "PDF with embedded images",
		},
	}

	validator := pdf.NewStructureValidator()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testFile := filepath.Join("..", "..", "testdata", "pdf", tc.file)
			if _, err := os.Stat(testFile); os.IsNotExist(err) {
				t.Skipf("Test file not found: %s", testFile)
			}

			result, err := validator.ValidateFile(testFile)
			if err != nil {
				t.Fatalf("ValidateFile failed: %v", err)
			}

			if !result.Valid {
				t.Errorf("Expected valid PDF (%s), got invalid", tc.description)
				for _, e := range result.Errors {
					t.Logf("  Error: [%s] %s", e.Code, e.Message)
				}
			}
		})
	}
}

func TestPDFValidatorIntegration_PerformanceLargeFiles(t *testing.T) {
	testCases := []struct {
		name        string
		file        string
		description string
	}{
		{
			name:        "Large100Pages",
			file:        "valid/large_100_pages.pdf",
			description: "PDF with 100 pages",
		},
		{
			name:        "Large1000Pages",
			file:        "valid/large_1000_pages.pdf",
			description: "PDF with 1000 pages",
		},
	}

	validator := pdf.NewStructureValidator()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testFile := filepath.Join("..", "..", "testdata", "pdf", tc.file)
			if _, err := os.Stat(testFile); os.IsNotExist(err) {
				t.Skipf("Test file not found: %s", testFile)
			}

			result, err := validator.ValidateFile(testFile)
			if err != nil {
				t.Fatalf("ValidateFile failed: %v", err)
			}

			if !result.Valid {
				t.Errorf("Expected valid PDF (%s), got invalid", tc.description)
				for _, e := range result.Errors {
					t.Logf("  Error: [%s] %s", e.Code, e.Message)
				}
			}

			info, _ := os.Stat(testFile)
			t.Logf("Validated %s (size: %d bytes)", tc.description, info.Size())
		})
	}
}

func TestPDFValidatorIntegration_EdgeCases(t *testing.T) {
	testCases := []struct {
		name        string
		file        string
		shouldFail  bool
		description string
	}{
		{
			name:        "VeryLarge10MBPlus",
			file:        "edge_cases/large_10mb_plus.pdf",
			shouldFail:  false,
			description: "Very large PDF > 10MB",
		},
		{
			name:        "CorruptPDF",
			file:        "invalid/corrupt.pdf",
			shouldFail:  true,
			description: "Truncated/corrupt PDF file",
		},
		{
			name:        "TruncatedStream",
			file:        "invalid/truncated_stream.pdf",
			shouldFail:  true,
			description: "PDF with truncated stream",
		},
		{
			name:        "MalformedObjects",
			file:        "invalid/malformed_objects.pdf",
			shouldFail:  true,
			description: "PDF with malformed objects",
		},
	}

	validator := pdf.NewStructureValidator()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testFile := filepath.Join("..", "..", "testdata", "pdf", tc.file)
			if _, err := os.Stat(testFile); os.IsNotExist(err) {
				t.Skipf("Test file not found: %s", testFile)
			}

			result, err := validator.ValidateFile(testFile)
			if err != nil {
				t.Fatalf("ValidateFile failed: %v", err)
			}

			if tc.shouldFail && result.Valid {
				t.Errorf("Expected invalid PDF (%s), got valid", tc.description)
			}

			if !tc.shouldFail && !result.Valid {
				t.Errorf("Expected valid PDF (%s), got invalid", tc.description)
				for _, e := range result.Errors {
					t.Logf("  Error: [%s] %s", e.Code, e.Message)
				}
			}
		})
	}
}

func TestPDFValidatorIntegration_EncryptedPDF(t *testing.T) {
	validator := pdf.NewStructureValidator()

	testFile := filepath.Join("..", "..", "testdata", "pdf", "edge_cases", "encrypted.pdf")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skipf("Test file not found: %s", testFile)
	}

	result, err := validator.ValidateFile(testFile)
	if err != nil {
		t.Fatalf("ValidateFile failed: %v", err)
	}

	t.Logf("Encrypted PDF validation result: Valid=%v, Errors=%d", result.Valid, len(result.Errors))
	for _, e := range result.Errors {
		t.Logf("  Error: [%s] %s", e.Code, e.Message)
	}
}

func TestPDFValidatorIntegration_CorruptionScenarios(t *testing.T) {
	testCases := []struct {
		name        string
		file        string
		description string
	}{
		{
			name:        "NoHeader",
			file:        "invalid/no_header.pdf",
			description: "PDF missing header signature",
		},
		{
			name:        "InvalidVersion",
			file:        "invalid/invalid_version.pdf",
			description: "PDF with unsupported version",
		},
		{
			name:        "NoEOF",
			file:        "invalid/no_eof.pdf",
			description: "PDF missing %%EOF marker",
		},
		{
			name:        "NoStartxref",
			file:        "invalid/no_startxref.pdf",
			description: "PDF missing startxref",
		},
		{
			name:        "CorruptXref",
			file:        "invalid/corrupt_xref.pdf",
			description: "PDF with corrupt xref table",
		},
		{
			name:        "NoCatalog",
			file:        "invalid/no_catalog.pdf",
			description: "PDF missing catalog",
		},
		{
			name:        "InvalidCatalog",
			file:        "invalid/invalid_catalog.pdf",
			description: "PDF with invalid catalog",
		},
	}

	validator := pdf.NewStructureValidator()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testFile := filepath.Join("..", "..", "testdata", "pdf", tc.file)
			if _, err := os.Stat(testFile); os.IsNotExist(err) {
				t.Skipf("Test file not found: %s", testFile)
			}

			result, err := validator.ValidateFile(testFile)
			if err != nil {
				t.Fatalf("ValidateFile failed: %v", err)
			}

			if result.Valid {
				t.Errorf("Expected invalid PDF (%s), got valid", tc.description)
			}

			if len(result.Errors) == 0 {
				t.Errorf("Expected at least one error for %s", tc.description)
			}

			t.Logf("Corruption scenario '%s' detected %d errors:", tc.description, len(result.Errors))
			for _, e := range result.Errors {
				t.Logf("  [%s] %s", e.Code, e.Message)
			}
		})
	}
}

func TestPDFValidatorIntegration_ResultStructure(t *testing.T) {
	validator := pdf.NewStructureValidator()

	testFile := filepath.Join("..", "..", "testdata", "pdf", "valid", "minimal.pdf")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skipf("Test file not found: %s", testFile)
	}

	result, err := validator.ValidateFile(testFile)
	if err != nil {
		t.Fatalf("ValidateFile failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result to be non-nil")
	}

	if result.Errors == nil {
		t.Error("Expected Errors slice to be initialized")
	}

	if !result.Valid && len(result.Errors) == 0 {
		t.Error("If Valid is false, Errors should contain at least one error")
	}
}

func TestPDFValidatorIntegration_ErrorDetails(t *testing.T) {
	validator := pdf.NewStructureValidator()

	testFile := filepath.Join("..", "..", "testdata", "pdf", "invalid", "no_header.pdf")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skipf("Test file not found: %s", testFile)
	}

	result, err := validator.ValidateFile(testFile)
	if err != nil {
		t.Fatalf("ValidateFile failed: %v", err)
	}

	if len(result.Errors) == 0 {
		t.Fatal("Expected at least one error")
	}

	for i, e := range result.Errors {
		if e.Code == "" {
			t.Errorf("Error %d: Code should not be empty", i)
		}

		if e.Message == "" {
			t.Errorf("Error %d: Message should not be empty", i)
		}

		if e.Details == nil {
			t.Logf("Error %d: Details is nil (may be valid for some errors)", i)
		}

		t.Logf("Error %d: [%s] %s (details: %v)", i, e.Code, e.Message, e.Details)
	}
}

func TestPDFValidatorIntegration_Systematic_Coverage(t *testing.T) {
	errorCodeTests := []struct {
		errorCode   string
		file        string
		description string
	}{
		{"PDF-HEADER-001", "invalid/not_pdf.pdf", "Invalid or missing PDF header"},
		{"PDF-HEADER-002", "invalid/invalid_version.pdf", "Invalid PDF version number"},
		{"PDF-TRAILER-001", "invalid/no_startxref.pdf", "Invalid or missing startxref"},
		{"PDF-TRAILER-003", "invalid/no_eof.pdf", "Missing %%EOF marker"},
		{"PDF-XREF-001", "invalid/corrupt_xref.pdf", "Invalid or damaged cross-reference table"},
		{"PDF-CATALOG-001", "invalid/no_catalog.pdf", "Missing or invalid catalog object"},
		{"PDF-CATALOG-003", "invalid/invalid_catalog.pdf", "Catalog missing /Pages entry"},
	}

	validator := pdf.NewStructureValidator()

	for _, tc := range errorCodeTests {
		t.Run(tc.errorCode, func(t *testing.T) {
			testFile := filepath.Join("..", "..", "testdata", "pdf", tc.file)
			if _, err := os.Stat(testFile); os.IsNotExist(err) {
				t.Skipf("Test file not found: %s", testFile)
			}

			result, err := validator.ValidateFile(testFile)
			if err != nil {
				t.Fatalf("ValidateFile failed: %v", err)
			}

			foundCode := false
			for _, e := range result.Errors {
				if e.Code == tc.errorCode {
					foundCode = true
					t.Logf("Successfully detected [%s]: %s", e.Code, e.Message)
					break
				}
			}

			if !foundCode {
				t.Errorf("Expected to find error code %s (%s)", tc.errorCode, tc.description)
				t.Logf("Found errors:")
				for _, e := range result.Errors {
					t.Logf("  [%s] %s", e.Code, e.Message)
				}
			}
		})
	}
}
