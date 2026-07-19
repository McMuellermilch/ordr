//go:build !darwin

package launchd

import "fmt"

// AgentStatus describes the current state of a service manager agent.
type AgentStatus struct {
	Installed bool
	Running   bool
	PID       int
}

var errNotSupported = fmt.Errorf("watch agent management is only supported on macOS (launchd)")

func PlistPath() (string, error)  { return "", errNotSupported }
func LogPath() (string, error)    { return "", errNotSupported }
func Install() error              { return errNotSupported }
func Uninstall() error            { return errNotSupported }
func Status() (AgentStatus, error) { return AgentStatus{}, errNotSupported }
