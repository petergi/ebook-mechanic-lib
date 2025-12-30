package main

import (
	"context"
	"fmt"
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
			"  ebm-cli repair book.epub",
			"  ebm-cli repair document.pdf --in-place --backup",
			"  ebm-cli repair book.epub --output fixed.epub",
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

			var reportErr error
			if report != nil {
				reportErr = writeValidationReport(ctx, cmd, root, report)
			}

			if result != nil {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Repaired: %v\n", result.Success)
				if flags.inPlace {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Output: %s\n", args[0])
				} else if flags.output != "" {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Output: %s\n", flags.output)
				} else {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Output: %s\n", cli.DefaultRepairedPath(args[0]))
				}
				if result.BackupPath != "" {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Backup: %s\n", result.BackupPath)
				}
				if len(result.ActionsApplied) > 0 {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Actions applied:")
					for _, action := range result.ActionsApplied {
						_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  - %s\n", action.Description)
					}
				}
			}

			if reportErr != nil {
				return reportErr
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&flags.output, "output", "o", "", "Output path for repaired file")
	cmd.Flags().BoolVar(&flags.inPlace, "in-place", false, "Repair file in place using atomic replace")
	cmd.Flags().BoolVar(&flags.backup, "backup", false, "Create backup before in-place repair")
	cmd.Flags().StringVar(&flags.backupDir, "backup-dir", "", "Directory to place backups")

	return cmd
}
