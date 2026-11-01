package main

import (
	"strconv"
	"strings"
)

// softGLDir is the folder (next to envui.exe) holding a second copy of
// envui.exe together with Mesa's software OpenGL renderer. It is used
// automatically when the machine has no usable OpenGL driver.
const softGLDir = "softgl"

// glUsable reports whether an OpenGL context with the given GL_VERSION and
// GL_RENDERER strings can run envui (Fyne needs OpenGL 2.1). Windows'
// built-in "GDI Generic" renderer only offers OpenGL 1.1.
func glUsable(version, renderer string) bool {
	if strings.Contains(strings.ToLower(renderer), "gdi generic") {
		return false
	}
	// GL_VERSION starts with "major.minor", optionally followed by more
	// text ("4.6.0 NVIDIA 551.23", "2.1 Mesa 23.0")
	fields := strings.Fields(version)
	if len(fields) == 0 {
		return false
	}
	parts := strings.SplitN(fields[0], ".", 3)
	if len(parts) < 2 {
		return false
	}
	major, err1 := strconv.Atoi(parts[0])
	minor, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return false
	}
	return major > 2 || (major == 2 && minor >= 1)
}
