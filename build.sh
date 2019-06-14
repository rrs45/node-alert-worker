#!/usr/bin/env bash

set -ex

godir=/tmp/go/src/github.com

mkdir -p $godir/box-node-alert-worker
export GOPATH=/tmp/go

mv cmd $godir/box-node-alert-worker/
mv pkg $godir/box-node-alert-worker/
cd $godir/box-node-alert-worker
mkdir bin
go get ./...
go test -v ./pkg/...
CGO_ENABLED=1 GOOS=linux go build -o bin/node-alert-worker cmd/node-alert-worker.go

mkdir -p /git-root/build
mv bin/node-alert-worker /git-root/build