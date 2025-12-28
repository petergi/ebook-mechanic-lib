package epub

import (
	"os"
	"path/filepath"
	"testing"
)

func createValidNavDocument() string {
	return `<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops">
<head>
  <title>Table of Contents</title>
</head>
<body>
  <nav epub:type="toc">
    <h1>Table of Contents</h1>
    <ol>
      <li><a href="chapter1.xhtml">Chapter 1</a></li>
      <li><a href="chapter2.xhtml">Chapter 2</a></li>
      <li>
        <a href="chapter3.xhtml">Chapter 3</a>
        <ol>
          <li><a href="chapter3.xhtml#section1">Section 3.1</a></li>
          <li><a href="chapter3.xhtml#section2">Section 3.2</a></li>
        </ol>
      </li>
    </ol>
  </nav>
</body>
</html>`
}

func createNavWithLandmarks() string {
	return `<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops">
<head>
  <title>Navigation</title>
</head>
<body>
  <nav epub:type="toc">
    <h1>Table of Contents</h1>
    <ol>
      <li><a href="chapter1.xhtml">Chapter 1</a></li>
      <li><a href="chapter2.xhtml">Chapter 2</a></li>
    </ol>
  </nav>
  <nav epub:type="landmarks">
    <h2>Landmarks</h2>
    <ol>
      <li><a href="cover.xhtml" epub:type="cover">Cover</a></li>
      <li><a href="toc.xhtml" epub:type="toc">Table of Contents</a></li>
      <li><a href="chapter1.xhtml" epub:type="bodymatter">Start of Content</a></li>
    </ol>
  </nav>
</body>
</html>`
}

func createNavMissingTOC() string {
	return `<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops">
<head>
  <title>Navigation</title>
</head>
<body>
  <nav epub:type="landmarks">
    <h2>Landmarks</h2>
    <ol>
      <li><a href="cover.xhtml">Cover</a></li>
    </ol>
  </nav>
</body>
</html>`
}

func createNavMissingOL() string {
	return `<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops">
<head>
  <title>Navigation</title>
</head>
<body>
  <nav epub:type="toc">
    <h1>Table of Contents</h1>
    <p>This should be an ordered list</p>
  </nav>
</body>
</html>`
}

func createNavWithInvalidLinks() string {
	return `<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops">
<head>
  <title>Navigation</title>
</head>
<body>
  <nav epub:type="toc">
    <h1>Table of Contents</h1>
    <ol>
      <li><a href="chapter1.xhtml">Valid Chapter 1</a></li>
      <li><a href="http://example.com/chapter2.xhtml">Absolute URL</a></li>
      <li><a href="/absolute/path.xhtml">Absolute Path</a></li>
      <li><a href="../outside/book.xhtml">Outside Path</a></li>
    </ol>
  </nav>
</body>
</html>`
}

func createNavMalformed() string {
	return `<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops">
<head>
  <title>Navigation</title>
</head>
<body>
  <nav epub:type="toc">
    <h1>Table of Contents
    <ol>
      <li><a href="chapter1.xhtml">Chapter 1</a>
    </ol>
  </nav>
</body>
</html>`
}

func createNavMissingNavElement() string {
	return `<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops">
<head>
  <title>Navigation</title>
</head>
<body>
  <div>
    <h1>Table of Contents</h1>
    <ol>
      <li><a href="chapter1.xhtml">Chapter 1</a></li>
    </ol>
  </div>
</body>
</html>`
}

func createNavWithEmptyLinks() string {
	return `<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops">
<head>
  <title>Navigation</title>
</head>
<body>
  <nav epub:type="toc">
    <h1>Table of Contents</h1>
    <ol>
      <li><a href="">Empty Link</a></li>
      <li><a href="chapter1.xhtml">Valid Link</a></li>
    </ol>
  </nav>
</body>
</html>`
}

func createNavWithProtocolRelativeURL() string {
	return `<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops">
<head>
  <title>Navigation</title>
</head>
<body>
  <nav epub:type="toc">
    <h1>Table of Contents</h1>
    <ol>
      <li><a href="//example.com/chapter.xhtml">Protocol-relative URL</a></li>
    </ol>
  </nav>
</body>
</html>`
}

func createNavLandmarksMissingOL() string {
	return `<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops">
<head>
  <title>Navigation</title>
</head>
<body>
  <nav epub:type="toc">
    <h1>Table of Contents</h1>
    <ol>
      <li><a href="chapter1.xhtml">Chapter 1</a></li>
    </ol>
  </nav>
  <nav epub:type="landmarks">
    <h2>Landmarks</h2>
    <p>This should be an ordered list</p>
  </nav>
</body>
</html>`
}

func createNavComplexValid() string {
	return `<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops">
<head>
  <title>Navigation</title>
</head>
<body>
  <nav epub:type="toc" id="toc">
    <h1>Contents</h1>
    <ol>
      <li><a href="frontmatter/cover.xhtml">Cover</a></li>
      <li><a href="frontmatter/titlepage.xhtml">Title Page</a></li>
      <li>
        <a href="part1/chapter1.xhtml">Part I</a>
        <ol>
          <li><a href="part1/chapter1.xhtml">Chapter 1</a></li>
          <li>
            <a href="part1/chapter2.xhtml">Chapter 2</a>
            <ol>
              <li><a href="part1/chapter2.xhtml#s1">Section 2.1</a></li>
              <li><a href="part1/chapter2.xhtml#s2">Section 2.2</a></li>
            </ol>
          </li>
        </ol>
      </li>
      <li>
        <a href="part2/chapter3.xhtml">Part II</a>
        <ol>
          <li><a href="part2/chapter3.xhtml">Chapter 3</a></li>
          <li><a href="part2/chapter4.xhtml">Chapter 4</a></li>
        </ol>
      </li>
      <li><a href="backmatter/index.xhtml">Index</a></li>
    </ol>
  </nav>
  <nav epub:type="landmarks">
    <h2>Guide</h2>
    <ol>
      <li><a href="frontmatter/cover.xhtml" epub:type="cover">Cover</a></li>
      <li><a href="nav.xhtml" epub:type="toc">Table of Contents</a></li>
      <li><a href="part1/chapter1.xhtml" epub:type="bodymatter">Start of Content</a></li>
      <li><a href="backmatter/index.xhtml" epub:type="index">Index</a></li>
    </ol>
  </nav>
</body>
</html>`
}

func TestNavValidator_ValidateBytes_ValidNav(t *testing.T) {
	validator := NewNavValidator()
	navData := createValidNavDocument()

	result, err := validator.ValidateBytes([]byte(navData))

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !result.Valid {
		t.Errorf("Expected valid navigation, got invalid with errors: %v", result.Errors)
	}

	if len(result.Errors) != 0 {
		t.Errorf("Expected no errors, got %d: %v", len(result.Errors), result.Errors)
	}

	if !result.HasTOC {
		t.Error("Expected HasTOC to be true")
	}

	if result.HasLandmarks {
		t.Error("Expected HasLandmarks to be false")
	}

	if len(result.TOCLinks) != 5 {
		t.Errorf("Expected 5 TOC links, got %d", len(result.TOCLinks))
	}

	expectedLinks := []string{"chapter1.xhtml", "chapter2.xhtml", "chapter3.xhtml", "chapter3.xhtml#section1", "chapter3.xhtml#section2"}
	for i, link := range result.TOCLinks {
		if link.Href != expectedLinks[i] {
			t.Errorf("Expected link %d to be '%s', got '%s'", i, expectedLinks[i], link.Href)
		}
	}
}

func TestNavValidator_ValidateBytes_WithLandmarks(t *testing.T) {
	validator := NewNavValidator()
	navData := createNavWithLandmarks()

	result, err := validator.ValidateBytes([]byte(navData))

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !result.Valid {
		t.Errorf("Expected valid navigation, got invalid with errors: %v", result.Errors)
	}

	if !result.HasTOC {
		t.Error("Expected HasTOC to be true")
	}

	if !result.HasLandmarks {
		t.Error("Expected HasLandmarks to be true")
	}

	if len(result.TOCLinks) != 2 {
		t.Errorf("Expected 2 TOC links, got %d", len(result.TOCLinks))
	}

	if len(result.LandmarkLinks) != 3 {
		t.Errorf("Expected 3 landmark links, got %d", len(result.LandmarkLinks))
	}
}

func TestNavValidator_ValidateBytes_MissingTOC(t *testing.T) {
	validator := NewNavValidator()
	navData := createNavMissingTOC()

	result, err := validator.ValidateBytes([]byte(navData))

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("Expected invalid navigation, got valid")
	}

	foundError := false
	for _, e := range result.Errors {
		if e.Code == ErrorCodeNavMissingTOC {
			foundError = true
			break
		}
	}

	if !foundError {
		t.Errorf("Expected error code %s, got errors: %v", ErrorCodeNavMissingTOC, result.Errors)
	}

	if result.HasTOC {
		t.Error("Expected HasTOC to be false")
	}
}

func TestNavValidator_ValidateBytes_MissingOL(t *testing.T) {
	validator := NewNavValidator()
	navData := createNavMissingOL()

	result, err := validator.ValidateBytes([]byte(navData))

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("Expected invalid navigation, got valid")
	}

	foundError := false
	for _, e := range result.Errors {
		if e.Code == ErrorCodeNavInvalidTOCStructure {
			foundError = true
			break
		}
	}

	if !foundError {
		t.Errorf("Expected error code %s, got errors: %v", ErrorCodeNavInvalidTOCStructure, result.Errors)
	}
}

func TestNavValidator_ValidateBytes_InvalidLinks(t *testing.T) {
	validator := NewNavValidator()
	navData := createNavWithInvalidLinks()

	result, err := validator.ValidateBytes([]byte(navData))

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("Expected invalid navigation, got valid")
	}

	invalidLinkCount := 0
	for _, e := range result.Errors {
		if e.Code == ErrorCodeNavInvalidLinks {
			invalidLinkCount++
		}
	}

	if invalidLinkCount != 3 {
		t.Errorf("Expected 3 invalid link errors, got %d", invalidLinkCount)
	}
}

func TestNavValidator_ValidateBytes_Malformed(t *testing.T) {
	validator := NewNavValidator()
	navData := createNavMalformed()

	result, err := validator.ValidateBytes([]byte(navData))

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("Expected valid navigation (HTML parser is lenient)")
	}
}

func TestNavValidator_ValidateBytes_MissingNavElement(t *testing.T) {
	validator := NewNavValidator()
	navData := createNavMissingNavElement()

	result, err := validator.ValidateBytes([]byte(navData))

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("Expected invalid navigation, got valid")
	}

	foundError := false
	for _, e := range result.Errors {
		if e.Code == ErrorCodeNavMissingNavElement {
			foundError = true
			break
		}
	}

	if !foundError {
		t.Errorf("Expected error code %s, got errors: %v", ErrorCodeNavMissingNavElement, result.Errors)
	}
}

func TestNavValidator_ValidateBytes_EmptyLinks(t *testing.T) {
	validator := NewNavValidator()
	navData := createNavWithEmptyLinks()

	result, err := validator.ValidateBytes([]byte(navData))

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("Expected invalid navigation, got valid")
	}

	foundError := false
	for _, e := range result.Errors {
		if e.Code == ErrorCodeNavInvalidLinks && e.Details["href"] == "" {
			foundError = true
			break
		}
	}

	if !foundError {
		t.Errorf("Expected error code %s for empty link, got errors: %v", ErrorCodeNavInvalidLinks, result.Errors)
	}
}

func TestNavValidator_ValidateBytes_ProtocolRelativeURL(t *testing.T) {
	validator := NewNavValidator()
	navData := createNavWithProtocolRelativeURL()

	result, err := validator.ValidateBytes([]byte(navData))

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("Expected invalid navigation, got valid")
	}

	foundError := false
	for _, e := range result.Errors {
		if e.Code == ErrorCodeNavInvalidLinks {
			foundError = true
			break
		}
	}

	if !foundError {
		t.Errorf("Expected error code %s for protocol-relative URL, got errors: %v", ErrorCodeNavInvalidLinks, result.Errors)
	}
}

func TestNavValidator_ValidateBytes_LandmarksMissingOL(t *testing.T) {
	validator := NewNavValidator()
	navData := createNavLandmarksMissingOL()

	result, err := validator.ValidateBytes([]byte(navData))

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("Expected invalid navigation, got valid")
	}

	foundError := false
	for _, e := range result.Errors {
		if e.Code == ErrorCodeNavInvalidLandmarks {
			foundError = true
			break
		}
	}

	if !foundError {
		t.Errorf("Expected error code %s, got errors: %v", ErrorCodeNavInvalidLandmarks, result.Errors)
	}
}

func TestNavValidator_ValidateBytes_ComplexValid(t *testing.T) {
	validator := NewNavValidator()
	navData := createNavComplexValid()

	result, err := validator.ValidateBytes([]byte(navData))

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !result.Valid {
		t.Errorf("Expected valid navigation, got invalid with errors: %v", result.Errors)
	}

	if len(result.Errors) != 0 {
		t.Errorf("Expected no errors, got %d: %v", len(result.Errors), result.Errors)
	}

	if !result.HasTOC {
		t.Error("Expected HasTOC to be true")
	}

	if !result.HasLandmarks {
		t.Error("Expected HasLandmarks to be true")
	}

	if len(result.TOCLinks) != 11 {
		t.Errorf("Expected 11 TOC links, got %d", len(result.TOCLinks))
	}

	if len(result.LandmarkLinks) != 4 {
		t.Errorf("Expected 4 landmark links, got %d", len(result.LandmarkLinks))
	}
}

func TestNavValidator_ValidateFile(t *testing.T) {
	validator := NewNavValidator()
	navData := createValidNavDocument()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "nav.xhtml")

	if err := os.WriteFile(tmpFile, []byte(navData), 0644); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	result, err := validator.ValidateFile(tmpFile)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !result.Valid {
		t.Errorf("Expected valid navigation, got invalid with errors: %v", result.Errors)
	}

	if len(result.Errors) != 0 {
		t.Errorf("Expected no errors, got %d: %v", len(result.Errors), result.Errors)
	}
}

func TestNavValidator_ValidateFile_NonExistent(t *testing.T) {
	validator := NewNavValidator()

	result, err := validator.ValidateFile("/nonexistent/nav.xhtml")

	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}

	if result != nil {
		t.Errorf("Expected nil result for error case, got %v", result)
	}
}

func TestNavValidator_ValidateBytes_EmptyContent(t *testing.T) {
	validator := NewNavValidator()
	navData := ""

	result, err := validator.ValidateBytes([]byte(navData))

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("Expected invalid navigation, got valid")
	}

	if len(result.Errors) == 0 {
		t.Error("Expected errors for empty content")
	}
}

func TestNavValidator_ValidateBytes_OnlyHTML(t *testing.T) {
	validator := NewNavValidator()
	navData := `<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml">
<head><title>Test</title></head>
<body></body>
</html>`

	result, err := validator.ValidateBytes([]byte(navData))

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("Expected invalid navigation, got valid")
	}

	if !result.HasTOC {
		if result.Valid {
			t.Error("Expected validation to fail when TOC is missing")
		}
	}
}

func TestNavErrorCodes(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"Not Well-Formed", ErrorCodeNavNotWellFormed, "EPUB-NAV-001"},
		{"Missing TOC", ErrorCodeNavMissingTOC, "EPUB-NAV-002"},
		{"Invalid TOC Structure", ErrorCodeNavInvalidTOCStructure, "EPUB-NAV-003"},
		{"Invalid Links", ErrorCodeNavInvalidLinks, "EPUB-NAV-004"},
		{"Invalid Landmarks", ErrorCodeNavInvalidLandmarks, "EPUB-NAV-005"},
		{"Missing Nav Element", ErrorCodeNavMissingNavElement, "EPUB-NAV-006"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.code != tt.expected {
				t.Errorf("Expected error code %s, got %s", tt.expected, tt.code)
			}
		})
	}
}

func TestNavValidator_LinkValidation_EdgeCases(t *testing.T) {
	validator := NewNavValidator()

	tests := []struct {
		name        string
		href        string
		expectValid bool
	}{
		{"Relative path", "chapter1.xhtml", true},
		{"Relative with subdirectory", "content/chapter1.xhtml", true},
		{"With fragment", "chapter1.xhtml#section1", true},
		{"HTTP URL", "http://example.com/page.html", false},
		{"HTTPS URL", "https://example.com/page.html", false},
		{"Protocol-relative", "//example.com/page.html", false},
		{"Absolute path", "/content/chapter1.xhtml", false},
		{"Parent directory", "../chapter1.xhtml", false},
		{"Empty", "", false},
		{"Complex relative", "part1/subdir/chapter.xhtml", true},
		{"Fragment only", "#section1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.isValidRelativeLink(tt.href)
			if result != tt.expectValid {
				t.Errorf("For href '%s': expected valid=%v, got valid=%v", tt.href, tt.expectValid, result)
			}
		})
	}
}
