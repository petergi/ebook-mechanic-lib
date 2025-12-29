package ports

import (
	"context"
	"io"

	"github.com/example/project/internal/domain"
)

// OutputFormat identifies the output format for reports.
type OutputFormat string

const (
	// FormatJSON renders reports as JSON.
	FormatJSON OutputFormat = "json"
	// FormatText renders reports as plain text.
	FormatText OutputFormat = "text"
	// FormatHTML renders reports as HTML.
	FormatHTML OutputFormat = "html"
	// FormatXML renders reports as XML.
	FormatXML OutputFormat = "xml"
	// FormatMarkdown renders reports as Markdown.
	FormatMarkdown OutputFormat = "markdown"
)

// ReportOptions configures report formatting and filtering.
type ReportOptions struct {
	Format          OutputFormat
	IncludeWarnings bool
	IncludeInfo     bool
	Verbose         bool
	ColorEnabled    bool
	MaxErrors       int
}

// Reporter formats and writes single validation reports.
type Reporter interface {
	Format(ctx context.Context, report *domain.ValidationReport, options *ReportOptions) (string, error)
	Write(ctx context.Context, report *domain.ValidationReport, writer io.Writer, options *ReportOptions) error
	WriteToFile(ctx context.Context, report *domain.ValidationReport, filePath string, options *ReportOptions) error
}

// MultiReporter formats and writes multiple reports plus summaries.
type MultiReporter interface {
	Reporter
	FormatMultiple(ctx context.Context, reports []*domain.ValidationReport, options *ReportOptions) (string, error)
	WriteMultiple(ctx context.Context, reports []*domain.ValidationReport, writer io.Writer, options *ReportOptions) error
	WriteSummary(ctx context.Context, reports []*domain.ValidationReport, writer io.Writer, options *ReportOptions) error
}
