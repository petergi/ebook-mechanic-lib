package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/petergi/ebook-mechanic-lib/internal/cli"
)

type batchFlags struct {
	jobs        int
	queue       int
	maxDepth    int
	extensions  []string
	ignore      []string
	progress    string
	backup      bool
	backupDir   string
	inPlace     bool
	summaryOnly bool
}

func newBatchCmd(root *rootFlags) *cobra.Command {
	flags := &batchFlags{}

	cmd := &cobra.Command{
		Use:   "batch",
		Short: "Batch validation and repair",
		Long:  "Run validation or repair across multiple files using a worker pool.",
	}

	cmd.PersistentFlags().IntVarP(&flags.jobs, "jobs", "j", 4, "Number of parallel workers")
	cmd.PersistentFlags().IntVar(&flags.queue, "queue", 64, "Job queue buffer size")
	cmd.PersistentFlags().IntVar(&flags.maxDepth, "max-depth", -1, "Maximum directory depth to traverse (-1 = unlimited)")
	cmd.PersistentFlags().StringSliceVar(&flags.extensions, "ext", []string{".epub", ".pdf"}, "File extensions to include")
	cmd.PersistentFlags().StringSliceVar(&flags.ignore, "ignore", nil, "Glob patterns to ignore")
	cmd.PersistentFlags().StringVar(&flags.progress, "progress", "auto", "Progress output: auto, simple, none")
	cmd.PersistentFlags().BoolVar(&flags.summaryOnly, "summary-only", false, "Only print summary output")

	validateCmd := &cobra.Command{
		Use:   "validate <paths...>",
		Short: "Validate multiple files",
		Example: strings.Join([]string{
			"  ebm-cli batch validate ./books",
			"  ebm-cli batch validate ./library --ext .epub --jobs 8",
			"  ebm-cli batch validate ./books/*.pdf --format json",
		}, "\n"),
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := withSignalContext(context.Background())
			defer cancel()

			options, filter, err := buildReportOptions(root)
			if err != nil {
				return err
			}

			batchOptions := cli.BatchOptions{
				Workers:     flags.jobs,
				QueueSize:   flags.queue,
				MaxDepth:    flags.maxDepth,
				Extensions:  flags.extensions,
				Ignore:      flags.ignore,
				Progress:    flags.progress,
				SummaryOnly: flags.summaryOnly,
				OutputPath:  root.output,
			}

			result, err := cli.RunBatchValidate(ctx, args, batchOptions, options, filter, cmd.OutOrStdout())
			if err != nil {
				return err
			}
			return cli.ExitWithBatchResult(result)
		},
	}

	repairCmd := &cobra.Command{
		Use:   "repair <paths...>",
		Short: "Repair multiple files",
		Example: strings.Join([]string{
			"  ebm-cli batch repair ./books --in-place --backup",
			"  ebm-cli batch repair ./library --jobs 4 --backup-dir ./backups",
		}, "\n"),
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := withSignalContext(context.Background())
			defer cancel()

			if flags.backupDir != "" && !flags.backup {
				return fmt.Errorf("--backup-dir requires --backup")
			}

			options, filter, err := buildReportOptions(root)
			if err != nil {
				return err
			}

			batchOptions := cli.BatchOptions{
				Workers:     flags.jobs,
				QueueSize:   flags.queue,
				MaxDepth:    flags.maxDepth,
				Extensions:  flags.extensions,
				Ignore:      flags.ignore,
				Progress:    flags.progress,
				SummaryOnly: flags.summaryOnly,
				OutputPath:  root.output,
				Repair: cli.RepairOptions{
					InPlace:   flags.inPlace,
					Backup:    flags.backup,
					BackupDir: flags.backupDir,
				},
			}

			result, err := cli.RunBatchRepair(ctx, args, batchOptions, options, filter, cmd.OutOrStdout())
			if err != nil {
				return err
			}
			return cli.ExitWithBatchResult(result)
		},
	}

	repairCmd.Flags().BoolVar(&flags.inPlace, "in-place", false, "Repair files in place using atomic replace")
	repairCmd.Flags().BoolVar(&flags.backup, "backup", false, "Create backup before in-place repair")
	repairCmd.Flags().StringVar(&flags.backupDir, "backup-dir", "", "Directory to place backups")
	cmd.AddCommand(validateCmd)
	cmd.AddCommand(repairCmd)

	return cmd
}
