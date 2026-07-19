package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// MoveFile moves src to dst, creating the destination directory if needed.
// Conflict resolution is applied based on the strategy parameter.
func MoveFile(src, dst, onConflict string) (string, error) {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return "", fmt.Errorf("create destination dir: %w", err)
	}

	// Check for conflict
	if _, err := os.Stat(dst); err == nil {
		switch onConflict {
		case "skip":
			return dst, ErrSkipped
		case "overwrite":
			// proceed, os.Rename will overwrite
		default: // "rename"
			dst = uniquePath(dst)
		}
	}

	if err := os.Rename(src, dst); err != nil {
		// Rename fails across devices — fall back to copy + delete
		if err := copyAndRemove(src, dst); err != nil {
			return "", err
		}
	}
	return dst, nil
}

// ErrSkipped is returned when a file is skipped due to conflict strategy "skip".
var ErrSkipped = fmt.Errorf("file skipped (conflict)")

// uniquePath appends a counter suffix to the path until a non-existing path is found.
func uniquePath(path string) string {
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(path, ext)
	for i := 1; ; i++ {
		candidate := fmt.Sprintf("%s (%d)%s", base, i, ext)
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}
}

// copyAndRemove copies src to dst byte-by-byte then removes src.
// Used when os.Rename fails across filesystem boundaries.
func copyAndRemove(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	srcInfo, err := in.Stat()
	if err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer out.Close()

	buf := make([]byte, 32*1024)
	for {
		n, readErr := in.Read(buf)
		if n > 0 {
			if _, writeErr := out.Write(buf[:n]); writeErr != nil {
				return writeErr
			}
		}
		if readErr != nil {
			break
		}
	}

	out.Close()
	return os.Remove(src)
}

// IsGitRepo returns true if the given directory contains a .git folder.
func IsGitRepo(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, ".git"))
	return err == nil
}
