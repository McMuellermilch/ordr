package watcher

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/florianmueller/ordr/internal/core/config"
	"github.com/florianmueller/ordr/internal/core/rule"
	"github.com/fsnotify/fsnotify"
)

// Event is produced when a file or directory matches a watch rule.
type Event struct {
	Path string
	Rule config.Rule
}

// Handler is called for each matched file event.
type Handler func(Event)

// Watcher monitors directories using OS-level file system events.
type Watcher struct {
	cfg      config.Config
	handler  Handler
	fw       *fsnotify.Watcher
	mu       sync.Mutex
	timers   map[string]*time.Timer
	debounce time.Duration
}

// New creates a Watcher. Call Start to begin watching.
func New(cfg config.Config, handler Handler) (*Watcher, error) {
	fw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &Watcher{
		cfg:      cfg,
		handler:  handler,
		fw:       fw,
		timers:   make(map[string]*time.Timer),
		debounce: time.Second,
	}, nil
}

// WatchDirs returns all unique absolute directories from scope.watch_dirs across all rules.
func WatchDirs(cfg config.Config) []string {
	seen := map[string]bool{}
	var dirs []string
	for _, r := range cfg.Rules {
		for _, d := range r.Scope.WatchDirs {
			abs, err := filepath.Abs(expandHome(d))
			if err != nil {
				continue
			}
			if !seen[abs] {
				seen[abs] = true
				dirs = append(dirs, abs)
			}
		}
	}
	return dirs
}

// Start registers all watch_dirs with the OS and blocks until Stop is called.
func (w *Watcher) Start() error {
	for _, dir := range WatchDirs(w.cfg) {
		if err := w.fw.Add(dir); err != nil {
			writeLog("ERROR  cannot watch " + dir + ": " + err.Error())
		} else {
			writeLog("WATCH  " + dir)
		}
	}

	for {
		select {
		case event, ok := <-w.fw.Events:
			if !ok {
				return nil
			}
			if event.Has(fsnotify.Create) || event.Has(fsnotify.Rename) {
				w.schedule(event.Name)
			}
		case err, ok := <-w.fw.Errors:
			if !ok {
				return nil
			}
			writeLog("ERROR  " + err.Error())
		}
	}
}

// Stop shuts down the watcher.
func (w *Watcher) Stop() {
	w.fw.Close()
}

func (w *Watcher) schedule(path string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if t, exists := w.timers[path]; exists {
		t.Reset(w.debounce)
		return
	}
	w.timers[path] = time.AfterFunc(w.debounce, func() {
		w.mu.Lock()
		delete(w.timers, path)
		w.mu.Unlock()
		w.process(path)
	})
}

func (w *Watcher) process(path string) {
	info, err := os.Stat(path)
	if err != nil {
		return // file gone by the time debounce fired
	}

	currentDir := filepath.Dir(path)
	fi := rule.FileInfo{
		Path:    path,
		Name:    filepath.Base(path),
		Size:    info.Size(),
		ModTime: info.ModTime(),
		IsDir:   info.IsDir(),
	}

	for _, r := range w.cfg.Rules {
		if result := rule.EvaluateWatch(fi, r, currentDir); result.Matched {
			w.handler(Event{Path: path, Rule: r})
			return
		}
	}
}

// writeLog appends a timestamped line to ~/.ordr/watch.log.
func writeLog(msg string) {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	logPath := filepath.Join(home, ".ordr", "watch.log")
	_ = os.MkdirAll(filepath.Dir(logPath), 0o755)
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	ts := time.Now().Format("2006-01-02 15:04:05")
	f.WriteString(ts + "  " + msg + "\n")
}

// WriteLog is the exported version for use by the CLI handler.
func WriteLog(msg string) {
	writeLog(msg)
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
