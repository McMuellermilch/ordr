//go:build darwin

package launchd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
)

const label = "com.ordr.watch"

// AgentStatus describes the current state of the launchd agent.
type AgentStatus struct {
	Installed bool
	Running   bool
	PID       int
}

// PlistPath returns the path to the agent's plist file.
func PlistPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "LaunchAgents", label+".plist"), nil
}

// LogPath returns the path to the watch log file.
func LogPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".ordr", "watch.log"), nil
}

// Install writes the plist and loads the agent. Safe to call when already installed.
func Install() error {
	plistPath, err := PlistPath()
	if err != nil {
		return err
	}

	// If already present, unload first to allow clean reinstall
	if _, err := os.Stat(plistPath); err == nil {
		uid := strconv.Itoa(os.Getuid())
		exec.Command("launchctl", "bootout", "gui/"+uid, plistPath).Run()
		os.Remove(plistPath)
	}

	binPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve binary path: %w", err)
	}
	binPath, err = filepath.EvalSymlinks(binPath)
	if err != nil {
		return fmt.Errorf("resolve symlinks: %w", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	logPath, err := LogPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		return err
	}

	content, err := renderPlist(plistData{
		Label:   label,
		BinPath: binPath,
		LogPath: logPath,
		Home:    home,
		Path:    "/opt/homebrew/bin:/usr/local/bin:/usr/bin:/bin",
	})
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(plistPath), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(plistPath, []byte(content), 0o644); err != nil {
		return err
	}

	uid := strconv.Itoa(os.Getuid())
	out, err := exec.Command("launchctl", "bootstrap", "gui/"+uid, plistPath).CombinedOutput()
	if err != nil {
		return fmt.Errorf("launchctl bootstrap: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

// Uninstall stops and removes the agent.
func Uninstall() error {
	plistPath, err := PlistPath()
	if err != nil {
		return err
	}
	uid := strconv.Itoa(os.Getuid())
	exec.Command("launchctl", "bootout", "gui/"+uid, plistPath).Run()
	if err := os.Remove(plistPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// Status returns the current state of the launchd agent.
func Status() (AgentStatus, error) {
	plistPath, err := PlistPath()
	if err != nil {
		return AgentStatus{}, err
	}

	_, statErr := os.Stat(plistPath)
	installed := statErr == nil

	// launchctl list | grep label → "<PID>\t<status>\t<label>"
	out, err := exec.Command("launchctl", "list").Output()
	if err != nil {
		return AgentStatus{Installed: installed}, nil
	}

	for _, line := range strings.Split(string(out), "\n") {
		if !strings.Contains(line, label) {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}
		if parts[0] == "-" {
			return AgentStatus{Installed: installed, Running: false}, nil
		}
		pid, _ := strconv.Atoi(parts[0])
		return AgentStatus{Installed: true, Running: true, PID: pid}, nil
	}

	return AgentStatus{Installed: installed, Running: false}, nil
}

type plistData struct {
	Label   string
	BinPath string
	LogPath string
	Home    string
	Path    string
}

const plistTmpl = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>{{.Label}}</string>

    <key>ProgramArguments</key>
    <array>
        <string>{{.BinPath}}</string>
        <string>watch</string>
        <string>--daemon</string>
    </array>

    <key>RunAtLoad</key>
    <true/>

    <key>KeepAlive</key>
    <true/>

    <key>StandardOutPath</key>
    <string>{{.LogPath}}</string>

    <key>StandardErrorPath</key>
    <string>{{.LogPath}}</string>

    <key>EnvironmentVariables</key>
    <dict>
        <key>HOME</key>
        <string>{{.Home}}</string>
        <key>PATH</key>
        <string>{{.Path}}</string>
    </dict>
</dict>
</plist>`

func renderPlist(data plistData) (string, error) {
	t, err := template.New("plist").Parse(plistTmpl)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}
