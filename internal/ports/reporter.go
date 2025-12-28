package ports

import (
	"context"
	"io"

	"github.com/example/project/internal/domain"
)

type OutputFormat string

const (
	FormatJSON     OutputFormat = "json"
	FormatText     OutputFormat = "text"
	FormatHTML     OutputFormat = "html"
	FormatXML      OutputFormat = "xml"
	FormatMarkdown OutputFormat = "markdown"
)

type ReportOptions struct {
	Format         OutputFormat
	IncludeWarnings bool
	IncludeInfo     bool
	Verbose         bool
	ColorEnabled    bool
	MaxErrors       int
}

type Reporter interface {
	Format(ctx context.Context, report *domain.ValidationReport, options *ReportOptions) (string, error)
	Write(ctx context.Context, report *domain.ValidationReport, writer io.Writer, options *ReportOptions) error
	WriteToFile(ctx context.Context, report *domain.ValidationReport, filePath string, options *ReportOptions) error
}

type MultiReporter interface {
	Reporter
	FormatMultiple(ctx context.Context, reports []*domain.ValidationReport, options *ReportOptions) (string, error)
	WriteMultiple(ctx context.Context, reports []*domain.ValidationReport, writer io.Writer, options *ReportOptions) error
	WriteSummary(ctx context.Context, reports []*domain.ValidationReport, writer io.Writer, options *ReportOptions) error
}
