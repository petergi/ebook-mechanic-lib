package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/example/project/internal/adapters/reporter"
	"github.com/example/project/internal/domain"
	"github.com/example/project/internal/ports"
	"github.com/example/project/pkg/ebmlib"
)

const (
	ExitCodeOK       = 0
	ExitCodeWarning  = 1
	ExitCodeError    = 2
	ExitCodeInternal = 3
)

type ExitError struct {
	Code int
	Err  error
}

func (e ExitError) Error() string {
	if e.Err == nil {
		return ""
	}
	return e.Err.Error()
}

func ParseFormat(raw string) (ports.OutputFormat, error) {
	switch strings.ToLower(raw) {
	case "json":
		return ports.FormatJSON, nil
	case "text":
		return ports.FormatText, nil
	case "markdown", "md":
		return ports.FormatMarkdown, nil
	default:
		return "", fmt.Errorf("unsupported format %q", raw)
	}
}

func ValidateFormat(raw string) error {
	_, err := ParseFormat(raw)
	return err
}

func BuildReporter(format ports.OutputFormat, filter *reporter.Filter) ports.Reporter {
	switch format {
	case ports.FormatJSON:
		if filter != nil {
			return reporter.NewJSONReporterWithFilter(filter)
		}
		return reporter.NewJSONReporter()
	case ports.FormatMarkdown:
		if filter != nil {
			return reporter.NewMarkdownReporterWithFilter(filter)
		}
		return reporter.NewMarkdownReporter()
	default:
		if filter != nil {
			return reporter.NewTextReporterWithFilter(filter)
		}
		return reporter.NewTextReporter()
	}
}

func BuildSeverityFilter(min string, severities []string) (*reporter.Filter, error) {
	if min == "" && len(severities) == 0 {
		return nil, nil
	}

	filter := reporter.NewFilter()
	if min != "" {
		sev, err := parseSeverity(min)
		if err != nil {
			return nil, err
		}
		filter.MinSeverity = sev
	}

	if len(severities) > 0 {
		for _, raw := range severities {
			sev, err := parseSeverity(raw)
			if err != nil {
				return nil, err
			}
			filter.Severities = append(filter.Severities, sev)
		}
	}

	return filter, nil
}

func parseSeverity(raw string) (domain.Severity, error) {
	switch strings.ToLower(raw) {
	case "error", "errors":
		return domain.SeverityError, nil
	case "warning", "warnings":
		return domain.SeverityWarning, nil
	case "info", "information":
		return domain.SeverityInfo, nil
	default:
		return "", fmt.Errorf("unsupported severity %q", raw)
	}
}

func ExitWithReport(report *domain.ValidationReport) error {
	if report == nil {
		return ExitError{Code: ExitCodeInternal, Err: errors.New("no report generated")}
	}

	if report.HasErrors() {
		return ExitError{Code: ExitCodeError}
	}
	if report.HasWarnings() {
		return ExitError{Code: ExitCodeWarning}
	}
	return ExitError{Code: ExitCodeOK}
}

func ExitWithBatchResult(result BatchResult) error {
	if result.InternalError != nil {
		return ExitError{Code: ExitCodeInternal, Err: result.InternalError}
	}
	if result.HasErrors {
		return ExitError{Code: ExitCodeError}
	}
	if result.HasWarnings {
		return ExitError{Code: ExitCodeWarning}
	}
	return ExitError{Code: ExitCodeOK}
}

func ValidateFile(ctx context.Context, path string) (*domain.ValidationReport, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".epub":
		return ebmlib.ValidateEPUBWithContext(ctx, path)
	case ".pdf":
		return ebmlib.ValidatePDFWithContext(ctx, path)
	default:
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}
}

func ValidateReader(ctx context.Context, reader io.Reader, size int64, fileType string) (*domain.ValidationReport, error) {
	switch strings.ToLower(fileType) {
	case "epub":
		return ebmlib.ValidateEPUBReaderWithContext(ctx, reader, size)
	case "pdf":
		return ebmlib.ValidatePDFReaderWithContext(ctx, reader)
	default:
		return nil, fmt.Errorf("unsupported file type: %s", fileType)
	}
}

func RepairFile(ctx context.Context, path string, opts RepairOptions) (*ports.RepairResult, *domain.ValidationReport, error) {
	fileType := strings.ToLower(filepath.Ext(path))
	report, err := ValidateFile(ctx, path)
	if err != nil {
		return nil, nil, err
	}

	if report.IsValid {
		return &ports.RepairResult{Success: true, ActionsApplied: []ports.RepairAction{}}, report, nil
	}

	var preview *ports.RepairPreview
	switch fileType {
	case ".epub":
		preview, err = ebmlib.PreviewEPUBRepairWithContext(ctx, path)
	case ".pdf":
		preview, err = ebmlib.PreviewPDFRepairWithContext(ctx, path)
	default:
		return nil, nil, fmt.Errorf("unsupported file type: %s", fileType)
	}
	if err != nil {
		return nil, report, err
	}

	outputPath := opts.OutputPath
	if !opts.InPlace && outputPath == "" {
		outputPath = defaultRepairedPath(path)
	}
	if opts.InPlace {
		outputPath, err = createTempPath(path)
		if err != nil {
			return nil, report, err
		}
	}

	var result *ports.RepairResult
	switch fileType {
	case ".epub":
		result, err = ebmlib.RepairEPUBWithPreviewContext(ctx, path, preview, outputPath)
	case ".pdf":
		result, err = ebmlib.RepairPDFWithPreviewContext(ctx, path, preview, outputPath)
	}
	if err != nil {
		return result, report, err
	}

	if opts.InPlace {
		backupPath := ""
		if opts.Backup {
			backupPath, err = backupFile(path, opts.BackupDir)
			if err != nil {
				return result, report, err
			}
		}
		if err := os.Rename(outputPath, path); err != nil {
			return result, report, err
		}
		if backupPath != "" {
			result.BackupPath = backupPath
		}
	}

	finalReport, err := ValidateFile(ctx, path)
	if err != nil {
		return result, report, err
	}

	return result, finalReport, nil
}

func defaultRepairedPath(path string) string {
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(path, ext)
	return base + ".repaired" + ext
}

func createTempPath(path string) (string, error) {
	dir := filepath.Dir(path)
	ext := filepath.Ext(path)
	file, err := os.CreateTemp(dir, "ebmlib-repair-*")
	if err != nil {
		return "", err
	}
	name := file.Name()
	if err := file.Close(); err != nil {
		return "", err
	}
	if ext != "" {
		rename := name + ext
		if err := os.Rename(name, rename); err != nil {
			return "", err
		}
		return rename, nil
	}
	return name, nil
}

func backupFile(path string, backupDir string) (string, error) {
	dir := filepath.Dir(path)
	if backupDir != "" {
		dir = backupDir
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	backupPath := filepath.Join(dir, filepath.Base(path)+".bak")
	input, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(backupPath, input, 0600); err != nil {
		return "", err
	}
	return backupPath, nil
}
