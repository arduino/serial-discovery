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
	"fmt"
	"io"

	discovery "github.com/arduino/pluggable-discovery-protocol-handler/v2"
	"github.com/s-urbaniak/uevent"
	"go.bug.st/serial/enumerator"
)

// Start the sync process, successful events will be passed to eventCB, errors to errorCB.
// Returns a channel used to stop the sync process.
// Returns error if sync process can't be started.
func Start(eventCB discovery.EventCallback, errorCB discovery.ErrorCallback) (chan<- bool, error) {
	// Get the current port list to send as initial "add" events
	current, err := enumerator.GetDetailedPortsList()
	if err != nil {
		return nil, err
	}

	current = filterValid(current)

	// Start sync reader from udev
	syncReader, err := uevent.NewReader()
	if err != nil {
		return nil, err
	}

	closeChan := make(chan bool)
	go func() {
		<-closeChan
		syncReader.Close()
	}()

	// Run synchronous event emitter
	go func() {
		// Output initial port state
		for _, port := range current {
			eventCB("add", toDiscoveryPort(port))
		}

		dec := uevent.NewDecoder(syncReader)
		for {
			evt, err := dec.Decode()
			if err == io.EOF {
				// The underlying syncReader has been closed
				// so there's nothing else to read
				return
			} else if err != nil {
				errorCB(fmt.Sprintf("Error decoding serial event: %s", err))
				return
			}
			if evt.Subsystem != "tty" {
				continue
			}
			changedPort := "/dev/" + evt.Vars["DEVNAME"]
			if evt.Action == "add" {
				portList, err := enumerator.GetDetailedPortsList()
				if err != nil {
					continue
				}

				portList = filterValid(portList)

				for _, port := range portList {
					if port.IsUSB && port.Name == changedPort {
						eventCB("add", toDiscoveryPort(port))
						break
					}
				}
			}
			if evt.Action == "remove" {
				eventCB("remove", &discovery.Port{
					Address:  changedPort,
					Protocol: "serial",
				})
			}
		}
	}()

	return closeChan, nil
}
