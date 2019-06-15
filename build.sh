#!/usr/bin/env bash

set -ex

godir=/tmp/go/src/github.com

mkdir -p $godir/box-node-alert-worker
export GOPATH=/tmp/go

mv cmd $godir/box-node-alert-worker/
mv options $godir/box-node-alert-worker/
mv pkg $godir/box-node-alert-worker/
mv workerpb $godir/box-node-alert-worker/

cd $godir/box-node-alert-worker
go get ./...
ls -la
mkdir bin
go get github.com/sirupsen/logrus
CGO_ENABLED=0 GOOS=linux go build -o bin/node-alert-worker -ldflags '-w' cmd/node-alert-worker.go

mkdir -p /git-root/build
mv bin/node-alert-worker /git-root/build