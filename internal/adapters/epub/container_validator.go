// Package epub provides EPUB validation and repair adapters.
package epub

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
)

// Container validation error codes.
const (
	ErrorCodeZIPInvalid          = "EPUB-CONTAINER-001"
	ErrorCodeMimetypeInvalid     = "EPUB-CONTAINER-002"
	ErrorCodeMimetypeNotFirst    = "EPUB-CONTAINER-003"
	ErrorCodeContainerXMLMissing = "EPUB-CONTAINER-004"
	ErrorCodeContainerXMLInvalid = "EPUB-CONTAINER-005"
)

// EPUB container constants.
const (
	ExpectedMimetype = "application/epub+zip"
	MimetypeFilename = "mimetype"
	ContainerXMLPath = "META-INF/container.xml"
)

// ContainerXML models META-INF/container.xml.
type ContainerXML struct {
	XMLName   xml.Name   `xml:"container"`
	Version   string     `xml:"version,attr"`
	Rootfiles []Rootfile `xml:"rootfiles>rootfile"`
}

// Rootfile describes a single rootfile entry in container.xml.
type Rootfile struct {
	FullPath  string `xml:"full-path,attr"`
	MediaType string `xml:"media-type,attr"`
}

// ValidationError captures container validation issues.
type ValidationError struct {
	Code    string
	Message string
	Details map[string]interface{}
}

// ValidationResult aggregates container validation findings.
type ValidationResult struct {
	Valid     bool
	Errors    []ValidationError
	Rootfiles []Rootfile
}

// ContainerValidator validates EPUB container structure.
type ContainerValidator struct{}

// NewContainerValidator returns a new container validator.
func NewContainerValidator() *ContainerValidator {
	return &ContainerValidator{}
}

// ValidateFile validates an EPUB file on disk.
func (v *ContainerValidator) ValidateFile(filePath string) (*ValidationResult, error) {
	file, err := os.Open(filePath) //nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	return v.Validate(file, fileInfo.Size())
}

// Validate validates EPUB container structure from a reader.
func (v *ContainerValidator) Validate(reader io.ReaderAt, size int64) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid:  true,
		Errors: make([]ValidationError, 0),
	}

	zipReader, zipErr := zip.NewReader(reader, size)
	if zipErr != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrorCodeZIPInvalid,
			Message: "File is not a valid ZIP archive",
			Details: map[string]interface{}{
				"error": zipErr.Error(),
			},
		})
		return result, nil //nolint:nilerr
	}

	if err := v.validateMimetype(zipReader, result); err != nil {
		return nil, err
	}

	if err := v.validateContainerXML(zipReader, result); err != nil {
		return nil, err
	}

	return result, nil
}

func (v *ContainerValidator) validateMimetype(zipReader *zip.Reader, result *ValidationResult) error {
	if len(zipReader.File) == 0 {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrorCodeMimetypeInvalid,
			Message: "EPUB container is empty",
			Details: map[string]interface{}{},
		})
		return nil
	}

	firstFile := zipReader.File[0]
	if firstFile.Name != MimetypeFilename {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrorCodeMimetypeNotFirst,
			Message: fmt.Sprintf("mimetype file must be first in ZIP archive, found '%s' instead", firstFile.Name),
			Details: map[string]interface{}{
				"first_file": firstFile.Name,
			},
		})
		return nil
	}

	if firstFile.Method != zip.Store {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrorCodeMimetypeInvalid,
			Message: "mimetype file must be stored uncompressed",
			Details: map[string]interface{}{
				"compression_method": firstFile.Method,
			},
		})
	}

	rc, err := firstFile.Open()
	if err != nil {
		return fmt.Errorf("failed to open mimetype file: %w", err)
	}
	defer func() {
		_ = rc.Close()
	}()

	mimetypeBytes, err := io.ReadAll(rc)
	if err != nil {
		return fmt.Errorf("failed to read mimetype file: %w", err)
	}

	mimetype := string(mimetypeBytes)
	if mimetype != ExpectedMimetype {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrorCodeMimetypeInvalid,
			Message: fmt.Sprintf("mimetype file must contain exactly '%s'", ExpectedMimetype),
			Details: map[string]interface{}{
				"expected": ExpectedMimetype,
				"found":    mimetype,
			},
		})
	}

	return nil
}

func (v *ContainerValidator) validateContainerXML(zipReader *zip.Reader, result *ValidationResult) error {
	var containerFile *zip.File
	for _, file := range zipReader.File {
		if file.Name == ContainerXMLPath {
			containerFile = file
			break
		}
	}

	if containerFile == nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrorCodeContainerXMLMissing,
			Message: fmt.Sprintf("Required file '%s' is missing", ContainerXMLPath),
			Details: map[string]interface{}{
				"expected_path": ContainerXMLPath,
			},
		})
		return nil
	}

	rc, err := containerFile.Open()
	if err != nil {
		return fmt.Errorf("failed to open container.xml: %w", err)
	}
	defer func() {
		_ = rc.Close()
	}()

	containerBytes, err := io.ReadAll(rc)
	if err != nil {
		return fmt.Errorf("failed to read container.xml: %w", err)
	}

	var container ContainerXML
	if err := xml.Unmarshal(containerBytes, &container); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrorCodeContainerXMLInvalid,
			Message: "META-INF/container.xml is not valid XML",
			Details: map[string]interface{}{
				"error": err.Error(),
			},
		})
		return nil
	}

	if len(container.Rootfiles) == 0 {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrorCodeContainerXMLInvalid,
			Message: "META-INF/container.xml must contain at least one rootfile",
			Details: map[string]interface{}{},
		})
		return nil
	}

	for i, rootfile := range container.Rootfiles {
		if strings.TrimSpace(rootfile.FullPath) == "" {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Code:    ErrorCodeContainerXMLInvalid,
				Message: fmt.Sprintf("Rootfile at index %d has empty full-path attribute", i),
				Details: map[string]interface{}{
					"rootfile_index": i,
				},
			})
		}
	}

	result.Rootfiles = container.Rootfiles

	return nil
}

// ValidateBytes validates EPUB container structure from in-memory data.
func (v *ContainerValidator) ValidateBytes(data []byte) (*ValidationResult, error) {
	reader := bytes.NewReader(data)
	return v.Validate(reader, int64(len(data)))
}
