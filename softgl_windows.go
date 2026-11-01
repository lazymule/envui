//go:build windows

package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"unsafe"

	"golang.org/x/sys/windows"
)

// relaunchWithSoftwareGL makes envui work on machines without a usable
// OpenGL driver (virtual machines, Remote Desktop, servers).
//
// Mesa's software renderer (opengl32.dll + libgallium_wgl.dll, from
// https://github.com/pal1000/mesa-dist-win) replaces the system OpenGL when
// it sits next to the exe, because Windows looks in the exe's folder first.
// It therefore lives in a separate "softgl" folder with its own copy of
// envui.exe, so normal machines keep using their graphics card:
//
//   - envui.exe probes the system OpenGL. If it is too old or missing (or
//     ENVUI_SOFTWARE_GL=1 is set), it starts softgl\envui.exe instead.
//   - an envui.exe that has Mesa next to it makes sure Mesa uses its CPU
//     renderer (llvmpipe). Mesa reads GALLIUM_DRIVER when its DLLs load at
//     process start, so that has to be set before the process starts.
//
// It returns true when another process ran envui and this one should exit.
func relaunchWithSoftwareGL() bool {
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	dir := filepath.Dir(exe)

	// Mesa next to this exe: ensure the llvmpipe driver is selected
	if fileExists(filepath.Join(dir, "libgallium_wgl.dll")) {
		if os.Getenv("GALLIUM_DRIVER") != "" {
			return false
		}
		return runChild(exe)
	}

	// normal exe: fall back to the bundled software renderer if needed
	soft := filepath.Join(dir, softGLDir, "envui.exe")
	if !fileExists(soft) || !fileExists(filepath.Join(dir, softGLDir, "libgallium_wgl.dll")) {
		return false
	}
	if os.Getenv("ENVUI_SOFTWARE_GL") != "1" && hardwareGLUsable() {
		return false
	}
	return runChild(soft)
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// runChild runs exe with our arguments and GALLIUM_DRIVER=llvmpipe, waits for
// it and exits with its exit code. It returns false if the child couldn't
// be started, so the caller can carry on in this process.
func runChild(exe string) bool {
	cmd := exec.Command(exe, os.Args[1:]...)
	env := os.Environ()
	if os.Getenv("GALLIUM_DRIVER") == "" {
		env = append(env, "GALLIUM_DRIVER=llvmpipe")
	}
	cmd.Env = env
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.ExitCode())
		}
		return false
	}
	return true
}

// pixelFormatDescriptor is PIXELFORMATDESCRIPTOR from wingdi.h.
type pixelFormatDescriptor struct {
	Size, Version                                  uint16
	Flags                                          uint32
	PixelType, ColorBits, RedBits, RedShift        uint8
	GreenBits, GreenShift, BlueBits, BlueShift     uint8
	AlphaBits, AlphaShift, AccumBits, AccumRedBits uint8
	AccumGreenBits, AccumBlueBits, AccumAlphaBits  uint8
	DepthBits, StencilBits, AuxBuffers, LayerType  uint8
	Reserved                                       uint8
	LayerMask, VisibleMask, DamageMask             uint32
}

// hardwareGLUsable creates a throw-away OpenGL context on a hidden window
// and checks its version. If anything unexpected goes wrong it answers
// true, i.e. "behave as before and let Fyne try".
func hardwareGLUsable() (ok bool) {
	defer func() {
		if recover() != nil {
			ok = true
		}
	}()
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	user32 := windows.NewLazySystemDLL("user32.dll")
	gdi32 := windows.NewLazySystemDLL("gdi32.dll")
	opengl32 := windows.NewLazyDLL("opengl32.dll") // the one already loaded

	createWindow := user32.NewProc("CreateWindowExW")
	destroyWindow := user32.NewProc("DestroyWindow")
	getDC := user32.NewProc("GetDC")
	releaseDC := user32.NewProc("ReleaseDC")
	choosePF := gdi32.NewProc("ChoosePixelFormat")
	setPF := gdi32.NewProc("SetPixelFormat")
	wglCreate := opengl32.NewProc("wglCreateContext")
	wglMakeCurrent := opengl32.NewProc("wglMakeCurrent")
	wglDelete := opengl32.NewProc("wglDeleteContext")
	glGetString := opengl32.NewProc("glGetString")

	class, _ := windows.UTF16PtrFromString("STATIC")
	title, _ := windows.UTF16PtrFromString("envui-gl-probe")
	const wsPopup = 0x80000000
	hwnd, _, _ := createWindow.Call(0, uintptr(unsafe.Pointer(class)), uintptr(unsafe.Pointer(title)),
		wsPopup, 0, 0, 1, 1, 0, 0, 0, 0)
	if hwnd == 0 {
		return true // can't even probe: don't interfere
	}
	defer destroyWindow.Call(hwnd)

	hdc, _, _ := getDC.Call(hwnd)
	if hdc == 0 {
		return true
	}
	defer releaseDC.Call(hwnd, hdc)

	pfd := pixelFormatDescriptor{
		Version:     1,
		Flags:       0x4 | 0x20 | 0x1, // DRAW_TO_WINDOW | SUPPORT_OPENGL | DOUBLEBUFFER
		ColorBits:   32,
		DepthBits:   24,
		StencilBits: 8,
	}
	pfd.Size = uint16(unsafe.Sizeof(pfd))
	format, _, _ := choosePF.Call(hdc, uintptr(unsafe.Pointer(&pfd)))
	if format == 0 {
		return false // no OpenGL pixel format at all
	}
	if r, _, _ := setPF.Call(hdc, format, uintptr(unsafe.Pointer(&pfd))); r == 0 {
		return false
	}
	ctx, _, _ := wglCreate.Call(hdc)
	if ctx == 0 {
		return false
	}
	defer wglDelete.Call(ctx)
	if r, _, _ := wglMakeCurrent.Call(hdc, ctx); r == 0 {
		return false
	}
	defer wglMakeCurrent.Call(0, 0)

	const glVersion, glRenderer = 0x1F02, 0x1F01
	v, _, _ := glGetString.Call(glVersion)
	r, _, _ := glGetString.Call(glRenderer)
	if v == 0 {
		return false
	}
	version := cString(v)
	renderer := ""
	if r != 0 {
		renderer = cString(r)
	}
	return glUsable(version, renderer)
}

// cString converts a C string returned as uintptr by a system call (memory
// owned by the OpenGL driver, not the Go heap).
func cString(p uintptr) string {
	return windows.BytePtrToString(*(**byte)(unsafe.Pointer(&p)))
}
