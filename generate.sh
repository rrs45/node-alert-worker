#!/bin/bash

protoc workerpb/worker.proto --go_out=plugins=grpc:.