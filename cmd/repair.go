package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/example/project/internal/cli"
)

type repairFlags struct {
	output    string
	inPlace   bool
	backup    bool
	backupDir string
}

func newRepairCmd(root *rootFlags) *cobra.Command {
	flags := &repairFlags{}

	cmd := &cobra.Command{
		Use:   "repair <file>",
		Short: "Repair an EPUB or PDF file",
		Long:  "Validate and attempt automatic repairs for a single EPUB or PDF file.",
		Example: strings.Join([]string{
			"  ebmlib repair book.epub",
			"  ebmlib repair document.pdf --in-place --backup",
			"  ebmlib repair book.epub --output fixed.epub",
		}, "\n"),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := withSignalContext(context.Background())
			defer cancel()

			if flags.inPlace && flags.output != "" {
				return fmt.Errorf("--in-place and --output cannot be used together")
			}
			if flags.backupDir != "" && !flags.backup {
				return fmt.Errorf("--backup-dir requires --backup")
			}

			result, report, err := cli.RepairFile(ctx, args[0], cli.RepairOptions{
				OutputPath: flags.output,
				InPlace:    flags.inPlace,
				Backup:     flags.backup,
				BackupDir:  flags.backupDir,
			})
			if err != nil {
				return err
			}

			if report != nil {
				options, filter, err := buildReportOptions(root)
				if err != nil {
					return err
				}
				if root.output != "" {
					if err := os.MkdirAll(filepath.Dir(root.output), 0755); err != nil {
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
			}

			if result != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "Repaired: %v\n", result.Success)
				if result.BackupPath != "" {
					fmt.Fprintf(cmd.OutOrStdout(), "Backup: %s\n", result.BackupPath)
				}
				if len(result.ActionsApplied) > 0 {
					fmt.Fprintln(cmd.OutOrStdout(), "Actions applied:")
					for _, action := range result.ActionsApplied {
						fmt.Fprintf(cmd.OutOrStdout(), "  - %s\n", action.Description)
					}
				}
			}

			return cli.ExitWithReport(report)
		},
	}

	cmd.Flags().StringVarP(&flags.output, "output", "o", "", "Output path for repaired file")
	cmd.Flags().BoolVar(&flags.inPlace, "in-place", false, "Repair file in place using atomic replace")
	cmd.Flags().BoolVar(&flags.backup, "backup", false, "Create backup before in-place repair")
	cmd.Flags().StringVar(&flags.backupDir, "backup-dir", "", "Directory to place backups")

	return cmd
}
