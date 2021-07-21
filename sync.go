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

package main

import (
	discovery "github.com/arduino/pluggable-discovery-protocol-handler"
	"go.bug.st/serial/enumerator"
)

// ProcessUpdates sends 'add' and 'remove' events by comparing two ports enumeration
// made at different times:
// - ports present in the new list but not in the old list are reported as 'added'
// - ports present in the old list but not in the new list are reported as 'removed'
func ProcessUpdates(old, new []*enumerator.PortDetails, eventCB discovery.EventCallback) {
	for _, oldPort := range old {
		if !PortListHas(new, oldPort) {
			eventCB("remove", &discovery.Port{
				Address:  oldPort.Name,
				Protocol: "serial",
			})
		}
	}

	for _, newPort := range new {
		if !PortListHas(old, newPort) {
			eventCB("add", toDiscoveryPort(newPort))
		}
	}
}

// PortListHas checks if port is contained in list. The port metadata are
// compared in particular the port address, and vid/pid if the port is a usb port.
func PortListHas(list []*enumerator.PortDetails, port *enumerator.PortDetails) bool {
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
