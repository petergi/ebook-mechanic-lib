// Package pdf provides PDF validation and repair adapters.
package pdf

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/example/project/internal/domain"
	"github.com/example/project/internal/ports"
)

const (
	repairSuffix = "_repaired.pdf"
)

// RepairServiceImpl implements PDF repair operations.
type RepairServiceImpl struct {
	validator *StructureValidator
}

// NewRepairService returns a repair service for PDF files.
func NewRepairService() ports.PDFRepairService {
	return &RepairServiceImpl{
		validator: NewStructureValidator(),
	}
}

// Preview builds a repair plan from a validation report.
func (r *RepairServiceImpl) Preview(_ context.Context, report *domain.ValidationReport) (*ports.RepairPreview, error) {
	if report == nil {
		return nil, fmt.Errorf("validation report is nil")
	}

	preview := &ports.RepairPreview{
		Actions:        make([]ports.RepairAction, 0),
		CanAutoRepair:  true,
		EstimatedTime:  500,
		BackupRequired: true,
		Warnings:       make([]string, 0),
	}

	for i := range report.Errors {
		actions := r.generateRepairActions(&report.Errors[i])
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

// Apply applies repairs and writes the repaired PDF to a default path.
func (r *RepairServiceImpl) Apply(ctx context.Context, filePath string, preview *ports.RepairPreview) (*ports.RepairResult, error) {
	outputPath := r.generateOutputPath(filePath)
	return r.ApplyWithBackup(ctx, filePath, preview, outputPath)
}

// ApplyWithBackup applies repairs and writes the repaired PDF to backupPath.
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

	data, err := os.ReadFile(filePath) //nolint:gosec
	if err != nil {
		result.Error = fmt.Errorf("failed to read PDF: %w", err)
		return result, nil
	}

	repairContext := &repairContext{
		actions: preview.Actions,
		data:    data,
		applied: make([]ports.RepairAction, 0),
	}

	if err := r.applyRepairs(ctx, repairContext); err != nil {
		result.Error = fmt.Errorf("repair failed: %w", err)
		return result, nil
	}

	if err := os.WriteFile(backupPath, repairContext.data, 0600); err != nil {
		result.Error = fmt.Errorf("failed to write repaired file: %w", err)
		return result, nil
	}

	result.Success = true
	result.ActionsApplied = repairContext.applied

	return result, nil
}

// CanRepair reports whether a validation error is repairable.
func (r *RepairServiceImpl) CanRepair(_ context.Context, err *domain.ValidationError) bool {
	if err == nil {
		return false
	}

	switch err.Code {
	case ErrorCodePDFTrailer003,
		ErrorCodePDFTrailer001:
		return true
	default:
		return false
	}
}

// CreateBackup creates a copy of the PDF at backupPath.
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
	data, err := os.ReadFile(filePath) //nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	validationResult, err := r.validator.ValidateBytes(data)
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	report := r.convertValidationResult(filePath, validationResult)
	preview, err := r.Preview(ctx, report)
	if err != nil {
		return nil, err
	}

	return r.Apply(ctx, filePath, preview)
}

// RepairMetadata runs metadata-focused repair steps.
func (r *RepairServiceImpl) RepairMetadata(_ context.Context, _ string) (*ports.RepairResult, error) {
	return &ports.RepairResult{
		Success:        true,
		ActionsApplied: make([]ports.RepairAction, 0),
		BackupPath:     "",
	}, nil
}

// OptimizeFile writes an optimized PDF stream to writer.
func (r *RepairServiceImpl) OptimizeFile(_ context.Context, _ io.Reader, _ io.Writer) error {
	return fmt.Errorf("PDF optimization not yet implemented")
}

func (r *RepairServiceImpl) generateRepairActions(err *domain.ValidationError) []ports.RepairAction {
	actions := make([]ports.RepairAction, 0)

	switch err.Code {
	case ErrorCodePDFTrailer003:
		actions = append(actions, ports.RepairAction{
			Type:        "append_eof_marker",
			Description: "Append missing %%EOF marker to end of file",
			Target:      "trailer",
			Details: map[string]interface{}{
				"marker": "%%EOF",
			},
			Automated: true,
		})

	case ErrorCodePDFTrailer001:
		actions = append(actions, ports.RepairAction{
			Type:        "recompute_startxref",
			Description: "Recompute startxref offset by scanning for xref table",
			Target:      "trailer",
			Details:     map[string]interface{}{},
			Automated:   true,
		})

	case ErrorCodePDFTrailer002:
		actions = append(actions, ports.RepairAction{
			Type:        "fix_trailer_typos",
			Description: "Attempt to fix minor trailer dictionary typos",
			Target:      "trailer",
			Details:     map[string]interface{}{},
			Automated:   true,
		})

	case ErrorCodePDFHeader001,
		ErrorCodePDFHeader002:
		actions = append(actions, ports.RepairAction{
			Type:        "manual_header_fix",
			Description: "Header modification requires manual intervention (unsafe)",
			Target:      "header",
			Details: map[string]interface{}{
				"reason": "Modifying file header can corrupt document structure",
			},
			Automated: false,
		})

	case ErrorCodePDFXref001,
		ErrorCodePDFXref002,
		ErrorCodePDFXref003:
		actions = append(actions, ports.RepairAction{
			Type:        "manual_xref_rebuild",
			Description: "Cross-reference table rebuild requires manual intervention (unsafe)",
			Target:      "xref",
			Details: map[string]interface{}{
				"reason": "Cross-reference rebuild may alter object structure",
			},
			Automated: false,
		})

	case ErrorCodePDFCatalog001,
		ErrorCodePDFCatalog002,
		ErrorCodePDFCatalog003:
		actions = append(actions, ports.RepairAction{
			Type:        "manual_catalog_fix",
			Description: "Catalog repair requires manual intervention (unsafe)",
			Target:      "catalog",
			Details: map[string]interface{}{
				"reason": "Catalog modifications affect document structure",
			},
			Automated: false,
		})

	default:
		actions = append(actions, ports.RepairAction{
			Type:        "manual_review",
			Description: fmt.Sprintf("Requires manual review: %s", err.Message),
			Target:      "unknown",
			Details:     err.Details,
			Automated:   false,
		})
	}

	return actions
}

type repairContext struct {
	actions []ports.RepairAction
	data    []byte
	applied []ports.RepairAction
}

func (r *RepairServiceImpl) applyRepairs(_ context.Context, repairCtx *repairContext) error {
	actionsByType := make(map[string][]ports.RepairAction)
	for _, action := range repairCtx.actions {
		if action.Automated {
			actionsByType[action.Type] = append(actionsByType[action.Type], action)
		}
	}

	for _, action := range actionsByType["append_eof_marker"] {
		r.appendEOFMarker(repairCtx)
		repairCtx.applied = append(repairCtx.applied, action)
	}

	for _, action := range actionsByType["recompute_startxref"] {
		if err := r.recomputeStartxref(repairCtx); err != nil {
			return fmt.Errorf("failed to recompute startxref: %w", err)
		}
		repairCtx.applied = append(repairCtx.applied, action)
	}

	for _, action := range actionsByType["fix_trailer_typos"] {
		r.fixTrailerTypos(repairCtx)
		repairCtx.applied = append(repairCtx.applied, action)
	}

	return nil
}

func (r *RepairServiceImpl) appendEOFMarker(repairCtx *repairContext) {
	eofMarker := []byte("%%EOF")
	lastBytes := repairCtx.data
	if len(repairCtx.data) > 1024 {
		lastBytes = repairCtx.data[len(repairCtx.data)-1024:]
	}

	if bytes.Contains(lastBytes, eofMarker) {
		return
	}

	if len(repairCtx.data) > 0 && repairCtx.data[len(repairCtx.data)-1] != '\n' {
		repairCtx.data = append(repairCtx.data, '\n')
	}

	repairCtx.data = append(repairCtx.data, eofMarker...)
	repairCtx.data = append(repairCtx.data, '\n')
}

func (r *RepairServiceImpl) recomputeStartxref(repairCtx *repairContext) error {
	xrefPattern := regexp.MustCompile(`\bxref\s+\d+\s+\d+`)
	matches := xrefPattern.FindAllIndex(repairCtx.data, -1)

	if len(matches) == 0 {
		return fmt.Errorf("no xref table found in document")
	}

	lastXrefOffset := int64(matches[len(matches)-1][0])

	startxrefPattern := regexp.MustCompile(`startxref\s+\d+`)
	startxrefMatch := startxrefPattern.FindIndex(repairCtx.data)

	if startxrefMatch == nil {
		eofPattern := []byte("%%EOF")
		eofIndex := bytes.LastIndex(repairCtx.data, eofPattern)
		if eofIndex == -1 {
			return fmt.Errorf("cannot add startxref: no %%EOF marker found")
		}

		newStartxref := fmt.Sprintf("startxref\n%d\n", lastXrefOffset)
		repairCtx.data = append(repairCtx.data[:eofIndex], append([]byte(newStartxref), repairCtx.data[eofIndex:]...)...)
	} else {
		oldStartxref := repairCtx.data[startxrefMatch[0]:startxrefMatch[1]]
		newStartxref := fmt.Sprintf("startxref\n%d", lastXrefOffset)

		repairCtx.data = bytes.Replace(repairCtx.data, oldStartxref, []byte(newStartxref), 1)
	}

	return nil
}

func (r *RepairServiceImpl) fixTrailerTypos(repairCtx *repairContext) {
	trailerPattern := regexp.MustCompile(`trailer\s*<<`)
	if !trailerPattern.Match(repairCtx.data) {
		typoPattern := regexp.MustCompile(`(?i)traler\s*<<|trailer\s*<[^<]|trailer\s+<<`)
		repairCtx.data = typoPattern.ReplaceAll(repairCtx.data, []byte("trailer <<"))
	}

	sizePattern := regexp.MustCompile(`/Size\s+\d+`)
	if !sizePattern.Match(repairCtx.data) {
		typoPattern := regexp.MustCompile(`(?i)/Sise\s+\d+|/size\s+\d+`)
		repairCtx.data = typoPattern.ReplaceAllFunc(repairCtx.data, func(match []byte) []byte {
			parts := regexp.MustCompile(`\d+`).FindSubmatch(match)
			if len(parts) > 0 {
				return []byte(fmt.Sprintf("/Size %s", parts[0]))
			}
			return match
		})
	}

	rootPattern := regexp.MustCompile(`/Root\s+\d+\s+\d+\s+R`)
	if !rootPattern.Match(repairCtx.data) {
		typoPattern := regexp.MustCompile(`(?i)/root\s+\d+\s+\d+\s+R`)
		repairCtx.data = typoPattern.ReplaceAllFunc(repairCtx.data, func(match []byte) []byte {
			return []byte(strings.Replace(string(match), "/root", "/Root", 1))
		})
	}
}

func (r *RepairServiceImpl) generateOutputPath(filePath string) string {
	ext := filepath.Ext(filePath)
	base := strings.TrimSuffix(filePath, ext)
	return base + repairSuffix
}

func (r *RepairServiceImpl) convertValidationResult(filePath string, result *StructureValidationResult) *domain.ValidationReport {
	report := &domain.ValidationReport{
		FilePath: filePath,
		FileType: "PDF",
		IsValid:  result.Valid,
		Errors:   make([]domain.ValidationError, 0),
	}

	for _, err := range result.Errors {
		report.Errors = append(report.Errors, domain.ValidationError{
			Code:    err.Code,
			Message: err.Message,
			Details: err.Details,
		})
	}

	return report
}
