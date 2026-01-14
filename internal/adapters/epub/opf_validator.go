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

// OPF validation error codes.
const (
	ErrorCodeOPFXMLInvalid          = "EPUB-OPF-001"
	ErrorCodeOPFMissingTitle        = "EPUB-OPF-002"
	ErrorCodeOPFMissingIdentifier   = "EPUB-OPF-003"
	ErrorCodeOPFMissingLanguage     = "EPUB-OPF-004"
	ErrorCodeOPFMissingModified     = "EPUB-OPF-005"
	ErrorCodeOPFInvalidUniqueID     = "EPUB-OPF-006"
	ErrorCodeOPFMissingManifest     = "EPUB-OPF-007"
	ErrorCodeOPFMissingSpine        = "EPUB-OPF-008"
	ErrorCodeOPFMissingNavDocument  = "EPUB-OPF-009"
	ErrorCodeOPFInvalidManifestItem = "EPUB-OPF-010"
	ErrorCodeOPFInvalidSpineItem    = "EPUB-OPF-011"
	ErrorCodeOPFMissingMetadata     = "EPUB-OPF-012"
	ErrorCodeOPFInvalidPackage      = "EPUB-OPF-013"
	ErrorCodeOPFDuplicateID         = "EPUB-OPF-014"
	ErrorCodeOPFFileNotFound        = "EPUB-OPF-015"
	ErrorCodeOPFMissingNCX          = "EPUB-OPF-016"
	ErrorCodeOPFInvalidSpineTOC     = "EPUB-OPF-017"
)

// OPF namespace constants.
const (
	DCNamespace     = "http://purl.org/dc/elements/1.1/"
	OPFNamespace    = "http://www.idpf.org/2007/opf"
	DCTermsProperty = "dcterms:modified"
)

// Package models an OPF package document.
type Package struct {
	XMLName  xml.Name `xml:"package"`
	Version  string   `xml:"version,attr"`
	UniqueID string   `xml:"unique-identifier,attr"`
	Metadata Metadata `xml:"metadata"`
	Manifest Manifest `xml:"manifest"`
	Spine    Spine    `xml:"spine"`
}

// Metadata captures OPF metadata entries.
type Metadata struct {
	XMLName     xml.Name       `xml:"metadata"`
	Titles      []DCElement    `xml:"title"`
	Identifiers []DCIdentifier `xml:"identifier"`
	Languages   []DCElement    `xml:"language"`
	Meta        []MetaElement  `xml:"meta"`
}

// DCElement represents a Dublin Core element value.
type DCElement struct {
	XMLName xml.Name `xml:""`
	Value   string   `xml:",chardata"`
	ID      string   `xml:"id,attr,omitempty"`
}

// DCIdentifier represents a Dublin Core identifier value.
type DCIdentifier struct {
	XMLName xml.Name `xml:"identifier"`
	Value   string   `xml:",chardata"`
	ID      string   `xml:"id,attr,omitempty"`
}

// MetaElement represents an OPF meta element.
type MetaElement struct {
	XMLName  xml.Name `xml:"meta"`
	Property string   `xml:"property,attr,omitempty"`
	Value    string   `xml:",chardata"`
}

// Manifest captures the list of content items.
type Manifest struct {
	XMLName xml.Name       `xml:"manifest"`
	Items   []ManifestItem `xml:"item"`
}

// ManifestItem describes a single manifest entry.
type ManifestItem struct {
	XMLName    xml.Name `xml:"item"`
	ID         string   `xml:"id,attr"`
	Href       string   `xml:"href,attr"`
	MediaType  string   `xml:"media-type,attr"`
	Properties string   `xml:"properties,attr,omitempty"`
}

// Spine describes the reading order.
type Spine struct {
	XMLName xml.Name    `xml:"spine"`
	Toc     string      `xml:"toc,attr,omitempty"`
	Items   []SpineItem `xml:"itemref"`
}

// SpineItem references a manifest item in the spine.
type SpineItem struct {
	XMLName xml.Name `xml:"itemref"`
	IDRef   string   `xml:"idref,attr"`
}

// OPFValidationResult aggregates OPF validation findings.
type OPFValidationResult struct {
	Valid   bool
	Errors  []ValidationError
	Package *Package
}

// OPFValidator validates OPF package documents.
type OPFValidator struct{}

// NewOPFValidator returns a new OPF validator.
func NewOPFValidator() *OPFValidator {
	return &OPFValidator{}
}

// ValidateFile validates an OPF file from disk.
func (v *OPFValidator) ValidateFile(filePath string) (*OPFValidationResult, error) {
	data, err := os.ReadFile(filePath) //nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return v.ValidateBytes(data)
}

// ValidateFromEPUB validates an OPF file inside an EPUB.
func (v *OPFValidator) ValidateFromEPUB(epubPath string, opfPath string) (*OPFValidationResult, error) {
	file, err := os.Open(epubPath) //nolint:gosec
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

// ValidateBytes validates OPF data from memory.
func (v *OPFValidator) ValidateBytes(data []byte) (*OPFValidationResult, error) {
	result := &OPFValidationResult{
		Valid:  true,
		Errors: make([]ValidationError, 0),
	}

	var pkg Package
	if unmarshalErr := xml.Unmarshal(data, &pkg); unmarshalErr != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrorCodeOPFXMLInvalid,
			Message: "OPF file is not valid XML",
			Details: map[string]interface{}{
				"error": unmarshalErr.Error(),
			},
		})
		return result, nil //nolint:nilerr
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

	if result.Package != nil && isEPUB2(result.Package.Version) {
		v.validateNCXReference(spine, manifest, result)
	}
}

func (v *OPFValidator) validateNCXReference(spine *Spine, manifest *Manifest, result *OPFValidationResult) {
	ncxIDs := make([]string, 0)
	for _, item := range manifest.Items {
		if strings.EqualFold(strings.TrimSpace(item.MediaType), "application/x-dtbncx+xml") {
			if strings.TrimSpace(item.ID) != "" {
				ncxIDs = append(ncxIDs, item.ID)
			}
		}
	}

	if len(ncxIDs) == 0 {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrorCodeOPFMissingNCX,
			Message: "EPUB 2 package must include an NCX item in the manifest",
			Details: map[string]interface{}{
				"media_type": "application/x-dtbncx+xml",
			},
		})
		return
	}

	toc := strings.TrimSpace(spine.Toc)
	if toc == "" || !stringInSlice(toc, ncxIDs) {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrorCodeOPFInvalidSpineTOC,
			Message: "EPUB 2 spine must reference the NCX item via the toc attribute",
			Details: map[string]interface{}{
				"toc":     toc,
				"ncx_id":  ncxIDs[0],
				"ncx_ids": ncxIDs,
			},
		})
	}
}

func isEPUB2(version string) bool {
	return strings.HasPrefix(strings.TrimSpace(version), "2")
}

func stringInSlice(value string, values []string) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
}

// ValidateBytesReader validates OPF data from an io.ReaderAt.
func (v *OPFValidator) ValidateBytesReader(reader io.ReaderAt, size int64) (*OPFValidationResult, error) {
	buf := make([]byte, size)
	_, err := reader.ReadAt(buf, 0)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read data: %w", err)
	}
	return v.ValidateBytes(buf)
}

// ValidateBytesBuffer validates OPF data from a byte slice.
func (v *OPFValidator) ValidateBytesBuffer(data []byte) (*OPFValidationResult, error) {
	reader := bytes.NewReader(data)
	return v.ValidateBytesReader(reader, int64(len(data)))
}
