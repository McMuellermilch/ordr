package cli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/florianmueller/ordr/internal/core/config"
	"github.com/florianmueller/ordr/internal/core/organizer"
	corewatcher "github.com/florianmueller/ordr/internal/core/watcher"
	infraconfig "github.com/florianmueller/ordr/internal/infra/config"
	infrafs "github.com/florianmueller/ordr/internal/infra/fs"
	"github.com/florianmueller/ordr/internal/infra/launchd"
	"github.com/florianmueller/ordr/pkg/display"
	"github.com/spf13/cobra"
)

func newWatchCmd() *cobra.Command {
	var flagDaemon bool

	cmd := &cobra.Command{
		Use:   "watch",
		Short: "Auto-organize files as they appear (macOS launchd agent)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if flagDaemon {
				return runDaemon()
			}
			return cmd.Help()
		},
	}

	cmd.Flags().BoolVar(&flagDaemon, "daemon", false, "Run the watch loop (used by launchd, not meant for direct use)")
	cmd.Flag("daemon").Hidden = true

	cmd.AddCommand(
		newWatchInstallCmd(),
		newWatchUninstallCmd(),
		newWatchRestartCmd(),
		newWatchStatusCmd(),
		newWatchLogsCmd(),
	)

	return cmd
}

func newWatchInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Install and start the background watch agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadWatchConfig()
			if err != nil {
				return err
			}

			dirs := corewatcher.WatchDirs(cfg)
			if len(dirs) == 0 {
				display.PrintError("no watch_dirs found in config — add scope.watch_dirs to at least one rule")
				return nil
			}

			if err := launchd.Install(); err != nil {
				return fmt.Errorf("installing agent: %w", err)
			}

			display.PrintSuccess2("Watch agent installed and started.")
			fmt.Println()
			for _, d := range dirs {
				display.PrintInfo("watching  " + d)
			}
			fmt.Println()
			display.PrintInfo("Run 'ordr watch status' to check the agent.")
			display.PrintInfo("Run 'ordr watch uninstall' to stop.")
			return nil
		},
	}
}

func newWatchUninstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall",
		Short: "Stop and remove the background watch agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := launchd.Uninstall(); err != nil {
				return fmt.Errorf("uninstalling agent: %w", err)
			}
			display.PrintSuccess2("Watch agent stopped and removed.")
			return nil
		},
	}
}

func newWatchRestartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "restart",
		Short: "Restart the watch agent (picks up config changes)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := launchd.Uninstall(); err != nil {
				return fmt.Errorf("stopping agent: %w", err)
			}
			if err := launchd.Install(); err != nil {
				return fmt.Errorf("starting agent: %w", err)
			}
			display.PrintSuccess2("Watch agent restarted.")
			return nil
		},
	}
}

func newWatchStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show the status of the watch agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			status, err := launchd.Status()
			if err != nil {
				return err
			}

			cfg, _ := loadWatchConfig()
			var dirs []string
			if cfg.Rules != nil {
				dirs = corewatcher.WatchDirs(cfg)
			}

			logPath, _ := launchd.LogPath()

			display.PrintWatchStatus(display.WatchStatusDisplay{
				Installed: status.Installed,
				Running:   status.Running,
				PID:       status.PID,
				WatchDirs: dirs,
				LogPath:   logPath,
			})
			return nil
		},
	}
}

func newWatchLogsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logs",
		Short: "Tail the watch log (~/.ordr/watch.log)",
		RunE: func(cmd *cobra.Command, args []string) error {
			logPath, err := launchd.LogPath()
			if err != nil {
				return err
			}
			if _, err := os.Stat(logPath); os.IsNotExist(err) {
				display.PrintInfo("No log file yet. The log is created when the agent first runs.")
				return nil
			}
			c := exec.Command("tail", "-f", logPath)
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			return c.Run()
		},
	}
}

// runDaemon is the actual watch loop started by launchd.
func runDaemon() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	result, err := infraconfig.Load(home)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	if result == nil {
		corewatcher.WriteLog("ERROR  no config found at " + home)
		return fmt.Errorf("no config found")
	}

	cfg := result.Config
	dirs := corewatcher.WatchDirs(cfg)
	if len(dirs) == 0 {
		corewatcher.WriteLog("ERROR  no watch_dirs configured — nothing to watch")
		return fmt.Errorf("no watch_dirs configured")
	}

	corewatcher.WriteLog(fmt.Sprintf("START  watching %d director(y/ies)", len(dirs)))

	w, err := corewatcher.New(cfg, func(event corewatcher.Event) {
		handleWatchEvent(event)
	})
	if err != nil {
		return err
	}
	defer w.Stop()

	return w.Start()
}

// handleWatchEvent executes the move for a matched watch event.
func handleWatchEvent(event corewatcher.Event) {
	r := event.Rule
	srcDir := filepath.Dir(event.Path)
	target := organizer.ResolveTarget(r.Target, srcDir, event.Path)
	onConflict := config.DefaultOnConflict

	action := r.Options.Action
	if action == "" {
		action = config.ActionMove
	}

	info, err := os.Stat(event.Path)
	if err != nil {
		corewatcher.WriteLog("ERROR  stat " + event.Path + ": " + err.Error())
		return
	}

	var undoMoves []infrafs.UndoMove

	switch {
	case info.IsDir() && action == config.ActionFlatten:
		moves, err := infrafs.FlattenDir(event.Path, target, onConflict, r.Options.RemoveEmpty)
		undoMoves = append(undoMoves, moves...)
		if err != nil && !errors.Is(err, infrafs.ErrSkipped) {
			corewatcher.WriteLog("ERROR  flatten " + event.Path + ": " + err.Error())
			return
		}
		corewatcher.WriteLog(fmt.Sprintf("FLATTEN  %s  →  %s  [%s]", event.Path, target, r.Name))

	case info.IsDir():
		actualDst, err := infrafs.MoveDir(event.Path, target, onConflict)
		if err != nil {
			if !errors.Is(err, infrafs.ErrSkipped) {
				corewatcher.WriteLog("ERROR  move dir " + event.Path + ": " + err.Error())
			}
			return
		}
		undoMoves = append(undoMoves, infrafs.UndoMove{From: event.Path, To: actualDst})
		corewatcher.WriteLog(fmt.Sprintf("MOVED  %s  →  %s  [%s]", event.Path, actualDst, r.Name))

	default:
		actualDst, err := infrafs.MoveFile(event.Path, target, onConflict)
		if err != nil {
			if !errors.Is(err, infrafs.ErrSkipped) {
				corewatcher.WriteLog("ERROR  move " + event.Path + ": " + err.Error())
			}
			return
		}
		undoMoves = append(undoMoves, infrafs.UndoMove{From: event.Path, To: actualDst})
		corewatcher.WriteLog(fmt.Sprintf("MOVED  %s  →  %s  [%s]", event.Path, actualDst, r.Name))
	}

	if len(undoMoves) > 0 {
		if err := infrafs.AppendUndoEntry(undoMoves); err != nil {
			corewatcher.WriteLog("WARN   could not write undo log: " + err.Error())
		}
	}
}

// loadWatchConfig loads the global config (watch daemon always uses ~/.ordrrc).
func loadWatchConfig() (config.Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return config.Config{}, err
	}
	result, err := infraconfig.Load(home)
	if err != nil {
		return config.Config{}, err
	}
	if result == nil {
		return config.Config{}, nil
	}
	return result.Config, nil
}
