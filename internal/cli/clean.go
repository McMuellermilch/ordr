package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/florianmueller/ordr/internal/core/config"
	"github.com/florianmueller/ordr/internal/core/organizer"
	infraconfig "github.com/florianmueller/ordr/internal/infra/config"
	infrafs "github.com/florianmueller/ordr/internal/infra/fs"
	"github.com/florianmueller/ordr/pkg/display"
	"github.com/spf13/cobra"
)

func newCleanCmd() *cobra.Command {
	var (
		flagGlobal    bool
		flagRecursive bool
		flagDryRun    bool
		flagRule      string
		flagYes       bool
		flagVerbose   bool
	)

	cmd := &cobra.Command{
		Use:   "clean [directory]",
		Short: "Organize files in a directory by applying rules",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := currentOrArg(args)
			return runClean(dir, cleanOptions{
				global:    flagGlobal,
				recursive: flagRecursive,
				dryRun:    flagDryRun,
				ruleName:  flagRule,
				yes:       flagYes,
				verbose:   flagVerbose,
			})
		},
	}

	cmd.Flags().BoolVarP(&flagGlobal, "global", "g", false, "Also run rules that have explicit scope.dirs defined")
	cmd.Flags().BoolVarP(&flagRecursive, "recursive", "r", false, "Recurse into subdirectories")
	cmd.Flags().BoolVarP(&flagDryRun, "dry-run", "n", false, "Preview what would happen without moving files")
	cmd.Flags().StringVar(&flagRule, "rule", "", "Run only the rule with this name")
	cmd.Flags().BoolVarP(&flagYes, "yes", "y", false, "Skip all confirmation prompts")
	cmd.Flags().BoolVarP(&flagVerbose, "verbose", "v", false, "Show each file evaluated")

	return cmd
}

type cleanOptions struct {
	global    bool
	recursive bool
	dryRun    bool
	ruleName  string
	yes       bool
	verbose   bool
}

func runClean(dir string, opts cleanOptions) error {
	result, err := infraconfig.Load(dir)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	if result == nil {
		display.PrintError("no config found — run 'ordr init' to create one")
		return nil
	}

	cfg := result.Config

	// Filter to a single rule if --rule is set
	if opts.ruleName != "" {
		cfg = filterByRuleName(cfg, opts.ruleName)
		if len(cfg.Rules) == 0 {
			display.PrintError(fmt.Sprintf("no rule named %q found", opts.ruleName))
			return nil
		}
	}

	// Git repo warning
	if infrafs.IsGitRepo(dir) && !opts.yes {
		confirmed, err := confirmPrompt("Current directory appears to be a git repo — continue?")
		if err != nil || !confirmed {
			display.PrintInfo("Clean aborted.")
			return nil
		}
	}

	plan, err := organizer.Plan(dir, cfg, opts.recursive)
	if err != nil {
		return fmt.Errorf("planning: %w", err)
	}

	display.PrintPlan(plan, opts.dryRun)

	if opts.dryRun {
		display.PrintInfo("Run without --dry-run to apply.")
		return nil
	}

	if !plan.HasWork() {
		return nil
	}

	if !opts.yes {
		confirmed, err := confirmPrompt(fmt.Sprintf("Move %d file(s)?", len(plan.Moves)))
		if err != nil || !confirmed {
			display.PrintInfo("Clean aborted.")
			return nil
		}
	}

	return executePlan(plan, opts.verbose)
}

func executePlan(plan *organizer.ExecutionPlan, verbose bool) error {
	var undoMoves []infrafs.UndoMove
	var errs []error

	for _, move := range plan.Moves {
		onConflict := config.DefaultOnConflict
		actualDst, err := infrafs.MoveFile(move.From, move.To, onConflict)
		if err != nil {
			if errors.Is(err, infrafs.ErrSkipped) {
				if verbose {
					display.PrintSkipped(move.From, "conflict")
				}
				continue
			}
			display.PrintError(fmt.Sprintf("moving %s: %v", move.From, err))
			errs = append(errs, err)
			continue
		}
		undoMoves = append(undoMoves, infrafs.UndoMove{From: move.From, To: actualDst})
		if verbose {
			display.PrintSuccess(move.From, actualDst)
		}
	}

	if len(undoMoves) > 0 {
		if err := infrafs.AppendUndoEntry(undoMoves); err != nil {
			display.PrintWarning("could not write undo log: " + err.Error())
		}
		display.PrintSuccess2(fmt.Sprintf("Moved %d file(s). Run 'ordr undo' to reverse.", len(undoMoves)))
	}

	if len(errs) > 0 {
		return fmt.Errorf("%d error(s) occurred during clean", len(errs))
	}
	return nil
}

func filterByRuleName(cfg config.Config, name string) config.Config {
	filtered := config.Config{Version: cfg.Version}
	for _, r := range cfg.Rules {
		if r.Name == name {
			filtered.Rules = append(filtered.Rules, r)
		}
	}
	return filtered
}

func currentOrArg(args []string) string {
	if len(args) > 0 {
		return args[0]
	}
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	return dir
}

func confirmPrompt(msg string) (bool, error) {
	var confirmed bool
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(msg).
				Value(&confirmed),
		),
	)
	if err := form.Run(); err != nil {
		return false, err
	}
	return confirmed, nil
}
