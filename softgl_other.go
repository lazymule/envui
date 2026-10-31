//go:build !windows

package main

// relaunchWithSoftwareGL is only needed on Windows (see softgl_windows.go).
func relaunchWithSoftwareGL() bool { return false }
