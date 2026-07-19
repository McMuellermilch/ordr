package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	coreconfig "github.com/florianmueller/ordr/internal/core/config"
)

const configFilename = ".ordrrc"

// LoadResult holds the loaded config and its source path.
type LoadResult struct {
	Config coreconfig.Config
	Path   string
}

// Load finds and merges config files. Local configs (walked up from startDir)
// take precedence; the global ~/.ordrrc is the final fallback.
// Rules are merged: local rules first, then global rules.
func Load(startDir string) (*LoadResult, error) {
	local, localPath, err := findLocal(startDir)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	global, globalPath, err := loadGlobal()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	if local == nil && global == nil {
		return nil, nil
	}

	merged := coreconfig.Config{Version: 1}

	// Local rules take precedence
	if local != nil {
		merged.Rules = append(merged.Rules, local.Rules...)
	}
	// Global rules appended after
	if global != nil {
		merged.Rules = append(merged.Rules, global.Rules...)
	}

	sourcePath := globalPath
	if localPath != "" {
		sourcePath = localPath
	}

	return &LoadResult{Config: merged, Path: sourcePath}, nil
}

// LoadGlobalOnly loads only the global ~/.ordrrc without walking up.
func LoadGlobalOnly() (*LoadResult, error) {
	cfg, path, err := loadGlobal()
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		return nil, nil
	}
	return &LoadResult{Config: *cfg, Path: path}, nil
}

// GlobalPath returns the expected path of the global config file.
func GlobalPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, configFilename), nil
}

// LocalPath returns the expected path for a local config in the given dir.
func LocalPath(dir string) string {
	return filepath.Join(dir, configFilename)
}

func findLocal(startDir string) (*coreconfig.Config, string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return nil, "", err
	}

	home, _ := os.UserHomeDir()

	for {
		candidate := filepath.Join(dir, configFilename)
		cfg, err := parseFile(candidate)
		if err == nil {
			return cfg, candidate, nil
		}
		if !errors.Is(err, os.ErrNotExist) {
			return nil, "", err
		}

		// Stop at home directory or filesystem root
		if dir == home || dir == filepath.Dir(dir) {
			break
		}
		dir = filepath.Dir(dir)
	}
	return nil, "", os.ErrNotExist
}

func loadGlobal() (*coreconfig.Config, string, error) {
	path, err := GlobalPath()
	if err != nil {
		return nil, "", err
	}
	cfg, err := parseFile(path)
	if err != nil {
		return nil, "", err
	}
	return cfg, path, nil
}

func parseFile(path string) (*coreconfig.Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg coreconfig.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
