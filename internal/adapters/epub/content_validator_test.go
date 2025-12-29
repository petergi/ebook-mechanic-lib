package epub

import (
	"os"
	"path/filepath"
	"testing"
)

func createValidXHTML() string {
	return `<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
  <title>Test Chapter</title>
  <meta charset="utf-8"/>
</head>
<body>
  <h1>Chapter 1</h1>
  <p>This is a test paragraph.</p>
</body>
</html>`
}

func createXHTMLMalformedHTML() string {
	return `<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
  <title>Test Chapter</title>
</head>
<body>
  <p>Unclosed paragraph
  <div>Another unclosed div
</body>
</html>`
}

func createXHTMLMissingDoctype() string {
	return `<html xmlns="http://www.w3.org/1999/xhtml">
<head>
  <title>Test Chapter</title>
</head>
<body>
  <p>Content without DOCTYPE</p>
</body>
</html>`
}

func createXHTMLInvalidDoctype() string {
	return `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.1//EN" "http://www.w3.org/TR/xhtml11/DTD/xhtml11.dtd">
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
  <title>Test Chapter</title>
</head>
<body>
  <p>Content with XHTML 1.1 DOCTYPE</p>
</body>
</html>`
}

func createXHTMLMissingHTML() string {
	return `<!DOCTYPE html>
<head>
  <title>Test Chapter</title>
</head>
<body>
  <p>Content without html element</p>
</body>`
}

func createXHTMLMissingHead() string {
	return `<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml">
<body>
  <p>Content without head element</p>
</body>
</html>`
}

func createXHTMLMissingBody() string {
	return `<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
  <title>Test Chapter</title>
</head>
</html>`
}

func createXHTMLInvalidNamespace() string {
	return `<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/html">
<head>
  <title>Test Chapter</title>
</head>
<body>
  <p>Content with wrong namespace</p>
</body>
</html>`
}

func createXHTMLMissingNamespace() string {
	return `<!DOCTYPE html>
<html>
<head>
  <title>Test Chapter</title>
</head>
<body>
  <p>Content without namespace</p>
</body>
</html>`
}

func createXHTMLMultipleErrors() string {
	return `<html>
<body>
  <p>Missing DOCTYPE, head, and namespace</p>
</body>
</html>`
}

func createXHTMLComplexValid() string {
	return `<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops" lang="en" xml:lang="en">
<head>
  <meta charset="utf-8"/>
  <title>Complex Chapter</title>
  <link rel="stylesheet" type="text/css" href="styles.css"/>
</head>
<body>
  <section epub:type="chapter">
    <h1>Chapter Title</h1>
    <p>First paragraph with <em>emphasis</em> and <strong>strong</strong> text.</p>
    <p>Second paragraph with <a href="chapter2.xhtml">a link</a>.</p>
    <figure>
      <img src="image.jpg" alt="Description"/>
      <figcaption>Image caption</figcaption>
    </figure>
    <ul>
      <li>Item 1</li>
      <li>Item 2</li>
    </ul>
  </section>
</body>
</html>`
}

func TestContentValidator_ValidateBytes_Valid(t *testing.T) {
	validator := NewContentValidator()
	xhtmlData := createValidXHTML()

	result, err := validator.ValidateBytes([]byte(xhtmlData))

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !result.Valid {
		t.Errorf("Expected valid XHTML, got invalid with errors: %v", result.Errors)
	}

	if len(result.Errors) != 0 {
		t.Errorf("Expected no errors, got %d: %v", len(result.Errors), result.Errors)
	}

	if !result.HasDoctype {
		t.Error("Expected HasDoctype to be true")
	}

	if !result.HasHTML {
		t.Error("Expected HasHTML to be true")
	}

	if !result.HasHead {
		t.Error("Expected HasHead to be true")
	}

	if !result.HasBody {
		t.Error("Expected HasBody to be true")
	}

	if result.Namespace != XHTMLNamespace {
		t.Errorf("Expected namespace '%s', got '%s'", XHTMLNamespace, result.Namespace)
	}
}

func TestContentValidator_ValidateBytes_MalformedHTML(t *testing.T) {
	validator := NewContentValidator()
	xhtmlData := createXHTMLMalformedHTML()

	result, err := validator.ValidateBytes([]byte(xhtmlData))

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !result.Valid {
		t.Errorf("Expected valid result (HTML parser is lenient), got errors: %v", result.Errors)
	}
}

func TestContentValidator_ValidateBytes_MissingDoctype(t *testing.T) {
	validator := NewContentValidator()
	xhtmlData := createXHTMLMissingDoctype()

	result, err := validator.ValidateBytes([]byte(xhtmlData))

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("Expected invalid XHTML, got valid")
	}

	foundError := false
	for _, e := range result.Errors {
		if e.Code == ErrorCodeContentMissingDoctype {
			foundError = true
			break
		}
	}

	if !foundError {
		t.Errorf("Expected error code %s, got errors: %v", ErrorCodeContentMissingDoctype, result.Errors)
	}

	if result.HasDoctype {
		t.Error("Expected HasDoctype to be false")
	}
}

func TestContentValidator_ValidateBytes_InvalidDoctype(t *testing.T) {
	validator := NewContentValidator()
	xhtmlData := createXHTMLInvalidDoctype()

	result, err := validator.ValidateBytes([]byte(xhtmlData))

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("Expected invalid XHTML, got valid")
	}

	foundError := false
	for _, e := range result.Errors {
		if e.Code == ErrorCodeContentInvalidDoctype {
			foundError = true
			if e.Details["expected"] != "<!DOCTYPE html>" {
				t.Errorf("Expected 'expected' detail to be '<!DOCTYPE html>', got '%v'", e.Details["expected"])
			}
			break
		}
	}

	if !foundError {
		t.Errorf("Expected error code %s, got errors: %v", ErrorCodeContentInvalidDoctype, result.Errors)
	}

	if !result.HasDoctype {
		t.Error("Expected HasDoctype to be true (DOCTYPE is present, just wrong type)")
	}
}

func TestContentValidator_ValidateBytes_MissingHTML(t *testing.T) {
	validator := NewContentValidator()
	xhtmlData := createXHTMLMissingHTML()

	result, err := validator.ValidateBytes([]byte(xhtmlData))

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("Expected invalid XHTML, got valid")
	}

	foundError := false
	for _, e := range result.Errors {
		if e.Code == ErrorCodeContentMissingHTML {
			foundError = true
			break
		}
	}

	if !foundError {
		t.Errorf("Expected error code %s, got errors: %v", ErrorCodeContentMissingHTML, result.Errors)
	}

	if result.HasHTML {
		t.Error("Expected HasHTML to be false")
	}
}

func TestContentValidator_ValidateBytes_MissingHead(t *testing.T) {
	validator := NewContentValidator()
	xhtmlData := createXHTMLMissingHead()

	result, err := validator.ValidateBytes([]byte(xhtmlData))

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("Expected invalid XHTML, got valid")
	}

	foundError := false
	for _, e := range result.Errors {
		if e.Code == ErrorCodeContentMissingHead {
			foundError = true
			break
		}
	}

	if !foundError {
		t.Errorf("Expected error code %s, got errors: %v", ErrorCodeContentMissingHead, result.Errors)
	}

	if result.HasHead {
		t.Error("Expected HasHead to be false")
	}
}

func TestContentValidator_ValidateBytes_MissingBody(t *testing.T) {
	validator := NewContentValidator()
	xhtmlData := createXHTMLMissingBody()

	result, err := validator.ValidateBytes([]byte(xhtmlData))

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("Expected invalid XHTML, got valid")
	}

	foundError := false
	for _, e := range result.Errors {
		if e.Code == ErrorCodeContentMissingBody {
			foundError = true
			break
		}
	}

	if !foundError {
		t.Errorf("Expected error code %s, got errors: %v", ErrorCodeContentMissingBody, result.Errors)
	}

	if result.HasBody {
		t.Error("Expected HasBody to be false")
	}
}

func TestContentValidator_ValidateBytes_InvalidNamespace(t *testing.T) {
	validator := NewContentValidator()
	xhtmlData := createXHTMLInvalidNamespace()

	result, err := validator.ValidateBytes([]byte(xhtmlData))

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("Expected invalid XHTML, got valid")
	}

	foundError := false
	for _, e := range result.Errors {
		if e.Code == ErrorCodeContentInvalidNamespace {
			foundError = true
			if e.Details["expected"] != XHTMLNamespace {
				t.Errorf("Expected 'expected' detail to be '%s', got '%v'", XHTMLNamespace, e.Details["expected"])
			}
			if e.Details["found"] != "http://www.w3.org/1999/html" {
				t.Errorf("Expected 'found' detail to be 'http://www.w3.org/1999/html', got '%v'", e.Details["found"])
			}
			break
		}
	}

	if !foundError {
		t.Errorf("Expected error code %s, got errors: %v", ErrorCodeContentInvalidNamespace, result.Errors)
	}
}

func TestContentValidator_ValidateBytes_MissingNamespace(t *testing.T) {
	validator := NewContentValidator()
	xhtmlData := createXHTMLMissingNamespace()

	result, err := validator.ValidateBytes([]byte(xhtmlData))

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("Expected invalid XHTML, got valid")
	}

	foundError := false
	for _, e := range result.Errors {
		if e.Code == ErrorCodeContentInvalidNamespace {
			foundError = true
			if e.Details["expected"] != XHTMLNamespace {
				t.Errorf("Expected 'expected' detail to be '%s', got '%v'", XHTMLNamespace, e.Details["expected"])
			}
			if e.Details["found"] != "" {
				t.Errorf("Expected 'found' detail to be empty, got '%v'", e.Details["found"])
			}
			break
		}
	}

	if !foundError {
		t.Errorf("Expected error code %s, got errors: %v", ErrorCodeContentInvalidNamespace, result.Errors)
	}
}

func TestContentValidator_ValidateBytes_MultipleErrors(t *testing.T) {
	validator := NewContentValidator()
	xhtmlData := createXHTMLMultipleErrors()

	result, err := validator.ValidateBytes([]byte(xhtmlData))

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("Expected invalid XHTML, got valid")
	}

	if len(result.Errors) < 3 {
		t.Errorf("Expected at least 3 errors (missing DOCTYPE, head, namespace), got %d: %v", len(result.Errors), result.Errors)
	}

	errorCodes := make(map[string]bool)
	for _, e := range result.Errors {
		errorCodes[e.Code] = true
	}

	expectedCodes := []string{
		ErrorCodeContentMissingDoctype,
		ErrorCodeContentMissingHead,
		ErrorCodeContentInvalidNamespace,
	}

	for _, expectedCode := range expectedCodes {
		if !errorCodes[expectedCode] {
			t.Errorf("Expected error code %s to be present", expectedCode)
		}
	}
}

func TestContentValidator_ValidateBytes_ComplexValid(t *testing.T) {
	validator := NewContentValidator()
	xhtmlData := createXHTMLComplexValid()

	result, err := validator.ValidateBytes([]byte(xhtmlData))

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !result.Valid {
		t.Errorf("Expected valid XHTML, got invalid with errors: %v", result.Errors)
	}

	if len(result.Errors) != 0 {
		t.Errorf("Expected no errors, got %d: %v", len(result.Errors), result.Errors)
	}

	if result.Namespace != XHTMLNamespace {
		t.Errorf("Expected namespace '%s', got '%s'", XHTMLNamespace, result.Namespace)
	}
}

func TestContentValidator_ValidateFile(t *testing.T) {
	validator := NewContentValidator()
	xhtmlData := createValidXHTML()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "chapter.xhtml")

	if err := os.WriteFile(tmpFile, []byte(xhtmlData), 0600); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	result, err := validator.ValidateFile(tmpFile)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !result.Valid {
		t.Errorf("Expected valid XHTML, got invalid with errors: %v", result.Errors)
	}

	if len(result.Errors) != 0 {
		t.Errorf("Expected no errors, got %d: %v", len(result.Errors), result.Errors)
	}
}

func TestContentValidator_ValidateFile_NonExistent(t *testing.T) {
	validator := NewContentValidator()

	result, err := validator.ValidateFile("/nonexistent/chapter.xhtml")

	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}

	if result != nil {
		t.Errorf("Expected nil result for error case, got %v", result)
	}
}

func TestContentValidator_ValidateBytes_EmptyContent(t *testing.T) {
	validator := NewContentValidator()
	xhtmlData := ""

	result, err := validator.ValidateBytes([]byte(xhtmlData))

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("Expected invalid XHTML, got valid")
	}

	if len(result.Errors) == 0 {
		t.Error("Expected errors for empty content")
	}
}

func TestContentValidator_ValidateBytes_OnlyDoctype(t *testing.T) {
	validator := NewContentValidator()
	xhtmlData := "<!DOCTYPE html>"

	result, err := validator.ValidateBytes([]byte(xhtmlData))

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("Expected invalid XHTML, got valid")
	}

	if result.HasDoctype != true {
		t.Error("Expected HasDoctype to be true")
	}

	if result.HasHTML != false {
		t.Error("Expected HasHTML to be false")
	}
}

func TestContentValidator_ValidateBytes_CaseSensitivity(t *testing.T) {
	validator := NewContentValidator()
	xhtmlData := `<!DOCTYPE HTML>
<HTML xmlns="http://www.w3.org/1999/xhtml">
<HEAD>
  <title>Test Chapter</title>
</HEAD>
<BODY>
  <p>Content with uppercase tags</p>
</BODY>
</HTML>`

	result, err := validator.ValidateBytes([]byte(xhtmlData))

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !result.Valid {
		t.Errorf("Expected valid XHTML (parser normalizes case), got invalid with errors: %v", result.Errors)
	}
}

func TestContentErrorCodes(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"Not Well-Formed", ErrorCodeContentNotWellFormed, "EPUB-CONTENT-001"},
		{"Missing DOCTYPE", ErrorCodeContentMissingDoctype, "EPUB-CONTENT-002"},
		{"Invalid DOCTYPE", ErrorCodeContentInvalidDoctype, "EPUB-CONTENT-003"},
		{"Missing HTML", ErrorCodeContentMissingHTML, "EPUB-CONTENT-004"},
		{"Missing Head", ErrorCodeContentMissingHead, "EPUB-CONTENT-005"},
		{"Missing Body", ErrorCodeContentMissingBody, "EPUB-CONTENT-006"},
		{"Invalid Namespace", ErrorCodeContentInvalidNamespace, "EPUB-CONTENT-007"},
		{"Invalid Encoding", ErrorCodeContentInvalidEncoding, "EPUB-CONTENT-008"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.code != tt.expected {
				t.Errorf("Expected error code %s, got %s", tt.expected, tt.code)
			}
		})
	}
}
