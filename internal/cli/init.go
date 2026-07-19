package cli

import (
	"fmt"
	"os"

	coreconfig "github.com/florianmueller/ordr/internal/core/config"
	infraconfig "github.com/florianmueller/ordr/internal/infra/config"
	"github.com/florianmueller/ordr/pkg/display"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	var (
		flagLocal    bool
		flagTemplate bool
	)

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create a new .ordrrc config file",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if flagLocal {
				return runInit(infraconfig.LocalPath("."), flagTemplate)
			}
			path, err := infraconfig.GlobalPath()
			if err != nil {
				return err
			}
			return runInit(path, flagTemplate)
		},
	}

	cmd.Flags().BoolVarP(&flagLocal, "local", "l", false, "Create config in current directory instead of ~")
	cmd.Flags().BoolVar(&flagTemplate, "template", false, "Initialize with example rules")

	return cmd
}

func runInit(path string, withTemplate bool) error {
	if _, err := os.Stat(path); err == nil {
		display.PrintWarning(fmt.Sprintf("Config already exists at %s", path))
		display.PrintInfo("Edit it directly or use 'ordr rules add' to add new rules.")
		return nil
	}

	var cfg coreconfig.Config
	if withTemplate {
		cfg = infraconfig.TemplateConfig()
	} else {
		cfg = infraconfig.DefaultConfig()
	}

	if err := infraconfig.Write(path, cfg); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	display.PrintSuccess2(fmt.Sprintf("Created config at %s", path))
	display.PrintInfo("Use 'ordr rules add' to add your first rule.")
	display.PrintInfo("Use 'ordr preview' to see what would be moved.")
	return nil
}
