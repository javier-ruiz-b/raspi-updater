#!/bin/bash
set -euo pipefail

GOOS=linux GOARCH=arm GOARM=6   go build -ldflags '-w -extldflags "-static"' -o updater-armhf *.go &
GOOS=linux GOARCH=arm64         go build -ldflags '-w -extldflags "-static"' -o updater-arm64 *.go &
GOOS=windows                    go build -ldflags '-w -extldflags "-static"' -o updater-win   *.go &

wait