#!/bin/bash
set -euo pipefail

src_dir="$(pwd)"

sudo chown "$(id -u):$(id -g)" /cache
chown -R "$(id -u):$(id -g)" "$HOME"
mkdir -p "$HOME/go/pkg"
ln -sd /cache "$HOME/go/pkg/mod"

sudo mkdir /etc/raspi-updater

export DEVICE=/tmp/updater.img
cat <<EOF | sudo tee /etc/raspi-updater/raspi-updater.conf 
ID=raspberry
NET_INTERFACE=ens33
NET_DRIVER=e1000
SERVER=localhost:12345
CERT_FILE=$src_dir/pkg/testdata/cert.pem
DEVICE=$DEVICE
EOF

. /etc/raspi-updater/raspi-updater.conf 

cd "$src_dir"
sudo dpkg -i debian-packaging/output/raspi-updater*.deb

raspi-updater-config  # update configuration

(   
    set -x
    /usr/share/raspi-updater/raspi-updater \
        -address "$SERVER" \
        -certFile "$src_dir/pkg/testdata/cert.pem" \
        -keyFile "$src_dir/pkg/testdata/priv.key" \
        -images "$src_dir/test/images" \
        -updater "$src_dir/pkg/testdata/bin" \
        -verbose

    echo "ERROR: raspi-updater terminated"
    kill $$ || kill -9 $$
)   &
    
pid_server=$!
trap 'kill $pid_server' EXIT

sleep 1

tmp_dir="$(mktemp -d)"
cd "$tmp_dir"

EXPECTED_VERSION="1.0"
EXPECTED_IMAGE="$src_dir/test/images/raspberry_$EXPECTED_VERSION.img.lz4"
ACTUAL_IMAGE="$(pwd)/$DEVICE"

# copy MBR only
mkdir tmp
lz4cat "$EXPECTED_IMAGE" | dd bs=512 of="$(pwd)/$DEVICE" count=1 || true
dd if=/dev/zero bs=$(((64*1024*1024) - 512)) count=1 >> "$(pwd)/$DEVICE" || true

lz4 -d -c /boot/initrd.img* | cpio -id
cp /.dockerenv "$(pwd)"

echo running raspi-update initramfs hook

mkdir dev boot
cp /dev/null dev
sudo mv /boot/raspi-updater ./boot
sudo chroot "$(pwd)" sh -x scripts/init-premount/raspi-updater 

lz4 -c "$EXPECTED_IMAGE" > /tmp/expected.img

mkdir "$tmp_dir"/expected
cd  "$tmp_dir"/expected
7z x /tmp/expected.img

mkdir "$tmp_dir"/actual
cd  "$tmp_dir"/actual
7z x "$ACTUAL_IMAGE"

if ! diff "$tmp_dir"/actual/1.img "$tmp_dir"/expected/1.img; then
    echo "EXT4 partition differs:"
    echo "Expected:"
    7z l "$tmp_dir"/expected/1.img
    echo "Actual:"
    7z l "$tmp_dir"/actual/1.img
    exit 1
fi

cd "$tmp_dir"/actual
7z x 0.fat
cd "$tmp_dir"/expected
7z x 0.fat
cd ..
rm */0.fat */1.img

if [ "$EXPECTED_VERSION" != "$(cat actual/version)" ]; then
    echo "Version differs:"
    echo "  expected: $EXPECTED_VERSION"
    echo "  actual: $(cat actual/version)"
    exit 1
fi
rm actual/version

if ! diff -f actual/ expected/; then
    echo "FAT partition differs."
    exit 1
fi

echo "Acceptance test succeeded! Images contain same data"