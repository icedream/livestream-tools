//go:build !windows
// +build !windows

package main

// setIdlePriority normally sets priority to idle to not use CPU that is
// required for audio playback.
//
// On non-Windows OS this is a no-op.
func setIdlePriority() error {
	return nil
}
