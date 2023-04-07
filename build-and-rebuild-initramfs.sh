#!/bin/bash
set -euxo pipefail

cd "$(dirname "$(realpath "$0")")"


arch="$(dpkg --print-architecture)"
debian-packaging/run.sh "$arch"
sudo dpkg -i debian-packaging/output/raspi-*_amd64.deb
sudo update-initramfs -u -k "$(uname -r)"
sudo update-grub
sync