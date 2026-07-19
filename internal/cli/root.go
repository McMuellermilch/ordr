package cli

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ordr",
	Short: "File chaos, organized.",
	Long: `ordr is a rule-based file organizer.

Define rules in .ordrrc describing where files should go,
then run 'ordr clean' to apply them.

Get started:
  ordr init          Create a config file
  ordr rules add     Add your first rule
  ordr preview       See what would be moved`,
}

// Execute runs the root command.
func Execute(version string) error {
	rootCmd.Version = version
	rootCmd.SetVersionTemplate("ordr {{.Version}}\n")

	rootCmd.AddCommand(
		newCleanCmd(),
		newPreviewCmd(),
		newUndoCmd(),
		newRulesCmd(),
		newInitCmd(),
		newStatusCmd(),
		newWatchCmd(),
	)

	return rootCmd.Execute()
}
