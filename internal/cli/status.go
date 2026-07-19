package cli

import (
	"fmt"
	"os"

	"github.com/florianmueller/ordr/internal/core/organizer"
	infraconfig "github.com/florianmueller/ordr/internal/infra/config"
	infrafs "github.com/florianmueller/ordr/internal/infra/fs"
	"github.com/florianmueller/ordr/pkg/display"
	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status [directory]",
		Short: "Show a summary of the current directory relative to active rules",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := currentOrArg(args)
			return runStatus(dir)
		},
	}
}

func runStatus(dir string) error {
	result, err := infraconfig.Load(dir)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	display.PrintHeader(fmt.Sprintf("ordr status — %s", shortenPath(dir)))

	if result == nil {
		display.PrintInfo("Config:  none (run 'ordr init' to create one)")
		return nil
	}

	display.PrintInfo(fmt.Sprintf("Config:  %s", result.Path))
	display.PrintInfo(fmt.Sprintf("Rules:   %d active", len(result.Config.Rules)))

	plan, err := organizer.Plan(dir, result.Config, false)
	if err != nil {
		return fmt.Errorf("planning: %w", err)
	}

	entries, _ := os.ReadDir(dir)
	total := 0
	for _, e := range entries {
		if !e.IsDir() {
			total++
		}
	}

	fmt.Println()
	display.PrintInfo(fmt.Sprintf("Files:   %d total", total))
	display.PrintInfo(fmt.Sprintf("         %d would be moved  (run: ordr preview)", len(plan.Moves)))
	display.PrintInfo(fmt.Sprintf("         %d unmatched", len(plan.Skips)))

	isGit := infrafs.IsGitRepo(dir)
	display.PrintInfo(fmt.Sprintf("Git repo: %s", boolStr(isGit)))
	fmt.Println()

	return nil
}

func shortenPath(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if len(path) > len(home) && path[:len(home)] == home {
		return "~" + path[len(home):]
	}
	return path
}

func boolStr(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
