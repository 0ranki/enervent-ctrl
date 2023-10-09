#!/bin/bash

VERSION=$(grep -e 'version.*=' main.go | awk '{print $3}' | tr -d '"')

pushd TMP &> /dev/null || exit 1

rm -rf *
tar --exclude ../TMP -ch ../* | tar xf -

env GOOS=linux GOARCH=arm go build -o ../BUILD/enervent-ctrl-${VERSION}.linux-arm32 .

popd &> /dev/null