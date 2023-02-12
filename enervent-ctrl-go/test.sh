#!/bin/bash -e

SRCDIR="$PWD"
TMPTAR=/tmp/ec.tar
TESTDIR=/tmp/enervent-ctr-build

rm -rf $TMPTAR $TESTDIR
tar chf $TMPTAR *
mkdir -p $TESTDIR
pushd $TESTDIR
tar xf $TMPTAR
pushd pingvinKL
go test -v .
popd
popd
rm -rf $TMPTAR
