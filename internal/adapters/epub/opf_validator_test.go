package epub

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func createMinimalValidOPF() string {
	return `<?xml version="1.0" encoding="UTF-8"?>
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
}

func TestOPFValidator_ValidateBytes_MinimalValid(t *testing.T) {
	validator := NewOPFValidator()
	opfData := createMinimalValidOPF()

	result, err := validator.ValidateBytes([]byte(opfData))

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !result.Valid {
		t.Errorf("Expected valid OPF, got invalid with errors: %v", result.Errors)
	}

	if len(result.Errors) != 0 {
		t.Errorf("Expected no errors, got %d: %v", len(result.Errors), result.Errors)
	}

	if result.Package == nil {
		t.Fatal("Expected package to be parsed, got nil")
	}

	if result.Package.Version != "3.0" {
		t.Errorf("Expected version '3.0', got '%s'", result.Package.Version)
	}

	if result.Package.UniqueID != "book-id" {
		t.Errorf("Expected unique-identifier 'book-id', got '%s'", result.Package.UniqueID)
	}
}

func TestOPFValidator_ValidateBytes(t *testing.T) {
	tests := []struct {
		name          string
		opfContent    string
		expectValid   bool
		expectedCode  string
		checkError    func(*testing.T, []ValidationError)
	}{
		{
			name: "invalid XML",
			opfContent: `<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="book-id">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>Test</dc:title>
    <dc:identifier id="book-id">123</dc:identifier>
    <dc:language>en</dc:language>
    <meta property="dcterms:modified">2024-01-01T00:00:00Z</meta>
  </metadata>
  <manifest>
    <item id="nav" href="nav.xhtml" media-type="application/xhtml+xml" properties="nav"/>
  </manifest>
  <spine>
    <itemref idref="nav"/>
  </spine>
</unclosed>`,
			expectValid:  false,
			expectedCode: ErrorCodeOPFXMLInvalid,
		},
		{
			name: "missing dc:title",
			opfContent: `<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="book-id">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:identifier id="book-id">123</dc:identifier>
    <dc:language>en</dc:language>
    <meta property="dcterms:modified">2024-01-01T00:00:00Z</meta>
  </metadata>
  <manifest>
    <item id="nav" href="nav.xhtml" media-type="application/xhtml+xml" properties="nav"/>
  </manifest>
  <spine>
    <itemref idref="nav"/>
  </spine>
</package>`,
			expectValid:  false,
			expectedCode: ErrorCodeOPFMissingTitle,
		},
		{
			name: "empty dc:title",
			opfContent: `<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="book-id">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title></dc:title>
    <dc:identifier id="book-id">123</dc:identifier>
    <dc:language>en</dc:language>
    <meta property="dcterms:modified">2024-01-01T00:00:00Z</meta>
  </metadata>
  <manifest>
    <item id="nav" href="nav.xhtml" media-type="application/xhtml+xml" properties="nav"/>
  </manifest>
  <spine>
    <itemref idref="nav"/>
  </spine>
</package>`,
			expectValid:  false,
			expectedCode: ErrorCodeOPFMissingTitle,
		},
		{
			name: "missing dc:identifier",
			opfContent: `<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="book-id">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>Test Book</dc:title>
    <dc:language>en</dc:language>
    <meta property="dcterms:modified">2024-01-01T00:00:00Z</meta>
  </metadata>
  <manifest>
    <item id="nav" href="nav.xhtml" media-type="application/xhtml+xml" properties="nav"/>
  </manifest>
  <spine>
    <itemref idref="nav"/>
  </spine>
</package>`,
			expectValid:  false,
			expectedCode: ErrorCodeOPFMissingIdentifier,
		},
		{
			name: "missing dc:language",
			opfContent: `<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="book-id">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>Test Book</dc:title>
    <dc:identifier id="book-id">123</dc:identifier>
    <meta property="dcterms:modified">2024-01-01T00:00:00Z</meta>
  </metadata>
  <manifest>
    <item id="nav" href="nav.xhtml" media-type="application/xhtml+xml" properties="nav"/>
  </manifest>
  <spine>
    <itemref idref="nav"/>
  </spine>
</package>`,
			expectValid:  false,
			expectedCode: ErrorCodeOPFMissingLanguage,
		},
		{
			name: "missing dcterms:modified",
			opfContent: `<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="book-id">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>Test Book</dc:title>
    <dc:identifier id="book-id">123</dc:identifier>
    <dc:language>en</dc:language>
  </metadata>
  <manifest>
    <item id="nav" href="nav.xhtml" media-type="application/xhtml+xml" properties="nav"/>
  </manifest>
  <spine>
    <itemref idref="nav"/>
  </spine>
</package>`,
			expectValid:  false,
			expectedCode: ErrorCodeOPFMissingModified,
		},
		{
			name: "invalid unique-identifier",
			opfContent: `<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="wrong-id">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>Test Book</dc:title>
    <dc:identifier id="book-id">123</dc:identifier>
    <dc:language>en</dc:language>
    <meta property="dcterms:modified">2024-01-01T00:00:00Z</meta>
  </metadata>
  <manifest>
    <item id="nav" href="nav.xhtml" media-type="application/xhtml+xml" properties="nav"/>
  </manifest>
  <spine>
    <itemref idref="nav"/>
  </spine>
</package>`,
			expectValid:  false,
			expectedCode: ErrorCodeOPFInvalidUniqueID,
			checkError: func(t *testing.T, errors []ValidationError) {
				found := false
				for _, e := range errors {
					if e.Code == ErrorCodeOPFInvalidUniqueID {
						found = true
						if e.Details["unique_identifier"] != "wrong-id" {
							t.Errorf("Expected unique_identifier detail to be 'wrong-id', got '%v'", e.Details["unique_identifier"])
						}
					}
				}
				if !found {
					t.Error("Expected error with code EPUB-OPF-006")
				}
			},
		},
		{
			name: "missing manifest",
			opfContent: `<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="book-id">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>Test Book</dc:title>
    <dc:identifier id="book-id">123</dc:identifier>
    <dc:language>en</dc:language>
    <meta property="dcterms:modified">2024-01-01T00:00:00Z</meta>
  </metadata>
  <manifest>
  </manifest>
  <spine>
    <itemref idref="nav"/>
  </spine>
</package>`,
			expectValid:  false,
			expectedCode: ErrorCodeOPFMissingManifest,
		},
		{
			name: "missing spine",
			opfContent: `<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="book-id">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>Test Book</dc:title>
    <dc:identifier id="book-id">123</dc:identifier>
    <dc:language>en</dc:language>
    <meta property="dcterms:modified">2024-01-01T00:00:00Z</meta>
  </metadata>
  <manifest>
    <item id="nav" href="nav.xhtml" media-type="application/xhtml+xml" properties="nav"/>
  </manifest>
  <spine>
  </spine>
</package>`,
			expectValid:  false,
			expectedCode: ErrorCodeOPFMissingSpine,
		},
		{
			name: "missing nav document",
			opfContent: `<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="book-id">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>Test Book</dc:title>
    <dc:identifier id="book-id">123</dc:identifier>
    <dc:language>en</dc:language>
    <meta property="dcterms:modified">2024-01-01T00:00:00Z</meta>
  </metadata>
  <manifest>
    <item id="chapter1" href="chapter1.xhtml" media-type="application/xhtml+xml"/>
  </manifest>
  <spine>
    <itemref idref="chapter1"/>
  </spine>
</package>`,
			expectValid:  false,
			expectedCode: ErrorCodeOPFMissingNavDocument,
		},
		{
			name: "manifest item with empty id",
			opfContent: `<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="book-id">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>Test Book</dc:title>
    <dc:identifier id="book-id">123</dc:identifier>
    <dc:language>en</dc:language>
    <meta property="dcterms:modified">2024-01-01T00:00:00Z</meta>
  </metadata>
  <manifest>
    <item id="nav" href="nav.xhtml" media-type="application/xhtml+xml" properties="nav"/>
    <item id="" href="chapter1.xhtml" media-type="application/xhtml+xml"/>
  </manifest>
  <spine>
    <itemref idref="nav"/>
  </spine>
</package>`,
			expectValid:  false,
			expectedCode: ErrorCodeOPFInvalidManifestItem,
		},
		{
			name: "manifest item with empty href",
			opfContent: `<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="book-id">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>Test Book</dc:title>
    <dc:identifier id="book-id">123</dc:identifier>
    <dc:language>en</dc:language>
    <meta property="dcterms:modified">2024-01-01T00:00:00Z</meta>
  </metadata>
  <manifest>
    <item id="nav" href="nav.xhtml" media-type="application/xhtml+xml" properties="nav"/>
    <item id="chapter1" href="" media-type="application/xhtml+xml"/>
  </manifest>
  <spine>
    <itemref idref="nav"/>
  </spine>
</package>`,
			expectValid:  false,
			expectedCode: ErrorCodeOPFInvalidManifestItem,
		},
		{
			name: "manifest item with empty media-type",
			opfContent: `<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="book-id">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>Test Book</dc:title>
    <dc:identifier id="book-id">123</dc:identifier>
    <dc:language>en</dc:language>
    <meta property="dcterms:modified">2024-01-01T00:00:00Z</meta>
  </metadata>
  <manifest>
    <item id="nav" href="nav.xhtml" media-type="application/xhtml+xml" properties="nav"/>
    <item id="chapter1" href="chapter1.xhtml" media-type=""/>
  </manifest>
  <spine>
    <itemref idref="nav"/>
  </spine>
</package>`,
			expectValid:  false,
			expectedCode: ErrorCodeOPFInvalidManifestItem,
		},
		{
			name: "duplicate manifest item id",
			opfContent: `<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="book-id">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>Test Book</dc:title>
    <dc:identifier id="book-id">123</dc:identifier>
    <dc:language>en</dc:language>
    <meta property="dcterms:modified">2024-01-01T00:00:00Z</meta>
  </metadata>
  <manifest>
    <item id="nav" href="nav.xhtml" media-type="application/xhtml+xml" properties="nav"/>
    <item id="chapter1" href="chapter1.xhtml" media-type="application/xhtml+xml"/>
    <item id="chapter1" href="chapter2.xhtml" media-type="application/xhtml+xml"/>
  </manifest>
  <spine>
    <itemref idref="chapter1"/>
  </spine>
</package>`,
			expectValid:  false,
			expectedCode: ErrorCodeOPFDuplicateID,
		},
		{
			name: "spine itemref with empty idref",
			opfContent: `<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="book-id">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>Test Book</dc:title>
    <dc:identifier id="book-id">123</dc:identifier>
    <dc:language>en</dc:language>
    <meta property="dcterms:modified">2024-01-01T00:00:00Z</meta>
  </metadata>
  <manifest>
    <item id="nav" href="nav.xhtml" media-type="application/xhtml+xml" properties="nav"/>
  </manifest>
  <spine>
    <itemref idref=""/>
  </spine>
</package>`,
			expectValid:  false,
			expectedCode: ErrorCodeOPFInvalidSpineItem,
		},
		{
			name: "spine itemref references non-existent manifest item",
			opfContent: `<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="book-id">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>Test Book</dc:title>
    <dc:identifier id="book-id">123</dc:identifier>
    <dc:language>en</dc:language>
    <meta property="dcterms:modified">2024-01-01T00:00:00Z</meta>
  </metadata>
  <manifest>
    <item id="nav" href="nav.xhtml" media-type="application/xhtml+xml" properties="nav"/>
  </manifest>
  <spine>
    <itemref idref="nonexistent"/>
  </spine>
</package>`,
			expectValid:  false,
			expectedCode: ErrorCodeOPFInvalidSpineItem,
		},
		{
			name: "missing version attribute",
			opfContent: `<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" unique-identifier="book-id">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>Test Book</dc:title>
    <dc:identifier id="book-id">123</dc:identifier>
    <dc:language>en</dc:language>
    <meta property="dcterms:modified">2024-01-01T00:00:00Z</meta>
  </metadata>
  <manifest>
    <item id="nav" href="nav.xhtml" media-type="application/xhtml+xml" properties="nav"/>
  </manifest>
  <spine>
    <itemref idref="nav"/>
  </spine>
</package>`,
			expectValid:  false,
			expectedCode: ErrorCodeOPFInvalidPackage,
		},
		{
			name: "missing unique-identifier attribute",
			opfContent: `<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>Test Book</dc:title>
    <dc:identifier id="book-id">123</dc:identifier>
    <dc:language>en</dc:language>
    <meta property="dcterms:modified">2024-01-01T00:00:00Z</meta>
  </metadata>
  <manifest>
    <item id="nav" href="nav.xhtml" media-type="application/xhtml+xml" properties="nav"/>
  </manifest>
  <spine>
    <itemref idref="nav"/>
  </spine>
</package>`,
			expectValid:  false,
			expectedCode: ErrorCodeOPFInvalidPackage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewOPFValidator()
			result, err := validator.ValidateBytes([]byte(tt.opfContent))

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if result.Valid != tt.expectValid {
				t.Errorf("Expected valid=%v, got valid=%v with errors: %v", tt.expectValid, result.Valid, result.Errors)
			}

			if !tt.expectValid {
				foundExpectedError := false
				for _, e := range result.Errors {
					if e.Code == tt.expectedCode {
						foundExpectedError = true
						break
					}
				}
				if !foundExpectedError {
					t.Errorf("Expected error code %s, got errors: %v", tt.expectedCode, result.Errors)
				}

				if tt.checkError != nil {
					tt.checkError(t, result.Errors)
				}
			}
		})
	}
}

func TestOPFValidator_ValidateFile(t *testing.T) {
	validator := NewOPFValidator()
	opfData := createMinimalValidOPF()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "content.opf")

	if err := os.WriteFile(tmpFile, []byte(opfData), 0644); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	result, err := validator.ValidateFile(tmpFile)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !result.Valid {
		t.Errorf("Expected valid OPF, got invalid with errors: %v", result.Errors)
	}
}

func TestOPFValidator_ValidateFile_NonExistent(t *testing.T) {
	validator := NewOPFValidator()

	result, err := validator.ValidateFile("/nonexistent/content.opf")

	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}

	if result != nil {
		t.Errorf("Expected nil result for error case, got %v", result)
	}
}

func TestOPFValidator_ValidateFromEPUB(t *testing.T) {
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
	if _, err := opfWriter.Write([]byte(createMinimalValidOPF())); err != nil {
		t.Fatalf("Failed to write content.opf: %v", err)
	}

	if err := zipWriter.Close(); err != nil {
		t.Fatalf("Failed to close zip writer: %v", err)
	}

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.epub")
	if err := os.WriteFile(tmpFile, buf.Bytes(), 0644); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	validator := NewOPFValidator()
	result, err := validator.ValidateFromEPUB(tmpFile, "OEBPS/content.opf")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !result.Valid {
		t.Errorf("Expected valid OPF, got invalid with errors: %v", result.Errors)
	}
}

func TestOPFValidator_ValidateFromEPUB_MissingOPF(t *testing.T) {
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

	if err := zipWriter.Close(); err != nil {
		t.Fatalf("Failed to close zip writer: %v", err)
	}

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.epub")
	if err := os.WriteFile(tmpFile, buf.Bytes(), 0644); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	validator := NewOPFValidator()
	result, err := validator.ValidateFromEPUB(tmpFile, "OEBPS/content.opf")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("Expected invalid result, got valid")
	}

	foundError := false
	for _, e := range result.Errors {
		if e.Code == ErrorCodeOPFFileNotFound {
			foundError = true
			break
		}
	}

	if !foundError {
		t.Errorf("Expected error code %s, got errors: %v", ErrorCodeOPFFileNotFound, result.Errors)
	}
}

func TestOPFValidator_ComplexValid(t *testing.T) {
	opfData := `<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="book-id">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>Complex Test Book</dc:title>
    <dc:title>Subtitle</dc:title>
    <dc:identifier id="book-id">urn:isbn:123456789</dc:identifier>
    <dc:identifier>urn:uuid:12345</dc:identifier>
    <dc:language>en</dc:language>
    <dc:language>fr</dc:language>
    <dc:creator>John Doe</dc:creator>
    <meta property="dcterms:modified">2024-01-01T00:00:00Z</meta>
    <meta property="schema:accessMode">textual</meta>
  </metadata>
  <manifest>
    <item id="nav" href="nav.xhtml" media-type="application/xhtml+xml" properties="nav"/>
    <item id="chapter1" href="chapter1.xhtml" media-type="application/xhtml+xml"/>
    <item id="chapter2" href="chapter2.xhtml" media-type="application/xhtml+xml"/>
    <item id="css" href="styles.css" media-type="text/css"/>
    <item id="cover" href="cover.jpg" media-type="image/jpeg"/>
  </manifest>
  <spine>
    <itemref idref="chapter1"/>
    <itemref idref="chapter2"/>
  </spine>
</package>`

	validator := NewOPFValidator()
	result, err := validator.ValidateBytes([]byte(opfData))

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !result.Valid {
		t.Errorf("Expected valid OPF, got invalid with errors: %v", result.Errors)
	}

	if len(result.Errors) != 0 {
		t.Errorf("Expected no errors, got %d: %v", len(result.Errors), result.Errors)
	}

	if result.Package == nil {
		t.Fatal("Expected package to be parsed")
	}

	if len(result.Package.Metadata.Titles) != 2 {
		t.Errorf("Expected 2 titles, got %d", len(result.Package.Metadata.Titles))
	}

	if len(result.Package.Metadata.Identifiers) != 2 {
		t.Errorf("Expected 2 identifiers, got %d", len(result.Package.Metadata.Identifiers))
	}

	if len(result.Package.Metadata.Languages) != 2 {
		t.Errorf("Expected 2 languages, got %d", len(result.Package.Metadata.Languages))
	}

	if len(result.Package.Manifest.Items) != 5 {
		t.Errorf("Expected 5 manifest items, got %d", len(result.Package.Manifest.Items))
	}

	if len(result.Package.Spine.Items) != 2 {
		t.Errorf("Expected 2 spine items, got %d", len(result.Package.Spine.Items))
	}
}

func TestOPFErrorCodes(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"XML Invalid", ErrorCodeOPFXMLInvalid, "EPUB-OPF-001"},
		{"Missing Title", ErrorCodeOPFMissingTitle, "EPUB-OPF-002"},
		{"Missing Identifier", ErrorCodeOPFMissingIdentifier, "EPUB-OPF-003"},
		{"Missing Language", ErrorCodeOPFMissingLanguage, "EPUB-OPF-004"},
		{"Missing Modified", ErrorCodeOPFMissingModified, "EPUB-OPF-005"},
		{"Invalid Unique ID", ErrorCodeOPFInvalidUniqueID, "EPUB-OPF-006"},
		{"Missing Manifest", ErrorCodeOPFMissingManifest, "EPUB-OPF-007"},
		{"Missing Spine", ErrorCodeOPFMissingSpine, "EPUB-OPF-008"},
		{"Missing Nav Document", ErrorCodeOPFMissingNavDocument, "EPUB-OPF-009"},
		{"Invalid Manifest Item", ErrorCodeOPFInvalidManifestItem, "EPUB-OPF-010"},
		{"Invalid Spine Item", ErrorCodeOPFInvalidSpineItem, "EPUB-OPF-011"},
		{"Missing Metadata", ErrorCodeOPFMissingMetadata, "EPUB-OPF-012"},
		{"Invalid Package", ErrorCodeOPFInvalidPackage, "EPUB-OPF-013"},
		{"Duplicate ID", ErrorCodeOPFDuplicateID, "EPUB-OPF-014"},
		{"File Not Found", ErrorCodeOPFFileNotFound, "EPUB-OPF-015"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.code != tt.expected {
				t.Errorf("Expected error code %s, got %s", tt.expected, tt.code)
			}
		})
	}
}
