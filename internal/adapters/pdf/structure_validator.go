package pdf

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/unidoc/unipdf/v3/core"
	"github.com/unidoc/unipdf/v3/model"
)

const (
	ErrorCodePDFHeader001    = "PDF-HEADER-001"
	ErrorCodePDFHeader002    = "PDF-HEADER-002"
	ErrorCodePDFTrailer001   = "PDF-TRAILER-001"
	ErrorCodePDFTrailer002   = "PDF-TRAILER-002"
	ErrorCodePDFTrailer003   = "PDF-TRAILER-003"
	ErrorCodePDFXref001      = "PDF-XREF-001"
	ErrorCodePDFXref002      = "PDF-XREF-002"
	ErrorCodePDFXref003      = "PDF-XREF-003"
	ErrorCodePDFCatalog001   = "PDF-CATALOG-001"
	ErrorCodePDFCatalog002   = "PDF-CATALOG-002"
	ErrorCodePDFCatalog003   = "PDF-CATALOG-003"
	ErrorCodePDFStructure012 = "PDF-STRUCTURE-012"
)

type ValidationError struct {
	Code    string
	Message string
	Details map[string]interface{}
}

type StructureValidationResult struct {
	Valid  bool
	Errors []ValidationError
}

type StructureValidator struct{}

func NewStructureValidator() *StructureValidator {
	return &StructureValidator{}
}

func (v *StructureValidator) ValidateFile(filePath string) (*StructureValidationResult, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return v.ValidateBytes(data)
}

func (v *StructureValidator) ValidateReader(reader io.Reader) (*StructureValidationResult, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read from reader: %w", err)
	}
	return v.ValidateBytes(data)
}

func (v *StructureValidator) ValidateBytes(data []byte) (*StructureValidationResult, error) {
	result := &StructureValidationResult{
		Valid:  true,
		Errors: make([]ValidationError, 0),
	}

	v.validateHeader(data, result)
	v.validateTrailer(data, result)

	if len(result.Errors) > 0 {
		result.Valid = false
		return result, nil
	}

	if err := v.validateWithUnipdf(data, result); err != nil {
		return result, err
	}

	result.Valid = len(result.Errors) == 0
	return result, nil
}

func (v *StructureValidator) validateHeader(data []byte, result *StructureValidationResult) {
	if len(data) == 0 {
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrorCodePDFHeader001,
			Message: "File is empty",
			Details: map[string]interface{}{},
		})
		return
	}

	headerPattern := regexp.MustCompile(`^%PDF-1\.[0-7]`)
	if !headerPattern.Match(data) {
		if !bytes.HasPrefix(data, []byte("%PDF-")) {
			result.Errors = append(result.Errors, ValidationError{
				Code:    ErrorCodePDFHeader001,
				Message: "Invalid or missing PDF header",
				Details: map[string]interface{}{
					"expected": "%PDF-1.x where x=0-7",
				},
			})
		} else {
			versionStr := string(data[:10])
			if len(data) >= 10 {
				versionStr = strings.TrimSpace(versionStr)
			}
			result.Errors = append(result.Errors, ValidationError{
				Code:    ErrorCodePDFHeader002,
				Message: "Invalid PDF version number",
				Details: map[string]interface{}{
					"expected": "1.0 through 1.7",
					"found":    versionStr,
				},
			})
		}
	}
}

func (v *StructureValidator) validateTrailer(data []byte, result *StructureValidationResult) {
	if len(data) < 5 {
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrorCodePDFTrailer003,
			Message: "File too small to contain valid trailer",
			Details: map[string]interface{}{},
		})
		return
	}

	eofPattern := []byte("%%EOF")
	lastBytes := data
	if len(data) > 1024 {
		lastBytes = data[len(data)-1024:]
	}

	eofIndex := bytes.LastIndex(lastBytes, eofPattern)
	if eofIndex == -1 {
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrorCodePDFTrailer003,
			Message: "Missing %%EOF marker",
			Details: map[string]interface{}{
				"expected": "%%EOF at end of file",
			},
		})
		return
	}

	startxrefPattern := regexp.MustCompile(`startxref\s+(\d+)\s+%%EOF`)
	if !startxrefPattern.Match(lastBytes) {
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrorCodePDFTrailer001,
			Message: "Invalid or missing startxref",
			Details: map[string]interface{}{
				"expected": "startxref <offset> before %%EOF",
			},
		})
	}
}

func (v *StructureValidator) validateWithUnipdf(data []byte, result *StructureValidationResult) error {
	reader := bytes.NewReader(data)
	pdfReader, err := model.NewPdfReader(reader)
	if err != nil {
		if strings.Contains(err.Error(), "xref") || strings.Contains(err.Error(), "cross") {
			result.Errors = append(result.Errors, ValidationError{
				Code:    ErrorCodePDFXref001,
				Message: "Invalid or damaged cross-reference table",
				Details: map[string]interface{}{
					"error": err.Error(),
				},
			})
			return nil
		}
		if strings.Contains(err.Error(), "trailer") {
			result.Errors = append(result.Errors, ValidationError{
				Code:    ErrorCodePDFTrailer002,
				Message: "Invalid trailer dictionary",
				Details: map[string]interface{}{
					"error": err.Error(),
				},
			})
			return nil
		}
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrorCodePDFStructure012,
			Message: "Failed to parse PDF structure",
			Details: map[string]interface{}{
				"error": err.Error(),
			},
		})
		return nil
	}

	v.validateCrossReference(pdfReader, result)
	v.validateCatalog(pdfReader, result)
	v.validateObjectNumbering(pdfReader, result)

	return nil
}

func (v *StructureValidator) validateCrossReference(pdfReader *model.PdfReader, result *StructureValidationResult) {
	parser := pdfReader.GetParser()
	if parser == nil {
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrorCodePDFXref001,
			Message: "Unable to access PDF parser",
			Details: map[string]interface{}{},
		})
		return
	}

	xrefTable := parser.GetXrefTable()
	if xrefTable == nil {
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrorCodePDFXref001,
			Message: "Missing cross-reference table",
			Details: map[string]interface{}{},
		})
		return
	}

	if xrefTable.ObjectCount() == 0 {
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrorCodePDFXref002,
			Message: "Empty cross-reference table",
			Details: map[string]interface{}{},
		})
		return
	}

	offsets := make(map[int64][]int)
	for _, objNum := range xrefTable.GetObjectNums() {
		xrefObj, err := xrefTable.Get(objNum)
		if err != nil {
			continue
		}
		if xrefObj.Offset > 0 {
			offsets[xrefObj.Offset] = append(offsets[xrefObj.Offset], objNum)
		}
	}

	for offset, objNums := range offsets {
		if len(objNums) > 1 {
			result.Errors = append(result.Errors, ValidationError{
				Code:    ErrorCodePDFXref003,
				Message: "Cross-reference table has overlapping entries",
				Details: map[string]interface{}{
					"offset":  offset,
					"objects": objNums,
				},
			})
		}
	}
}

func (v *StructureValidator) validateCatalog(pdfReader *model.PdfReader, result *StructureValidationResult) {
	catalog := pdfReader.GetCatalog()
	if catalog == nil {
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrorCodePDFCatalog001,
			Message: "Missing or invalid catalog object",
			Details: map[string]interface{}{},
		})
		return
	}

	catalogDict := catalog.ToPdfObject()
	if catalogDict == nil {
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrorCodePDFCatalog001,
			Message: "Catalog object is not a dictionary",
			Details: map[string]interface{}{},
		})
		return
	}

	dict, ok := core.GetDict(catalogDict)
	if !ok {
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrorCodePDFCatalog001,
			Message: "Catalog object is not a valid dictionary",
			Details: map[string]interface{}{},
		})
		return
	}

	typeObj := dict.Get("Type")
	if typeObj == nil {
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrorCodePDFCatalog002,
			Message: "Catalog missing /Type entry",
			Details: map[string]interface{}{},
		})
	} else {
		typeName, ok := core.GetName(typeObj)
		if !ok || typeName.String() != "Catalog" {
			result.Errors = append(result.Errors, ValidationError{
				Code:    ErrorCodePDFCatalog002,
				Message: "Catalog /Type must be /Catalog",
				Details: map[string]interface{}{
					"found": typeObj.String(),
				},
			})
		}
	}

	pagesObj := dict.Get("Pages")
	if pagesObj == nil {
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrorCodePDFCatalog003,
			Message: "Catalog missing /Pages entry",
			Details: map[string]interface{}{},
		})
		return
	}

	pagesIndirect, ok := core.GetIndirect(pagesObj)
	if ok {
		pagesObj = pagesIndirect
	}

	if pagesObj == nil {
		result.Errors = append(result.Errors, ValidationError{
			Code:    ErrorCodePDFCatalog003,
			Message: "Catalog /Pages entry is invalid",
			Details: map[string]interface{}{},
		})
	}
}

func (v *StructureValidator) validateObjectNumbering(pdfReader *model.PdfReader, result *StructureValidationResult) {
	parser := pdfReader.GetParser()
	if parser == nil {
		return
	}

	xrefTable := parser.GetXrefTable()
	if xrefTable == nil {
		return
	}

	seenObjects := make(map[string]bool)
	for _, objNum := range xrefTable.GetObjectNums() {
		xrefObj, err := xrefTable.Get(objNum)
		if err != nil {
			continue
		}

		key := fmt.Sprintf("%d_%d", objNum, xrefObj.Generation)
		if seenObjects[key] {
			result.Errors = append(result.Errors, ValidationError{
				Code:    ErrorCodePDFStructure012,
				Message: "Duplicate object number/generation pair",
				Details: map[string]interface{}{
					"object_number": objNum,
					"generation":    xrefObj.Generation,
				},
			})
		}
		seenObjects[key] = true
	}
}
