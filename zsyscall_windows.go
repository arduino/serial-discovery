// Code generated by 'go generate'; DO NOT EDIT.

package main

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var _ unsafe.Pointer

// Do the interface allocations only once for common
// Errno values.
const (
	errnoERROR_IO_PENDING = 997
)

var (
	errERROR_IO_PENDING error = syscall.Errno(errnoERROR_IO_PENDING)
)

// errnoErr returns common boxed Errno values, to prevent
// allocations at runtime.
func errnoErr(e syscall.Errno) error {
	switch e {
	case 0:
		return nil
	case errnoERROR_IO_PENDING:
		return errERROR_IO_PENDING
	}
	// TODO: add more here, after collecting data on the common
	// error values see on Windows. (perhaps when running
	// all.bat?)
	return e
}

var (
	modkernel32 = windows.NewLazySystemDLL("kernel32.dll")
	moduser32   = windows.NewLazySystemDLL("user32.dll")

	procGetModuleHandleA            = modkernel32.NewProc("GetModuleHandleA")
	procRegisterClassA              = moduser32.NewProc("RegisterClassA")
	procDefWindowProcW              = moduser32.NewProc("DefWindowProcW")
	procCreateWindowExA             = moduser32.NewProc("CreateWindowExA")
	procRegisterDeviceNotificationA = moduser32.NewProc("RegisterDeviceNotificationA")
	procGetMessageA                 = moduser32.NewProc("GetMessageA")
	procTranslateMessage            = moduser32.NewProc("TranslateMessage")
	procDispatchMessageA            = moduser32.NewProc("DispatchMessageA")
)

func getModuleHandle(moduleName *byte) (handle syscall.Handle, err error) {
	r0, _, e1 := syscall.Syscall(procGetModuleHandleA.Addr(), 1, uintptr(unsafe.Pointer(moduleName)), 0, 0)
	handle = syscall.Handle(r0)
	if handle == 0 {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func registerClass(wndClass *wndClass) (atom uint16, err error) {
	r0, _, e1 := syscall.Syscall(procRegisterClassA.Addr(), 1, uintptr(unsafe.Pointer(wndClass)), 0, 0)
	atom = uint16(r0)
	if atom == 0 {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func defWindowProc(hwnd syscall.Handle, msg uint32, wParam uintptr, lParam uintptr) (lResult uintptr) {
	r0, _, _ := syscall.Syscall6(procDefWindowProcW.Addr(), 4, uintptr(hwnd), uintptr(msg), uintptr(wParam), uintptr(lParam), 0, 0)
	lResult = uintptr(r0)
	return
}

func createWindowEx(exstyle uint32, className *byte, windowText *byte, style uint32, x int32, y int32, width int32, height int32, parent syscall.Handle, menu syscall.Handle, hInstance syscall.Handle, lpParam uintptr) (hwnd syscall.Handle, err error) {
	r0, _, e1 := syscall.Syscall12(procCreateWindowExA.Addr(), 12, uintptr(exstyle), uintptr(unsafe.Pointer(className)), uintptr(unsafe.Pointer(windowText)), uintptr(style), uintptr(x), uintptr(y), uintptr(width), uintptr(height), uintptr(parent), uintptr(menu), uintptr(hInstance), uintptr(lpParam))
	hwnd = syscall.Handle(r0)
	if hwnd == 0 {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func registerDeviceNotification(recipient syscall.Handle, filter *devBroadcastDeviceInterface, flags uint32) (devHandle syscall.Handle, err error) {
	r0, _, e1 := syscall.Syscall(procRegisterDeviceNotificationA.Addr(), 3, uintptr(recipient), uintptr(unsafe.Pointer(filter)), uintptr(flags))
	devHandle = syscall.Handle(r0)
	if devHandle == 0 {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func getMessage(msg *msg, hwnd syscall.Handle, msgFilterMin uint32, msgFilterMax uint32) (res int32, err error) {
	r0, _, e1 := syscall.Syscall6(procGetMessageA.Addr(), 4, uintptr(unsafe.Pointer(msg)), uintptr(hwnd), uintptr(msgFilterMin), uintptr(msgFilterMax), 0, 0)
	res = int32(r0)
	if res == 0 {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

func translateMessage(msg *msg) (res bool) {
	r0, _, _ := syscall.Syscall(procTranslateMessage.Addr(), 1, uintptr(unsafe.Pointer(msg)), 0, 0)
	res = r0 != 0
	return
}

func dispatchMessage(msg *msg) (res int32, err error) {
	r0, _, e1 := syscall.Syscall(procDispatchMessageA.Addr(), 1, uintptr(unsafe.Pointer(msg)), 0, 0)
	res = int32(r0)
	if res == 0 {
		if e1 != 0 {
			err = errnoErr(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}
