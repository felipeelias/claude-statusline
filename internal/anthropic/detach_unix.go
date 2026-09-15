//go:build !windows

package anthropic

import "syscall"

// detachAttr puts the refresh in its own session, so it survives the status
// line process exiting and never inherits its controlling terminal.
func detachAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setsid: true}
}
