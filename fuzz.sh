#!/bin/bash
set -x
docker run -tid -v $(pwd):/go/src/github.com/moby/buildkit/ -w /go/src/github.com/moby/buildkit/ golang:1.25 go test $*
