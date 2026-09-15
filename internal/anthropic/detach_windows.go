//go:build windows

package anthropic

import "syscall"

// detachedProcess is Windows' CREATE_NO_WINDOW: the refresh runs without a
// console of its own, which is the closest equivalent to setsid here. It is
// not in the syscall package, so the value is spelled out.
const detachedProcess = 0x08000000

// detachAttr keeps the refresh from flashing a console window and from dying
// with the status line process that started it.
func detachAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{CreationFlags: detachedProcess | syscall.CREATE_NEW_PROCESS_GROUP}
}
