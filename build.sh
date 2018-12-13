#!/bin/bash

# Build for every architecture and emit json and shasums

rm -rf distrib

# TODO: replace me with proper go modules stuff
export GOPATH=$PWD

export CGO_ENABLED=0
GOOS=linux GOARCH=amd64 go build -o distrib/linux64/serial-discovery
GOOS=linux GOARCH=386 go build -o distrib/linux32/serial-discovery
GOOS=linux GOARCH=arm go build -o distrib/linuxarm/serial-discovery
GOOS=linux GOARCH=arm64 go build -o distrib/linuxarm64/serial-discovery
GOOS=windows GOARCH=386 GO386=387 go build -o distrib/windows/serial-discovery.exe
CGO_ENABLED=1 CC=o64-clang GOOS=darwin GOARCH=amd64 go build -ldflags="-extldflags=-mmacosx-version-min=10.9" -o distrib/darwin/serial-discovery

cd distrib
zip -r ../serial-discovery-${VERSION}.zip *
cd ..

shasum serial-discovery-${VERSION}.zip
sha256sum serial-discovery-${VERSION}.zip
ls -la serial-discovery-${VERSION}.zip
