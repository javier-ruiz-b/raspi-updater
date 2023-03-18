#!/bin/bash
set -euxo pipefail

cd "$(dirname "$(realpath "$0")")/debian-packaging"


arch="$(dpkg --print-architecture)"
./build.sh "$arch"
sudo dpkg -i output/raspi-*_amd64.deb
sudo update-initramfs -u -k "$(uname -r)"
sudo update-grub
sync