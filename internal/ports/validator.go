package ports

import (
	"context"
	"io"

	"github.com/example/project/internal/domain"
)

type EPUBValidator interface {
	ValidateFile(ctx context.Context, filePath string) (*domain.ValidationReport, error)
	ValidateReader(ctx context.Context, reader io.Reader, size int64) (*domain.ValidationReport, error)
	ValidateStructure(ctx context.Context, filePath string) (*domain.ValidationReport, error)
	ValidateMetadata(ctx context.Context, filePath string) (*domain.ValidationReport, error)
	ValidateContent(ctx context.Context, filePath string) (*domain.ValidationReport, error)
}

type PDFValidator interface {
	ValidateFile(ctx context.Context, filePath string) (*domain.ValidationReport, error)
	ValidateReader(ctx context.Context, reader io.Reader, size int64) (*domain.ValidationReport, error)
	ValidateStructure(ctx context.Context, filePath string) (*domain.ValidationReport, error)
	ValidateMetadata(ctx context.Context, filePath string) (*domain.ValidationReport, error)
	ValidateCompliance(ctx context.Context, filePath string, standard string) (*domain.ValidationReport, error)
}
