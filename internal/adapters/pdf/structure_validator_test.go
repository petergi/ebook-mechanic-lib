package pdf

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func createMinimalValidPDF() []byte {
	return []byte(`%PDF-1.4
1 0 obj
<<
/Type /Catalog
/Pages 2 0 R
>>
endobj
2 0 obj
<<
/Type /Pages
/Kids [3 0 R]
/Count 1
>>
endobj
3 0 obj
<<
/Type /Page
/Parent 2 0 R
/MediaBox [0 0 612 792]
/Contents 4 0 R
/Resources <<
/Font <<
/F1 <<
/Type /Font
/Subtype /Type1
/BaseFont /Helvetica
>>
>>
>>
>>
endobj
4 0 obj
<<
/Length 44
>>
stream
BT
/F1 12 Tf
100 700 Td
(Hello World) Tj
ET
endstream
endobj
xref
0 5
0000000000 65535 f 
0000000009 00000 n 
0000000058 00000 n 
0000000115 00000 n 
0000000317 00000 n 
trailer
<<
/Size 5
/Root 1 0 R
>>
startxref
410
%%EOF
`)
}

func createPDFWithInvalidHeader() []byte {
	return []byte(`%PD-1.4
1 0 obj
<<
/Type /Catalog
/Pages 2 0 R
>>
endobj
xref
0 2
0000000000 65535 f 
0000000008 00000 n 
trailer
<<
/Size 2
/Root 1 0 R
>>
startxref
57
%%EOF
`)
}

func createPDFWithInvalidVersion() []byte {
	return []byte(`%PDF-1.9
1 0 obj
<<
/Type /Catalog
/Pages 2 0 R
>>
endobj
xref
0 2
0000000000 65535 f 
0000000009 00000 n 
trailer
<<
/Size 2
/Root 1 0 R
>>
startxref
58
%%EOF
`)
}

func createPDFWithMissingEOF() []byte {
	return []byte(`%PDF-1.4
1 0 obj
<<
/Type /Catalog
/Pages 2 0 R
>>
endobj
xref
0 2
0000000000 65535 f 
0000000009 00000 n 
trailer
<<
/Size 2
/Root 1 0 R
>>
startxref
58
`)
}

func createTruncatedPDF() []byte {
	return []byte(`%PDF-1.4
1 0 obj
<<
/Type /Catalog
/Pages 2 0 R
>>
endobj
xref
0 2
0000000000 65535 f 
0000000009 00000 n 
trailer
<<
/Size 2
/Root 1 0 R
>>`)
}

func createPDFWithMissingStartXref() []byte {
	return []byte(`%PDF-1.4
1 0 obj
<<
/Type /Catalog
/Pages 2 0 R
>>
endobj
xref
0 2
0000000000 65535 f 
0000000009 00000 n 
trailer
<<
/Size 2
/Root 1 0 R
>>
%%EOF
`)
}

func createPDFWithDamagedXref() []byte {
	return []byte(`%PDF-1.4
1 0 obj
<<
/Type /Catalog
/Pages 2 0 R
>>
endobj
xref
0 2
DAMAGED XREF TABLE
trailer
<<
/Size 2
/Root 1 0 R
>>
startxref
58
%%EOF
`)
}

func createPDFWithMissingCatalog() []byte {
	return []byte(`%PDF-1.4
1 0 obj
<<
/Type /NotCatalog
/Pages 2 0 R
>>
endobj
xref
0 2
0000000000 65535 f 
0000000009 00000 n 
trailer
<<
/Size 2
/Root 1 0 R
>>
startxref
64
%%EOF
`)
}

func createPDFWithoutCatalogType() []byte {
	return []byte(`%PDF-1.4
1 0 obj
<<
/Pages 2 0 R
>>
endobj
xref
0 2
0000000000 65535 f 
0000000009 00000 n 
trailer
<<
/Size 2
/Root 1 0 R
>>
startxref
45
%%EOF
`)
}

func createPDFWithoutPages() []byte {
	return []byte(`%PDF-1.4
1 0 obj
<<
/Type /Catalog
>>
endobj
xref
0 2
0000000000 65535 f 
0000000009 00000 n 
trailer
<<
/Size 2
/Root 1 0 R
>>
startxref
38
%%EOF
`)
}

func createEmptyPDF() []byte {
	return []byte{}
}

func TestStructureValidator_ValidateBytes_ValidPDF(t *testing.T) {
	validator := NewStructureValidator()
	data := createMinimalValidPDF()

	result, err := validator.ValidateBytes(data)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !result.Valid {
		t.Errorf("Expected valid PDF, got invalid")
		for _, e := range result.Errors {
			t.Logf("Error: %s - %s", e.Code, e.Message)
		}
	}

	if len(result.Errors) != 0 {
		t.Errorf("Expected 0 errors, got %d", len(result.Errors))
	}
}

func TestStructureValidator_ValidateBytes_InvalidHeader(t *testing.T) {
	validator := NewStructureValidator()
	data := createPDFWithInvalidHeader()

	result, err := validator.ValidateBytes(data)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.Valid {
		t.Errorf("Expected invalid PDF, got valid")
	}

	if len(result.Errors) == 0 {
		t.Fatalf("Expected errors, got none")
	}

	foundHeaderError := false
	for _, e := range result.Errors {
		if e.Code == ErrorCodePDFHeader001 {
			foundHeaderError = true
			if e.Message == "" {
				t.Errorf("Error message should not be empty")
			}
		}
	}

	if !foundHeaderError {
		t.Errorf("Expected error code %s", ErrorCodePDFHeader001)
	}
}

func TestStructureValidator_ValidateBytes_InvalidVersion(t *testing.T) {
	validator := NewStructureValidator()
	data := createPDFWithInvalidVersion()

	result, err := validator.ValidateBytes(data)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.Valid {
		t.Errorf("Expected invalid PDF, got valid")
	}

	foundVersionError := false
	for _, e := range result.Errors {
		if e.Code == ErrorCodePDFHeader002 {
			foundVersionError = true
		}
	}

	if !foundVersionError {
		t.Errorf("Expected error code %s", ErrorCodePDFHeader002)
	}
}

func TestStructureValidator_ValidateBytes_MissingEOF(t *testing.T) {
	validator := NewStructureValidator()
	data := createPDFWithMissingEOF()

	result, err := validator.ValidateBytes(data)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.Valid {
		t.Errorf("Expected invalid PDF, got valid")
	}

	foundEOFError := false
	for _, e := range result.Errors {
		if e.Code == ErrorCodePDFTrailer003 {
			foundEOFError = true
		}
	}

	if !foundEOFError {
		t.Errorf("Expected error code %s", ErrorCodePDFTrailer003)
	}
}

func TestStructureValidator_ValidateBytes_TruncatedFile(t *testing.T) {
	validator := NewStructureValidator()
	data := createTruncatedPDF()

	result, err := validator.ValidateBytes(data)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.Valid {
		t.Errorf("Expected invalid PDF, got valid")
	}

	foundTrailerError := false
	for _, e := range result.Errors {
		if e.Code == ErrorCodePDFTrailer001 || e.Code == ErrorCodePDFTrailer003 {
			foundTrailerError = true
		}
	}

	if !foundTrailerError {
		t.Errorf("Expected trailer error")
	}
}

func TestStructureValidator_ValidateBytes_MissingStartXref(t *testing.T) {
	validator := NewStructureValidator()
	data := createPDFWithMissingStartXref()

	result, err := validator.ValidateBytes(data)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.Valid {
		t.Errorf("Expected invalid PDF, got valid")
	}

	foundStartXrefError := false
	for _, e := range result.Errors {
		if e.Code == ErrorCodePDFTrailer001 {
			foundStartXrefError = true
		}
	}

	if !foundStartXrefError {
		t.Errorf("Expected error code %s", ErrorCodePDFTrailer001)
	}
}

func TestStructureValidator_ValidateBytes_DamagedXref(t *testing.T) {
	validator := NewStructureValidator()
	data := createPDFWithDamagedXref()

	result, err := validator.ValidateBytes(data)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.Valid {
		t.Errorf("Expected invalid PDF, got valid")
	}

	foundXrefError := false
	for _, e := range result.Errors {
		if e.Code == ErrorCodePDFXref001 || e.Code == ErrorCodePDFStructure012 {
			foundXrefError = true
		}
	}

	if !foundXrefError {
		t.Errorf("Expected xref error")
	}
}

func TestStructureValidator_ValidateBytes_MissingCatalog(t *testing.T) {
	validator := NewStructureValidator()
	data := createPDFWithMissingCatalog()

	result, err := validator.ValidateBytes(data)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.Valid {
		t.Errorf("Expected invalid PDF, got valid")
	}

	foundCatalogError := false
	for _, e := range result.Errors {
		if e.Code == ErrorCodePDFCatalog001 || e.Code == ErrorCodePDFCatalog002 {
			foundCatalogError = true
		}
	}

	if !foundCatalogError {
		t.Errorf("Expected catalog error")
	}
}

func TestStructureValidator_ValidateBytes_MissingCatalogType(t *testing.T) {
	validator := NewStructureValidator()
	data := createPDFWithoutCatalogType()

	result, err := validator.ValidateBytes(data)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.Valid {
		t.Errorf("Expected invalid PDF, got valid")
	}

	foundTypeError := false
	for _, e := range result.Errors {
		if e.Code == ErrorCodePDFCatalog002 {
			foundTypeError = true
		}
	}

	if !foundTypeError {
		t.Errorf("Expected error code %s", ErrorCodePDFCatalog002)
	}
}

func TestStructureValidator_ValidateBytes_MissingPages(t *testing.T) {
	validator := NewStructureValidator()
	data := createPDFWithoutPages()

	result, err := validator.ValidateBytes(data)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.Valid {
		t.Errorf("Expected invalid PDF, got valid")
	}

	foundPagesError := false
	for _, e := range result.Errors {
		if e.Code == ErrorCodePDFCatalog003 {
			foundPagesError = true
		}
	}

	if !foundPagesError {
		t.Errorf("Expected error code %s", ErrorCodePDFCatalog003)
	}
}

func TestStructureValidator_ValidateBytes_EmptyFile(t *testing.T) {
	validator := NewStructureValidator()
	data := createEmptyPDF()

	result, err := validator.ValidateBytes(data)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.Valid {
		t.Errorf("Expected invalid PDF, got valid")
	}

	if len(result.Errors) == 0 {
		t.Fatalf("Expected errors, got none")
	}

	foundHeaderError := false
	for _, e := range result.Errors {
		if e.Code == ErrorCodePDFHeader001 {
			foundHeaderError = true
		}
	}

	if !foundHeaderError {
		t.Errorf("Expected error code %s for empty file", ErrorCodePDFHeader001)
	}
}

func TestStructureValidator_ValidateFile(t *testing.T) {
	validator := NewStructureValidator()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.pdf")

	data := createMinimalValidPDF()
	if err := os.WriteFile(testFile, data, 0600); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	result, err := validator.ValidateFile(testFile)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !result.Valid {
		t.Errorf("Expected valid PDF, got invalid")
		for _, e := range result.Errors {
			t.Logf("Error: %s - %s", e.Code, e.Message)
		}
	}
}

func TestStructureValidator_ValidateFile_NonExistent(t *testing.T) {
	validator := NewStructureValidator()

	result, err := validator.ValidateFile("/nonexistent/file.pdf")
	if err == nil {
		t.Errorf("Expected error for nonexistent file, got nil")
	}
	if result != nil {
		t.Errorf("Expected nil result for nonexistent file, got %+v", result)
	}
}

func TestStructureValidator_ValidateReader(t *testing.T) {
	validator := NewStructureValidator()
	data := createMinimalValidPDF()
	reader := bytes.NewReader(data)

	result, err := validator.ValidateReader(reader)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !result.Valid {
		t.Errorf("Expected valid PDF, got invalid")
		for _, e := range result.Errors {
			t.Logf("Error: %s - %s", e.Code, e.Message)
		}
	}
}

func TestErrorCodes_Coverage(t *testing.T) {
	expectedCodes := []string{
		ErrorCodePDFHeader001,
		ErrorCodePDFHeader002,
		ErrorCodePDFTrailer001,
		ErrorCodePDFTrailer002,
		ErrorCodePDFTrailer003,
		ErrorCodePDFXref001,
		ErrorCodePDFXref002,
		ErrorCodePDFXref003,
		ErrorCodePDFCatalog001,
		ErrorCodePDFCatalog002,
		ErrorCodePDFCatalog003,
		ErrorCodePDFStructure012,
	}

	for _, code := range expectedCodes {
		if code == "" {
			t.Errorf("Error code is empty")
		}
		if len(code) < 5 {
			t.Errorf("Error code %s is too short", code)
		}
	}
}

func TestValidationError_Structure(t *testing.T) {
	err := ValidationError{
		Code:    ErrorCodePDFHeader001,
		Message: "Test message",
		Details: map[string]interface{}{
			"key": "value",
		},
	}

	if err.Code != ErrorCodePDFHeader001 {
		t.Errorf("Expected code %s, got %s", ErrorCodePDFHeader001, err.Code)
	}

	if err.Message != "Test message" {
		t.Errorf("Expected message 'Test message', got '%s'", err.Message)
	}

	if err.Details["key"] != "value" {
		t.Errorf("Expected details key='value', got %v", err.Details["key"])
	}
}

func TestStructureValidator_MultipleErrors(t *testing.T) {
	validator := NewStructureValidator()

	pdfData := []byte(`%PD-1.9
1 0 obj
<<
/Type /NotCatalog
>>
endobj
xref
0 2
0000000000 65535 f 
0000000008 00000 n 
trailer
<<
/Size 2
/Root 1 0 R
>>
startxref
52
`)

	result, err := validator.ValidateBytes(pdfData)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.Valid {
		t.Errorf("Expected invalid PDF, got valid")
	}

	if len(result.Errors) < 2 {
		t.Errorf("Expected multiple errors, got %d", len(result.Errors))
	}
}

func TestStructureValidator_AllVersions(t *testing.T) {
	validator := NewStructureValidator()

	versions := []string{"1.0", "1.1", "1.2", "1.3", "1.4", "1.5", "1.6", "1.7"}

	for _, version := range versions {
		pdfData := []byte(fmt.Sprintf(`%%PDF-%s
1 0 obj
<<
/Type /Catalog
/Pages 2 0 R
>>
endobj
2 0 obj
<<
/Type /Pages
/Kids []
/Count 0
>>
endobj
xref
0 3
0000000000 65535 f 
0000000010 00000 n 
0000000059 00000 n 
trailer
<<
/Size 3
/Root 1 0 R
>>
startxref
108
%%%%EOF
`, version))

		result, err := validator.ValidateBytes(pdfData)
		if err != nil {
			t.Errorf("Version %s: unexpected error: %v", version, err)
			continue
		}

		if !result.Valid {
			t.Errorf("Version %s: expected valid, got invalid", version)
			for _, e := range result.Errors {
				t.Logf("  Error: %s - %s", e.Code, e.Message)
			}
		}
	}
}
