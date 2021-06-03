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
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/arduino/go-properties-orderedmap"
	"github.com/arduino/serial-discovery/version"
	"go.bug.st/serial/enumerator"
)

func main() {
	parseArgs()
	if args.showVersion {
		fmt.Printf("serial-discovery %s (build timestamp: %s)\n", version.Tag, version.Timestamp)
		return
	}

	syncStarted := false
	var syncCloseChan chan<- bool

	reader := bufio.NewReader(os.Stdin)
	for {
		fullCmd, err := reader.ReadString('\n')
		if err != nil {
			output(&genericMessageJSON{
				EventType: "command_error",
				Error:     true,
				Message:   err.Error(),
			})
			os.Exit(1)
		}
		split := strings.Split(fullCmd, " ")
		cmd := strings.ToUpper(strings.TrimSpace(split[0]))
		switch cmd {
		case "HELLO":
			re := regexp.MustCompile(`(\d+) "([^"]+)"`)
			matches := re.FindStringSubmatch(fullCmd[6:])
			if len(matches) != 3 {
				output(&genericMessageJSON{
					EventType: "command_error",
					Error:     true,
					Message:   "Invalid HELLO command",
				})
				continue
			}
			// userAgent := matches[2]
			// reqProtocolVersion, err := strconv.ParseUint(matches[1], 10, 64)
			_, err := strconv.ParseUint(matches[1], 10, 64)
			if err != nil {
				output(&genericMessageJSON{
					EventType: "command_error",
					Error:     true,
					Message:   "Invalid protocol version: " + matches[2],
				})
			}
			// fmt.Println("User agent:", userAgent)
			// fmt.Println("Req. Protocol version:", reqProtocolVersion)
			output(&helloMessageJSON{
				EventType:       "hello",
				ProtocolVersion: "1", // Protocol version 1 is the only supported for now...
				Message:         "OK",
			})
		case "START":
			output(&genericMessageJSON{
				EventType: "start",
				Message:   "OK",
			})
		case "STOP":
			if syncStarted {
				syncCloseChan <- true
				syncStarted = false
			}
			output(&genericMessageJSON{
				EventType: "stop",
				Message:   "OK",
			})
		case "LIST":
			outputList()
		case "QUIT":
			output(&genericMessageJSON{
				EventType: "quit",
				Message:   "OK",
			})
			os.Exit(0)
		case "START_SYNC":
			if syncStarted {
				// sync already started, just acknowledge again...
				output(&genericMessageJSON{
					EventType: "start_sync",
					Message:   "OK",
				})
			} else if close, err := startSync(); err != nil {
				output(&genericMessageJSON{
					EventType: "start_sync",
					Error:     true,
					Message:   err.Error(),
				})
			} else {
				syncCloseChan = close
				syncStarted = true
			}
		default:
			output(&genericMessageJSON{
				EventType: "command_error",
				Error:     true,
				Message:   fmt.Sprintf("Command %s not supported", cmd),
			})
		}
	}
}

type boardPortJSON struct {
	Address       string          `json:"address"`
	Label         string          `json:"label,omitempty"`
	Protocol      string          `json:"protocol,omitempty"`
	ProtocolLabel string          `json:"protocolLabel,omitempty"`
	Properties    *properties.Map `json:"properties,omitempty"`
}

type listOutputJSON struct {
	EventType string           `json:"eventType"`
	Ports     []*boardPortJSON `json:"ports"`
}

func outputList() {
	list, err := enumerator.GetDetailedPortsList()
	if err != nil {
		output(&genericMessageJSON{
			EventType: "list",
			Error:     true,
			Message:   err.Error(),
		})
		return
	}
	portsJSON := []*boardPortJSON{}
	for _, port := range list {
		portJSON := newBoardPortJSON(port)
		portsJSON = append(portsJSON, portJSON)
	}
	output(&listOutputJSON{
		EventType: "list",
		Ports:     portsJSON,
	})
}

func newBoardPortJSON(port *enumerator.PortDetails) *boardPortJSON {
	prefs := properties.NewMap()
	portJSON := &boardPortJSON{
		Address:       port.Name,
		Label:         port.Name,
		Protocol:      "serial",
		ProtocolLabel: "Serial Port",
		Properties:    prefs,
	}
	if port.IsUSB {
		portJSON.ProtocolLabel = "Serial Port (USB)"
		portJSON.Properties.Set("vid", "0x"+port.VID)
		portJSON.Properties.Set("pid", "0x"+port.PID)
		portJSON.Properties.Set("serialNumber", port.SerialNumber)
	}
	return portJSON
}

type helloMessageJSON struct {
	EventType       string `json:"eventType"`
	ProtocolVersion string `json:"protocolVersion"`
	Message         string `json:"message"`
}

type genericMessageJSON struct {
	EventType string `json:"eventType"`
	Error     bool   `json:"error,omitempty"`
	Message   string `json:"message"`
}

func output(msg interface{}) {
	d, err := json.MarshalIndent(msg, "", "  ")
	if err != nil {
		output(&genericMessageJSON{
			EventType: "command_error",
			Error:     true,
			Message:   err.Error(),
		})
	} else {
		syncronizedPrintLn(string(d))
	}
}

var stdoutMutext sync.Mutex

func syncronizedPrintLn(a ...interface{}) {
	stdoutMutext.Lock()
	fmt.Println(a...)
	stdoutMutext.Unlock()
}
