#!/bin/bash

AUTHOR=nscc-gz.cn/starlight-v4
IMAGE=cwl-runner
VERSION=0.8.0

set -e && set -x

cd proto && bash ./generate.sh && cd ..

go mod tidy
CGO_ENABLED=0 go build -o cwl.go
goupx cwl.go

docker build -t ${AUTHOR}/${IMAGE}:${VERSION} .
docker save ${AUTHOR}/${IMAGE}:${VERSION} -o DockerImage_${AUTHOR}_${IMAGE}_${VERSION}.tar

rm cwl.go