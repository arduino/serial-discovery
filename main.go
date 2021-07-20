//
// This file is part of serial-discovery.
//
// Copyright 2021 ARDUINO SA (http://www.arduino.cc/)
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
	"os"

	"github.com/arduino/go-properties-orderedmap"
	discovery "github.com/arduino/pluggable-discovery-protocol-handler"
	"github.com/arduino/serial-discovery/version"
	"go.bug.st/serial/enumerator"
)

func main() {
	parseArgs()
	if args.showVersion {
		fmt.Printf("serial-discovery %s (build timestamp: %s)\n", version.Tag, version.Timestamp)
		return
	}

	serialDisc := &SerialDiscovery{}
	disc := discovery.NewDiscoveryServer(serialDisc)
	if err := disc.Run(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(1)
	}
}

type SerialDiscovery struct {
	closeChan chan<- bool
}

func (d *SerialDiscovery) Hello(userAgent string, protocolVersion int) error {
	return nil
}

func (d *SerialDiscovery) Quit() {
}

func (d *SerialDiscovery) Start() error {
	return nil
}

func (d *SerialDiscovery) Stop() error {
	if d.closeChan != nil {
		d.closeChan <- true
		close(d.closeChan)
		d.closeChan = nil
	}
	return nil
}

func (d *SerialDiscovery) List() ([]*discovery.Port, error) {
	list, err := enumerator.GetDetailedPortsList()
	if err != nil {
		return nil, err
	}
	ports := []*discovery.Port{}
	for _, port := range list {
		ports = append(ports, toDiscoveryPort(port))
	}
	return ports, nil
}

func (d *SerialDiscovery) StartSync(eventCB discovery.EventCallback, errorCB discovery.ErrorCallback) error {
	close, err := startSync(eventCB, errorCB)
	if err != nil {
		return err
	}
	d.closeChan = close
	return nil
}

func toDiscoveryPort(port *enumerator.PortDetails) *discovery.Port {
	protocolLabel := "Serial Port"
	props := properties.NewMap()
	if port.IsUSB {
		protocolLabel += " (USB)"
		props.Set("vid", "0x"+port.VID)
		props.Set("pid", "0x"+port.PID)
		props.Set("serialNumber", port.SerialNumber)
	}
	res := &discovery.Port{
		Address:       port.Name,
		AddressLabel:  port.Name,
		Protocol:      "serial",
		ProtocolLabel: protocolLabel,
		Properties:    props,
	}
	return res
}
