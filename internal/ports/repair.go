// Package ports defines interfaces for adapters and services.
package ports

import (
	"context"
	"io"

	"github.com/petergi/ebook-mechanic-lib/internal/domain"
)

// RepairAction describes a single repair step.
type RepairAction struct {
	Type        string
	Description string
	Target      string
	Details     map[string]interface{}
	Automated   bool
}

// RepairPreview summarizes candidate repairs and constraints.
type RepairPreview struct {
	Actions        []RepairAction
	CanAutoRepair  bool
	EstimatedTime  int64
	BackupRequired bool
	Warnings       []string
}

// RepairResult captures applied repairs and resulting report.
type RepairResult struct {
	Success        bool
	ActionsApplied []RepairAction
	Report         *domain.ValidationReport
	BackupPath     string
	Error          error
}

// RepairService defines the generic repair workflow.
type RepairService interface {
	Preview(ctx context.Context, report *domain.ValidationReport) (*RepairPreview, error)
	Apply(ctx context.Context, filePath string, preview *RepairPreview) (*RepairResult, error)
	ApplyWithBackup(ctx context.Context, filePath string, preview *RepairPreview, backupPath string) (*RepairResult, error)
	CanRepair(ctx context.Context, err *domain.ValidationError) bool
	CreateBackup(ctx context.Context, filePath string, backupPath string) error
	RestoreBackup(ctx context.Context, backupPath string, originalPath string) error
}

// EPUBRepairService specializes repairs for EPUB files.
type EPUBRepairService interface {
	RepairService
	RepairStructure(ctx context.Context, filePath string) (*RepairResult, error)
	RepairMetadata(ctx context.Context, filePath string) (*RepairResult, error)
	RepairContent(ctx context.Context, filePath string) (*RepairResult, error)
}

// PDFRepairService specializes repairs for PDF files.
type PDFRepairService interface {
	RepairService
	RepairStructure(ctx context.Context, filePath string) (*RepairResult, error)
	RepairMetadata(ctx context.Context, filePath string) (*RepairResult, error)
	OptimizeFile(ctx context.Context, reader io.Reader, writer io.Writer) error
}
