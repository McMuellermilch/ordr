package fs

import (
	"fmt"
	"io"
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

// MoveDir moves an entire directory from src to dst.
// If dst already exists, the move is treated as a conflict (rename by default).
func MoveDir(src, dst, onConflict string) (string, error) {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return "", fmt.Errorf("create destination parent: %w", err)
	}

	if _, err := os.Stat(dst); err == nil {
		switch onConflict {
		case "skip":
			return dst, ErrSkipped
		case "overwrite":
			if err := os.RemoveAll(dst); err != nil {
				return "", fmt.Errorf("remove existing dir: %w", err)
			}
		default: // "rename"
			dst = uniquePath(dst)
		}
	}

	if err := os.Rename(src, dst); err != nil {
		// Cross-device: copy recursively then remove source
		if err := copyDirAndRemove(src, dst); err != nil {
			return "", err
		}
	}
	return dst, nil
}

// FlattenDir moves all files from src directory into dst directory (one level deep),
// then optionally removes src if it is empty.
// Returns UndoMoves for each file moved and any errors encountered.
func FlattenDir(src, dst, onConflict string, removeEmpty bool) ([]UndoMove, error) {
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return nil, fmt.Errorf("create destination dir: %w", err)
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return nil, fmt.Errorf("reading source dir: %w", err)
	}

	var moves []UndoMove
	for _, entry := range entries {
		if entry.IsDir() {
			continue // only flatten top-level files
		}
		srcFile := filepath.Join(src, entry.Name())
		dstFile := filepath.Join(dst, entry.Name())
		actualDst, err := MoveFile(srcFile, dstFile, onConflict)
		if err != nil {
			return moves, err
		}
		moves = append(moves, UndoMove{From: srcFile, To: actualDst})
	}

	if removeEmpty {
		// Remove source dir only if now empty
		remaining, _ := os.ReadDir(src)
		if len(remaining) == 0 {
			_ = os.Remove(src)
		}
	}

	return moves, nil
}

// copyDirAndRemove recursively copies src directory to dst, then removes src.
func copyDirAndRemove(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		if err := copyFile(path, target, info.Mode()); err != nil {
			return err
		}
		return os.Remove(path)
	})
}

// copyFile copies a single file preserving permissions.
func copyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

// IsGitRepo returns true if the given directory contains a .git folder.
func IsGitRepo(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, ".git"))
	return err == nil
}
