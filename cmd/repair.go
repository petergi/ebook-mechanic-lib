package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/example/project/internal/cli"
	"github.com/example/project/internal/ports"
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
				outputPath := ""
				if flags.inPlace {
					outputPath = args[0]
				} else if flags.output != "" {
					outputPath = flags.output
				} else {
					outputPath = cli.DefaultRepairedPath(args[0])
				}

				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Repaired: %v\n", result.Success)
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Output: %s\n", outputPath)
				if flags.backup && result.BackupPath != "" {
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Backup: %s\n", result.BackupPath)
				}
				if len(result.ActionsApplied) > 0 {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Actions applied:")
					for _, action := range result.ActionsApplied {
						_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  - %s\n", action.Description)
					}
				}

				if outputPath != "" && appliedAction(result, "append_eof_marker") {
					if err := verifyEOFMarker(outputPath); err != nil {
						return err
					}
					_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Verified: %s marker present\n", "%%EOF")
				}

				if report != nil && !report.IsValid {
					warnings := manualRepairWarnings(result)
					if len(warnings) > 0 {
						_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Manual repair required:")
						for _, warning := range warnings {
							_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  - %s\n", warning)
						}
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

func appliedAction(result *ports.RepairResult, actionType string) bool {
	if result == nil {
		return false
	}
	for _, action := range result.ActionsApplied {
		if action.Type == actionType {
			return true
		}
	}
	return false
}

func verifyEOFMarker(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read repaired output: %w", err)
	}
	if len(data) == 0 {
		return fmt.Errorf("repaired output is empty")
	}
	last := data
	if len(data) > 1024 {
		last = data[len(data)-1024:]
	}
	if !bytes.Contains(last, []byte("%%EOF")) {
		return fmt.Errorf("missing %%EOF marker in repaired output")
	}
	return nil
}

func manualRepairWarnings(result *ports.RepairResult) []string {
	if result == nil {
		return nil
	}
	unique := make(map[string]struct{})
	for _, action := range result.ActionsApplied {
		if action.Automated {
			continue
		}
		if action.Description == "" {
			continue
		}
		unique[action.Description] = struct{}{}
	}
	if len(unique) == 0 {
		return nil
	}
	warnings := make([]string, 0, len(unique))
	for warning := range unique {
		warnings = append(warnings, warning)
	}
	sort.Strings(warnings)
	return warnings
}
