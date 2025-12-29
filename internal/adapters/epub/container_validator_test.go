package epub

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

type epubBuildOptions struct {
	mimetype      string
	storeMimetype bool
	containerXML  *string
	preMimetype   func(*zip.Writer) error
}

func buildEPUB(t *testing.T, opts epubBuildOptions) []byte {
	t.Helper()

	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	if opts.preMimetype != nil {
		if err := opts.preMimetype(zipWriter); err != nil {
			t.Fatalf("Failed to create pre-mimetype files: %v", err)
		}
	}

	var mimetypeWriter io.Writer
	var err error
	if opts.storeMimetype {
		mimetypeWriter, err = zipWriter.CreateHeader(&zip.FileHeader{
			Name:   MimetypeFilename,
			Method: zip.Store,
		})
	} else {
		mimetypeWriter, err = zipWriter.Create(MimetypeFilename)
	}
	if err != nil {
		t.Fatalf("Failed to create mimetype header: %v", err)
	}
	if _, err := mimetypeWriter.Write([]byte(opts.mimetype)); err != nil {
		t.Fatalf("Failed to write mimetype: %v", err)
	}

	if opts.containerXML != nil {
		containerWriter, err := zipWriter.Create(ContainerXMLPath)
		if err != nil {
			t.Fatalf("Failed to create container.xml: %v", err)
		}
		if _, err := containerWriter.Write([]byte(*opts.containerXML)); err != nil {
			t.Fatalf("Failed to write container.xml: %v", err)
		}
	}

	if err := zipWriter.Close(); err != nil {
		t.Fatalf("Failed to close zip writer: %v", err)
	}

	return buf.Bytes()
}

func createValidEPUB(t *testing.T) []byte {
	t.Helper()

	containerXML := `<?xml version="1.0" encoding="UTF-8"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`
	return buildEPUB(t, epubBuildOptions{
		mimetype:      ExpectedMimetype,
		storeMimetype: true,
		containerXML:  &containerXML,
	})
}

func createEPUBWithInvalidZIP(t *testing.T) []byte {
	t.Helper()
	return []byte("This is not a valid ZIP file")
}

func createEPUBWithWrongMimetypeContent(t *testing.T) []byte {
	t.Helper()

	containerXML := `<?xml version="1.0" encoding="UTF-8"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`
	return buildEPUB(t, epubBuildOptions{
		mimetype:      "application/wrong",
		storeMimetype: true,
		containerXML:  &containerXML,
	})
}

func createEPUBWithCompressedMimetype(t *testing.T) []byte {
	t.Helper()

	containerXML := `<?xml version="1.0" encoding="UTF-8"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`
	return buildEPUB(t, epubBuildOptions{
		mimetype:      ExpectedMimetype,
		storeMimetype: false,
		containerXML:  &containerXML,
	})
}

func createEPUBWithMimetypeNotFirst(t *testing.T) []byte {
	t.Helper()

	containerXML := `<?xml version="1.0" encoding="UTF-8"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`
	return buildEPUB(t, epubBuildOptions{
		mimetype:      ExpectedMimetype,
		storeMimetype: true,
		containerXML:  &containerXML,
		preMimetype: func(zipWriter *zip.Writer) error {
			otherWriter, err := zipWriter.Create("other.txt")
			if err != nil {
				return err
			}
			_, err = otherWriter.Write([]byte("some content"))
			return err
		},
	})
}

func createEPUBWithoutContainerXML(t *testing.T) []byte {
	t.Helper()
	return buildEPUB(t, epubBuildOptions{
		mimetype:      ExpectedMimetype,
		storeMimetype: true,
	})
}

func createEPUBWithInvalidContainerXML(t *testing.T) []byte {
	t.Helper()

	containerXML := "This is not valid XML <>"
	return buildEPUB(t, epubBuildOptions{
		mimetype:      ExpectedMimetype,
		storeMimetype: true,
		containerXML:  &containerXML,
	})
}

func createEPUBWithNoRootfiles(t *testing.T) []byte {
	t.Helper()

	containerXML := `<?xml version="1.0" encoding="UTF-8"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
  </rootfiles>
</container>`
	return buildEPUB(t, epubBuildOptions{
		mimetype:      ExpectedMimetype,
		storeMimetype: true,
		containerXML:  &containerXML,
	})
}

func createEPUBWithEmptyRootfilePath(t *testing.T) []byte {
	t.Helper()

	containerXML := `<?xml version="1.0" encoding="UTF-8"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`
	return buildEPUB(t, epubBuildOptions{
		mimetype:      ExpectedMimetype,
		storeMimetype: true,
		containerXML:  &containerXML,
	})
}

func createEPUBWithMultipleRootfiles(t *testing.T) []byte {
	t.Helper()

	containerXML := `<?xml version="1.0" encoding="UTF-8"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
    <rootfile full-path="OEBPS/fallback.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`
	return buildEPUB(t, epubBuildOptions{
		mimetype:      ExpectedMimetype,
		storeMimetype: true,
		containerXML:  &containerXML,
	})
}

func TestContainerValidator_ValidateBytes_ValidEPUB(t *testing.T) {
	validator := NewContainerValidator()
	epubData := createValidEPUB(t)

	result, err := validator.ValidateBytes(epubData)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !result.Valid {
		t.Errorf("Expected valid EPUB, got invalid with errors: %v", result.Errors)
	}

	if len(result.Errors) != 0 {
		t.Errorf("Expected no errors, got %d: %v", len(result.Errors), result.Errors)
	}

	if len(result.Rootfiles) != 1 {
		t.Errorf("Expected 1 rootfile, got %d", len(result.Rootfiles))
	}

	if len(result.Rootfiles) > 0 && result.Rootfiles[0].FullPath != "OEBPS/content.opf" {
		t.Errorf("Expected rootfile path 'OEBPS/content.opf', got '%s'", result.Rootfiles[0].FullPath)
	}
}

func TestContainerValidator_ValidateBytes_InvalidZIP(t *testing.T) {
	validator := NewContainerValidator()
	epubData := createEPUBWithInvalidZIP(t)

	result, err := validator.ValidateBytes(epubData)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("Expected invalid EPUB, got valid")
	}

	if len(result.Errors) == 0 {
		t.Error("Expected errors, got none")
	}

	foundError := false
	for _, e := range result.Errors {
		if e.Code == ErrorCodeZIPInvalid {
			foundError = true
			break
		}
	}

	if !foundError {
		t.Errorf("Expected error code %s, got errors: %v", ErrorCodeZIPInvalid, result.Errors)
	}
}

func TestContainerValidator_ValidateBytes_WrongMimetypeContent(t *testing.T) {
	validator := NewContainerValidator()
	epubData := createEPUBWithWrongMimetypeContent(t)

	result, err := validator.ValidateBytes(epubData)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("Expected invalid EPUB, got valid")
	}

	foundError := false
	for _, e := range result.Errors {
		if e.Code == ErrorCodeMimetypeInvalid {
			foundError = true
			if e.Details["expected"] != ExpectedMimetype {
				t.Errorf("Expected 'expected' detail to be '%s', got '%v'", ExpectedMimetype, e.Details["expected"])
			}
			if e.Details["found"] != "application/wrong" {
				t.Errorf("Expected 'found' detail to be 'application/wrong', got '%v'", e.Details["found"])
			}
			break
		}
	}

	if !foundError {
		t.Errorf("Expected error code %s, got errors: %v", ErrorCodeMimetypeInvalid, result.Errors)
	}
}

func TestContainerValidator_ValidateBytes_CompressedMimetype(t *testing.T) {
	validator := NewContainerValidator()
	epubData := createEPUBWithCompressedMimetype(t)

	result, err := validator.ValidateBytes(epubData)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("Expected invalid EPUB, got valid")
	}

	foundError := false
	for _, e := range result.Errors {
		if e.Code == ErrorCodeMimetypeInvalid && e.Message == "mimetype file must be stored uncompressed" {
			foundError = true
			break
		}
	}

	if !foundError {
		t.Errorf("Expected error code %s for compressed mimetype, got errors: %v", ErrorCodeMimetypeInvalid, result.Errors)
	}
}

func TestContainerValidator_ValidateBytes_MimetypeNotFirst(t *testing.T) {
	validator := NewContainerValidator()
	epubData := createEPUBWithMimetypeNotFirst(t)

	result, err := validator.ValidateBytes(epubData)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("Expected invalid EPUB, got valid")
	}

	foundError := false
	for _, e := range result.Errors {
		if e.Code == ErrorCodeMimetypeNotFirst {
			foundError = true
			if e.Details["first_file"] != "other.txt" {
				t.Errorf("Expected 'first_file' detail to be 'other.txt', got '%v'", e.Details["first_file"])
			}
			break
		}
	}

	if !foundError {
		t.Errorf("Expected error code %s, got errors: %v", ErrorCodeMimetypeNotFirst, result.Errors)
	}
}

func TestContainerValidator_ValidateBytes_MissingContainerXML(t *testing.T) {
	validator := NewContainerValidator()
	epubData := createEPUBWithoutContainerXML(t)

	result, err := validator.ValidateBytes(epubData)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("Expected invalid EPUB, got valid")
	}

	foundError := false
	for _, e := range result.Errors {
		if e.Code == ErrorCodeContainerXMLMissing {
			foundError = true
			if e.Details["expected_path"] != ContainerXMLPath {
				t.Errorf("Expected 'expected_path' detail to be '%s', got '%v'", ContainerXMLPath, e.Details["expected_path"])
			}
			break
		}
	}

	if !foundError {
		t.Errorf("Expected error code %s, got errors: %v", ErrorCodeContainerXMLMissing, result.Errors)
	}
}

func TestContainerValidator_ValidateBytes_InvalidContainerXML(t *testing.T) {
	validator := NewContainerValidator()
	epubData := createEPUBWithInvalidContainerXML(t)

	result, err := validator.ValidateBytes(epubData)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("Expected invalid EPUB, got valid")
	}

	foundError := false
	for _, e := range result.Errors {
		if e.Code == ErrorCodeContainerXMLInvalid {
			foundError = true
			break
		}
	}

	if !foundError {
		t.Errorf("Expected error code %s, got errors: %v", ErrorCodeContainerXMLInvalid, result.Errors)
	}
}

func TestContainerValidator_ValidateBytes_NoRootfiles(t *testing.T) {
	validator := NewContainerValidator()
	epubData := createEPUBWithNoRootfiles(t)

	result, err := validator.ValidateBytes(epubData)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("Expected invalid EPUB, got valid")
	}

	foundError := false
	for _, e := range result.Errors {
		if e.Code == ErrorCodeContainerXMLInvalid && e.Message == "META-INF/container.xml must contain at least one rootfile" {
			foundError = true
			break
		}
	}

	if !foundError {
		t.Errorf("Expected error code %s for no rootfiles, got errors: %v", ErrorCodeContainerXMLInvalid, result.Errors)
	}
}

func TestContainerValidator_ValidateBytes_EmptyRootfilePath(t *testing.T) {
	validator := NewContainerValidator()
	epubData := createEPUBWithEmptyRootfilePath(t)

	result, err := validator.ValidateBytes(epubData)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("Expected invalid EPUB, got valid")
	}

	foundError := false
	for _, e := range result.Errors {
		if e.Code == ErrorCodeContainerXMLInvalid && e.Details["rootfile_index"] == 0 {
			foundError = true
			break
		}
	}

	if !foundError {
		t.Errorf("Expected error code %s for empty rootfile path, got errors: %v", ErrorCodeContainerXMLInvalid, result.Errors)
	}
}

func TestContainerValidator_ValidateBytes_MultipleRootfiles(t *testing.T) {
	validator := NewContainerValidator()
	epubData := createEPUBWithMultipleRootfiles(t)

	result, err := validator.ValidateBytes(epubData)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !result.Valid {
		t.Errorf("Expected valid EPUB with multiple rootfiles, got invalid with errors: %v", result.Errors)
	}

	if len(result.Rootfiles) != 2 {
		t.Errorf("Expected 2 rootfiles, got %d", len(result.Rootfiles))
	}

	if len(result.Rootfiles) > 0 && result.Rootfiles[0].FullPath != "OEBPS/content.opf" {
		t.Errorf("Expected first rootfile path 'OEBPS/content.opf', got '%s'", result.Rootfiles[0].FullPath)
	}

	if len(result.Rootfiles) > 1 && result.Rootfiles[1].FullPath != "OEBPS/fallback.opf" {
		t.Errorf("Expected second rootfile path 'OEBPS/fallback.opf', got '%s'", result.Rootfiles[1].FullPath)
	}
}

func TestContainerValidator_ValidateFile(t *testing.T) {
	validator := NewContainerValidator()
	epubData := createValidEPUB(t)

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.epub")

	if err := os.WriteFile(tmpFile, epubData, 0600); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}

	result, err := validator.ValidateFile(tmpFile)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !result.Valid {
		t.Errorf("Expected valid EPUB, got invalid with errors: %v", result.Errors)
	}

	if len(result.Errors) != 0 {
		t.Errorf("Expected no errors, got %d: %v", len(result.Errors), result.Errors)
	}
}

func TestContainerValidator_ValidateFile_NonExistent(t *testing.T) {
	validator := NewContainerValidator()

	result, err := validator.ValidateFile("/nonexistent/file.epub")

	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}

	if result != nil {
		t.Errorf("Expected nil result for error case, got %v", result)
	}
}

func TestErrorCodes(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected string
	}{
		{"ZIP Invalid", ErrorCodeZIPInvalid, "EPUB-CONTAINER-001"},
		{"Mimetype Invalid", ErrorCodeMimetypeInvalid, "EPUB-CONTAINER-002"},
		{"Mimetype Not First", ErrorCodeMimetypeNotFirst, "EPUB-CONTAINER-003"},
		{"Container XML Missing", ErrorCodeContainerXMLMissing, "EPUB-CONTAINER-004"},
		{"Container XML Invalid", ErrorCodeContainerXMLInvalid, "EPUB-CONTAINER-005"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.code != tt.expected {
				t.Errorf("Expected error code %s, got %s", tt.expected, tt.code)
			}
		})
	}
}
