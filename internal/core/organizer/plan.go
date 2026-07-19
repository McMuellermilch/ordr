package organizer

// MoveOperation describes a single move (file or directory).
type MoveOperation struct {
	From        string // absolute source path
	To          string // absolute destination path
	RuleName    string
	IsDir       bool   // true when moving a whole directory
	Action      string // "move" | "flatten"
	RemoveEmpty bool   // flatten only: remove source dir if empty after extraction
}

// SkipOperation describes a file that was evaluated but not matched.
type SkipOperation struct {
	Path   string
	Reason string // "no matching rule" | "skipped (conflict)" | ...
}

// ExecutionPlan holds the full result of planning a clean operation.
// It is produced without touching the filesystem.
type ExecutionPlan struct {
	Moves []MoveOperation
	Skips []SkipOperation
}

// HasWork returns true if there is at least one move to perform.
func (p *ExecutionPlan) HasWork() bool {
	return len(p.Moves) > 0
}
