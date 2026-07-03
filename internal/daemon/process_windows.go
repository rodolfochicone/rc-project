//go:build windows

package daemon

import (
	"log/slog"
	"math"

	"golang.org/x/sys/windows"
)

// ProcessAlive reports whether a process with pid is currently alive.
func ProcessAlive(pid int) bool {
	if pid <= 0 || uint64(pid) > math.MaxUint32 {
		return false
	}

	handle, err := windows.OpenProcess(windows.SYNCHRONIZE, false, uint32(pid))
	if err != nil {
		return false
	}
	defer func() {
		if cerr := windows.CloseHandle(handle); cerr != nil {
			slog.Warn("close process handle", "pid", pid, "error", cerr)
		}
	}()

	event, err := windows.WaitForSingleObject(handle, 0)
	if err != nil {
		return false
	}
	return event == uint32(windows.WAIT_TIMEOUT)
}
