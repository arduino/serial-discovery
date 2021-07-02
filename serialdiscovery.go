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
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"go.bug.st/serial/enumerator"
)

const maximumProtocolVersion uint64 = 1

type SerialDiscovery struct {
	initialized   bool
	started       bool
	syncStarted   bool
	syncInterrupt chan bool
	outputChan    chan interface{}

	userAgent       string
	protocolVersion uint64
}

// NewSerialDiscovery creates a new SerialDiscovery and starts the goroutine
// that handles printing messages to stdout.
func NewSerialDiscovery() *SerialDiscovery {
	s := &SerialDiscovery{
		initialized: false,
		started:     false,
		syncStarted: false,
		outputChan:  make(chan interface{}),
	}
	// Start go routine to serialize messages printing
	go func() {
		for message := range s.outputChan {
			data, err := json.MarshalIndent(message, "", "  ")
			if err != nil {
				// We are certain that this will be marshalled correctly
				// so we don't handle the error
				data, _ = json.MarshalIndent(&genericMessageJSON{
					EventType: "command_error",
					Error:     true,
					Message:   err.Error(),
				}, "", "  ")
			}
			fmt.Println(string(data))
		}
	}()
	return s
}

// CloseOutput closes the channel that receives messages for printing.
func (s *SerialDiscovery) CloseOutput() {
	close(s.outputChan)
}

// Hello initializes the SerialDiscovery by setting the protocol version and the user agent
func (s *SerialDiscovery) Hello(args []string) {
	if s.initialized {
		s.outputChan <- &genericMessageJSON{
			EventType: "hello",
			Error:     true,
			Message:   "HELLO already called",
		}
	}

	if len(args) < 2 {
		s.outputChan <- &genericMessageJSON{
			EventType: "hello",
			Error:     true,
			Message:   "invalid HELLO command",
		}
		return
	}

	protocolVersionStr := args[0]
	protocolVersion, err := strconv.ParseUint(protocolVersionStr, 10, 64)
	if err != nil {
		s.outputChan <- &genericMessageJSON{
			EventType: "hello",
			Error:     true,
			Message:   fmt.Sprintf("invalid protocol version: %s", protocolVersionStr),
		}
		return
	}

	if protocolVersion > maximumProtocolVersion {
		s.outputChan <- &genericMessageJSON{
			EventType: "hello",
			Error:     true,
			Message:   fmt.Sprintf("protocol version %d not supported, maximum version supported: %d", protocolVersion, maximumProtocolVersion),
		}
		return
	}

	s.protocolVersion = protocolVersion
	s.userAgent = strings.Join(args[1:], " ")
	s.initialized = true
	s.outputChan <- &helloMessageJSON{
		EventType:       "hello",
		ProtocolVersion: maximumProtocolVersion,
		Message:         "OK",
	}
}

// Start sets the Discovery state to started so LIST can be called.
// In the case of the SerialDiscovery there is nothing to do when starting other
// than set the state.
func (s *SerialDiscovery) Start() {
	if !s.initialized {
		s.outputChan <- &genericMessageJSON{
			EventType: "command_error",
			Error:     true,
			Message:   "discovery not initialized, please call HELLO to initialize it",
		}
		return
	}

	if s.started {
		s.outputChan <- &genericMessageJSON{
			EventType: "start",
			Error:     true,
			Message:   "already STARTed",
		}
		return
	} else if s.syncStarted {
		s.outputChan <- &genericMessageJSON{
			EventType: "start",
			Error:     true,
			Message:   "discovery already START_SYNCed, cannot START",
		}
		return
	}

	s.started = true
	s.outputChan <- &genericMessageJSON{
		EventType: "start",
		Message:   "OK",
	}
}

func (s *SerialDiscovery) Stop() {
	if !s.initialized {
		s.outputChan <- &genericMessageJSON{
			EventType: "command_error",
			Error:     true,
			Message:   "discovery not initialized, please call HELLO to initialize it",
		}
		return
	}
	if !s.syncStarted && !s.started {
		s.outputChan <- &genericMessageJSON{
			EventType: "stop",
			Error:     true,
			Message:   "already STOPped",
		}
		return
	}
	if s.started {
		s.started = false
	} else if s.syncStarted {
		s.syncInterrupt <- true
		if s.syncInterrupt != nil {
			close(s.syncInterrupt)
			s.syncInterrupt = nil
		}
		s.syncStarted = false
	}

	s.outputChan <- &genericMessageJSON{
		EventType: "stop",
		Message:   "OK",
	}
}

func (s *SerialDiscovery) List() {
	if !s.initialized {
		s.outputChan <- &genericMessageJSON{
			EventType: "command_error",
			Error:     true,
			Message:   "discovery not initialized, please call HELLO to initialize it",
		}
		return
	}

	if !s.started {
		s.outputChan <- &genericMessageJSON{
			EventType: "list",
			Error:     true,
			Message:   "discovery not STARTed",
		}
		return
	}
	if s.syncStarted {
		s.outputChan <- &genericMessageJSON{
			EventType: "list",
			Error:     true,
			Message:   "discovery already START_SYNCed, LIST not allowed",
		}
		return
	}

	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		s.outputChan <- &genericMessageJSON{
			EventType: "list",
			Error:     true,
			Message:   err.Error(),
		}
		return
	}

	portsJSON := []*boardPortJSON{}
	for _, port := range ports {
		portsJSON = append(portsJSON, newBoardPortJSON(port))
	}
	s.outputChan <- &listOutputJSON{
		EventType: "list",
		Ports:     portsJSON,
	}
}

func (s *SerialDiscovery) StartSync() {
	if !s.initialized {
		s.outputChan <- &genericMessageJSON{
			EventType: "command_error",
			Error:     true,
			Message:   "discovery not initialized, please call HELLO to initialize it",
		}
		return
	}

	if s.syncStarted {
		s.outputChan <- &genericMessageJSON{
			EventType: "start_sync",
			Error:     true,
			Message:   "discovery already START_SYNCed",
		}
		return
	}
	if s.started {
		s.outputChan <- &genericMessageJSON{
			EventType: "start_sync",
			Error:     true,
			Message:   "discovery already STARTed, cannot START_SYNC",
		}
		return
	}
	s.syncInterrupt = make(chan bool, 1)
	syncChan, err := sync(s.syncInterrupt)
	if err != nil {
		close(s.syncInterrupt)
		s.syncInterrupt = nil

		s.outputChan <- &genericMessageJSON{
			EventType: "start_sync",
			Error:     true,
			Message:   fmt.Sprintf("error START_SYNCing: %s", err),
		}
		return
	}

	s.syncStarted = true
	s.outputChan <- &genericMessageJSON{
		EventType: "start_sync",
		Message:   "OK",
	}

	go func() {
		for message := range syncChan {
			s.outputChan <- message
		}
	}()
}

// Quit closes the SerialDiscovery
func (s *SerialDiscovery) Quit() {
	s.outputChan <- &genericMessageJSON{
		EventType: "quit",
		Message:   "OK",
	}
	os.Exit(0)
}

// UnknownCommand prints a message telling the user the command is unknown
func (s *SerialDiscovery) UnknownCommand(command string) {
	s.outputChan <- &genericMessageJSON{
		EventType: "command_error",
		Error:     true,
		Message:   fmt.Sprintf("Command %s not supported", command),
	}
}
