package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/petergi/ebook-mechanic-lib/internal/adapters/reporter"
	"github.com/petergi/ebook-mechanic-lib/internal/batch"
	"github.com/petergi/ebook-mechanic-lib/internal/domain"
	"github.com/petergi/ebook-mechanic-lib/internal/ports"
)

type progressStyle struct {
	useColor bool
	mode     string
}

// RunBatchValidate validates a collection of files.
func RunBatchValidate(ctx context.Context, targets []string, opts BatchOptions, reportOpts *ports.ReportOptions, filter *reporter.Filter, out io.Writer) (BatchResult, error) {
	items, err := collectTargets(targets, opts)
	if err != nil {
		return BatchResult{InternalError: err}, err
	}
	if len(items) == 0 {
		return BatchResult{InternalError: fmt.Errorf("no files matched")}, fmt.Errorf("no files matched")
	}

	rep := BuildReporter(reportOpts.Format, filter)
	multi := buildMultiReporter(rep)

	style := progressStyle{useColor: reportOpts.ColorEnabled, mode: opts.Progress}
	progress := buildProgress(out, style)

	engineResult := batch.Run(ctx, items, batch.Config{Workers: opts.Workers, QueueSize: opts.QueueSize}, func(ctx context.Context, path string) batch.ItemResult {
		start := time.Now()
		report, err := ValidateFile(ctx, path)
		return batch.ItemResult{Path: path, Value: report, Err: err, Duration: time.Since(start)}
	}, progress)

	result := summarizeBatch(engineResult.Items)
	if result.InternalError != nil {
		return result, result.InternalError
	}

	if !opts.SummaryOnly {
		for _, report := range result.Reports {
			if err := rep.Write(ctx, report, out, reportOpts); err != nil {
				return BatchResult{InternalError: err}, err
			}
		}
	}

	if multi != nil {
		if opts.OutputPath != "" {
			if err := os.MkdirAll(filepath.Dir(opts.OutputPath), 0750); err != nil {
				return BatchResult{InternalError: err}, err
			}
			file, err := os.Create(opts.OutputPath)
			if err != nil {
				return BatchResult{InternalError: err}, err
			}
			defer func() {
				_ = file.Close()
			}()
			if err := multi.WriteSummary(ctx, result.Reports, file, reportOpts); err != nil {
				return BatchResult{InternalError: err}, err
			}
		} else if err := multi.WriteSummary(ctx, result.Reports, out, reportOpts); err != nil {
			return BatchResult{InternalError: err}, err
		}
	}

	return result, nil
}

// RunBatchRepair repairs a collection of files.
func RunBatchRepair(ctx context.Context, targets []string, opts BatchOptions, reportOpts *ports.ReportOptions, filter *reporter.Filter, out io.Writer) (BatchResult, error) {
	items, err := collectTargets(targets, opts)
	if err != nil {
		return BatchResult{InternalError: err}, err
	}
	if len(items) == 0 {
		return BatchResult{InternalError: fmt.Errorf("no files matched")}, fmt.Errorf("no files matched")
	}

	rep := BuildReporter(reportOpts.Format, filter)
	multi := buildMultiReporter(rep)

	style := progressStyle{useColor: reportOpts.ColorEnabled, mode: opts.Progress}
	progress := buildProgress(out, style)

	engineResult := batch.Run(ctx, items, batch.Config{Workers: opts.Workers, QueueSize: opts.QueueSize}, func(ctx context.Context, path string) batch.ItemResult {
		start := time.Now()
		localOpts := opts.Repair
		if !localOpts.InPlace && localOpts.OutputPath == "" {
			localOpts.OutputPath = defaultRepairedPath(path)
		}
		result, report, err := RepairFile(ctx, path, localOpts)
		if report == nil {
			return batch.ItemResult{Path: path, Value: result, Err: err, Duration: time.Since(start)}
		}
		return batch.ItemResult{Path: path, Value: report, Err: err, Duration: time.Since(start)}
	}, progress)

	result := summarizeBatch(engineResult.Items)
	if result.InternalError != nil {
		return result, result.InternalError
	}

	if !opts.SummaryOnly {
		for _, report := range result.Reports {
			if err := rep.Write(ctx, report, out, reportOpts); err != nil {
				return BatchResult{InternalError: err}, err
			}
		}
	}

	if multi != nil {
		if opts.OutputPath != "" {
			if err := os.MkdirAll(filepath.Dir(opts.OutputPath), 0750); err != nil {
				return BatchResult{InternalError: err}, err
			}
			file, err := os.Create(opts.OutputPath)
			if err != nil {
				return BatchResult{InternalError: err}, err
			}
			defer func() {
				_ = file.Close()
			}()
			if err := multi.WriteSummary(ctx, result.Reports, file, reportOpts); err != nil {
				return BatchResult{InternalError: err}, err
			}
		} else if err := multi.WriteSummary(ctx, result.Reports, out, reportOpts); err != nil {
			return BatchResult{InternalError: err}, err
		}
	}

	return result, nil
}

func collectTargets(targets []string, opts BatchOptions) ([]string, error) {
	expanded, err := batch.ExpandTargets(targets)
	if err != nil {
		return nil, err
	}
	return batch.DiscoverFiles(expanded, batch.DiscoverOptions{MaxDepth: opts.MaxDepth, Extensions: opts.Extensions, Ignore: opts.Ignore})
}

func summarizeBatch(items []batch.ItemResult) BatchResult {
	result := BatchResult{Reports: make([]*domain.ValidationReport, 0, len(items))}
	for _, item := range items {
		result.Total++
		if item.Err != nil {
			result.Failed++
			if result.InternalError == nil {
				result.InternalError = item.Err
			}
			continue
		}
		report, ok := item.Value.(*domain.ValidationReport)
		if !ok || report == nil {
			result.Skipped++
			continue
		}
		result.Reports = append(result.Reports, report)
		result.Processed++
		if report.HasErrors() {
			result.HasErrors = true
		}
		if report.HasWarnings() {
			result.HasWarnings = true
		}
	}
	return result
}

func buildMultiReporter(rep ports.Reporter) ports.MultiReporter {
	multi, ok := rep.(ports.MultiReporter)
	if !ok {
		return nil
	}
	return multi
}

func buildProgress(out io.Writer, style progressStyle) batch.ProgressFunc {
	mode := style.mode
	if mode == "" || mode == "auto" {
		mode = "simple"
	}
	if mode == "none" {
		return nil
	}
	return func(update batch.ProgressUpdate) {
		status := "OK"
		if update.Err != nil {
			status = "ERR"
		} else if report, ok := update.Value.(*domain.ValidationReport); ok {
			if report.HasErrors() {
				status = "ERR"
			} else if report.HasWarnings() {
				status = "WARN"
			}
		}
		line := fmt.Sprintf("[%s] %d/%d %s", status, update.Completed, update.Total, update.Path)
		if style.useColor {
			line = colorize(status, line)
		}
		_, _ = fmt.Fprintln(out, line)
	}
}

func colorize(status string, line string) string {
	switch status {
	case "OK":
		return "\033[32m" + line + "\033[0m"
	case "ERR":
		return "\033[31m" + line + "\033[0m"
	case "WARN":
		return "\033[33m" + line + "\033[0m"
	default:
		return line
	}
}
