package epub

import (
	"archive/zip"
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
)

func createCompleteValidEPUB(t *testing.T) []byte {
	t.Helper()

	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	mimetypeWriter, err := zipWriter.CreateHeader(&zip.FileHeader{
		Name:   MimetypeFilename,
		Method: zip.Store,
	})
	if err != nil {
		t.Fatalf("Failed to create mimetype header: %v", err)
	}
	if _, err := mimetypeWriter.Write([]byte(ExpectedMimetype)); err != nil {
		t.Fatalf("Failed to write mimetype: %v", err)
	}

	containerXML := `<?xml version="1.0" encoding="UTF-8"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`

	containerWriter, err := zipWriter.Create(ContainerXMLPath)
	if err != nil {
		t.Fatalf("Failed to create container.xml: %v", err)
	}
	if _, err := containerWriter.Write([]byte(containerXML)); err != nil {
		t.Fatalf("Failed to write container.xml: %v", err)
	}

	opfContent := `<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="book-id">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>Complete Test Book</dc:title>
    <dc:identifier id="book-id">urn:isbn:123456789</dc:identifier>
    <dc:language>en</dc:language>
    <meta property="dcterms:modified">2024-01-01T00:00:00Z</meta>
  </metadata>
  <manifest>
    <item id="nav" href="nav.xhtml" media-type="application/xhtml+xml" properties="nav"/>
    <item id="chapter1" href="chapter1.xhtml" media-type="application/xhtml+xml"/>
    <item id="chapter2" href="chapter2.xhtml" media-type="application/xhtml+xml"/>
  </manifest>
  <spine>
    <itemref idref="chapter1"/>
    <itemref idref="chapter2"/>
  </spine>
</package>`

	opfWriter, err := zipWriter.Create("OEBPS/content.opf")
	if err != nil {
		t.Fatalf("Failed to create content.opf: %v", err)
	}
	if _, err := opfWriter.Write([]byte(opfContent)); err != nil {
		t.Fatalf("Failed to write content.opf: %v", err)
	}

	navContent := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html>
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
</body>
</html>`

	navWriter, err := zipWriter.Create("OEBPS/nav.xhtml")
	if err != nil {
		t.Fatalf("Failed to create nav.xhtml: %v", err)
	}
	if _, err := navWriter.Write([]byte(navContent)); err != nil {
		t.Fatalf("Failed to write nav.xhtml: %v", err)
	}

	chapter1Content := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
  <title>Chapter 1</title>
</head>
<body>
  <h1>Chapter 1</h1>
  <p>This is the content of chapter 1.</p>
</body>
</html>`

	chapter1Writer, err := zipWriter.Create("OEBPS/chapter1.xhtml")
	if err != nil {
		t.Fatalf("Failed to create chapter1.xhtml: %v", err)
	}
	if _, err := chapter1Writer.Write([]byte(chapter1Content)); err != nil {
		t.Fatalf("Failed to write chapter1.xhtml: %v", err)
	}

	chapter2Content := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
  <title>Chapter 2</title>
</head>
<body>
  <h1>Chapter 2</h1>
  <p>This is the content of chapter 2.</p>
</body>
</html>`

	chapter2Writer, err := zipWriter.Create("OEBPS/chapter2.xhtml")
	if err != nil {
		t.Fatalf("Failed to create chapter2.xhtml: %v", err)
	}
	if _, err := chapter2Writer.Write([]byte(chapter2Content)); err != nil {
		t.Fatalf("Failed to write chapter2.xhtml: %v", err)
	}

	if err := zipWriter.Close(); err != nil {
		t.Fatalf("Failed to close zip writer: %v", err)
	}

	return buf.Bytes()
}

func createEPUBWithInvalidContainer(t *testing.T) []byte {
	t.Helper()

	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	mimetypeWriter, err := zipWriter.CreateHeader(&zip.FileHeader{
		Name:   MimetypeFilename,
		Method: zip.Store,
	})
	if err != nil {
		t.Fatalf("Failed to create mimetype header: %v", err)
	}
	if _, err := mimetypeWriter.Write([]byte("application/wrong")); err != nil {
		t.Fatalf("Failed to write mimetype: %v", err)
	}

	if err := zipWriter.Close(); err != nil {
		t.Fatalf("Failed to close zip writer: %v", err)
	}

	return buf.Bytes()
}

func createEPUBWithInvalidOPF(t *testing.T) []byte {
	t.Helper()

	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	mimetypeWriter, err := zipWriter.CreateHeader(&zip.FileHeader{
		Name:   MimetypeFilename,
		Method: zip.Store,
	})
	if err != nil {
		t.Fatalf("Failed to create mimetype header: %v", err)
	}
	if _, err := mimetypeWriter.Write([]byte(ExpectedMimetype)); err != nil {
		t.Fatalf("Failed to write mimetype: %v", err)
	}

	containerXML := `<?xml version="1.0" encoding="UTF-8"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`

	containerWriter, err := zipWriter.Create(ContainerXMLPath)
	if err != nil {
		t.Fatalf("Failed to create container.xml: %v", err)
	}
	if _, err := containerWriter.Write([]byte(containerXML)); err != nil {
		t.Fatalf("Failed to write container.xml: %v", err)
	}

	opfContent := `<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="book-id">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:identifier id="book-id">urn:isbn:123456789</dc:identifier>
    <dc:language>en</dc:language>
    <meta property="dcterms:modified">2024-01-01T00:00:00Z</meta>
  </metadata>
  <manifest>
    <item id="nav" href="nav.xhtml" media-type="application/xhtml+xml" properties="nav"/>
  </manifest>
  <spine>
    <itemref idref="nav"/>
  </spine>
</package>`

	opfWriter, err := zipWriter.Create("OEBPS/content.opf")
	if err != nil {
		t.Fatalf("Failed to create content.opf: %v", err)
	}
	if _, err := opfWriter.Write([]byte(opfContent)); err != nil {
		t.Fatalf("Failed to write content.opf: %v", err)
	}

	if err := zipWriter.Close(); err != nil {
		t.Fatalf("Failed to close zip writer: %v", err)
	}

	return buf.Bytes()
}

type testFile struct {
	path    string
	content string
}

func buildEPUBWithOPFAndFiles(t *testing.T, opfContent, navContent string, extraFiles []testFile) []byte {
	t.Helper()

	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	mimetypeWriter, err := zipWriter.CreateHeader(&zip.FileHeader{
		Name:   MimetypeFilename,
		Method: zip.Store,
	})
	if err != nil {
		t.Fatalf("Failed to create mimetype header: %v", err)
	}
	if _, err := mimetypeWriter.Write([]byte(ExpectedMimetype)); err != nil {
		t.Fatalf("Failed to write mimetype: %v", err)
	}

	containerXML := `<?xml version="1.0" encoding="UTF-8"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`

	containerWriter, err := zipWriter.Create(ContainerXMLPath)
	if err != nil {
		t.Fatalf("Failed to create container.xml: %v", err)
	}
	if _, err := containerWriter.Write([]byte(containerXML)); err != nil {
		t.Fatalf("Failed to write container.xml: %v", err)
	}

	opfWriter, err := zipWriter.Create("OEBPS/content.opf")
	if err != nil {
		t.Fatalf("Failed to create content.opf: %v", err)
	}
	if _, err := opfWriter.Write([]byte(opfContent)); err != nil {
		t.Fatalf("Failed to write content.opf: %v", err)
	}

	if navContent != "" {
		navWriter, err := zipWriter.Create("OEBPS/nav.xhtml")
		if err != nil {
			t.Fatalf("Failed to create nav.xhtml: %v", err)
		}
		if _, err := navWriter.Write([]byte(navContent)); err != nil {
			t.Fatalf("Failed to write nav.xhtml: %v", err)
		}
	}

	for _, file := range extraFiles {
		writer, err := zipWriter.Create(file.path)
		if err != nil {
			t.Fatalf("Failed to create %s: %v", file.path, err)
		}
		if _, err := writer.Write([]byte(file.content)); err != nil {
			t.Fatalf("Failed to write %s: %v", file.path, err)
		}
	}

	if err := zipWriter.Close(); err != nil {
		t.Fatalf("Failed to close zip writer: %v", err)
	}

	return buf.Bytes()
}

func createEPUBWithInvalidNav(t *testing.T) []byte {
	t.Helper()

	opfContent := `<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="book-id">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>Test Book</dc:title>
    <dc:identifier id="book-id">urn:isbn:123456789</dc:identifier>
    <dc:language>en</dc:language>
    <meta property="dcterms:modified">2024-01-01T00:00:00Z</meta>
  </metadata>
  <manifest>
    <item id="nav" href="nav.xhtml" media-type="application/xhtml+xml" properties="nav"/>
    <item id="chapter1" href="chapter1.xhtml" media-type="application/xhtml+xml"/>
  </manifest>
  <spine>
    <itemref idref="chapter1"/>
  </spine>
</package>`

	navContent := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
  <title>Navigation</title>
</head>
<body>
  <div>No nav element here</div>
</body>
</html>`

	chapter1Content := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
  <title>Chapter 1</title>
</head>
<body>
  <h1>Chapter 1</h1>
</body>
</html>`
	return buildEPUBWithOPFAndFiles(t, opfContent, navContent, []testFile{
		{path: "OEBPS/chapter1.xhtml", content: chapter1Content},
	})
}

func createEPUBWithInvalidContent(t *testing.T) []byte {
	t.Helper()

	opfContent := `<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="book-id">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>Test Book</dc:title>
    <dc:identifier id="book-id">urn:isbn:123456789</dc:identifier>
    <dc:language>en</dc:language>
    <meta property="dcterms:modified">2024-01-01T00:00:00Z</meta>
  </metadata>
  <manifest>
    <item id="nav" href="nav.xhtml" media-type="application/xhtml+xml" properties="nav"/>
    <item id="chapter1" href="chapter1.xhtml" media-type="application/xhtml+xml"/>
  </manifest>
  <spine>
    <itemref idref="chapter1"/>
  </spine>
</package>`

	navContent := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops">
<head>
  <title>Navigation</title>
</head>
<body>
  <nav epub:type="toc">
    <ol>
      <li><a href="chapter1.xhtml">Chapter 1</a></li>
    </ol>
  </nav>
</body>
</html>`

	chapter1Content := `<html>
<head>
  <title>Chapter 1</title>
</head>
<body>
  <h1>Chapter 1</h1>
</body>
</html>`
	return buildEPUBWithOPFAndFiles(t, opfContent, navContent, []testFile{
		{path: "OEBPS/chapter1.xhtml", content: chapter1Content},
	})
}

func createEPUBWithMissingManifestFile(t *testing.T) []byte {
	t.Helper()

	opfContent := `<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="book-id">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>Test Book</dc:title>
    <dc:identifier id="book-id">urn:isbn:123456789</dc:identifier>
    <dc:language>en</dc:language>
    <meta property="dcterms:modified">2024-01-01T00:00:00Z</meta>
  </metadata>
  <manifest>
    <item id="nav" href="nav.xhtml" media-type="application/xhtml+xml" properties="nav"/>
    <item id="chapter1" href="chapter1.xhtml" media-type="application/xhtml+xml"/>
  </manifest>
  <spine>
    <itemref idref="chapter1"/>
  </spine>
</package>`

	navContent := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops">
<head>
  <title>Navigation</title>
</head>
<body>
  <nav epub:type="toc">
    <ol>
      <li><a href="chapter1.xhtml">Chapter 1</a></li>
    </ol>
  </nav>
</body>
</html>`
	return buildEPUBWithOPFAndFiles(t, opfContent, navContent, nil)
}

func createEPUBWithMultipleErrors(t *testing.T) []byte {
	t.Helper()

	opfContent := `<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="book-id">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:identifier id="book-id">urn:isbn:123456789</dc:identifier>
    <dc:language>en</dc:language>
    <meta property="dcterms:modified">2024-01-01T00:00:00Z</meta>
  </metadata>
  <manifest>
    <item id="nav" href="nav.xhtml" media-type="application/xhtml+xml" properties="nav"/>
    <item id="chapter1" href="chapter1.xhtml" media-type="application/xhtml+xml"/>
  </manifest>
  <spine>
    <itemref idref="nav"/>
    <itemref idref="chapter1"/>
  </spine>
</package>`

	navContent := `<?xml version="1.0" encoding="UTF-8"?>
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
  <title>Navigation</title>
</head>
<body>
  <div>No nav element</div>
</body>
</html>`

	content := `<html xmlns="http://www.w3.org/1999/xhtml">
<head>
  <title>Chapter 1</title>
</head>
<body>
  <p>Missing DOCTYPE</p>
</body>
</html>`
	return buildEPUBWithOPFAndFiles(t, opfContent, navContent, []testFile{
		{path: "OEBPS/chapter1.xhtml", content: content},
	})
}

func TestEPUBValidator_ValidateFile_CompleteValid(t *testing.T) {
	validator := NewEPUBValidator()
	epubData := createCompleteValidEPUB(t)

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.epub")

	if err := os.WriteFile(tmpFile, epubData, 0600); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	ctx := context.Background()
	report, err := validator.ValidateFile(ctx, tmpFile)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !report.IsValid {
		t.Errorf("Expected valid EPUB, got invalid")
		for _, e := range report.Errors {
			t.Logf("Error: [%s] %s (file: %s)", e.Code, e.Message, e.Location.Path)
		}
	}

	if len(report.Errors) != 0 {
		t.Errorf("Expected no errors, got %d", len(report.Errors))
	}

	if report.FileType != "EPUB" {
		t.Errorf("Expected FileType 'EPUB', got '%s'", report.FileType)
	}
}

func TestEPUBValidator_ValidateFile_InvalidContainer(t *testing.T) {
	validator := NewEPUBValidator()
	epubData := createEPUBWithInvalidContainer(t)

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.epub")

	if err := os.WriteFile(tmpFile, epubData, 0600); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	ctx := context.Background()
	report, err := validator.ValidateFile(ctx, tmpFile)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if report.IsValid {
		t.Error("Expected invalid EPUB, got valid")
	}

	foundMimetypeError := false
	for _, e := range report.Errors {
		if e.Code == ErrorCodeMimetypeInvalid {
			foundMimetypeError = true
			break
		}
	}

	if !foundMimetypeError {
		t.Errorf("Expected mimetype error, got errors: %v", report.Errors)
	}
}

func TestEPUBValidator_ValidateFile_InvalidOPF(t *testing.T) {
	validator := NewEPUBValidator()
	epubData := createEPUBWithInvalidOPF(t)

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.epub")

	if err := os.WriteFile(tmpFile, epubData, 0600); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	ctx := context.Background()
	report, err := validator.ValidateFile(ctx, tmpFile)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if report.IsValid {
		t.Error("Expected invalid EPUB, got valid")
	}

	foundOPFError := false
	for _, e := range report.Errors {
		if e.Code == ErrorCodeOPFMissingTitle {
			foundOPFError = true
			if e.Location.Path != "OEBPS/content.opf" {
				t.Errorf("Expected error location 'OEBPS/content.opf', got '%s'", e.Location.Path)
			}
			break
		}
	}

	if !foundOPFError {
		t.Errorf("Expected OPF error for missing title, got errors: %v", report.Errors)
	}
}

func TestEPUBValidator_ValidateFile_InvalidNav(t *testing.T) {
	validator := NewEPUBValidator()
	epubData := createEPUBWithInvalidNav(t)

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.epub")

	if err := os.WriteFile(tmpFile, epubData, 0600); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	ctx := context.Background()
	report, err := validator.ValidateFile(ctx, tmpFile)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if report.IsValid {
		t.Error("Expected invalid EPUB, got valid")
	}

	foundNavError := false
	for _, e := range report.Errors {
		if e.Code == ErrorCodeNavMissingNavElement {
			foundNavError = true
			if e.Location.Path != "OEBPS/nav.xhtml" {
				t.Errorf("Expected error location 'OEBPS/nav.xhtml', got '%s'", e.Location.Path)
			}
			break
		}
	}

	if !foundNavError {
		t.Errorf("Expected nav error, got errors: %v", report.Errors)
	}
}

func TestEPUBValidator_ValidateFile_InvalidContent(t *testing.T) {
	validator := NewEPUBValidator()
	epubData := createEPUBWithInvalidContent(t)

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.epub")

	if err := os.WriteFile(tmpFile, epubData, 0600); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	ctx := context.Background()
	report, err := validator.ValidateFile(ctx, tmpFile)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if report.IsValid {
		t.Error("Expected invalid EPUB, got valid")
	}

	foundContentError := false
	for _, e := range report.Errors {
		if e.Code == ErrorCodeContentMissingDoctype ||
			e.Code == ErrorCodeContentInvalidNamespace {
			foundContentError = true
			if e.Location.Path != "OEBPS/chapter1.xhtml" {
				t.Errorf("Expected error location 'OEBPS/chapter1.xhtml', got '%s'", e.Location.Path)
			}
			if e.Details["manifest_id"] != "chapter1" {
				t.Errorf("Expected manifest_id 'chapter1', got '%v'", e.Details["manifest_id"])
			}
			break
		}
	}

	if !foundContentError {
		t.Errorf("Expected content error, got errors: %v", report.Errors)
	}
}

func TestEPUBValidator_ValidateFile_MissingManifestFile(t *testing.T) {
	validator := NewEPUBValidator()
	epubData := createEPUBWithMissingManifestFile(t)

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.epub")

	if err := os.WriteFile(tmpFile, epubData, 0600); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	ctx := context.Background()
	report, err := validator.ValidateFile(ctx, tmpFile)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if report.IsValid {
		t.Error("Expected invalid EPUB, got valid")
	}

	foundMissingFileError := false
	for _, e := range report.Errors {
		if e.Code == ErrorCodeOPFFileNotFound {
			foundMissingFileError = true
			if e.Details["manifest_id"] != "chapter1" {
				t.Errorf("Expected manifest_id 'chapter1', got '%v'", e.Details["manifest_id"])
			}
			break
		}
	}

	if !foundMissingFileError {
		t.Errorf("Expected missing file error, got errors: %v", report.Errors)
	}
}

func TestEPUBValidator_ValidateFile_MultipleErrors(t *testing.T) {
	validator := NewEPUBValidator()
	epubData := createEPUBWithMultipleErrors(t)

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.epub")

	if err := os.WriteFile(tmpFile, epubData, 0600); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	ctx := context.Background()
	report, err := validator.ValidateFile(ctx, tmpFile)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if report.IsValid {
		t.Error("Expected invalid EPUB, got valid")
	}

	if len(report.Errors) < 3 {
		t.Errorf("Expected at least 3 errors (OPF, nav, content), got %d", len(report.Errors))
	}

	errorCodes := make(map[string]bool)
	for _, e := range report.Errors {
		errorCodes[e.Code] = true
	}

	if !errorCodes[ErrorCodeOPFMissingTitle] {
		t.Error("Expected OPF missing title error")
	}

	if !errorCodes[ErrorCodeNavMissingNavElement] {
		t.Error("Expected nav missing element error")
	}

	if !errorCodes[ErrorCodeContentMissingDoctype] {
		t.Error("Expected content missing doctype error")
	}
}

func TestEPUBValidator_ValidateReader(t *testing.T) {
	validator := NewEPUBValidator()
	epubData := createCompleteValidEPUB(t)

	reader := bytes.NewReader(epubData)
	ctx := context.Background()
	report, err := validator.ValidateReader(ctx, reader, int64(len(epubData)))

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !report.IsValid {
		t.Errorf("Expected valid EPUB, got invalid with errors: %v", report.Errors)
	}

	if len(report.Errors) != 0 {
		t.Errorf("Expected no errors, got %d", len(report.Errors))
	}
}

func TestEPUBValidator_ValidateStructure(t *testing.T) {
	validator := NewEPUBValidator()
	epubData := createCompleteValidEPUB(t)

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.epub")

	if err := os.WriteFile(tmpFile, epubData, 0600); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	ctx := context.Background()
	report, err := validator.ValidateStructure(ctx, tmpFile)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !report.IsValid {
		t.Errorf("Expected valid structure, got invalid with errors: %v", report.Errors)
	}

	if len(report.Errors) != 0 {
		t.Errorf("Expected no errors, got %d", len(report.Errors))
	}
}

func TestEPUBValidator_ValidateMetadata(t *testing.T) {
	validator := NewEPUBValidator()
	epubData := createCompleteValidEPUB(t)

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.epub")

	if err := os.WriteFile(tmpFile, epubData, 0600); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	ctx := context.Background()
	report, err := validator.ValidateMetadata(ctx, tmpFile)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !report.IsValid {
		t.Errorf("Expected valid metadata, got invalid with errors: %v", report.Errors)
	}

	if len(report.Errors) != 0 {
		t.Errorf("Expected no errors, got %d", len(report.Errors))
	}
}

func TestEPUBValidator_ValidateMetadata_InvalidOPF(t *testing.T) {
	validator := NewEPUBValidator()
	epubData := createEPUBWithInvalidOPF(t)

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.epub")

	if err := os.WriteFile(tmpFile, epubData, 0600); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	ctx := context.Background()
	report, err := validator.ValidateMetadata(ctx, tmpFile)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if report.IsValid {
		t.Error("Expected invalid metadata, got valid")
	}

	foundOPFError := false
	for _, e := range report.Errors {
		if e.Code == ErrorCodeOPFMissingTitle {
			foundOPFError = true
			break
		}
	}

	if !foundOPFError {
		t.Errorf("Expected OPF error, got errors: %v", report.Errors)
	}
}

func TestEPUBValidator_ValidateContent_CompleteValidation(t *testing.T) {
	validator := NewEPUBValidator()
	epubData := createCompleteValidEPUB(t)

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.epub")

	if err := os.WriteFile(tmpFile, epubData, 0600); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	ctx := context.Background()
	report, err := validator.ValidateContent(ctx, tmpFile)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !report.IsValid {
		t.Errorf("Expected valid content, got invalid")
		for _, e := range report.Errors {
			t.Logf("Error: [%s] %s (file: %s)", e.Code, e.Message, e.Location.Path)
		}
	}

	if len(report.Errors) != 0 {
		t.Errorf("Expected no errors, got %d", len(report.Errors))
	}
}

func TestEPUBValidator_ErrorAggregation(t *testing.T) {
	validator := NewEPUBValidator()
	epubData := createEPUBWithMultipleErrors(t)

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.epub")

	if err := os.WriteFile(tmpFile, epubData, 0600); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	ctx := context.Background()
	report, err := validator.ValidateFile(ctx, tmpFile)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	errorsByFile := make(map[string][]string)
	for _, e := range report.Errors {
		errorsByFile[e.Location.Path] = append(errorsByFile[e.Location.Path], e.Code)
	}

	if len(errorsByFile) == 0 {
		t.Error("Expected errors grouped by file")
	}

	for file, codes := range errorsByFile {
		t.Logf("File: %s, Errors: %v", file, codes)
	}

	for _, e := range report.Errors {
		if e.Location == nil {
			t.Error("Expected all errors to have location information")
		}
		if e.Location.File == "" {
			t.Error("Expected all errors to have file information")
		}
		if e.Code == "" {
			t.Error("Expected all errors to have error codes")
		}
		if e.Severity != "error" {
			t.Errorf("Expected severity 'error', got '%s'", e.Severity)
		}
	}
}

func TestEPUBValidator_NonExistentFile(t *testing.T) {
	validator := NewEPUBValidator()

	ctx := context.Background()
	report, err := validator.ValidateFile(ctx, "/nonexistent/file.epub")

	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}

	if report != nil {
		t.Errorf("Expected nil report for error case, got %v", report)
	}
}

func TestEPUBValidator_ReportMetadata(t *testing.T) {
	validator := NewEPUBValidator()
	epubData := createCompleteValidEPUB(t)

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.epub")

	if err := os.WriteFile(tmpFile, epubData, 0600); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	ctx := context.Background()
	report, err := validator.ValidateFile(ctx, tmpFile)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if report.FilePath != tmpFile {
		t.Errorf("Expected FilePath '%s', got '%s'", tmpFile, report.FilePath)
	}

	if report.FileType != "EPUB" {
		t.Errorf("Expected FileType 'EPUB', got '%s'", report.FileType)
	}

	if report.ValidationTime.IsZero() {
		t.Error("Expected ValidationTime to be set")
	}

	if report.Duration == 0 {
		t.Error("Expected Duration to be set")
	}

	if report.Metadata == nil {
		t.Error("Expected Metadata map to be initialized")
	}
}
