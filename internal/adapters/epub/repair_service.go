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

	"github.com/petergi/ebook-mechanic-lib/internal/domain"
	"github.com/petergi/ebook-mechanic-lib/internal/ports"
)

const (
	repairSuffix = "_repaired.epub"
)

// RepairServiceImpl implements EPUB repair operations.
type RepairServiceImpl struct {
	containerValidator *ContainerValidator
	opfValidator       *OPFValidator
	navValidator       *NavValidator
	contentValidator   *ContentValidator
}

// NewRepairService returns a repair service for EPUB files.
func NewRepairService() ports.EPUBRepairService {
	return &RepairServiceImpl{
		containerValidator: NewContainerValidator(),
		opfValidator:       NewOPFValidator(),
		navValidator:       NewNavValidator(),
		contentValidator:   NewContentValidator(),
	}
}

// Preview builds a repair plan from a validation report.
func (r *RepairServiceImpl) Preview(ctx context.Context, report *domain.ValidationReport) (*ports.RepairPreview, error) {
	return r.PreviewWithOptions(ctx, report, RepairOptions{})
}

// RepairOptions configures repair behavior for EPUB.
type RepairOptions struct {
	Aggressive bool
}

// PreviewWithOptions builds a repair plan using the provided options.
func (r *RepairServiceImpl) PreviewWithOptions(_ context.Context, report *domain.ValidationReport, options RepairOptions) (*ports.RepairPreview, error) {
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

	for i := range report.Errors {
		actions := r.generateRepairActions(&report.Errors[i], options)
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

// Apply applies repairs and writes the repaired EPUB to a default path.
func (r *RepairServiceImpl) Apply(ctx context.Context, filePath string, preview *ports.RepairPreview) (*ports.RepairResult, error) {
	outputPath := r.generateOutputPath(filePath)
	return r.ApplyWithBackup(ctx, filePath, preview, outputPath)
}

// ApplyWithBackup applies repairs and writes the repaired EPUB to backupPath.
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

	file, err := os.Open(filePath) //nolint:gosec
	if err != nil {
		result.Error = fmt.Errorf("failed to open EPUB: %w", err)
		return result, nil
	}
	defer func() {
		_ = file.Close()
	}()

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

	outputFile, err := os.Create(backupPath) //nolint:gosec
	if err != nil {
		result.Error = fmt.Errorf("failed to create output file: %w", err)
		return result, nil
	}
	defer func() {
		_ = outputFile.Close()
	}()

	zipWriter := zip.NewWriter(outputFile)
	defer func() {
		_ = zipWriter.Close()
	}()

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

// CanRepair reports whether a validation error can be repaired automatically.
func (r *RepairServiceImpl) CanRepair(_ context.Context, err *domain.ValidationError) bool {
	if err == nil {
		return false
	}

	switch err.Code {
	case ErrorCodeMimetypeInvalid,
		ErrorCodeMimetypeNotFirst,
		ErrorCodeContainerXMLMissing,
		ErrorCodeContentMissingDoctype,
		ErrorCodeContentMissingHTML,
		ErrorCodeContentMissingHead,
		ErrorCodeContentMissingBody,
		ErrorCodeContentInvalidNamespace,
		ErrorCodeOPFMissingTitle,
		ErrorCodeOPFMissingIdentifier,
		ErrorCodeOPFMissingLanguage,
		ErrorCodeOPFMissingModified,
		ErrorCodeOPFMissingNavDocument,
		ErrorCodeOPFInvalidUniqueID,
		ErrorCodeNavMissingNavElement,
		ErrorCodeNavMissingTOC,
		ErrorCodeNavInvalidTOCStructure,
		ErrorCodeOPFFileNotFound:
		return true
	default:
		return false
	}
}

// CreateBackup creates a copy of the EPUB at backupPath.
func (r *RepairServiceImpl) CreateBackup(_ context.Context, filePath string, backupPath string) error {
	sourceFile, err := os.Open(filePath) //nolint:gosec
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer func() {
		_ = sourceFile.Close()
	}()

	destFile, err := os.Create(backupPath) //nolint:gosec
	if err != nil {
		return fmt.Errorf("failed to create backup file: %w", err)
	}
	defer func() {
		_ = destFile.Close()
	}()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return nil
}

// RestoreBackup restores the original file from backupPath.
func (r *RepairServiceImpl) RestoreBackup(ctx context.Context, backupPath string, originalPath string) error {
	return r.CreateBackup(ctx, backupPath, originalPath)
}

// RepairStructure runs structure-focused repair steps.
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

// RepairMetadata runs metadata-focused repair steps.
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

// RepairContent runs content-focused repair steps.
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

func (r *RepairServiceImpl) generateRepairActions(err *domain.ValidationError, options RepairOptions) []ports.RepairAction {
	actions := make([]ports.RepairAction, 0)

	switch err.Code {
	case ErrorCodeMimetypeInvalid:
		if err.Details != nil {
			if _, ok := err.Details["compression_method"]; ok {
				actions = append(actions, ports.RepairAction{
					Type:        "fix_mimetype_compression",
					Description: "Rewrite mimetype entry as uncompressed ZIP store",
					Target:      "mimetype",
					Details:     map[string]interface{}{},
					Automated:   true,
				})
				return actions
			}
		}
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

	case ErrorCodeContentMissingHTML,
		ErrorCodeContentMissingHead,
		ErrorCodeContentMissingBody,
		ErrorCodeContentInvalidNamespace:
		target := "content.xhtml"
		if err.Location != nil && err.Location.Path != "" {
			target = err.Location.Path
		}
		actions = append(actions, ports.RepairAction{
			Type:        "repair_content_structure",
			Description: "Replace content document with minimal valid XHTML structure",
			Target:      target,
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

	case ErrorCodeOPFInvalidUniqueID:
		actions = append(actions, ports.RepairAction{
			Type:        "fix_opf_unique_id",
			Description: "Align unique-identifier with an existing dc:identifier",
			Target:      err.Location.Path,
			Details:     err.Details,
			Automated:   true,
		})

	case ErrorCodeOPFInvalidSpineTOC:
		ncxID, _ := err.Details["ncx_id"].(string)
		if ncxID != "" {
			actions = append(actions, ports.RepairAction{
				Type:        "fix_spine_toc",
				Description: "Set OPF spine toc attribute to reference the NCX item",
				Target:      err.Location.Path,
				Details:     map[string]interface{}{"ncx_id": ncxID},
				Automated:   true,
			})
		} else {
			actions = append(actions, ports.RepairAction{
				Type:        "manual_review",
				Description: fmt.Sprintf("Requires manual review: %s", err.Message),
				Target:      err.Location.Path,
				Details:     err.Details,
				Automated:   false,
			})
		}

	case ErrorCodeOPFInvalidSpineItem,
		ErrorCodeOPFMissingSpine,
		ErrorCodeOPFInvalidManifestItem,
		ErrorCodeOPFMissingManifest:
		if options.Aggressive {
			target := err.Location.Path
			if target == "" {
				target = "OEBPS/content.opf"
			}
			actions = append(actions, ports.RepairAction{
				Type:        "rebuild_opf",
				Description: "Aggressively rebuild OPF manifest and spine from available content",
				Target:      target,
				Details:     map[string]interface{}{},
				Automated:   true,
			})
		} else {
			actions = append(actions, ports.RepairAction{
				Type:        "manual_review",
				Description: fmt.Sprintf("Requires manual review: %s", err.Message),
				Target:      err.Location.Path,
				Details:     err.Details,
				Automated:   false,
			})
		}

	case ErrorCodeOPFMissingNavDocument:
		actions = append(actions, ports.RepairAction{
			Type:        "add_nav_document",
			Description: "Add navigation document item to OPF manifest",
			Target:      err.Location.Path,
			Details:     map[string]interface{}{},
			Automated:   true,
		})

	case ErrorCodeNavMissingNavElement, ErrorCodeNavMissingTOC, ErrorCodeNavInvalidTOCStructure:
		navTarget := ""
		if err.Location != nil {
			navTarget = err.Location.Path
		}
		if navTarget == "" {
			navTarget = "OEBPS/nav.xhtml"
		}
		actions = append(actions, ports.RepairAction{
			Type:        "repair_nav_document",
			Description: "Replace navigation document with a minimal valid nav",
			Target:      navTarget,
			Details:     map[string]interface{}{},
			Automated:   true,
		})

	case ErrorCodeOPFFileNotFound:
		if err.Location != nil {
			pathLower := strings.ToLower(err.Location.Path)
			switch {
			case strings.HasSuffix(pathLower, ".opf"):
				actions = append(actions, ports.RepairAction{
					Type:        "create_opf",
					Description: "Create minimal OPF package document",
					Target:      err.Location.Path,
					Details:     map[string]interface{}{},
					Automated:   true,
				})
			case strings.HasSuffix(pathLower, ".xhtml") || strings.HasSuffix(pathLower, ".html"):
				actions = append(actions, ports.RepairAction{
					Type:        "create_xhtml_stub",
					Description: "Create minimal XHTML content document",
					Target:      err.Location.Path,
					Details:     map[string]interface{}{},
					Automated:   true,
				})
			default:
				actions = append(actions, ports.RepairAction{
					Type:        "manual_review",
					Description: fmt.Sprintf("Missing referenced file: %s", err.Location.Path),
					Target:      err.Location.Path,
					Details:     err.Details,
					Automated:   false,
				})
			}
		}

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

func (r *RepairServiceImpl) applyRepairs(_ context.Context, repairCtx *repairContext) error {
	actionsByType := make(map[string][]ports.RepairAction)
	for _, action := range repairCtx.actions {
		if action.Automated {
			actionsByType[action.Type] = append(actionsByType[action.Type], action)
		}
	}

	needsMimetypeFix := len(actionsByType["fix_mimetype_content"]) > 0 ||
		len(actionsByType["fix_mimetype_order"]) > 0 ||
		len(actionsByType["fix_mimetype_compression"]) > 0
	needsContainerFix := len(actionsByType["create_container_xml"]) > 0

	filesProcessed := make(map[string]bool)

	if err := r.writeMimetype(repairCtx.zipWriter); err != nil {
		return err
	}
	filesProcessed["mimetype"] = true
	if needsMimetypeFix {
		repairCtx.applied = append(
			repairCtx.applied,
			append(actionsByType["fix_mimetype_content"],
				append(actionsByType["fix_mimetype_order"], actionsByType["fix_mimetype_compression"]...)...)...,
		)
	}

	contentActions := make(map[string]ports.RepairAction)
	for _, action := range actionsByType["add_doctype"] {
		contentActions[action.Target] = action
	}
	for _, action := range actionsByType["repair_content_structure"] {
		contentActions[action.Target] = action
	}

	opfActions := make(map[string][]ports.RepairAction)
	for _, actionType := range []string{"add_metadata_title", "add_metadata_identifier",
		"add_metadata_language", "add_metadata_modified", "add_nav_document", "fix_opf_unique_id", "fix_spine_toc"} {
		for _, action := range actionsByType[actionType] {
			opfActions[action.Target] = append(opfActions[action.Target], action)
		}
	}

	opfRebuildPlans := make(map[string]ports.RepairAction)
	for _, action := range actionsByType["rebuild_opf"] {
		if _, exists := opfRebuildPlans[action.Target]; !exists {
			opfRebuildPlans[action.Target] = action
		}
	}

	for path := range opfRebuildPlans {
		delete(opfActions, path)
	}

	opfCreatePlans := make(map[string]ports.RepairAction)
	for _, action := range actionsByType["create_opf"] {
		if _, exists := opfCreatePlans[action.Target]; !exists {
			opfCreatePlans[action.Target] = action
		}
	}

	xhtmlCreatePlans := make(map[string]ports.RepairAction)
	for _, action := range actionsByType["create_xhtml_stub"] {
		if _, exists := xhtmlCreatePlans[action.Target]; !exists {
			xhtmlCreatePlans[action.Target] = action
		}
	}

	navRepairPlans := make(map[string]ports.RepairAction)
	for _, action := range actionsByType["repair_nav_document"] {
		if _, exists := navRepairPlans[action.Target]; !exists {
			navRepairPlans[action.Target] = action
		}
	}

	navPlans := make(map[string]navPlan)
	for _, action := range actionsByType["add_nav_document"] {
		if _, exists := navPlans[action.Target]; exists {
			continue
		}
		href, navPath, navExists := resolveNavPath(repairCtx.zipReader, action.Target)
		navPlans[action.Target] = navPlan{
			Href:   href,
			Path:   navPath,
			Exists: navExists,
			Action: action,
		}
	}

	for _, f := range repairCtx.zipReader.File {
		if filesProcessed[f.Name] {
			continue
		}

		if _, shouldRepair := navRepairPlans[f.Name]; shouldRepair {
			filesProcessed[f.Name] = true
			continue
		}

		if needsContainerFix && f.Name == ContainerXMLPath {
			if err := r.writeContainerXML(repairCtx.zipWriter); err != nil {
				return err
			}
			filesProcessed[f.Name] = true
			repairCtx.applied = append(repairCtx.applied, actionsByType["create_container_xml"]...)
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("failed to open file %s: %w", f.Name, err)
		}
		defer func() {
			_ = rc.Close()
		}()

		data, err := io.ReadAll(rc)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", f.Name, err)
		}

		if action, exists := contentActions[f.Name]; exists {
			if action.Type == "repair_content_structure" {
				data = minimalXHTMLContent()
			} else {
				data = r.addDoctype(data)
			}
			repairCtx.applied = append(repairCtx.applied, action)
		}

		if _, shouldRebuild := opfRebuildPlans[f.Name]; shouldRebuild {
			filesProcessed[f.Name] = true
			continue
		}

		if actions, exists := opfActions[f.Name]; exists && len(actions) > 0 {
			navHref := ""
			if plan, ok := navPlans[f.Name]; ok {
				navHref = plan.Href
			}
			data, err = r.repairOPF(data, actions, navHref)
			if err != nil {
				return fmt.Errorf("failed to repair OPF in %s: %w", f.Name, err)
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
		repairCtx.applied = append(repairCtx.applied, actionsByType["create_container_xml"]...)
		filesProcessed[ContainerXMLPath] = true
	}

	for _, plan := range navPlans {
		if plan.Exists || filesProcessed[plan.Path] {
			continue
		}
		if err := r.writeNavDocument(repairCtx.zipWriter, plan.Path); err != nil {
			return err
		}
		filesProcessed[plan.Path] = true
	}

	for navPath, action := range navRepairPlans {
		if err := r.writeNavDocument(repairCtx.zipWriter, navPath); err != nil {
			return err
		}
		filesProcessed[navPath] = true
		repairCtx.applied = append(repairCtx.applied, action)
	}

	for opfPath, action := range opfRebuildPlans {
		if filesProcessed[opfPath] || zipHasFile(repairCtx.zipReader, opfPath) {
			continue
		}
		contentHrefs := collectContentHrefs(repairCtx.zipReader, opfPath)
		navHref := "nav.xhtml"
		opfContent := r.buildAggressiveOPF(opfPath, contentHrefs, navHref)
		if err := r.writeFile(repairCtx.zipWriter, opfPath, opfContent); err != nil {
			return err
		}
		filesProcessed[opfPath] = true
		repairCtx.applied = append(repairCtx.applied, action)

		navPath := resolveNavPathFromHref(opfPath, navHref)
		if err := r.writeFile(repairCtx.zipWriter, navPath, buildNavFromSpine(contentHrefs)); err != nil {
			return err
		}
		filesProcessed[navPath] = true
	}

	if !filesProcessed[ContainerXMLPath] && !zipHasFile(repairCtx.zipReader, ContainerXMLPath) {
		if err := r.writeContainerXML(repairCtx.zipWriter); err != nil {
			return err
		}
		filesProcessed[ContainerXMLPath] = true
	}

	if !zipHasFile(repairCtx.zipReader, "OEBPS/content.opf") && !filesProcessed["OEBPS/content.opf"] {
		opfContent := r.buildMinimalOPF("OEBPS/content.opf")
		if err := r.writeFile(repairCtx.zipWriter, "OEBPS/content.opf", opfContent); err != nil {
			return err
		}
		filesProcessed["OEBPS/content.opf"] = true
		for _, stubPath := range r.minimalContentPaths("OEBPS/content.opf") {
			if filesProcessed[stubPath] || zipHasFile(repairCtx.zipReader, stubPath) {
				continue
			}
			if strings.HasSuffix(strings.ToLower(stubPath), ".xhtml") {
				if err := r.writeFile(repairCtx.zipWriter, stubPath, minimalXHTMLContent()); err != nil {
					return err
				}
				filesProcessed[stubPath] = true
			}
		}
	}

	for opfPath, action := range opfCreatePlans {
		if filesProcessed[opfPath] || zipHasFile(repairCtx.zipReader, opfPath) {
			continue
		}
		opfContent := r.buildMinimalOPF(opfPath)
		if err := r.writeFile(repairCtx.zipWriter, opfPath, opfContent); err != nil {
			return err
		}
		filesProcessed[opfPath] = true
		repairCtx.applied = append(repairCtx.applied, action)

		for _, stubPath := range r.minimalContentPaths(opfPath) {
			if filesProcessed[stubPath] || zipHasFile(repairCtx.zipReader, stubPath) {
				continue
			}
			if strings.HasSuffix(strings.ToLower(stubPath), ".xhtml") {
				if err := r.writeFile(repairCtx.zipWriter, stubPath, minimalXHTMLContent()); err != nil {
					return err
				}
				filesProcessed[stubPath] = true
			}
		}
	}

	for xhtmlPath, action := range xhtmlCreatePlans {
		if filesProcessed[xhtmlPath] || zipHasFile(repairCtx.zipReader, xhtmlPath) {
			continue
		}
		if strings.EqualFold(path.Base(xhtmlPath), "nav.xhtml") {
			if err := r.writeNavDocument(repairCtx.zipWriter, xhtmlPath); err != nil {
				return err
			}
		} else {
			if err := r.writeFile(repairCtx.zipWriter, xhtmlPath, minimalXHTMLContent()); err != nil {
				return err
			}
		}
		filesProcessed[xhtmlPath] = true
		repairCtx.applied = append(repairCtx.applied, action)
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

func (r *RepairServiceImpl) addDoctype(data []byte) []byte {
	content := string(data)

	trimmed := strings.TrimSpace(content)
	if strings.HasPrefix(strings.ToLower(trimmed), "<!doctype") {
		return data
	}

	doctype := "<!DOCTYPE html>\n"

	if strings.HasPrefix(trimmed, "<?xml") {
		xmlDeclEnd := strings.Index(trimmed, "?>")
		if xmlDeclEnd != -1 {
			xmlDecl := trimmed[:xmlDeclEnd+2]
			rest := strings.TrimLeft(trimmed[xmlDeclEnd+2:], " \t\n\r")
			return []byte(xmlDecl + "\n" + doctype + rest)
		}
	}

	return []byte(doctype + content)
}

func (r *RepairServiceImpl) repairOPF(data []byte, actions []ports.RepairAction, navHref string) ([]byte, error) {
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

	if actionTypes["fix_opf_unique_id"] {
		normalizeUniqueIdentifier(&pkg)
	}

	if actionTypes["fix_spine_toc"] {
		ncxID := extractRepairDetail(actions, "ncx_id")
		if ncxID == "" {
			ncxID = findFirstNCXID(&pkg)
		}
		if ncxID != "" {
			pkg.Spine.Toc = ncxID
		}
	}

	if actionTypes["add_nav_document"] {
		if err := ensureNavItem(&pkg, navHref); err != nil {
			return nil, err
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

func extractRepairDetail(actions []ports.RepairAction, key string) string {
	for _, action := range actions {
		if action.Details == nil {
			continue
		}
		if value, ok := action.Details[key].(string); ok {
			return value
		}
	}
	return ""
}

func findFirstNCXID(pkg *Package) string {
	for _, item := range pkg.Manifest.Items {
		if strings.EqualFold(strings.TrimSpace(item.MediaType), "application/x-dtbncx+xml") {
			return item.ID
		}
	}
	return ""
}

type navPlan struct {
	Href   string
	Path   string
	Exists bool
	Action ports.RepairAction
}

func resolveNavPath(zipReader *zip.Reader, opfPath string) (string, string, bool) {
	opfDir := path.Dir(opfPath)
	if opfDir == "." {
		opfDir = ""
	}

	defaultPath := path.Join(opfDir, "nav.xhtml")
	if zipHasFile(zipReader, defaultPath) {
		return "nav.xhtml", defaultPath, true
	}

	for _, file := range zipReader.File {
		if path.Base(file.Name) != "nav.xhtml" {
			continue
		}
		rel := file.Name
		if opfDir != "" {
			prefix := opfDir + "/"
			if strings.HasPrefix(file.Name, prefix) {
				rel = strings.TrimPrefix(file.Name, prefix)
			}
		}
		if strings.HasPrefix(rel, "../") {
			rel = path.Base(file.Name)
		}
		return rel, file.Name, true
	}

	return "nav.xhtml", defaultPath, false
}

func zipHasFile(zipReader *zip.Reader, name string) bool {
	for _, file := range zipReader.File {
		if file.Name == name {
			return true
		}
	}
	return false
}

func ensureNavItem(pkg *Package, navHref string) error {
	if strings.TrimSpace(navHref) == "" {
		return fmt.Errorf("nav href is empty")
	}

	for _, item := range pkg.Manifest.Items {
		if strings.Contains(item.Properties, "nav") {
			return nil
		}
	}

	for i, item := range pkg.Manifest.Items {
		if item.Href == navHref {
			pkg.Manifest.Items[i].Properties = addPropertyToken(item.Properties, "nav")
			return nil
		}
	}

	navID := uniqueManifestID(pkg.Manifest.Items, "nav")
	pkg.Manifest.Items = append(pkg.Manifest.Items, ManifestItem{
		ID:         navID,
		Href:       navHref,
		MediaType:  "application/xhtml+xml",
		Properties: "nav",
	})
	return nil
}

func uniqueManifestID(items []ManifestItem, base string) string {
	used := make(map[string]bool)
	for _, item := range items {
		used[item.ID] = true
	}
	if !used[base] {
		return base
	}
	for i := 1; ; i++ {
		candidate := fmt.Sprintf("%s-%d", base, i)
		if !used[candidate] {
			return candidate
		}
	}
}

func addPropertyToken(props string, token string) string {
	parts := strings.Fields(props)
	for _, part := range parts {
		if part == token {
			return props
		}
	}
	if len(parts) == 0 {
		return token
	}
	return strings.TrimSpace(props) + " " + token
}

func (r *RepairServiceImpl) writeNavDocument(zipWriter *zip.Writer, navPath string) error {
	const navTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops">
  <head>
    <title>Navigation</title>
  </head>
  <body>
    <nav epub:type="toc" id="toc">
      <h1>Table of Contents</h1>
      <ol>
        <li><a href="#">Start</a></li>
      </ol>
    </nav>
  </body>
</html>
`

	w, err := zipWriter.Create(navPath)
	if err != nil {
		return fmt.Errorf("failed to create nav document: %w", err)
	}

	if _, err := w.Write([]byte(navTemplate)); err != nil {
		return fmt.Errorf("failed to write nav document: %w", err)
	}

	return nil
}

func (r *RepairServiceImpl) writeFile(zipWriter *zip.Writer, filePath string, contents []byte) error {
	w, err := zipWriter.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create %s: %w", filePath, err)
	}
	if _, err := w.Write(contents); err != nil {
		return fmt.Errorf("failed to write %s: %w", filePath, err)
	}
	return nil
}

func (r *RepairServiceImpl) buildMinimalOPF(opfPath string) []byte {
	contentHref := "content.xhtml"
	navHref := "nav.xhtml"

	pkg := Package{
		XMLName:  xml.Name{Space: OPFNamespace, Local: "package"},
		Version:  "3.0",
		UniqueID: "bookid",
		Metadata: Metadata{
			XMLName: xml.Name{Local: "metadata"},
			Titles: []DCElement{
				{
					XMLName: xml.Name{Space: DCNamespace, Local: "title"},
					Value:   "Untitled",
				},
			},
			Identifiers: []DCIdentifier{
				{
					XMLName: xml.Name{Space: DCNamespace, Local: "identifier"},
					Value:   fmt.Sprintf("urn:uuid:%d", time.Now().UnixNano()),
					ID:      "bookid",
				},
			},
			Languages: []DCElement{
				{
					XMLName: xml.Name{Space: DCNamespace, Local: "language"},
					Value:   "en",
				},
			},
			Meta: []MetaElement{
				{
					XMLName:  xml.Name{Local: "meta"},
					Property: DCTermsProperty,
					Value:    time.Now().Format("2006-01-02T15:04:05Z"),
				},
			},
		},
		Manifest: Manifest{
			XMLName: xml.Name{Local: "manifest"},
			Items: []ManifestItem{
				{
					ID:         "nav",
					Href:       navHref,
					MediaType:  "application/xhtml+xml",
					Properties: "nav",
				},
				{
					ID:        "content",
					Href:      contentHref,
					MediaType: "application/xhtml+xml",
				},
			},
		},
		Spine: Spine{
			XMLName: xml.Name{Local: "spine"},
			Items: []SpineItem{
				{IDRef: "content"},
			},
		},
	}

	var buf bytes.Buffer
	buf.WriteString(xml.Header)
	encoder := xml.NewEncoder(&buf)
	encoder.Indent("", "  ")
	_ = encoder.Encode(pkg)
	return buf.Bytes()
}

func (r *RepairServiceImpl) minimalContentPaths(opfPath string) []string {
	dir := path.Dir(opfPath)
	if dir == "." {
		dir = ""
	}
	if dir == "" {
		return []string{"content.xhtml", "nav.xhtml"}
	}
	return []string{path.Join(dir, "content.xhtml"), path.Join(dir, "nav.xhtml")}
}

func minimalXHTMLContent() []byte {
	return []byte(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml">
  <head>
    <title>Content</title>
  </head>
  <body>
    <p>Repaired content placeholder.</p>
  </body>
</html>
`)
}

func collectContentHrefs(zipReader *zip.Reader, opfPath string) []string {
	opfDir := path.Dir(opfPath)
	if opfDir == "." {
		opfDir = ""
	}
	seen := make(map[string]bool)
	var hrefs []string

	for _, file := range zipReader.File {
		lower := strings.ToLower(file.Name)
		if !strings.HasSuffix(lower, ".xhtml") && !strings.HasSuffix(lower, ".html") {
			continue
		}
		if strings.HasSuffix(lower, "/nav.xhtml") || strings.HasSuffix(lower, "/toc.xhtml") || strings.HasSuffix(lower, "/navigation.xhtml") {
			continue
		}

		rel := file.Name
		if opfDir != "" {
			prefix := opfDir + "/"
			if strings.HasPrefix(file.Name, prefix) {
				rel = strings.TrimPrefix(file.Name, prefix)
			} else {
				continue
			}
		}

		if rel == "" || seen[rel] {
			continue
		}
		seen[rel] = true
		hrefs = append(hrefs, rel)
	}

	if len(hrefs) == 0 {
		hrefs = []string{"content.xhtml"}
	}

	return hrefs
}

func (r *RepairServiceImpl) buildAggressiveOPF(opfPath string, contentHrefs []string, navHref string) []byte {
	if navHref == "" {
		navHref = "nav.xhtml"
	}

	pkg := Package{
		XMLName:  xml.Name{Space: OPFNamespace, Local: "package"},
		Version:  "3.0",
		UniqueID: "bookid",
		Metadata: Metadata{
			XMLName: xml.Name{Local: "metadata"},
			Titles: []DCElement{
				{
					XMLName: xml.Name{Space: DCNamespace, Local: "title"},
					Value:   "Untitled",
				},
			},
			Identifiers: []DCIdentifier{
				{
					XMLName: xml.Name{Space: DCNamespace, Local: "identifier"},
					Value:   fmt.Sprintf("urn:uuid:%d", time.Now().UnixNano()),
					ID:      "bookid",
				},
			},
			Languages: []DCElement{
				{
					XMLName: xml.Name{Space: DCNamespace, Local: "language"},
					Value:   "en",
				},
			},
			Meta: []MetaElement{
				{
					XMLName:  xml.Name{Local: "meta"},
					Property: DCTermsProperty,
					Value:    time.Now().Format("2006-01-02T15:04:05Z"),
				},
			},
		},
		Manifest: Manifest{
			XMLName: xml.Name{Local: "manifest"},
			Items: []ManifestItem{
				{
					ID:         "nav",
					Href:       navHref,
					MediaType:  "application/xhtml+xml",
					Properties: "nav",
				},
			},
		},
		Spine: Spine{
			XMLName: xml.Name{Local: "spine"},
			Items:   []SpineItem{},
		},
	}

	for idx, href := range contentHrefs {
		itemID := fmt.Sprintf("item-%d", idx+1)
		pkg.Manifest.Items = append(pkg.Manifest.Items, ManifestItem{
			ID:        itemID,
			Href:      href,
			MediaType: "application/xhtml+xml",
		})
		pkg.Spine.Items = append(pkg.Spine.Items, SpineItem{IDRef: itemID})
	}

	var buf bytes.Buffer
	buf.WriteString(xml.Header)
	encoder := xml.NewEncoder(&buf)
	encoder.Indent("", "  ")
	_ = encoder.Encode(pkg)
	return buf.Bytes()
}

func resolveNavPathFromHref(opfPath, navHref string) string {
	dir := path.Dir(opfPath)
	if dir == "." || dir == "" {
		return navHref
	}
	return path.Join(dir, navHref)
}

func buildNavFromSpine(contentHrefs []string) []byte {
	var list strings.Builder
	for i, href := range contentHrefs {
		title := fmt.Sprintf("Section %d", i+1)
		list.WriteString(fmt.Sprintf("        <li><a href=\"%s\">%s</a></li>\n", href, title))
	}

	if list.Len() == 0 {
		list.WriteString("        <li><a href=\"#\">Start</a></li>\n")
	}

	return []byte(fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops">
  <head>
    <title>Navigation</title>
  </head>
  <body>
    <nav epub:type="toc" id="toc">
      <h1>Table of Contents</h1>
      <ol>
%s      </ol>
    </nav>
  </body>
</html>
`, list.String()))
}
func normalizeUniqueIdentifier(pkg *Package) {
	if pkg == nil {
		return
	}

	if pkg.UniqueID == "" {
		if len(pkg.Metadata.Identifiers) > 0 && pkg.Metadata.Identifiers[0].ID != "" {
			pkg.UniqueID = pkg.Metadata.Identifiers[0].ID
		} else {
			pkg.UniqueID = "bookid"
		}
	}

	for i := range pkg.Metadata.Identifiers {
		if pkg.Metadata.Identifiers[i].ID == pkg.UniqueID {
			return
		}
	}

	if len(pkg.Metadata.Identifiers) > 0 {
		pkg.Metadata.Identifiers[0].ID = pkg.UniqueID
		return
	}

	pkg.Metadata.Identifiers = append(pkg.Metadata.Identifiers, DCIdentifier{
		XMLName: xml.Name{Space: DCNamespace, Local: "identifier"},
		Value:   fmt.Sprintf("urn:uuid:%d", time.Now().UnixNano()),
		ID:      pkg.UniqueID,
	})
}

func (r *RepairServiceImpl) generateOutputPath(filePath string) string {
	ext := filepath.Ext(filePath)
	base := strings.TrimSuffix(filePath, ext)
	return base + repairSuffix
}
