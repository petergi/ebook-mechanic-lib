package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/petergi/ebook-mechanic-lib/internal/cli"
	"github.com/petergi/ebook-mechanic-lib/internal/domain"
)

type validateFlags struct {
	fileType string
}

func writeValidationReport(ctx context.Context, cmd *cobra.Command, root *rootFlags, report *domain.ValidationReport) error {
	options, filter, err := buildReportOptions(root)
	if err != nil {
		return err
	}
	if root.output != "" {
		if err := os.MkdirAll(filepath.Dir(root.output), 0750); err != nil {
			return err
		}
	}

	rep := cli.BuildReporter(options.Format, filter)

	if root.output != "" {
		if err := rep.WriteToFile(ctx, report, root.output, options); err != nil {
			return err
		}
	} else {
		if err := rep.Write(ctx, report, cmd.OutOrStdout(), options); err != nil {
			return err
		}
	}

	return cli.ExitWithReport(report)
}

func newValidateCmd(root *rootFlags) *cobra.Command {
	flags := &validateFlags{}

	cmd := &cobra.Command{
		Use:   "validate <file>|-",
		Short: "Validate a single EPUB or PDF file",
		Long:  "Validate an EPUB or PDF file, or read from stdin when path is '-'.",
		Example: strings.Join([]string{
			"  ebm-cli validate book.epub",
			"  ebm-cli validate document.pdf --format json",
			"  cat book.epub | ebm-cli validate - --type epub",
		}, "\n"),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := withSignalContext(context.Background())
			defer cancel()

			target := args[0]
			var err error
			var report *domain.ValidationReport

			if target == "-" {
				if flags.fileType == "" {
					return fmt.Errorf("stdin requires --type epub or pdf")
				}
				data, err := io.ReadAll(bufio.NewReader(os.Stdin))
				if err != nil {
					return fmt.Errorf("read stdin: %w", err)
				}
				report, err = cli.ValidateReader(ctx, bytes.NewReader(data), int64(len(data)), flags.fileType)
				if err != nil {
					return err
				}
			} else {
				report, err = cli.ValidateFile(ctx, target)
				if err != nil {
					return err
				}
			}

			return writeValidationReport(ctx, cmd, root, report)
		},
	}

	cmd.Flags().StringVar(&flags.fileType, "type", "", "Specify file type when reading from stdin (epub, pdf)")
	return cmd
}
