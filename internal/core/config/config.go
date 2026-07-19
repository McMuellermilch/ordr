package config

// Config is the top-level structure of an .ordrrc file.
type Config struct {
	Version int    `yaml:"version"`
	Rules   []Rule `yaml:"rules"`
}

// Rule defines how files are matched and where they are moved.
type Rule struct {
	Name    string      `yaml:"name"`
	Target  string      `yaml:"target"`
	Match   MatchConfig `yaml:"match"`
	Scope   ScopeConfig `yaml:"scope"`
	Options RuleOptions `yaml:"options"`
}

// MatchConfig holds all matchers for a rule. All specified matchers must pass (AND logic).
type MatchConfig struct {
	Extensions []string `yaml:"extensions"`
	Pattern    string   `yaml:"pattern"`
	MinSize    string   `yaml:"min_size"`
	MaxSize    string   `yaml:"max_size"`
	OlderThan  string   `yaml:"older_than"`
	NewerThan  string   `yaml:"newer_than"`
}

// ScopeConfig restricts which directories a rule applies to.
type ScopeConfig struct {
	Dirs []string `yaml:"dirs"`
}

// RuleOptions controls per-rule behavior.
type RuleOptions struct {
	OnConflict string `yaml:"on_conflict"` // "rename" | "skip" | "overwrite"
	Recursive  bool   `yaml:"recursive"`
}

// OnConflict strategies.
const (
	ConflictRename    = "rename"
	ConflictSkip      = "skip"
	ConflictOverwrite = "overwrite"
)

// DefaultOnConflict is used when a rule's OnConflict field is empty.
const DefaultOnConflict = ConflictRename
