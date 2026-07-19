package organizer

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/florianmueller/ordr/internal/core/config"
	"github.com/florianmueller/ordr/internal/core/rule"
)

// Plan builds an ExecutionPlan for the given directory and config without
// touching the filesystem. It is safe to call repeatedly (pure function).
func Plan(dir string, cfg config.Config, recursive bool) (*ExecutionPlan, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	plan := &ExecutionPlan{}
	err = walk(absDir, absDir, cfg, recursive, plan)
	return plan, err
}

func walk(rootDir, currentDir string, cfg config.Config, recursive bool, plan *ExecutionPlan) error {
	entries, err := os.ReadDir(currentDir)
	if err != nil {
		return err
	}

	// Track which source paths already have a move planned (first-rule-wins).
	planned := map[string]bool{}
	for _, m := range plan.Moves {
		planned[m.From] = true
	}

	for _, entry := range entries {
		entryPath := filepath.Join(currentDir, entry.Name())

		if entry.IsDir() {
			if recursive && entry.Name() != ".git" {
				if err := walk(rootDir, entryPath, cfg, recursive, plan); err != nil {
					return err
				}
			}
			continue
		}

		if planned[entryPath] {
			continue
		}

		fi, err := entry.Info()
		if err != nil {
			continue
		}

		fileInfo := rule.FileInfo{
			Path:    entryPath,
			Name:    entry.Name(),
			Size:    fi.Size(),
			ModTime: fi.ModTime(),
			IsDir:   false,
		}

		matched := false
		for _, r := range cfg.Rules {
			result := rule.Evaluate(fileInfo, r, currentDir)
			if result.Matched {
				target := resolveTarget(r.Target, rootDir, entryPath)
				plan.Moves = append(plan.Moves, MoveOperation{
					From:     entryPath,
					To:       target,
					RuleName: r.Name,
				})
				planned[entryPath] = true
				matched = true
				break // first matching rule wins
			}
		}

		if !matched {
			plan.Skips = append(plan.Skips, SkipOperation{
				Path:   entryPath,
				Reason: "no matching rule",
			})
		}
	}

	return nil
}

// resolveTarget resolves the destination path for a file.
// If target is absolute or starts with ~, it is used as-is (with home expansion).
// Otherwise it is relative to the directory being cleaned.
func resolveTarget(target, rootDir, srcPath string) string {
	target = expandHome(target)
	var destDir string
	if filepath.IsAbs(target) {
		destDir = target
	} else {
		destDir = filepath.Join(rootDir, target)
	}
	return filepath.Join(destDir, filepath.Base(srcPath))
}

func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}
