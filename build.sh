#!/usr/bin/env bash

set -ex
ls -la

godir=/tmp/go/src/github.com

mkdir -p $godir/box-node-alert-worker
export GOPATH=/tmp/go

mv cmd $godir/box-node-alert-worker/
mv options $godir/box-node-alert-worker/
mv pkg $godir/box-node-alert-worker/
mv workerpb $godir/box-node-alert-worker/
mv go.mod $godir/box-node-alert-worker/

#go get k8s.io/klog && cd $GOPATH/src/k8s.io/klog && git checkout v0.4.0
cd $godir/box-node-alert-worker
export GO111MODULE=on
go get ./...
ls -la
cat go.mod
mkdir bin
go get github.com/sirupsen/logrus
CGO_ENABLED=0 GOOS=linux go build -o bin/node-alert-worker -ldflags '-w' cmd/node-alert-worker.go

mkdir -p /git-root/build
mv bin/node-alert-worker /git-root/build