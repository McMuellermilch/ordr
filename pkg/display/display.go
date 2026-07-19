package display

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/florianmueller/ordr/internal/core/organizer"
)

var (
	styleBold    = lipgloss.NewStyle().Bold(true)
	styleGreen   = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	styleYellow  = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	styleRed     = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	styleGray    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	styleCyan    = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
	styleMuted   = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	styleHeader  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
)

// PrintPlan renders an ExecutionPlan to stdout.
func PrintPlan(plan *organizer.ExecutionPlan, dryRun bool) {
	if dryRun {
		fmt.Println()
		fmt.Println(styleHeader.Render("  Preview") + styleMuted.Render(fmt.Sprintf(" — %d file(s) would be moved", len(plan.Moves))))
	} else {
		fmt.Println()
		fmt.Println(styleHeader.Render("  Plan") + styleMuted.Render(fmt.Sprintf(" — %d file(s) to move", len(plan.Moves))))
	}
	fmt.Println()

	if len(plan.Moves) == 0 {
		fmt.Println(styleMuted.Render("  Nothing to do."))
		fmt.Println()
		return
	}

	for _, m := range plan.Moves {
		from := shortenPath(m.From)
		to := shortenPath(m.To)
		arrow := styleGray.Render("→")
		label := ""
		if m.RuleName != "" {
			label = styleMuted.Render(fmt.Sprintf(" [%s]", m.RuleName))
		}
		fmt.Printf("  %s  %s  %s%s\n",
			styleCyan.Render(from),
			arrow,
			styleGreen.Render(to),
			label,
		)
	}

	if len(plan.Skips) > 0 {
		fmt.Println()
		fmt.Println(styleMuted.Render(fmt.Sprintf("  %d file(s) skipped (no matching rule)", len(plan.Skips))))
	}
	fmt.Println()
}

// PrintSuccess prints a move result.
func PrintSuccess(from, to string) {
	fmt.Printf("  %s  %s  %s\n",
		styleCyan.Render(shortenPath(from)),
		styleGray.Render("→"),
		styleGreen.Render(shortenPath(to)),
	)
}

// PrintSkipped prints a skipped file.
func PrintSkipped(path, reason string) {
	fmt.Printf("  %s  %s\n",
		styleMuted.Render(shortenPath(path)),
		styleYellow.Render(fmt.Sprintf("(skipped: %s)", reason)),
	)
}

// PrintError prints an error message.
func PrintError(msg string) {
	fmt.Fprintln(os.Stderr, styleRed.Render("  error: ")+msg)
}

// PrintWarning prints a warning message.
func PrintWarning(msg string) {
	fmt.Println(styleYellow.Render("  warning: ") + msg)
}

// PrintInfo prints an info message.
func PrintInfo(msg string) {
	fmt.Println(styleGray.Render("  ") + msg)
}

// PrintSuccess2 prints a success message.
func PrintSuccess2(msg string) {
	fmt.Println(styleGreen.Render("  ✓ ") + msg)
}

// PrintHeader prints a bold section header.
func PrintHeader(msg string) {
	fmt.Println()
	fmt.Println(styleHeader.Render("  " + msg))
	fmt.Println()
}

// PrintUndoEntry renders the moves that would be undone.
func PrintUndoEntry(moves []UndoMoveDisplay) {
	fmt.Println()
	fmt.Println(styleHeader.Render(fmt.Sprintf("  Last clean — %d move(s)", len(moves))))
	fmt.Println()
	for _, m := range moves {
		fmt.Printf("  %s  %s  %s\n",
			styleGreen.Render(shortenPath(m.To)),
			styleGray.Render("→"),
			styleCyan.Render(shortenPath(m.From)),
		)
	}
	fmt.Println()
}

// UndoMoveDisplay is a display-friendly undo move pair.
type UndoMoveDisplay struct {
	From string
	To   string
}

// RuleRow is a display-friendly rule.
type RuleRow struct {
	Name   string
	Match  string
	Target string
	Source string
}

// PrintRules renders a table of rules.
func PrintRules(rows []RuleRow) {
	if len(rows) == 0 {
		fmt.Println(styleMuted.Render("  No rules configured."))
		fmt.Println()
		return
	}

	// Compute column widths
	nameW, matchW, targetW, sourceW := 4, 5, 6, 6
	for _, r := range rows {
		if len(r.Name) > nameW {
			nameW = len(r.Name)
		}
		if len(r.Match) > matchW {
			matchW = len(r.Match)
		}
		if len(r.Target) > targetW {
			targetW = len(r.Target)
		}
		if len(r.Source) > sourceW {
			sourceW = len(r.Source)
		}
	}

	sep := styleGray.Render("│")
	header := fmt.Sprintf("  %s %s %s %s %s %s %s",
		styleBold.Render(pad("Name", nameW)),
		sep,
		styleBold.Render(pad("Match", matchW)),
		sep,
		styleBold.Render(pad("Target", targetW)),
		sep,
		styleBold.Render(pad("Source", sourceW)),
	)
	divider := "  " + strings.Repeat("─", nameW+matchW+targetW+sourceW+9)

	fmt.Println(header)
	fmt.Println(divider)
	for _, r := range rows {
		fmt.Printf("  %s %s %s %s %s %s %s\n",
			pad(r.Name, nameW),
			sep,
			styleMuted.Render(pad(r.Match, matchW)),
			sep,
			styleGreen.Render(pad(r.Target, targetW)),
			sep,
			styleMuted.Render(r.Source),
		)
	}
	fmt.Println()
}

// shortenPath replaces the home directory with ~ for readability.
func shortenPath(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	rel, err := filepath.Rel(home, path)
	if err != nil || strings.HasPrefix(rel, "..") {
		return path
	}
	return "~/" + rel
}

func pad(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}
