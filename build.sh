#!/bin/bash

# Build for every architecture and emit json and shasums

rm -rf distrib

export CGO_ENABLED=0
GOOS=linux GOARCH=amd64 go build -o distrib/linux64/bin/serial-discovery
GOOS=linux GOARCH=386 go build -o distrib/linux32/bin/serial-discovery
GOOS=linux GOARCH=arm go build -o distrib/linuxarm/bin/serial-discovery
GOOS=linux GOARCH=arm64 go build -o distrib/linuxarm64/bin/serial-discovery
GOOS=windows GOARCH=386 GO386=387 go build -o distrib/windows/bin/serial-discovery.exe
CGO_ENABLED=1 CC=o64-clang GOOS=darwin GOARCH=amd64 go build -ldflags="-extldflags=-mmacosx-version-min=10.9" -o distrib/darwin/bin/serial-discovery

VERSION=`git describe --tag`
cd distrib
cd windows && zip -r ../../serial-discovery-windows-$VERSION.zip * && cd ..
cd osx && tar cjf ../../serial-discovery-macosx-$VERSION.tar.bz2 * && cd ..
cd linuxarm && tar cjf ../../serial-discovery-linuxarm-$VERSION.tar.bz2 * && cd ..
cd linuxarm64 && tar cjf ../../serial-discovery-linuxarm64-$VERSION.tar.bz2 * && cd ..
cd linux32 && tar cjf ../../serial-discovery-linux32-$VERSION.tar.bz2 * && cd ..
cd linux64 && tar cjf ../../serial-discovery-linux64-$VERSION.tar.bz2 * && cd ..
cd ..

shasum serial-discovery*-${VERSION}.*
sha256sum serial-discovery*-${VERSION}.*
ls -la serial-discovery*-${VERSION}.*
