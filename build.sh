#!/bin/bash

# Build for every architecture and emit json and shasums

rm -rf distrib

export CGO_ENABLED=0
GOOS=linux GOARCH=amd64 go build -o distrib/linux64/bin/serial-discovery
GOOS=linux GOARCH=386 go build -o distrib/linux32/bin/serial-discovery
GOOS=linux GOARCH=arm go build -o distrib/linuxarm/bin/serial-discovery
GOOS=linux GOARCH=arm64 go build -o distrib/linuxarm64/bin/serial-discovery
CGO_ENABLED=1 CC=o64-clang GOOS=darwin GOARCH=amd64 go build -ldflags="-extldflags=-mmacosx-version-min=10.9" -o distrib/macosx/bin/serial-discovery
GOOS=windows GOARCH=386 GO386=387 go build -o distrib/windows32/bin/serial-discovery.exe
GOOS=windows GOARCH=amd64 GO386=387 go build -o distrib/windows64/bin/serial-discovery.exe

VERSION=`git describe --tag`
cd distrib
cd windows32 && zip -r ../../serial-discovery_${VERSION}_Windows_32bit.zip * && cd ..
cd windows64 && zip -r ../../serial-discovery_${VERSION}_Windows_64bit.zip * && cd ..
cd macosx && tar cjf ../../serial-discovery_${VERSION}_macOS_64bit.tar.bz2 * && cd ..
cd linuxarm && tar cjf ../../serial-discovery_${VERSION}_Linux_ARM.tar.bz2 * && cd ..
cd linuxarm64 && tar cjf ../../serial-discovery_${VERSION}_Linux_ARM64.tar.bz2 * && cd ..
cd linux32 && tar cjf ../../serial-discovery_${VERSION}_Linux_32bit.tar.bz2 * && cd ..
cd linux64 && tar cjf ../../serial-discovery_${VERSION}_Linux_64bit.tar.bz2 * && cd ..
cd ..

shasum -a 256 serial-discovery_${VERSION}*
ls -la serial-discovery_${VERSION}*
