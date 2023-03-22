#!/bin/bash
set -eu

cd "$(dirname "$(realpath "$0")")"
src_dir="$(pwd)"

PATH="$PATH:$(pwd)/tools_win"

cd "cmd/updater"
GOOS=windows GOARCH=amd64 go build -race -o windows-amd64 
./windows-amd64 \
    -address "0.0.0.0:31416" \
    -certFile "$src_dir/server_images/local/lange.fritz.box.crt" \
    -keyFile "$src_dir/server_images/local/lange.fritz.box.key" \
    -images "$src_dir/server_images" \
    -updater "$src_dir/cmd/updater" \
    -verbose \
    "$@"