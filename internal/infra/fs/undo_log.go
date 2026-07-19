package fs

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const undoLogFilename = "undo.json"

// UndoEntry represents a single clean session that can be undone.
type UndoEntry struct {
	ID    string      `json:"id"`
	Moves []UndoMove  `json:"moves"`
}

// UndoMove holds the original source and actual destination of a moved file.
type UndoMove struct {
	From string `json:"from"` // original location
	To   string `json:"to"`   // where the file was moved
}

// AppendUndoEntry adds a new entry to the undo log stack.
func AppendUndoEntry(moves []UndoMove) error {
	log, err := readUndoLog()
	if err != nil {
		log = []UndoEntry{}
	}
	entry := UndoEntry{
		ID:    time.Now().UTC().Format(time.RFC3339),
		Moves: moves,
	}
	log = append(log, entry)
	return writeUndoLog(log)
}

// PopUndoEntry removes and returns the most recent undo entry.
// Returns nil if the log is empty.
func PopUndoEntry() (*UndoEntry, error) {
	log, err := readUndoLog()
	if err != nil || len(log) == 0 {
		return nil, err
	}
	last := log[len(log)-1]
	log = log[:len(log)-1]
	return &last, writeUndoLog(log)
}

// PeekUndoEntry returns the most recent entry without removing it.
func PeekUndoEntry() (*UndoEntry, error) {
	log, err := readUndoLog()
	if err != nil || len(log) == 0 {
		return nil, err
	}
	last := log[len(log)-1]
	return &last, nil
}

func undoLogPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".ordr")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(dir, undoLogFilename), nil
}

func readUndoLog() ([]UndoEntry, error) {
	path, err := undoLogPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return []UndoEntry{}, nil
	}
	if err != nil {
		return nil, err
	}
	var log []UndoEntry
	if err := json.Unmarshal(data, &log); err != nil {
		return nil, err
	}
	return log, nil
}

func writeUndoLog(log []UndoEntry) error {
	path, err := undoLogPath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(log, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
