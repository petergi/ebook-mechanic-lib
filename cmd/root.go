package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/example/project/internal/adapters/reporter"
	"github.com/example/project/internal/cli"
	"github.com/example/project/internal/ports"
)

const (
	appName = "ebmlib"
)

type rootFlags struct {
	format      string
	output      string
	verbose     bool
	color       bool
	minSeverity string
	severities  []string
	maxErrors   int
}

func newRootCmd() *cobra.Command {
	flags := &rootFlags{}

	cmd := &cobra.Command{
		Use:   appName,
		Short: "Validate and repair EPUB and PDF files",
		Long:  "EBMLib CLI validates and repairs EPUB and PDF files with configurable output formats.",
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			return cli.ValidateFormat(flags.format)
		},
	}

	cmd.PersistentFlags().StringVarP(&flags.format, "format", "f", "text", "Output format: text, json, markdown")
	cmd.PersistentFlags().StringVarP(&flags.output, "output", "o", "", "Write report to file instead of stdout")
	cmd.PersistentFlags().BoolVarP(&flags.verbose, "verbose", "v", false, "Enable verbose output")
	cmd.PersistentFlags().BoolVar(&flags.color, "color", true, "Enable colorized output")
	cmd.PersistentFlags().StringVar(&flags.minSeverity, "min-severity", "", "Minimum severity to include (info, warning, error)")
	cmd.PersistentFlags().StringSliceVar(&flags.severities, "severity", nil, "Include only specific severities (repeatable)")
	cmd.PersistentFlags().IntVar(&flags.maxErrors, "max-errors", 0, "Limit number of errors per report (0 = unlimited)")

	cmd.AddCommand(newValidateCmd(flags))
	cmd.AddCommand(newRepairCmd(flags))
	cmd.AddCommand(newBatchCmd(flags))

	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)

	return cmd
}

func buildReportOptions(flags *rootFlags) (*ports.ReportOptions, *reporter.Filter, error) {
	format, err := cli.ParseFormat(flags.format)
	if err != nil {
		return nil, nil, err
	}

	filter, err := cli.BuildSeverityFilter(flags.minSeverity, flags.severities)
	if err != nil {
		return nil, nil, err
	}

	return &ports.ReportOptions{
		Format:          format,
		IncludeWarnings: true,
		IncludeInfo:     true,
		Verbose:         flags.verbose,
		ColorEnabled:    flags.color,
		MaxErrors:       flags.maxErrors,
	}, filter, nil
}

func withSignalContext(parent context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(parent)
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		select {
		case <-ctx.Done():
			return
		case <-ch:
			cancel()
		}
	}()
	return ctx, cancel
}

func Execute() {
	cmd := newRootCmd()
	if err := cmd.Execute(); err != nil {
		var exitErr cli.ExitError
		if errors.As(err, &exitErr) {
			if exitErr.Err != nil {
				fmt.Fprintln(os.Stderr, exitErr.Err)
			}
			os.Exit(exitErr.Code)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(cli.ExitCodeInternal)
	}
}
