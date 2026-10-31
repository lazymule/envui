//go:build windows

package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
)

// relaunchWithSoftwareGL supports running on machines without an OpenGL
// driver (virtual machines, Remote Desktop, servers).
//
// If Mesa's software renderer (opengl32.dll + libgallium_wgl.dll from
// https://github.com/pal1000/mesa-dist-win) sits next to envui.exe, Windows
// loads it instead of the system OpenGL. Mesa reads GALLIUM_DRIVER when its
// DLLs are loaded at process start, which is before main runs, so we
// re-launch ourselves once with GALLIUM_DRIVER=llvmpipe (the CPU renderer).
//
// It returns true when the child process has run and this process should exit.
func relaunchWithSoftwareGL() bool {
	if os.Getenv("GALLIUM_DRIVER") != "" {
		return false // already set (by us or by the user)
	}
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(exe), "libgallium_wgl.dll")); err != nil {
		return false // no bundled Mesa: use the normal system OpenGL
	}

	cmd := exec.Command(exe, os.Args[1:]...)
	cmd.Env = append(os.Environ(), "GALLIUM_DRIVER=llvmpipe")
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.ExitCode())
		}
		return false // could not start the child: try in this process instead
	}
	return true
}
