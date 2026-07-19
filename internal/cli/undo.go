package cli

import (
	"errors"
	"fmt"

	"github.com/florianmueller/ordr/internal/infra/fs"
	"github.com/florianmueller/ordr/pkg/display"
	"github.com/spf13/cobra"
)

func newUndoCmd() *cobra.Command {
	var (
		flagList bool
		flagYes  bool
	)

	cmd := &cobra.Command{
		Use:   "undo",
		Short: "Reverse the last 'ordr clean' operation",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if flagList {
				return runUndoList()
			}
			return runUndo(flagYes)
		},
	}

	cmd.Flags().BoolVar(&flagList, "list", false, "Show what the last clean did, without undoing")
	cmd.Flags().BoolVarP(&flagYes, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}

func runUndoList() error {
	entry, err := fs.PeekUndoEntry()
	if err != nil {
		return err
	}
	if entry == nil {
		display.PrintInfo("Nothing to undo.")
		return nil
	}

	moves := toDisplayMoves(entry.Moves)
	display.PrintUndoEntry(moves)
	display.PrintInfo("Run 'ordr undo' to reverse these moves.")
	return nil
}

func runUndo(yes bool) error {
	entry, err := fs.PeekUndoEntry()
	if err != nil {
		return err
	}
	if entry == nil {
		display.PrintInfo("Nothing to undo.")
		return nil
	}

	moves := toDisplayMoves(entry.Moves)
	display.PrintUndoEntry(moves)

	if !yes {
		confirmed, err := confirmPrompt(fmt.Sprintf("Undo %d move(s)?", len(entry.Moves)))
		if err != nil || !confirmed {
			display.PrintInfo("Undo aborted.")
			return nil
		}
	}

	// Pop from log before executing so partial failures don't cause double-undo
	if _, err := fs.PopUndoEntry(); err != nil {
		return fmt.Errorf("updating undo log: %w", err)
	}

	var errs []error
	for _, m := range entry.Moves {
		_, err := fs.MoveFile(m.To, m.From, "rename")
		if err != nil && !errors.Is(err, fs.ErrSkipped) {
			display.PrintError(fmt.Sprintf("restoring %s: %v", m.To, err))
			errs = append(errs, err)
			continue
		}
		display.PrintSuccess(m.To, m.From)
	}

	if len(errs) > 0 {
		return fmt.Errorf("%d error(s) during undo", len(errs))
	}

	display.PrintSuccess2(fmt.Sprintf("Restored %d file(s).", len(entry.Moves)))
	return nil
}

func toDisplayMoves(moves []fs.UndoMove) []display.UndoMoveDisplay {
	result := make([]display.UndoMoveDisplay, len(moves))
	for i, m := range moves {
		result[i] = display.UndoMoveDisplay{From: m.From, To: m.To}
	}
	return result
}
