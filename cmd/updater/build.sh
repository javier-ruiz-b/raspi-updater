#!/bin/bash
set -euo pipefail

cd "$(dirname "$(realpath "$0")")"

GOOS=linux GOARCH=arm GOARM=6   go build -ldflags '-w -extldflags "-static"' -o linux-armhf *.go &
GOOS=linux GOARCH=arm64         go build -ldflags '-w -extldflags "-static"' -o linux-arm64 *.go &
GOOS=windows GOARCH=amd64       go build -ldflags '-w -extldflags "-static"' -o windows-amd64   *.go &
GOOS=linux GOARCH=amd64         go build -ldflags '-w -extldflags "-static"' -o linux-amd64   *.go &

wait