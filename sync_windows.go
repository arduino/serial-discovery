//
// This file is part of serial-discovery.
//
// Copyright 2018 ARDUINO SA (http://www.arduino.cc/)
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

package main

import (
	"fmt"
	"runtime"
	"syscall"
	"time"
	"unsafe"

	discovery "github.com/arduino/pluggable-discovery-protocol-handler"
	"go.bug.st/serial/enumerator"
)

//go:generate go run golang.org/x/sys/windows/mkwinsyscall -output zsyscall_windows.go sync_windows.go

//sys getModuleHandle(moduleName *byte) (handle syscall.Handle, err error) = GetModuleHandleA
//sys registerClass(wndClass *wndClass) (atom uint16, err error) = user32.RegisterClassA
//sys defWindowProc(hwnd syscall.Handle, msg uint32, wParam uintptr, lParam uintptr) (lResult uintptr) = user32.DefWindowProcW
//sys createWindowEx(exstyle uint32, className *byte, windowText *byte, style uint32, x int32, y int32, width int32, height int32, parent syscall.Handle, menu syscall.Handle, hInstance syscall.Handle, lpParam uintptr) (hwnd syscall.Handle, err error) = user32.CreateWindowExA
//sys registerDeviceNotification(recipient syscall.Handle, filter *devBroadcastDeviceInterface, flags uint32) (devHandle syscall.Handle, err error) = user32.RegisterDeviceNotificationA
//sys getMessage(msg *msg, hwnd syscall.Handle, msgFilterMin uint32, msgFilterMax uint32) (res int32, err error) = user32.GetMessageA
//sys translateMessage(msg *msg) (res bool) = user32.TranslateMessage
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

const wsExDlgModalFrame = 0x00000001
const wsExTopmost = 0x00000008
const wsExTransparent = 0x00000020
const wsExMDIChild = 0x00000040
const wsExToolWindow = 0x00000080
const wsExAppWindow = 0x00040000
const wsExLayered = 0x00080000

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

//var usbEventGUID = guid{???} // TODO

const deviceNotifyWindowHandle = 0
const deviceNotifySserviceHandle = 1
const deviceNotifyAllInterfaceClasses = 4

const dbtDevtypeDeviceInterface = 5

func init() {
	runtime.LockOSThread()
}

func startSync(eventCB discovery.EventCallback, errorCB discovery.ErrorCallback) (chan<- bool, error) {
	startResult := make(chan error)
	event := make(chan bool, 1)
	go func() {
		initAndRunWindowHandler(startResult, event)
	}()
	if err := <-startResult; err != nil {
		return nil, err
	}
	go func() {
		current, err := enumerator.GetDetailedPortsList()
		if err != nil {
			errorCB(fmt.Sprintf("Error enumarating serial ports: %s", err))
			return
		}
		for _, port := range current {
			eventCB("add", toDiscoveryPort(port))
		}

		for {
			<-event

			// Wait 100 ms to pile up events
			time.Sleep(100 * time.Millisecond)
			select {
			case <-event:
				// Just one event could be queued because the channel has size 1
				// (more events coming after this one are discarded on send)
			default:
			}

			// Send updates

			updates, err := enumerator.GetDetailedPortsList()
			if err != nil {
				errorCB(fmt.Sprintf("Error enumarating serial ports: %s", err))
				return
			}

			portListHas := func(list []*enumerator.PortDetails, port *enumerator.PortDetails) bool {
				for _, p := range list {
					if port.Name == p.Name && port.IsUSB == p.IsUSB {
						if p.IsUSB &&
							port.VID == p.VID &&
							port.PID == p.PID &&
							port.SerialNumber == p.SerialNumber {
							return true
						}
						if !p.IsUSB {
							return true
						}
					}
				}
				return false
			}

			for _, port := range current {
				if !portListHas(updates, port) {
					eventCB("remove", &discovery.Port{
						Address:  port.Name,
						Protocol: "serial",
					})
				}
			}

			for _, port := range updates {
				if !portListHas(current, port) {
					eventCB("add", toDiscoveryPort(port))
				}
			}

			current = updates
		}
	}()
	quit := make(chan bool)
	go func() {
		<-quit
		// TODO: implement termination channel
	}()
	return quit, nil
}

func initAndRunWindowHandler(startResult chan<- error, event chan<- bool) {
	handle, err := getModuleHandle(nil)
	if err != nil {
		startResult <- err
		return
	}

	wndProc := func(hwnd syscall.Handle, msg uint32, wParam uintptr, lParam uintptr) uintptr {
		select {
		case event <- true:
		default:
		}
		return defWindowProc(hwnd, msg, wParam, lParam)
	}

	className := syscall.StringBytePtr("serialdiscovery")
	windowClass := &wndClass{
		instance:  handle,
		className: className,
		wndProc:   syscall.NewCallback(wndProc),
	}
	if _, err := registerClass(windowClass); err != nil {
		startResult <- fmt.Errorf("registering new window: %s", err)
		return
	}

	hwnd, err := createWindowEx(wsExTopmost, className, className, 0, 0, 0, 0, 0, 0, 0, 0, 0)
	if err != nil {
		startResult <- fmt.Errorf("creating window: %s", err)
		return
	}

	notificationFilter := devBroadcastDeviceInterface{
		dwDeviceType: dbtDevtypeDeviceInterface,
		// TODO: Filter USB events using the correct GUID
	}
	notificationFilter.dwSize = uint32(unsafe.Sizeof(notificationFilter))

	if _, err := registerDeviceNotification(
		hwnd,
		&notificationFilter,
		deviceNotifyWindowHandle|deviceNotifyAllInterfaceClasses); err != nil {
		startResult <- fmt.Errorf("registering for devices notification: %s", err)
		return
	}

	startResult <- nil

	var m msg
	for {
		if res, err := getMessage(&m, hwnd, 0, 0); res == 0 || res == -1 {
			if err != nil {
				// TODO: send err and stop sync mode.
				// fmt.Println(err)
			}
			break
		}
		translateMessage(&m)
		dispatchMessage(&m)
	}
}
