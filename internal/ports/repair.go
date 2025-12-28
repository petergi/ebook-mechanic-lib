package ports

import (
	"context"
	"io"

	"github.com/example/project/internal/domain"
)

type RepairAction struct {
	Type        string
	Description string
	Target      string
	Details     map[string]interface{}
	Automated   bool
}

type RepairPreview struct {
	Actions       []RepairAction
	CanAutoRepair bool
	EstimatedTime int64
	BackupRequired bool
	Warnings      []string
}

type RepairResult struct {
	Success       bool
	ActionsApplied []RepairAction
	Report        *domain.ValidationReport
	BackupPath    string
	Error         error
}

type RepairService interface {
	Preview(ctx context.Context, report *domain.ValidationReport) (*RepairPreview, error)
	Apply(ctx context.Context, filePath string, preview *RepairPreview) (*RepairResult, error)
	ApplyWithBackup(ctx context.Context, filePath string, preview *RepairPreview, backupPath string) (*RepairResult, error)
	CanRepair(ctx context.Context, err *domain.ValidationError) bool
	CreateBackup(ctx context.Context, filePath string, backupPath string) error
	RestoreBackup(ctx context.Context, backupPath string, originalPath string) error
}

type EPUBRepairService interface {
	RepairService
	RepairStructure(ctx context.Context, filePath string) (*RepairResult, error)
	RepairMetadata(ctx context.Context, filePath string) (*RepairResult, error)
	RepairContent(ctx context.Context, filePath string) (*RepairResult, error)
}

type PDFRepairService interface {
	RepairService
	RepairStructure(ctx context.Context, filePath string) (*RepairResult, error)
	RepairMetadata(ctx context.Context, filePath string) (*RepairResult, error)
	OptimizeFile(ctx context.Context, reader io.Reader, writer io.Writer) error
}
