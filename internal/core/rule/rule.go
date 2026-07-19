package rule

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/florianmueller/ordr/internal/core/config"
)

// FileInfo describes a file to be evaluated against rules.
type FileInfo struct {
	Path    string
	Name    string
	Size    int64
	ModTime time.Time
	IsDir   bool
}

// MatchResult holds the result of evaluating a file against a rule.
type MatchResult struct {
	Matched bool
	Rule    config.Rule
	Reason  string // human-readable explanation for dry-run output
}

// Evaluate checks whether a file or directory matches a rule. All specified matchers must pass.
func Evaluate(f FileInfo, r config.Rule, currentDir string) MatchResult {
	// Type matching: default is "file"
	matchType := r.Match.Type
	if matchType == "" {
		matchType = config.MatchTypeFile
	}

	switch matchType {
	case config.MatchTypeDir:
		if !f.IsDir {
			return MatchResult{Matched: false}
		}
	default: // "file"
		if f.IsDir {
			return MatchResult{Matched: false}
		}
	}

	if !isScopeMatch(r, currentDir) {
		return MatchResult{Matched: false}
	}

	if !hasAnyMatcher(r.Match) {
		return MatchResult{Matched: false, Reason: fmt.Sprintf("rule %q has no matchers", r.Name)}
	}

	if !isPatternMatch(r.Match, f.Name) {
		return MatchResult{Matched: false}
	}

	// Extension matching only applies to files
	if !f.IsDir && !isExtensionMatch(r.Match, f.Name) {
		return MatchResult{Matched: false}
	}

	// Size and age matching only apply to files
	if !f.IsDir {
		if !isSizeMatch(r.Match, f.Size) {
			return MatchResult{Matched: false}
		}
		if !isAgeMatch(r.Match, f.ModTime) {
			return MatchResult{Matched: false}
		}
	}

	return MatchResult{Matched: true, Rule: r}
}

// hasAnyMatcher returns true if at least one matcher is configured.
func hasAnyMatcher(m config.MatchConfig) bool {
	return len(m.Extensions) > 0 ||
		m.Pattern != "" ||
		m.MinSize != "" ||
		m.MaxSize != "" ||
		m.OlderThan != "" ||
		m.NewerThan != ""
}

func isScopeMatch(r config.Rule, currentDir string) bool {
	if len(r.Scope.Dirs) == 0 {
		return true
	}
	return matchesAnyDir(r.Scope.Dirs, currentDir)
}

func matchesAnyDir(dirs []string, currentDir string) bool {
	absCurrentDir, err := filepath.Abs(currentDir)
	if err != nil {
		return false
	}
	for _, dir := range dirs {
		expanded := expandHome(dir)
		absDir, err := filepath.Abs(expanded)
		if err != nil {
			continue
		}
		if absCurrentDir == absDir {
			return true
		}
	}
	return false
}

// EvaluateWatch checks whether a file matches a rule using scope.watch_dirs.
// Returns false immediately if the rule has no watch_dirs defined.
func EvaluateWatch(f FileInfo, r config.Rule, currentDir string) MatchResult {
	if len(r.Scope.WatchDirs) == 0 {
		return MatchResult{Matched: false}
	}

	matchType := r.Match.Type
	if matchType == "" {
		matchType = config.MatchTypeFile
	}

	switch matchType {
	case config.MatchTypeDir:
		if !f.IsDir {
			return MatchResult{Matched: false}
		}
	default:
		if f.IsDir {
			return MatchResult{Matched: false}
		}
	}

	if !matchesAnyDir(r.Scope.WatchDirs, currentDir) {
		return MatchResult{Matched: false}
	}

	if !hasAnyMatcher(r.Match) {
		return MatchResult{Matched: false}
	}

	if !isPatternMatch(r.Match, f.Name) {
		return MatchResult{Matched: false}
	}

	if !f.IsDir && !isExtensionMatch(r.Match, f.Name) {
		return MatchResult{Matched: false}
	}

	if !f.IsDir {
		if !isSizeMatch(r.Match, f.Size) {
			return MatchResult{Matched: false}
		}
		if !isAgeMatch(r.Match, f.ModTime) {
			return MatchResult{Matched: false}
		}
	}

	return MatchResult{Matched: true, Rule: r}
}

func isExtensionMatch(m config.MatchConfig, filename string) bool {
	if len(m.Extensions) == 0 {
		return true
	}
	ext := strings.ToLower(filepath.Ext(filename))
	for _, e := range m.Extensions {
		if strings.ToLower(e) == ext {
			return true
		}
	}
	return false
}

func isPatternMatch(m config.MatchConfig, filename string) bool {
	if m.Pattern == "" {
		return true
	}
	re, err := regexp.Compile(m.Pattern)
	if err != nil {
		return false
	}
	return re.MatchString(filename)
}

func isSizeMatch(m config.MatchConfig, size int64) bool {
	if m.MinSize != "" {
		min, err := parseSize(m.MinSize)
		if err == nil && size < min {
			return false
		}
	}
	if m.MaxSize != "" {
		max, err := parseSize(m.MaxSize)
		if err == nil && size > max {
			return false
		}
	}
	return true
}

func isAgeMatch(m config.MatchConfig, modTime time.Time) bool {
	now := time.Now()
	if m.OlderThan != "" {
		d, err := parseDuration(m.OlderThan)
		if err == nil && now.Sub(modTime) < d {
			return false
		}
	}
	if m.NewerThan != "" {
		d, err := parseDuration(m.NewerThan)
		if err == nil && now.Sub(modTime) > d {
			return false
		}
	}
	return true
}

// parseSize parses size strings like "100MB", "1GB", "500KB".
func parseSize(s string) (int64, error) {
	s = strings.TrimSpace(strings.ToUpper(s))
	units := map[string]int64{
		"B":   1,
		"KB":  1024,
		"MB":  1024 * 1024,
		"GB":  1024 * 1024 * 1024,
		"TB":  1024 * 1024 * 1024 * 1024,
	}
	for suffix, multiplier := range units {
		if strings.HasSuffix(s, suffix) {
			numStr := strings.TrimSuffix(s, suffix)
			num, err := strconv.ParseFloat(strings.TrimSpace(numStr), 64)
			if err != nil {
				return 0, err
			}
			return int64(num * float64(multiplier)), nil
		}
	}
	// Fallback: plain bytes
	num, err := strconv.ParseInt(s, 10, 64)
	return num, err
}

// parseDuration parses duration strings like "30d", "2w", "24h".
func parseDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	if strings.HasSuffix(s, "d") {
		n, err := strconv.Atoi(strings.TrimSuffix(s, "d"))
		if err != nil {
			return 0, err
		}
		return time.Duration(n) * 24 * time.Hour, nil
	}
	if strings.HasSuffix(s, "w") {
		n, err := strconv.Atoi(strings.TrimSuffix(s, "w"))
		if err != nil {
			return 0, err
		}
		return time.Duration(n) * 7 * 24 * time.Hour, nil
	}
	return time.ParseDuration(s)
}

// expandHome replaces a leading ~ with the user's home directory.
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
