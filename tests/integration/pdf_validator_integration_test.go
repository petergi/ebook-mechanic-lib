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

func TestPDFValidatorIntegration_AllErrorCodes(t *testing.T) {
	testCases := []struct {
		name         string
		file         string
		expectedCode string
	}{
		{"NotPDF", "invalid/not_pdf.pdf", "PDF-HEADER-001"},
		{"NoHeader", "invalid/no_header.pdf", "PDF-HEADER-001"},
		{"InvalidVersion", "invalid/invalid_version.pdf", "PDF-HEADER-002"},
		{"NoEOF", "invalid/no_eof.pdf", "PDF-TRAILER-003"},
		{"NoStartxref", "invalid/no_startxref.pdf", "PDF-TRAILER-001"},
		{"CorruptXref", "invalid/corrupt_xref.pdf", "PDF-XREF-001"},
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
				t.Error("Expected invalid PDF, got valid")
			}

			foundExpectedError := false
			for _, e := range result.Errors {
				if e.Code == tc.expectedCode {
					foundExpectedError = true
					break
				}
			}

			if !foundExpectedError {
				t.Errorf("Expected error code %s, got errors: %v", tc.expectedCode, result.Errors)
			}
		})
	}
}
