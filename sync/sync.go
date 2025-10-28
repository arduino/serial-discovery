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

// Package sync provides functions for synchronizing and processing updates
// related to serial port discovery.
//
// This package includes functions for comparing two lists of serial ports and
// sending 'add' and 'remove' events based on the differences between the lists.
// It also includes utility functions for converting port details to the discovery
// protocol format.
//
// The main function in this package is `processUpdates`, which takes in two lists
// of serial port details and an event callback function. It compares the two lists
// and sends 'add' and 'remove' events based on the differences. The `portListHas`
// function is used to check if a port is contained in a list, and the `toDiscoveryPort`
// function is used to convert port details to the discovery protocol format.
package sync

import (
	"fmt"
	"regexp"
	"strings"
	"os"
	"path/filepath"
	"github.com/arduino/go-properties-orderedmap"
	discovery "github.com/arduino/pluggable-discovery-protocol-handler/v2"
	"go.bug.st/serial/enumerator"
)

var loaded = false
var filter = ""

func load() string {
	if loaded {
		return filter
	}

	loaded = true
	thepath, err := filepath.Abs(filepath.Dir(os.Args[0]))

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		return filter
	}

	data, err := os.ReadFile(filepath.Join(thepath, "skip.txt"))

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		return filter
	}

	filter = strings.Trim(string(data), " \t\r\n")

	return filter
}

func filterValid(ports []*enumerator.PortDetails) (ret []*enumerator.PortDetails) {
	filter := load()

	if len(filter) <= 0 {
		ret = ports
		return
	}

	for _, port := range ports {
		if isValid(port, filter) {
			ret = append(ret, port)
		}
	}
	return
}

func isValid(port *enumerator.PortDetails, filter string) bool {
	match, _ := regexp.MatchString(filter, port.Name)

	return !match;
}

// nolint
// processUpdates sends 'add' and 'remove' events by comparing two ports enumeration
// made at different times:
// - ports present in the new list but not in the old list are reported as 'added'
// - ports present in the old list but not in the new list are reported as 'removed'
func processUpdates(old, new []*enumerator.PortDetails, eventCB discovery.EventCallback) {
	for _, oldPort := range old {
		if !portListHas(new, oldPort) {
			eventCB("remove", &discovery.Port{
				Address:  oldPort.Name,
				Protocol: "serial",
			})
		}
	}

	for _, newPort := range new {
		if !portListHas(old, newPort) {
			eventCB("add", toDiscoveryPort(newPort))
		}
	}
}

// nolint
// portListHas checks if port is contained in list. The port metadata are
// compared in particular the port address, and vid/pid if the port is a usb port.
func portListHas(list []*enumerator.PortDetails, port *enumerator.PortDetails) bool {
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

func toDiscoveryPort(port *enumerator.PortDetails) *discovery.Port {
	protocolLabel := "Serial Port"
	hardwareID := ""
	props := properties.NewMap()
	if port.IsUSB {
		protocolLabel += " (USB)"
		props.Set("vid", "0x"+port.VID)
		props.Set("pid", "0x"+port.PID)
		props.Set("serialNumber", port.SerialNumber)
		hardwareID = port.SerialNumber
	}
	res := &discovery.Port{
		Address:       port.Name,
		AddressLabel:  port.Name,
		Protocol:      "serial",
		ProtocolLabel: protocolLabel,
		Properties:    props,
		HardwareID:    hardwareID,
	}
	return res
}
