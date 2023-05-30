#!/bin/bash

protoc --go_out=plugins=grpc:. cwl.go.proto
