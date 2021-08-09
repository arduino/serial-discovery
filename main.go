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
	"fmt"
	"os"

	discovery "github.com/arduino/pluggable-discovery-protocol-handler/v2"
	"github.com/arduino/serial-discovery/args"
	"github.com/arduino/serial-discovery/sync"
	"github.com/arduino/serial-discovery/version"
)

func main() {
	args.Parse()
	if args.ShowVersion {
		fmt.Printf("%s\n", version.VersionInfo)
		return
	}

	serialDisc := &SerialDiscovery{}
	disc := discovery.NewServer(serialDisc)
	if err := disc.Run(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(1)
	}
}

// SerialDiscovery is the implementation of the serial ports pluggable-discovery
type SerialDiscovery struct {
	closeChan chan<- bool
}

// Hello is the handler for the pluggable-discovery HELLO command
func (d *SerialDiscovery) Hello(userAgent string, protocolVersion int) error {
	return nil
}

// Quit is the handler for the pluggable-discovery QUIT command
func (d *SerialDiscovery) Quit() {
}

// Stop is the handler for the pluggable-discovery STOP command
func (d *SerialDiscovery) Stop() error {
	if d.closeChan != nil {
		d.closeChan <- true
		close(d.closeChan)
		d.closeChan = nil
	}
	return nil
}

// StartSync is the handler for the pluggable-discovery START_SYNC command
func (d *SerialDiscovery) StartSync(eventCB discovery.EventCallback, errorCB discovery.ErrorCallback) error {
	close, err := sync.Start(eventCB, errorCB)
	if err != nil {
		return err
	}
	d.closeChan = close
	return nil
}
