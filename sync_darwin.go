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
	"syscall"

	discovery "github.com/arduino/pluggable-discovery-protocol-handler"
	"go.bug.st/serial/enumerator"
)

func startSync(eventCB discovery.EventCallback) (chan<- bool, error) {
	// create kqueue
	kq, err := syscall.Kqueue()
	if err != nil {
		return nil, err
	}

	// open folder
	fd, err := syscall.Open("/dev", syscall.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}

	// build kevent
	ev1 := syscall.Kevent_t{
		Ident:  uint64(fd),
		Filter: syscall.EVFILT_VNODE,
		Flags:  syscall.EV_ADD | syscall.EV_ENABLE | syscall.EV_ONESHOT,
		Fflags: syscall.NOTE_DELETE | syscall.NOTE_WRITE,
		Data:   0,
		Udata:  nil,
	}

	// Helper function to avoid decoging kqueue event messages
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

	// Run synchronous event emitter
	closeChan := make(chan bool)

	go func() {
		// Ouput initial port state: get the current port list to send as initial "add" events
		current, err := enumerator.GetDetailedPortsList()
		if err != nil {
			// TODO: how to handle errors? should we just retry silently?
		}
		for _, port := range current {
			eventCB("add", toDiscoveryPort(port))
		}

		// wait for events
		events := make([]syscall.Kevent_t, 10)
		retries := 0

		for {
			for {
				t100ms := syscall.Timespec{Nsec: 100000000, Sec: 0}
				n, err := syscall.Kevent(kq, []syscall.Kevent_t{ev1}, events, &t100ms)
				select {
				case <-closeChan:
					syscall.Close(kq)
					syscall.Close(fd)
					return
				default:
				}
				if err == syscall.EINTR {
					continue
				}
				if err != nil {
					// output(&genericMessageJSON{
					// 	EventType: "start_sync",
					// 	Error:     true,
					// 	Message:   fmt.Sprintf("error decoding START_SYNC event: %s", err),
					// })
					// TODO: how to handle errors? should we just retry silently?
				}
				// if there is an event retry up to 5 times
				if n > 0 {
					retries = 5
				}
				break
			}

			for retries > 0 {
				retries--
				updates, _ := enumerator.GetDetailedPortsList()

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
		}
	}()

	return closeChan, nil
}
