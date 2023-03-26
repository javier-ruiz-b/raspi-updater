#!/bin/bash
set -eu

cd "$(dirname "$(realpath "$0")")"
src_dir="$(pwd)"

PATH="$PATH:$(pwd)/tools_win"

cd "cmd/updater"
go build -race -o server 
./server \
    -address "0.0.0.0:31416" \
    -certFile "$src_dir/server_images/local/"*.crt \
    -keyFile "$src_dir/server_images/local/"*.key \
    -images "$src_dir/server_images" \
    -updater "$src_dir/cmd/updater" \
    -verbose \
    "$@"
