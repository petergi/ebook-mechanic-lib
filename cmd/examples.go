package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newExamplesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "examples",
		Short: "Show CLI examples for common workflows",
		Long:  "Prints example commands covering validation, repair, batch processing, reporting, and filtering.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			lines := []string{
				"Validate a single file:",
				"  ebm-cli validate book.epub",
				"  ebm-cli validate document.pdf --format json",
				"",
				"Validate from stdin:",
				"  cat book.epub | ebm-cli validate - --type epub",
				"",
				"Repair a file (writes output to a new file):",
				"  ebm-cli repair broken.pdf --output repaired.pdf",
				"  ebm-cli repair broken.epub --backup backup.epub",
				"",
				"Batch validate (glob + directory):",
				"  ebm-cli batch validate \"library/**/*.epub\"",
				"  ebm-cli batch validate ./samples --max-depth 2 --ext .epub --ext .pdf",
				"",
				"Batch repair with atomic writes and backups:",
				"  ebm-cli batch repair ./inbox --ext .epub --backup-dir ./backups",
				"",
				"Output formats:",
				"  ebm-cli validate book.epub --format text",
				"  ebm-cli validate book.epub --format markdown",
				"  ebm-cli validate book.epub --format json --output report.json",
				"",
				"Verbosity and severity filtering:",
				"  ebm-cli validate book.epub --verbose",
				"  ebm-cli validate book.epub --min-severity warning",
				"  ebm-cli validate book.epub --severity error --severity warning",
				"",
				"Progress output and parallelism:",
				"  ebm-cli batch validate ./samples --progress bar",
				"  ebm-cli batch validate ./samples --workers 8",
				"",
				"Exit codes:",
				"  0 = clean, 1 = warnings, 2 = errors, 3 = internal failure",
			}

			_, err := fmt.Fprintln(cmd.OutOrStdout(), strings.Join(lines, "\n"))
			return err
		},
	}

	return cmd
}
