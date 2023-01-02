module github.com/arduino/serial-discovery

replace go.bug.st/serial => github.com/cmaglie/go-serial v0.0.0-20230102134456-e6cff1a986e7

require (
	github.com/arduino/go-properties-orderedmap v1.6.0
	github.com/arduino/pluggable-discovery-protocol-handler/v2 v2.0.2
	github.com/s-urbaniak/uevent v1.0.1
	go.bug.st/serial v1.3.5
	golang.org/x/sys v0.0.0-20220829200755-d48e67d00261
)

go 1.16
