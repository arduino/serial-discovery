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
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/arduino/go-properties-orderedmap"
	"go.bug.st/serial.v1/enumerator"
)

func main() {
	syncStarted := false
	var syncCloseChan chan<- bool

	reader := bufio.NewReader(os.Stdin)
	for {
		cmd, err := reader.ReadString('\n')
		if err != nil {
			outputError(err)
			os.Exit(1)
		}
		cmd = strings.ToUpper(strings.TrimSpace(cmd))
		switch cmd {
		case "START":
			outputMessage("start", "OK")
		case "STOP":
			if syncStarted {
				syncCloseChan <- true
				syncStarted = false
			}
			outputMessage("stop", "OK")
		case "LIST":
			outputList()
		case "QUIT":
			outputMessage("quit", "OK")
			os.Exit(0)
		case "START_SYNC":
			if syncStarted {
				outputMessage("startSync", "OK")
			} else if close, err := startSync(); err != nil {
				outputError(err)
			} else {
				syncCloseChan = close
				syncStarted = true
			}
		default:
			outputError(fmt.Errorf("Command %s not supported", cmd))
		}
	}
}

type boardPortJSON struct {
	Address             string          `json:"address"`
	Label               string          `json:"label"`
	Prefs               *properties.Map `json:"prefs"`
	IdentificationPrefs *properties.Map `json:"identificationPrefs"`
	Protocol            string          `json:"protocol"`
	ProtocolLabel       string          `json:"protocolLabel"`
}

type listOutputJSON struct {
	EventType string           `json:"eventType"`
	Ports     []*boardPortJSON `json:"ports"`
}

func outputList() {
	list, err := enumerator.GetDetailedPortsList()
	if err != nil {
		outputError(err)
		return
	}
	portsJSON := []*boardPortJSON{}
	for _, port := range list {
		portJSON := newBoardPortJSON(port)
		portsJSON = append(portsJSON, portJSON)
	}
	d, err := json.MarshalIndent(&listOutputJSON{
		EventType: "list",
		Ports:     portsJSON,
	}, "", "  ")
	if err != nil {
		outputError(err)
		return
	}
	syncronizedPrintLn(string(d))
}

func newBoardPortJSON(port *enumerator.PortDetails) *boardPortJSON {
	prefs := properties.NewMap()
	identificationPrefs := properties.NewMap()
	portJSON := &boardPortJSON{
		Address:             port.Name,
		Label:               port.Name,
		Protocol:            "serial",
		ProtocolLabel:       "Serial Port",
		Prefs:               prefs,
		IdentificationPrefs: identificationPrefs,
	}
	if port.IsUSB {
		portJSON.ProtocolLabel = "USB Serial Port"
		portJSON.Prefs.Set("vendorId", "0x"+port.VID)
		portJSON.Prefs.Set("productId", "0x"+port.PID)
		portJSON.Prefs.Set("serialNumber", port.SerialNumber)
		portJSON.IdentificationPrefs.Set("pid", "0x"+port.PID)
		portJSON.IdentificationPrefs.Set("vid", "0x"+port.VID)
	}
	return portJSON
}

type messageOutputJSON struct {
	EventType string `json:"eventType"`
	Message   string `json:"message"`
}

func outputMessage(eventType, message string) {
	d, err := json.MarshalIndent(&messageOutputJSON{
		EventType: eventType,
		Message:   message,
	}, "", "  ")
	if err != nil {
		outputError(err)
	} else {
		syncronizedPrintLn(string(d))
	}
}

func outputError(err error) {
	outputMessage("error", err.Error())
}

var stdoutMutext sync.Mutex

func syncronizedPrintLn(a ...interface{}) {
	stdoutMutext.Lock()
	fmt.Println(a...)
	stdoutMutext.Unlock()
}
