package cli

import (
	"github.com/spf13/cobra"
)

func newPreviewCmd() *cobra.Command {
	var (
		flagRecursive bool
		flagRule      string
	)

	cmd := &cobra.Command{
		Use:   "preview [directory]",
		Short: "Preview what 'ordr clean' would do (dry-run)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := currentOrArg(args)
			return runClean(dir, cleanOptions{
				recursive: flagRecursive,
				dryRun:    true,
				ruleName:  flagRule,
				yes:       true,
			})
		},
	}

	cmd.Flags().BoolVarP(&flagRecursive, "recursive", "r", false, "Recurse into subdirectories")
	cmd.Flags().StringVar(&flagRule, "rule", "", "Preview only the rule with this name")

	return cmd
}
