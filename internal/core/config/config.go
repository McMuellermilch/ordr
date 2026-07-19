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
	Type       string   `yaml:"type"` // "file" (default) | "dir"
}

// ScopeConfig restricts which directories a rule applies to.
type ScopeConfig struct {
	Dirs      []string `yaml:"dirs"`       // used by `ordr clean`
	WatchDirs []string `yaml:"watch_dirs"` // used by `ordr watch` (falls back to Dirs if empty)
}

// RuleOptions controls per-rule behavior.
type RuleOptions struct {
	OnConflict  string `yaml:"on_conflict"`  // "rename" | "skip" | "overwrite"
	Recursive   bool   `yaml:"recursive"`
	Action      string `yaml:"action"`       // "move" (default) | "flatten"
	RemoveEmpty bool   `yaml:"remove_empty"` // after flatten: remove source dir if empty
}

// Action values for RuleOptions.
const (
	ActionMove    = "move"
	ActionFlatten = "flatten"
)

// MatchType values for MatchConfig.
const (
	MatchTypeFile = "file"
	MatchTypeDir  = "dir"
)

// OnConflict strategies.
const (
	ConflictRename    = "rename"
	ConflictSkip      = "skip"
	ConflictOverwrite = "overwrite"
)

// DefaultOnConflict is used when a rule's OnConflict field is empty.
const DefaultOnConflict = ConflictRename
