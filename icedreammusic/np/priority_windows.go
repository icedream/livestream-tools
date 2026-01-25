//go:build windows
// +build windows

package main

import (
	"os"

	"golang.org/x/sys/windows"
)

// https://learn.microsoft.com/en-us/windows/win32/procthread/process-security-and-access-rights
const PROCESS_ALL_ACCESS = windows.STANDARD_RIGHTS_REQUIRED | windows.SYNCHRONIZE | 0xffff

func SetPriorityWindows(pid int, priority uint32) error {
	handle, err := windows.OpenProcess(PROCESS_ALL_ACCESS, false, uint32(pid))
	if err != nil {
		return err
	}
	defer windows.CloseHandle(handle) // Technically this can fail, but we ignore it if it does

	return windows.SetPriorityClass(handle, priority)
}

// setIdlePriority sets priority to idle to not use CPU that is required for audio playback
func setIdlePriority() error {
	return SetPriorityWindows(os.Getpid(), windows.IDLE_PRIORITY_CLASS)
}
