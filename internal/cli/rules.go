package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	coreconfig "github.com/florianmueller/ordr/internal/core/config"
	"github.com/florianmueller/ordr/internal/core/rule"
	infraconfig "github.com/florianmueller/ordr/internal/infra/config"
	"github.com/florianmueller/ordr/pkg/display"
	"github.com/spf13/cobra"
)

func newRulesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rules",
		Short: "Manage ordr rules",
	}

	cmd.AddCommand(
		newRulesListCmd(),
		newRulesAddCmd(),
		newRulesRemoveCmd(),
		newRulesTestCmd(),
	)

	return cmd
}

// --- list ---

func newRulesListCmd() *cobra.Command {
	var (
		flagLocal  bool
		flagGlobal bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "Show all configured rules",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRulesList(flagLocal, flagGlobal)
		},
	}

	cmd.Flags().BoolVarP(&flagLocal, "local", "l", false, "Show only local config rules")
	cmd.Flags().BoolVarP(&flagGlobal, "global", "g", false, "Show only global config rules")

	return cmd
}

func runRulesList(localOnly, globalOnly bool) error {
	dir, _ := os.Getwd()
	result, err := infraconfig.Load(dir)
	if err != nil {
		return err
	}
	if result == nil {
		display.PrintInfo("No config found. Run 'ordr init' to create one.")
		return nil
	}

	display.PrintHeader(fmt.Sprintf("Rules (%d total)", len(result.Config.Rules)))

	rows := make([]display.RuleRow, 0, len(result.Config.Rules))
	for _, r := range result.Config.Rules {
		rows = append(rows, display.RuleRow{
			Name:   r.Name,
			Match:  formatMatch(r.Match),
			Target: r.Target,
			Source: result.Path,
		})
	}
	display.PrintRules(rows)
	return nil
}

func formatMatch(m coreconfig.MatchConfig) string {
	var parts []string
	if len(m.Extensions) > 0 {
		parts = append(parts, "ext: "+strings.Join(m.Extensions, " "))
	}
	if m.Pattern != "" {
		parts = append(parts, "pattern: "+m.Pattern)
	}
	if m.MinSize != "" {
		parts = append(parts, ">"+m.MinSize)
	}
	if m.MaxSize != "" {
		parts = append(parts, "<"+m.MaxSize)
	}
	if m.OlderThan != "" {
		parts = append(parts, "older: "+m.OlderThan)
	}
	if m.NewerThan != "" {
		parts = append(parts, "newer: "+m.NewerThan)
	}
	if len(parts) == 0 {
		return "(no matchers)"
	}
	return strings.Join(parts, ", ")
}

// --- add ---

func newRulesAddCmd() *cobra.Command {
	var flagGlobal bool

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Interactively add a new rule",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRulesAdd(flagGlobal)
		},
	}

	cmd.Flags().BoolVarP(&flagGlobal, "global", "g", false, "Add rule to global config (~/.ordrrc)")

	return cmd
}

func runRulesAdd(global bool) error {
	var configPath string
	if global {
		p, err := infraconfig.GlobalPath()
		if err != nil {
			return err
		}
		configPath = p
	} else {
		dir, _ := os.Getwd()
		result, err := infraconfig.Load(dir)
		if err != nil {
			return err
		}
		if result == nil {
			p, err := infraconfig.GlobalPath()
			if err != nil {
				return err
			}
			configPath = p
			display.PrintInfo(fmt.Sprintf("No config found — adding to %s", configPath))
		} else {
			configPath = result.Path
		}
	}

	r, err := promptForRule()
	if err != nil {
		return err
	}
	if r == nil {
		display.PrintInfo("Aborted.")
		return nil
	}

	if err := infraconfig.AppendRule(configPath, *r); err != nil {
		return fmt.Errorf("saving rule: %w", err)
	}

	display.PrintSuccess2(fmt.Sprintf("Rule %q added to %s", r.Name, configPath))
	return nil
}

func promptForRule() (*coreconfig.Rule, error) {
	var (
		name       string
		target     string
		extensions string
		pattern    string
		minSize    string
		maxSize    string
		olderThan  string
	)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Rule name").
				Description("A human-readable label for this rule.").
				Value(&name).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("name is required")
					}
					return nil
				}),
			huh.NewInput().
				Title("Target directory").
				Description("Where matched files should be moved (relative or absolute, ~ supported).").
				Value(&target).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("target is required")
					}
					return nil
				}),
		),
		huh.NewGroup(
			huh.NewNote().
				Title("Matchers").
				Description("At least one matcher is required. Leave unused fields empty."),
			huh.NewInput().
				Title("File extensions").
				Description("Comma-separated, e.g. .pdf,.png").
				Value(&extensions),
			huh.NewInput().
				Title("Name pattern (regex)").
				Description("e.g. ^Screenshot.*").
				Value(&pattern),
			huh.NewInput().
				Title("Minimum size").
				Description("e.g. 1MB, 500KB — leave empty to ignore").
				Value(&minSize),
			huh.NewInput().
				Title("Maximum size").
				Description("e.g. 100MB — leave empty to ignore").
				Value(&maxSize),
			huh.NewInput().
				Title("Older than").
				Description("e.g. 30d, 2w — leave empty to ignore").
				Value(&olderThan),
		),
	)

	if err := form.Run(); err != nil {
		return nil, err
	}

	// Validate that at least one matcher is set
	exts := parseExtensions(extensions)
	if len(exts) == 0 && strings.TrimSpace(pattern) == "" && strings.TrimSpace(minSize) == "" && strings.TrimSpace(maxSize) == "" && strings.TrimSpace(olderThan) == "" {
		display.PrintError("at least one matcher (extension, pattern, size, or age) is required")
		return nil, nil
	}

	r := &coreconfig.Rule{
		Name:   strings.TrimSpace(name),
		Target: strings.TrimSpace(target),
		Match: coreconfig.MatchConfig{
			Extensions: exts,
			Pattern:    strings.TrimSpace(pattern),
			MinSize:    strings.TrimSpace(minSize),
			MaxSize:    strings.TrimSpace(maxSize),
			OlderThan:  strings.TrimSpace(olderThan),
		},
		Options: coreconfig.RuleOptions{
			OnConflict: coreconfig.DefaultOnConflict,
		},
	}
	return r, nil
}

func parseExtensions(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if !strings.HasPrefix(p, ".") {
			p = "." + p
		}
		result = append(result, p)
	}
	return result
}

// --- remove ---

func newRulesRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a rule by name",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRulesRemove(args[0])
		},
	}
}

func runRulesRemove(name string) error {
	dir, _ := os.Getwd()
	result, err := infraconfig.Load(dir)
	if err != nil {
		return err
	}
	if result == nil {
		display.PrintError("no config found")
		return nil
	}

	removed, err := infraconfig.RemoveRule(result.Path, name)
	if err != nil {
		return err
	}
	if removed == 0 {
		display.PrintWarning(fmt.Sprintf("No rule named %q found in %s", name, result.Path))
		return nil
	}

	display.PrintSuccess2(fmt.Sprintf("Removed rule %q from %s", name, result.Path))
	return nil
}

// --- test ---

func newRulesTestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "test <file>",
		Short: "Show which rules would match a specific file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRulesTest(args[0])
		},
	}
}

func runRulesTest(filePath string) error {
	fi, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	fileInfo := rule.FileInfo{
		Path:    filePath,
		Name:    fi.Name(),
		Size:    fi.Size(),
		ModTime: fi.ModTime(),
		IsDir:   fi.IsDir(),
	}

	dir, _ := os.Getwd()
	result, err := infraconfig.Load(dir)
	if err != nil {
		return err
	}
	if result == nil {
		display.PrintError("no config found")
		return nil
	}

	display.PrintHeader(fmt.Sprintf("Rule test — %s", fi.Name()))

	matched := false
	for _, r := range result.Config.Rules {
		res := rule.Evaluate(fileInfo, r, dir)
		if res.Matched {
			display.PrintSuccess2(fmt.Sprintf("matches rule %q → %s", r.Name, r.Target))
			matched = true
		}
	}

	if !matched {
		display.PrintInfo("No rules match this file.")
	}

	return nil
}
