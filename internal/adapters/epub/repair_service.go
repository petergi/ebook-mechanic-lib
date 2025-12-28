package epub

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/example/project/internal/domain"
	"github.com/example/project/internal/ports"
	"golang.org/x/net/html"
)

const (
	repairSuffix = "_repaired.epub"
)

type RepairServiceImpl struct {
	containerValidator *ContainerValidator
	opfValidator       *OPFValidator
	navValidator       *NavValidator
	contentValidator   *ContentValidator
}

func NewRepairService() ports.EPUBRepairService {
	return &RepairServiceImpl{
		containerValidator: NewContainerValidator(),
		opfValidator:       NewOPFValidator(),
		navValidator:       NewNavValidator(),
		contentValidator:   NewContentValidator(),
	}
}

func (r *RepairServiceImpl) Preview(ctx context.Context, report *domain.ValidationReport) (*ports.RepairPreview, error) {
	if report == nil {
		return nil, fmt.Errorf("validation report is nil")
	}

	preview := &ports.RepairPreview{
		Actions:        make([]ports.RepairAction, 0),
		CanAutoRepair:  true,
		EstimatedTime:  1000,
		BackupRequired: true,
		Warnings:       make([]string, 0),
	}

	for _, err := range report.Errors {
		actions := r.generateRepairActions(&err)
		for _, action := range actions {
			if !action.Automated {
				preview.CanAutoRepair = false
				preview.Warnings = append(preview.Warnings,
					fmt.Sprintf("Manual intervention may be required for: %s", action.Description))
			}
			preview.Actions = append(preview.Actions, action)
		}
	}

	if len(preview.Actions) == 0 {
		preview.BackupRequired = false
	}

	return preview, nil
}

func (r *RepairServiceImpl) Apply(ctx context.Context, filePath string, preview *ports.RepairPreview) (*ports.RepairResult, error) {
	outputPath := r.generateOutputPath(filePath)
	return r.ApplyWithBackup(ctx, filePath, preview, outputPath)
}

func (r *RepairServiceImpl) ApplyWithBackup(ctx context.Context, filePath string, preview *ports.RepairPreview, backupPath string) (*ports.RepairResult, error) {
	result := &ports.RepairResult{
		Success:        false,
		ActionsApplied: make([]ports.RepairAction, 0),
		BackupPath:     backupPath,
	}

	if preview == nil || len(preview.Actions) == 0 {
		result.Error = fmt.Errorf("no repair actions to apply")
		return result, nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		result.Error = fmt.Errorf("failed to open EPUB: %w", err)
		return result, nil
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		result.Error = fmt.Errorf("failed to stat EPUB: %w", err)
		return result, nil
	}

	zipReader, err := zip.NewReader(file, fileInfo.Size())
	if err != nil {
		result.Error = fmt.Errorf("failed to read EPUB as ZIP: %w", err)
		return result, nil
	}

	outputFile, err := os.Create(backupPath)
	if err != nil {
		result.Error = fmt.Errorf("failed to create output file: %w", err)
		return result, nil
	}
	defer outputFile.Close()

	zipWriter := zip.NewWriter(outputFile)
	defer zipWriter.Close()

	repairContext := &repairContext{
		actions:   preview.Actions,
		zipReader: zipReader,
		zipWriter: zipWriter,
		applied:   make([]ports.RepairAction, 0),
	}

	if err := r.applyRepairs(ctx, repairContext); err != nil {
		result.Error = fmt.Errorf("repair failed: %w", err)
		return result, nil
	}

	result.Success = true
	result.ActionsApplied = repairContext.applied

	return result, nil
}

func (r *RepairServiceImpl) CanRepair(ctx context.Context, err *domain.ValidationError) bool {
	if err == nil {
		return false
	}

	switch err.Code {
	case ErrorCodeMimetypeInvalid,
		ErrorCodeMimetypeNotFirst,
		ErrorCodeContainerXMLMissing,
		ErrorCodeContentMissingDoctype,
		ErrorCodeOPFMissingTitle,
		ErrorCodeOPFMissingIdentifier,
		ErrorCodeOPFMissingLanguage,
		ErrorCodeOPFMissingModified:
		return true
	default:
		return false
	}
}

func (r *RepairServiceImpl) CreateBackup(ctx context.Context, filePath string, backupPath string) error {
	sourceFile, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(backupPath)
	if err != nil {
		return fmt.Errorf("failed to create backup file: %w", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return nil
}

func (r *RepairServiceImpl) RestoreBackup(ctx context.Context, backupPath string, originalPath string) error {
	return r.CreateBackup(ctx, backupPath, originalPath)
}

func (r *RepairServiceImpl) RepairStructure(ctx context.Context, filePath string) (*ports.RepairResult, error) {
	validator := NewEPUBValidator()
	report, err := validator.ValidateStructure(ctx, filePath)
	if err != nil {
		return nil, err
	}

	preview, err := r.Preview(ctx, report)
	if err != nil {
		return nil, err
	}

	return r.Apply(ctx, filePath, preview)
}

func (r *RepairServiceImpl) RepairMetadata(ctx context.Context, filePath string) (*ports.RepairResult, error) {
	validator := NewEPUBValidator()
	report, err := validator.ValidateMetadata(ctx, filePath)
	if err != nil {
		return nil, err
	}

	preview, err := r.Preview(ctx, report)
	if err != nil {
		return nil, err
	}

	return r.Apply(ctx, filePath, preview)
}

func (r *RepairServiceImpl) RepairContent(ctx context.Context, filePath string) (*ports.RepairResult, error) {
	validator := NewEPUBValidator()
	report, err := validator.ValidateContent(ctx, filePath)
	if err != nil {
		return nil, err
	}

	preview, err := r.Preview(ctx, report)
	if err != nil {
		return nil, err
	}

	return r.Apply(ctx, filePath, preview)
}

func (r *RepairServiceImpl) generateRepairActions(err *domain.ValidationError) []ports.RepairAction {
	actions := make([]ports.RepairAction, 0)

	switch err.Code {
	case ErrorCodeMimetypeInvalid:
		actions = append(actions, ports.RepairAction{
			Type:        "fix_mimetype_content",
			Description: "Normalize mimetype file content to 'application/epub+zip'",
			Target:      "mimetype",
			Details: map[string]interface{}{
				"expected": ExpectedMimetype,
			},
			Automated: true,
		})

	case ErrorCodeMimetypeNotFirst:
		actions = append(actions, ports.RepairAction{
			Type:        "fix_mimetype_order",
			Description: "Rebuild ZIP to ensure mimetype is first and uncompressed",
			Target:      "mimetype",
			Details:     map[string]interface{}{},
			Automated:   true,
		})

	case ErrorCodeContainerXMLMissing:
		actions = append(actions, ports.RepairAction{
			Type:        "create_container_xml",
			Description: "Create minimal META-INF/container.xml with default OPF path",
			Target:      "META-INF/container.xml",
			Details: map[string]interface{}{
				"default_opf_path": "OEBPS/content.opf",
			},
			Automated: true,
		})

	case ErrorCodeContentMissingDoctype:
		actions = append(actions, ports.RepairAction{
			Type:        "add_doctype",
			Description: "Add HTML5 DOCTYPE declaration to content document",
			Target:      err.Location.Path,
			Details:     map[string]interface{}{},
			Automated:   true,
		})

	case ErrorCodeOPFMissingTitle:
		actions = append(actions, ports.RepairAction{
			Type:        "add_metadata_title",
			Description: "Add placeholder title to OPF metadata",
			Target:      err.Location.Path,
			Details: map[string]interface{}{
				"placeholder": "Untitled",
			},
			Automated: true,
		})

	case ErrorCodeOPFMissingIdentifier:
		actions = append(actions, ports.RepairAction{
			Type:        "add_metadata_identifier",
			Description: "Add generated UUID identifier to OPF metadata",
			Target:      err.Location.Path,
			Details: map[string]interface{}{
				"id_prefix": "bookid",
			},
			Automated: true,
		})

	case ErrorCodeOPFMissingLanguage:
		actions = append(actions, ports.RepairAction{
			Type:        "add_metadata_language",
			Description: "Add default language (en) to OPF metadata",
			Target:      err.Location.Path,
			Details: map[string]interface{}{
				"language": "en",
			},
			Automated: true,
		})

	case ErrorCodeOPFMissingModified:
		actions = append(actions, ports.RepairAction{
			Type:        "add_metadata_modified",
			Description: "Add dcterms:modified date to OPF metadata",
			Target:      err.Location.Path,
			Details: map[string]interface{}{
				"date": time.Now().Format("2006-01-02T15:04:05Z"),
			},
			Automated: true,
		})

	default:
		actions = append(actions, ports.RepairAction{
			Type:        "manual_review",
			Description: fmt.Sprintf("Requires manual review: %s", err.Message),
			Target:      err.Location.Path,
			Details:     err.Details,
			Automated:   false,
		})
	}

	return actions
}

type repairContext struct {
	actions   []ports.RepairAction
	zipReader *zip.Reader
	zipWriter *zip.Writer
	applied   []ports.RepairAction
}

func (r *RepairServiceImpl) applyRepairs(ctx context.Context, repairCtx *repairContext) error {
	actionsByType := make(map[string][]ports.RepairAction)
	for _, action := range repairCtx.actions {
		if action.Automated {
			actionsByType[action.Type] = append(actionsByType[action.Type], action)
		}
	}

	needsMimetypeFix := len(actionsByType["fix_mimetype_content"]) > 0 ||
		len(actionsByType["fix_mimetype_order"]) > 0
	needsContainerFix := len(actionsByType["create_container_xml"]) > 0

	filesProcessed := make(map[string]bool)

	if needsMimetypeFix {
		if err := r.writeMimetype(repairCtx.zipWriter); err != nil {
			return err
		}
		filesProcessed["mimetype"] = true
		for _, action := range append(actionsByType["fix_mimetype_content"], actionsByType["fix_mimetype_order"]...) {
			repairCtx.applied = append(repairCtx.applied, action)
		}
	}

	contentActions := make(map[string]ports.RepairAction)
	for _, action := range actionsByType["add_doctype"] {
		contentActions[action.Target] = action
	}

	opfActions := make(map[string][]ports.RepairAction)
	for _, actionType := range []string{"add_metadata_title", "add_metadata_identifier",
		"add_metadata_language", "add_metadata_modified"} {
		for _, action := range actionsByType[actionType] {
			opfActions[action.Target] = append(opfActions[action.Target], action)
		}
	}

	for _, f := range repairCtx.zipReader.File {
		if filesProcessed[f.Name] {
			continue
		}

		if needsContainerFix && f.Name == ContainerXMLPath {
			if err := r.writeContainerXML(repairCtx.zipWriter); err != nil {
				return err
			}
			filesProcessed[f.Name] = true
			for _, action := range actionsByType["create_container_xml"] {
				repairCtx.applied = append(repairCtx.applied, action)
			}
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("failed to open file %s: %w", f.Name, err)
		}

		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", f.Name, err)
		}

		if action, exists := contentActions[f.Name]; exists {
			data, err = r.addDoctype(data)
			if err != nil {
				return fmt.Errorf("failed to add DOCTYPE to %s: %w", f.Name, err)
			}
			repairCtx.applied = append(repairCtx.applied, action)
		}

		if actions, exists := opfActions[f.Name]; exists && len(actions) > 0 {
			data, err = r.repairOPFMetadata(data, actions)
			if err != nil {
				return fmt.Errorf("failed to repair OPF metadata in %s: %w", f.Name, err)
			}
			repairCtx.applied = append(repairCtx.applied, actions...)
		}

		w, err := repairCtx.zipWriter.Create(f.Name)
		if err != nil {
			return fmt.Errorf("failed to create file %s in output: %w", f.Name, err)
		}

		if _, err := w.Write(data); err != nil {
			return fmt.Errorf("failed to write file %s: %w", f.Name, err)
		}

		filesProcessed[f.Name] = true
	}

	if needsContainerFix && !filesProcessed[ContainerXMLPath] {
		if err := r.writeContainerXML(repairCtx.zipWriter); err != nil {
			return err
		}
		for _, action := range actionsByType["create_container_xml"] {
			repairCtx.applied = append(repairCtx.applied, action)
		}
	}

	return nil
}

func (r *RepairServiceImpl) writeMimetype(zipWriter *zip.Writer) error {
	header := &zip.FileHeader{
		Name:   MimetypeFilename,
		Method: zip.Store,
	}

	w, err := zipWriter.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("failed to create mimetype file: %w", err)
	}

	_, err = w.Write([]byte(ExpectedMimetype))
	if err != nil {
		return fmt.Errorf("failed to write mimetype content: %w", err)
	}

	return nil
}

func (r *RepairServiceImpl) writeContainerXML(zipWriter *zip.Writer) error {
	containerXML := ContainerXML{
		Version: "1.0",
		Rootfiles: []Rootfile{
			{
				FullPath:  "OEBPS/content.opf",
				MediaType: "application/oebps-package+xml",
			},
		},
	}

	containerXML.XMLName = xml.Name{Local: "container", Space: "urn:oasis:names:tc:opendocument:xmlns:container"}

	var buf bytes.Buffer
	buf.WriteString(xml.Header)

	encoder := xml.NewEncoder(&buf)
	encoder.Indent("", "  ")
	if err := encoder.Encode(containerXML); err != nil {
		return fmt.Errorf("failed to encode container.xml: %w", err)
	}

	w, err := zipWriter.Create(ContainerXMLPath)
	if err != nil {
		return fmt.Errorf("failed to create container.xml: %w", err)
	}

	_, err = w.Write(buf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to write container.xml: %w", err)
	}

	return nil
}

func (r *RepairServiceImpl) addDoctype(data []byte) ([]byte, error) {
	content := string(data)

	trimmed := strings.TrimSpace(content)
	if strings.HasPrefix(strings.ToLower(trimmed), "<!doctype") {
		return data, nil
	}

	doctype := "<!DOCTYPE html>\n"

	if strings.HasPrefix(trimmed, "<?xml") {
		xmlDeclEnd := strings.Index(trimmed, "?>")
		if xmlDeclEnd != -1 {
			xmlDecl := trimmed[:xmlDeclEnd+2]
			rest := strings.TrimLeft(trimmed[xmlDeclEnd+2:], " \t\n\r")
			return []byte(xmlDecl + "\n" + doctype + rest), nil
		}
	}

	return []byte(doctype + content), nil
}

func (r *RepairServiceImpl) repairOPFMetadata(data []byte, actions []ports.RepairAction) ([]byte, error) {
	var pkg Package
	if err := xml.Unmarshal(data, &pkg); err != nil {
		return nil, fmt.Errorf("failed to parse OPF: %w", err)
	}

	actionTypes := make(map[string]bool)
	for _, action := range actions {
		actionTypes[action.Type] = true
	}

	if actionTypes["add_metadata_title"] {
		if len(pkg.Metadata.Titles) == 0 {
			pkg.Metadata.Titles = append(pkg.Metadata.Titles, DCElement{
				XMLName: xml.Name{Space: DCNamespace, Local: "title"},
				Value:   "Untitled",
			})
		}
	}

	if actionTypes["add_metadata_identifier"] {
		if len(pkg.Metadata.Identifiers) == 0 {
			idValue := fmt.Sprintf("urn:uuid:%d", time.Now().Unix())
			pkg.Metadata.Identifiers = append(pkg.Metadata.Identifiers, DCIdentifier{
				XMLName: xml.Name{Space: DCNamespace, Local: "identifier"},
				Value:   idValue,
				ID:      "bookid",
			})
			if pkg.UniqueID == "" {
				pkg.UniqueID = "bookid"
			}
		}
	}

	if actionTypes["add_metadata_language"] {
		if len(pkg.Metadata.Languages) == 0 {
			pkg.Metadata.Languages = append(pkg.Metadata.Languages, DCElement{
				XMLName: xml.Name{Space: DCNamespace, Local: "language"},
				Value:   "en",
			})
		}
	}

	if actionTypes["add_metadata_modified"] {
		hasModified := false
		for _, meta := range pkg.Metadata.Meta {
			if meta.Property == DCTermsProperty {
				hasModified = true
				break
			}
		}
		if !hasModified {
			pkg.Metadata.Meta = append(pkg.Metadata.Meta, MetaElement{
				XMLName:  xml.Name{Local: "meta"},
				Property: DCTermsProperty,
				Value:    time.Now().Format("2006-01-02T15:04:05Z"),
			})
		}
	}

	pkg.XMLName = xml.Name{Space: OPFNamespace, Local: "package"}
	pkg.Metadata.XMLName = xml.Name{Local: "metadata"}

	var buf bytes.Buffer
	buf.WriteString(xml.Header)

	encoder := xml.NewEncoder(&buf)
	encoder.Indent("", "  ")
	if err := encoder.Encode(pkg); err != nil {
		return nil, fmt.Errorf("failed to encode OPF: %w", err)
	}

	return buf.Bytes(), nil
}

func (r *RepairServiceImpl) generateOutputPath(filePath string) string {
	ext := filepath.Ext(filePath)
	base := strings.TrimSuffix(filePath, ext)
	return base + repairSuffix
}

func (r *RepairServiceImpl) fixRelativePaths(href string, basePath string) string {
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return href
	}

	href = strings.TrimPrefix(href, "/")

	href = strings.ReplaceAll(href, "//", "/")

	if basePath != "" && basePath != "." {
		return path.Join(basePath, href)
	}

	return href
}

func (r *RepairServiceImpl) parseHTML(data []byte) (*html.Node, error) {
	doc, err := html.Parse(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return doc, nil
}

func (r *RepairServiceImpl) renderHTML(node *html.Node) ([]byte, error) {
	var buf bytes.Buffer
	if err := html.Render(&buf, node); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
