#!/bin/bash
set -eu

cd "$(dirname "$(realpath "$0")")"/..

arch="$(dpkg --print-architecture)"
debian-packaging/run.sh "$arch"

docker/run.sh acceptance_test/acceptance-test.sh