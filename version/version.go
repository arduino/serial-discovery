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

// Package version provides information about the version of the application.
// It includes the version string, commit hash, and timestamp.
// The package also defines a struct `Info` that represents the version information.
// The `newInfo` function creates a new `Info` instance with the provided application name.
// The `String` method of the `Info` struct returns a formatted string representation of the version information.
// The package also initializes the `Version` variable with a default version string if it is empty.
package version

import (
	"fmt"
	"os"
	"path/filepath"
)

// VersionInfo FIXMEDOC
var VersionInfo = newInfo(filepath.Base(os.Args[0]))

var (
	defaultVersionString = "0.0.0-git"
	// Version FIXMEDOC
	Version = ""
	// Commit FIXMEDOC
	Commit = ""
	// Timestamp FIXMEDOC
	Timestamp = ""
)

// Info FIXMEDOC
type Info struct {
	Application   string `json:"Application"`
	VersionString string `json:"VersionString"`
	Commit        string `json:"Commit"`
	Date          string `json:"Date"`
}

// NewInfo FIXMEDOC
func newInfo(application string) *Info {
	return &Info{
		Application:   application,
		VersionString: Version,
		Commit:        Commit,
		Date:          Timestamp,
	}
}

func (i *Info) String() string {
	return fmt.Sprintf("%s Version: %s Commit: %s Date: %s", i.Application, i.VersionString, i.Commit, i.Date)
}

//nolint:gochecknoinits
func init() {
	if Version == "" {
		Version = defaultVersionString
	}
}
