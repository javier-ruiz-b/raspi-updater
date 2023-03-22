#!/bin/bash
set -eu

cd "$(dirname "$(realpath "$0")")"
src_dir="$(pwd)"

PATH="$PATH:$(pwd)/tools_win"

mkdir -p "$src_dir/server_images"

cd "cmd/updater"

GOOS=windows GOARCH=amd64 go build -race -o windows-amd64 
./windows-amd64 \
    -address "0.0.0.0:31416" \
    -certFile "$src_dir/pkg/testdata/cert.pem" \
    -keyFile "$src_dir/pkg/testdata/priv.key" \
    -images "$src_dir/server_images" \
    -updater "$src_dir/cmd/updater" \
    -verbose \
    -log \
    "$@"