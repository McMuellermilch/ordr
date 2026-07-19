package config

import (
	"os"

	"gopkg.in/yaml.v3"

	coreconfig "github.com/florianmueller/ordr/internal/core/config"
)

// Write marshals a Config to YAML and writes it to the given path.
// The parent directory must already exist.
func Write(path string, cfg coreconfig.Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// DefaultConfig returns a minimal valid config ready to be written.
func DefaultConfig() coreconfig.Config {
	return coreconfig.Config{
		Version: 1,
		Rules:   []coreconfig.Rule{},
	}
}

// TemplateConfig returns a config with example rules.
func TemplateConfig() coreconfig.Config {
	return coreconfig.Config{
		Version: 1,
		Rules: []coreconfig.Rule{
			{
				Name:   "PDFs to documents",
				Target: "documents",
				Match:  coreconfig.MatchConfig{Extensions: []string{".pdf"}},
			},
			{
				Name:   "Images",
				Target: "images",
				Match:  coreconfig.MatchConfig{Extensions: []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".heic"}},
			},
			{
				Name:   "Archives",
				Target: "archives",
				Match:  coreconfig.MatchConfig{Extensions: []string{".zip", ".tar", ".gz", ".rar", ".7z"}},
			},
		},
	}
}

// AppendRule loads the config at path, appends the rule, and writes it back.
func AppendRule(path string, r coreconfig.Rule) error {
	cfg, err := parseFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg = &coreconfig.Config{Version: 1}
		} else {
			return err
		}
	}
	cfg.Rules = append(cfg.Rules, r)
	return Write(path, *cfg)
}

// RemoveRule loads the config at path, removes rules matching the name, and writes it back.
func RemoveRule(path string, name string) (removed int, err error) {
	cfg, err := parseFile(path)
	if err != nil {
		return 0, err
	}
	filtered := cfg.Rules[:0]
	for _, r := range cfg.Rules {
		if r.Name == name {
			removed++
		} else {
			filtered = append(filtered, r)
		}
	}
	cfg.Rules = filtered
	return removed, Write(path, *cfg)
}
