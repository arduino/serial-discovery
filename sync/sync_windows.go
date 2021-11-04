//
// This file is part of serial-discovery.
//
// Copyright 2018-2021 ARDUINO SA (http://www.arduino.cc/)
//
// This software is released under the GNU General Public License version 3,
// which covers the main part of arduino-cli.
// The terms of this license can be found at:
// https://www.gnu.org/licenses/gpl-3.0.en.html
//
// You can be released from the requirements of the above licenses by purchasing
// a commercial license. Buying such a license is mandatory if you want to modify or
// otherwise use the software for commercial activities involving the Arduino
// software without disclosing the source code of your own applications. To purchase
// a commercial license, send an email to license@arduino.cc.
//

package sync

import (
	"context"
	"fmt"
	"runtime"
	"syscall"
	"unsafe"

	discovery "github.com/arduino/pluggable-discovery-protocol-handler/v2"
	"go.bug.st/serial/enumerator"
)

//go:generate go run golang.org/x/sys/windows/mkwinsyscall -output zsyscall_windows.go sync_windows.go

//sys getModuleHandle(moduleName *byte) (handle syscall.Handle, err error) = GetModuleHandleA
//sys registerClass(wndClass *wndClass) (atom uint16, err error) = user32.RegisterClassA
//sys unregisterClass(className *byte) (err error) = user32.UnregisterClassA
//sys defWindowProc(hwnd syscall.Handle, msg uint32, wParam uintptr, lParam uintptr) (lResult uintptr) = user32.DefWindowProcW
//sys createWindowEx(exstyle uint32, className *byte, windowText *byte, style uint32, x int32, y int32, width int32, height int32, parent syscall.Handle, menu syscall.Handle, hInstance syscall.Handle, lpParam uintptr) (hwnd syscall.Handle, err error) = user32.CreateWindowExA
//sys destroyWindowEx(hwnd syscall.Handle) (err error) = user32.DestroyWindow
//sys registerDeviceNotification(recipient syscall.Handle, filter *devBroadcastDeviceInterface, flags uint32) (devHandle syscall.Handle, err error) = user32.RegisterDeviceNotificationA
//sys unregisterDeviceNotification(deviceHandle syscall.Handle) (err error) = user32.UnregisterDeviceNotification
//sys getMessage(msg *msg, hwnd syscall.Handle, msgFilterMin uint32, msgFilterMax uint32) (err error) = user32.GetMessageA
//sys dispatchMessage(msg *msg) (res int32, err error) = user32.DispatchMessageA

type wndClass struct {
	style        uint32
	wndProc      uintptr
	clsExtra     int32
	wndExtra     int32
	instance     syscall.Handle
	icon         syscall.Handle
	cursor       syscall.Handle
	brBackground syscall.Handle
	menuName     *byte
	className    *byte
}

type point struct {
	x int32
	y int32
}

type msg struct {
	hwnd     syscall.Handle
	message  uint32
	wParam   uintptr
	lParam   uintptr
	time     int32
	pt       point
	lPrivate int32
}

const wsExTopmost = 0x00000008

type guid struct {
	data1 uint32
	data2 uint16
	data3 uint16
	data4 [8]byte
}

type devBroadcastDeviceInterface struct {
	dwSize       uint32
	dwDeviceType uint32
	dwReserved   uint32
	classGUID    guid
	szName       uint16
}

// USB devices GUID used to filter notifications
var usbEventGUID guid = guid{
	data1: 0x10bfdca5,
	data2: 0x3065,
	data3: 0xd211,
	data4: [8]byte{0x90, 0x1f, 0x00, 0xc0, 0x4f, 0xb9, 0x51, 0xed},
}

const deviceNotifyWindowHandle = 0
const deviceNotifyAllInterfaceClasses = 4
const dbtDevtypeDeviceInterface = 5

type WindowProcCallback func(hwnd syscall.Handle, msg uint32, wParam uintptr, lParam uintptr) uintptr

// Start the sync process, successful events will be passed to eventCB, errors to errorCB.
// Returns a channel used to stop the sync process.
// Returns error if sync process can't be started.
func Start(eventCB discovery.EventCallback, errorCB discovery.ErrorCallback) (chan<- bool, error) {
	eventsChan := make(chan bool, 1)
	windowCallback := func(hwnd syscall.Handle, msg uint32, wParam uintptr, lParam uintptr) uintptr {
		select {
		case eventsChan <- true:
		default:
		}
		return defWindowProc(hwnd, msg, wParam, lParam)
	}

	go func() {
		current, err := enumerator.GetDetailedPortsList()
		if err != nil {
			errorCB(fmt.Sprintf("Error enumerating serial ports: %s", err))
			return
		}
		for _, port := range current {
			eventCB("add", toDiscoveryPort(port))
		}

		for {
			select {
			case ev := <-eventsChan:
				// Just one event could be queued because the channel has size 1
				// (more events coming after this one are discarded on send)
				if !ev {
					return
				}
			default:
			}
			updates, err := enumerator.GetDetailedPortsList()
			if err != nil {
				errorCB(fmt.Sprintf("Error enumerating serial ports: %s", err))
				return
			}
			processUpdates(current, updates, eventCB)
			current = updates
		}
	}()

	// Context used to stop the goroutine that consume the window messages
	ctx, cancel := context.WithCancel(context.Background())

	stopper := make(chan bool)
	go func() {
		// Lock this goroutine to the same OS thread for its whole execution,
		// if this is not done destruction of the windows will fail since
		// it must be done in the same thread that creates it
		runtime.LockOSThread()
		defer close(eventsChan)

		// We must create the window used to receive notifications in the same
		// thread that destroys it otherwise it would fail
		windowHandle, className, err := createWindow(windowCallback)
		if err != nil {
			errorCB(err.Error())
			return
		}
		defer func() {
			if err := destroyWindow(windowHandle, className); err != nil {
				errorCB(err.Error())
			}
		}()

		notificationsDevHandle, err := registerNotifications(windowHandle)
		if err != nil {
			errorCB(err.Error())
			return
		}
		defer func() {
			if err := unregisterNotifications(notificationsDevHandle); err != nil {
				errorCB(err.Error())
			}
		}()
		defer cancel()

		// To consume messages we need the window handle, so we must start
		// this goroutine in here and not outside the one that handles
		// creation and destruction of the window used to receive notifications
		go func() {
			if err := consumeMessages(ctx, windowHandle); err != nil {
				errorCB(err.Error())
			}
		}()

		<-stopper
	}()
	return stopper, nil
}

func createWindow(windowCallback WindowProcCallback) (syscall.Handle, *byte, error) {
	moduleHandle, err := getModuleHandle(nil)
	if err != nil {
		return syscall.InvalidHandle, nil, err
	}

	className, err := syscall.BytePtrFromString("arduino-serialdiscovery")
	if err != nil {
		return syscall.InvalidHandle, nil, err
	}
	windowClass := &wndClass{
		instance:  moduleHandle,
		className: className,
		wndProc:   syscall.NewCallback(windowCallback),
	}
	if _, err := registerClass(windowClass); err != nil {
		return syscall.InvalidHandle, nil, fmt.Errorf("registering new window: %s", err)
	}

	windowHandle, err := createWindowEx(wsExTopmost, className, className, 0, 0, 0, 0, 0, 0, 0, 0, 0)
	if err != nil {
		return syscall.InvalidHandle, nil, fmt.Errorf("creating window: %s", err)
	}
	return windowHandle, className, nil
}

func destroyWindow(windowHandle syscall.Handle, className *byte) error {
	if err := destroyWindowEx(windowHandle); err != nil {
		return fmt.Errorf("error destroying window: %s", err)
	}
	if err := unregisterClass(className); err != nil {
		return fmt.Errorf("error unregistering window class: %s", err)
	}
	return nil
}

func registerNotifications(windowHandle syscall.Handle) (syscall.Handle, error) {
	notificationFilter := devBroadcastDeviceInterface{
		dwDeviceType: dbtDevtypeDeviceInterface,
		classGUID:    usbEventGUID,
	}
	notificationFilter.dwSize = uint32(unsafe.Sizeof(notificationFilter))

	var flags uint32 = deviceNotifyWindowHandle | deviceNotifyAllInterfaceClasses
	notificationsDevHandle, err := registerDeviceNotification(windowHandle, &notificationFilter, flags)
	if err != nil {
		return syscall.InvalidHandle, err
	}

	return notificationsDevHandle, nil
}

func unregisterNotifications(notificationsDevHandle syscall.Handle) error {
	if err := unregisterDeviceNotification(notificationsDevHandle); err != nil {
		return fmt.Errorf("error unregistering device notifications: %s", err)
	}
	return nil
}

func consumeMessages(ctx context.Context, windowHandle syscall.Handle) error {
	var m msg
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}
		if err := getMessage(&m, windowHandle, 0, 0); err != nil {
			return fmt.Errorf("error consuming messages: %s", err)
		}
		dispatchMessage(&m)
	}
}
