#!/bin/bash
set -x
CID=$(docker run -tid -v $(pwd):/go/src/github.com/moby/buildkit/ -w /go/src/github.com/moby/buildkit/ golang:1.25 go test -run=^$ $*)
docker logs -f $CID
