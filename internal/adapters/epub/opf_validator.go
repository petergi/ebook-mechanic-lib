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

const (
	ErrorCodeOPFXMLInvalid            = "EPUB-OPF-001"
	ErrorCodeOPFMissingTitle          = "EPUB-OPF-002"
	ErrorCodeOPFMissingIdentifier     = "EPUB-OPF-003"
	ErrorCodeOPFMissingLanguage       = "EPUB-OPF-004"
	ErrorCodeOPFMissingModified       = "EPUB-OPF-005"
	ErrorCodeOPFInvalidUniqueID       = "EPUB-OPF-006"
	ErrorCodeOPFMissingManifest       = "EPUB-OPF-007"
	ErrorCodeOPFMissingSpine          = "EPUB-OPF-008"
	ErrorCodeOPFMissingNavDocument    = "EPUB-OPF-009"
	ErrorCodeOPFInvalidManifestItem   = "EPUB-OPF-010"
	ErrorCodeOPFInvalidSpineItem      = "EPUB-OPF-011"
	ErrorCodeOPFMissingMetadata       = "EPUB-OPF-012"
	ErrorCodeOPFInvalidPackage        = "EPUB-OPF-013"
	ErrorCodeOPFDuplicateID           = "EPUB-OPF-014"
	ErrorCodeOPFFileNotFound          = "EPUB-OPF-015"
)

const (
	DCNamespace      = "http://purl.org/dc/elements/1.1/"
	OPFNamespace     = "http://www.idpf.org/2007/opf"
	DCTermsProperty  = "dcterms:modified"
)

type Package struct {
	XMLName        xml.Name        `xml:"package"`
	Version        string          `xml:"version,attr"`
	UniqueID       string          `xml:"unique-identifier,attr"`
	Metadata       Metadata        `xml:"metadata"`
	Manifest       Manifest        `xml:"manifest"`
	Spine          Spine           `xml:"spine"`
}

type Metadata struct {
	XMLName    xml.Name         `xml:"metadata"`
	Titles     []DCElement      `xml:"title"`
	Identifiers []DCIdentifier  `xml:"identifier"`
	Languages  []DCElement      `xml:"language"`
	Meta       []MetaElement    `xml:"meta"`
}

type DCElement struct {
	XMLName xml.Name `xml:""`
	Value   string   `xml:",chardata"`
	ID      string   `xml:"id,attr,omitempty"`
}

type DCIdentifier struct {
	XMLName xml.Name `xml:"identifier"`
	Value   string   `xml:",chardata"`
	ID      string   `xml:"id,attr,omitempty"`
}

type MetaElement struct {
	XMLName  xml.Name `xml:"meta"`
	Property string   `xml:"property,attr,omitempty"`
	Value    string   `xml:",chardata"`
}

type Manifest struct {
	XMLName xml.Name       `xml:"manifest"`
	Items   []ManifestItem `xml:"item"`
}

type ManifestItem struct {
	XMLName    xml.Name `xml:"item"`
	ID         string   `xml:"id,attr"`
	Href       string   `xml:"href,attr"`
	MediaType  string   `xml:"media-type,attr"`
	Properties string   `xml:"properties,attr,omitempty"`
}

type Spine struct {
	XMLName xml.Name    `xml:"spine"`
	Items   []SpineItem `xml:"itemref"`
}

type SpineItem struct {
	XMLName xml.Name `xml:"itemref"`
	IDRef   string   `xml:"idref,attr"`
}

type OPFValidationResult struct {
	Valid      bool
	Errors     []ValidationError
	Package    *Package
}

type OPFValidator struct{}

func NewOPFValidator() *OPFValidator {
	return &OPFValidator{}
}

func (v *OPFValidator) ValidateFile(filePath string) (*OPFValidationResult, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return v.ValidateBytes(data)
}

func (v *OPFValidator) ValidateFromEPUB(epubPath string, opfPath string) (*OPFValidationResult, error) {
	file, err := os.Open(epubPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open EPUB file: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	zipReader, err := zip.NewReader(file, fileInfo.Size())
	if err != nil {
		return nil, fmt.Errorf("failed to read ZIP: %w", err)
	}

	var opfFile *zip.File
	for _, f := range zipReader.File {
		if f.Name == opfPath {
			opfFile = f
			break
		}
	}

	if opfFile == nil {
		result := &OPFValidationResult{
			Valid:  false,
			Errors: make([]ValidationError, 0),
		}
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrorCodeOPFFileNotFound,
			Message: fmt.Sprintf("OPF file not found at path: %s", opfPath),
			Details: map[string]interface{}{
				"path": opfPath,
			},
		})
		return result, nil
	}

	rc, err := opfFile.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open OPF file: %w", err)
	}
	defer func() {
		_ = rc.Close()
	}()

	opfData, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("failed to read OPF file: %w", err)
	}

	return v.ValidateBytes(opfData)
}

func (v *OPFValidator) ValidateBytes(data []byte) (*OPFValidationResult, error) {
	result := &OPFValidationResult{
		Valid:  true,
		Errors: make([]ValidationError, 0),
	}

	var pkg Package
	if err := xml.Unmarshal(data, &pkg); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrorCodeOPFXMLInvalid,
			Message: "OPF file is not valid XML",
			Details: map[string]interface{}{
				"error": err.Error(),
			},
		})
		return result, nil
	}

	result.Package = &pkg

	v.validatePackage(&pkg, result)
	v.validateMetadata(&pkg.Metadata, &pkg, result)
	v.validateManifest(&pkg.Manifest, result)
	v.validateSpine(&pkg.Spine, &pkg.Manifest, result)

	return result, nil
}

func (v *OPFValidator) validatePackage(pkg *Package, result *OPFValidationResult) {
	if pkg.Version == "" {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrorCodeOPFInvalidPackage,
			Message: "Package element must have a version attribute",
			Details: map[string]interface{}{},
		})
	}

	if strings.TrimSpace(pkg.UniqueID) == "" {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrorCodeOPFInvalidPackage,
			Message: "Package element must have a unique-identifier attribute",
			Details: map[string]interface{}{},
		})
	}
}

func (v *OPFValidator) validateMetadata(metadata *Metadata, pkg *Package, result *OPFValidationResult) {
	hasTitles := false
	for _, title := range metadata.Titles {
		if strings.TrimSpace(title.Value) != "" {
			hasTitles = true
			break
		}
	}
	if !hasTitles {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrorCodeOPFMissingTitle,
			Message: "OPF metadata must contain at least one dc:title element",
			Details: map[string]interface{}{},
		})
	}

	hasIdentifiers := false
	uniqueIDFound := false
	for _, identifier := range metadata.Identifiers {
		if strings.TrimSpace(identifier.Value) != "" {
			hasIdentifiers = true
			if identifier.ID == pkg.UniqueID {
				uniqueIDFound = true
			}
		}
	}
	if !hasIdentifiers {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrorCodeOPFMissingIdentifier,
			Message: "OPF metadata must contain at least one dc:identifier element",
			Details: map[string]interface{}{},
		})
	}

	if pkg.UniqueID != "" && !uniqueIDFound {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrorCodeOPFInvalidUniqueID,
			Message: fmt.Sprintf("unique-identifier '%s' does not match any dc:identifier id", pkg.UniqueID),
			Details: map[string]interface{}{
				"unique_identifier": pkg.UniqueID,
			},
		})
	}

	hasLanguages := false
	for _, language := range metadata.Languages {
		if strings.TrimSpace(language.Value) != "" {
			hasLanguages = true
			break
		}
	}
	if !hasLanguages {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrorCodeOPFMissingLanguage,
			Message: "OPF metadata must contain at least one dc:language element",
			Details: map[string]interface{}{},
		})
	}

	hasModified := false
	for _, meta := range metadata.Meta {
		if meta.Property == DCTermsProperty && strings.TrimSpace(meta.Value) != "" {
			hasModified = true
			break
		}
	}
	if !hasModified {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrorCodeOPFMissingModified,
			Message: "OPF metadata must contain meta element with property='dcterms:modified'",
			Details: map[string]interface{}{
				"property": DCTermsProperty,
			},
		})
	}
}

func (v *OPFValidator) validateManifest(manifest *Manifest, result *OPFValidationResult) {
	if len(manifest.Items) == 0 {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrorCodeOPFMissingManifest,
			Message: "OPF manifest must contain at least one item",
			Details: map[string]interface{}{},
		})
		return
	}

	seenIDs := make(map[string]bool)
	hasNavDocument := false

	for i, item := range manifest.Items {
		if strings.TrimSpace(item.ID) == "" {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Code:    ErrorCodeOPFInvalidManifestItem,
				Message: fmt.Sprintf("Manifest item at index %d has empty id attribute", i),
				Details: map[string]interface{}{
					"item_index": i,
				},
			})
		} else {
			if seenIDs[item.ID] {
				result.Valid = false
				result.Errors = append(result.Errors, ValidationError{
					Code:    ErrorCodeOPFDuplicateID,
					Message: fmt.Sprintf("Manifest contains duplicate id: %s", item.ID),
					Details: map[string]interface{}{
						"id":         item.ID,
						"item_index": i,
					},
				})
			}
			seenIDs[item.ID] = true
		}

		if strings.TrimSpace(item.Href) == "" {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Code:    ErrorCodeOPFInvalidManifestItem,
				Message: fmt.Sprintf("Manifest item at index %d has empty href attribute", i),
				Details: map[string]interface{}{
					"item_index": i,
					"id":         item.ID,
				},
			})
		}

		if strings.TrimSpace(item.MediaType) == "" {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Code:    ErrorCodeOPFInvalidManifestItem,
				Message: fmt.Sprintf("Manifest item at index %d has empty media-type attribute", i),
				Details: map[string]interface{}{
					"item_index": i,
					"id":         item.ID,
				},
			})
		}

		if strings.Contains(item.Properties, "nav") {
			hasNavDocument = true
		}
	}

	if !hasNavDocument {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrorCodeOPFMissingNavDocument,
			Message: "OPF manifest must contain at least one item with properties='nav'",
			Details: map[string]interface{}{},
		})
	}
}

func (v *OPFValidator) validateSpine(spine *Spine, manifest *Manifest, result *OPFValidationResult) {
	if len(spine.Items) == 0 {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrorCodeOPFMissingSpine,
			Message: "OPF spine must contain at least one itemref",
			Details: map[string]interface{}{},
		})
		return
	}

	manifestIDs := make(map[string]bool)
	for _, item := range manifest.Items {
		manifestIDs[item.ID] = true
	}

	for i, item := range spine.Items {
		if strings.TrimSpace(item.IDRef) == "" {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Code:    ErrorCodeOPFInvalidSpineItem,
				Message: fmt.Sprintf("Spine itemref at index %d has empty idref attribute", i),
				Details: map[string]interface{}{
					"item_index": i,
				},
			})
		} else if !manifestIDs[item.IDRef] {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Code:    ErrorCodeOPFInvalidSpineItem,
				Message: fmt.Sprintf("Spine itemref at index %d references non-existent manifest id: %s", i, item.IDRef),
				Details: map[string]interface{}{
					"item_index": i,
					"idref":      item.IDRef,
				},
			})
		}
	}
}

func (v *OPFValidator) ValidateBytesReader(reader io.ReaderAt, size int64) (*OPFValidationResult, error) {
	buf := make([]byte, size)
	_, err := reader.ReadAt(buf, 0)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read data: %w", err)
	}
	return v.ValidateBytes(buf)
}

func (v *OPFValidator) ValidateBytesBuffer(data []byte) (*OPFValidationResult, error) {
	reader := bytes.NewReader(data)
	return v.ValidateBytesReader(reader, int64(len(data)))
}
