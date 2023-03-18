#!/bin/bash
set -euo pipefail

src_dir="$(pwd)"

sudo chown "$(id -u):$(id -g)" /cache
chown -R "$(id -u):$(id -g)" "$HOME"
mkdir -p "$HOME/go/pkg"
ln -sd /cache "$HOME/go/pkg/mod"

cd "$src_dir"
arch="$(dpkg --print-architecture)"
debian-packaging/build.sh "$arch"
sudo dpkg -i debian-packaging/output/raspi-updater*.deb

export DEVICE=/tmp/updater.img

sudo mkdir /etc/raspi-updater
cat <<EOF | sudo tee /etc/raspi-updater/raspi-updater.conf 
ID=raspberry
NET_INTERFACE=ens33
NET_DRIVER=e1000
SERVER=localhost:12345
CERT_FILE=$src_dir/pkg/testdata/cert.pem
DEVICE=$DEVICE
COMPRESSION_TOOL=lz4
EOF

. /etc/raspi-updater/raspi-updater.conf 

raspi-updater-config  # update configuration

/usr/share/raspi-updater/raspi-updater \
    -address "$SERVER" \
    -certFile "$src_dir/pkg/testdata/cert.pem" \
    -keyFile "$src_dir/pkg/testdata/priv.key" \
    -images "$src_dir/test/images" \
    -updater "$src_dir/pkg/testdata/bin" \
    -verbose &
pid_server=$!

sleep 1

tmp_dir="$(mktemp -d)"
cd "$tmp_dir"


EXPECTED_IMAGE="$src_dir/test/images/raspberry_1.0.img.lz4"
ACTUAL_IMAGE="$(pwd)/$DEVICE"

# copy MBR only
mkdir tmp
lz4cat "$EXPECTED_IMAGE" | dd bs=512 of="$(pwd)/$DEVICE" count=1 || true
dd if=/dev/zero bs=$(((64*1024*1024) - 512)) count=1 >> "$(pwd)/$DEVICE" || true

if lz4cat "$EXPECTED_IMAGE" | diff "$EXPECTED_IMAGE" -; then
    echo "Error: images are identical before running test"
    exit 1
fi

lz4 -d -c /boot/initrd.img* | cpio -id
cp /.dockerenv "$(pwd)"

echo running raspi-update initramfs hook

mkdir dev
cp /dev/null dev
sudo chroot "$(pwd)" sh -x scripts/init-premount/raspi-updater 

if lz4cat "$EXPECTED_IMAGE" | diff "$ACTUAL_IMAGE" -; then
    echo "Test succesful"
else
    echo "Error: images are NOT identical"
    bash
fi