#!/bin/bash
set -eu

cd "$(dirname "$(realpath "$0")")"/..

docker/run.sh debian-packaging/build-packages.sh "$@"