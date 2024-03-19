#!/bin/bash -x

pwd

if [[ "$1" == "-h" ]] || [[ "$1" == "--help" ]]; then
    echo -e "Usage: $0 [ARCH|-h|--help]"
    echo -e "\tARCH: amd64 (default), arm, arm64"
    exit
fi

ARCH=${1:-"amd64"}

VERSION=$(grep -e 'version.*=' main.go | awk '{print $3}' | tr -d '"')

CGO_ENABLED=0 GOOS=linux GOARCH="$ARCH" go build -o "BUILD/enervent-ctrl-${VERSION}.linux-$ARCH" .
