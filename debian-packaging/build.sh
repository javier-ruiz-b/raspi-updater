#!/bin/bash
set -euxo pipefail
cd "$(dirname "$(realpath "$0")")"

# https://www.internalpointers.com/post/build-binary-deb-package-practical-guide
VERSION=$(cat ../version.txt)
output_dir="$(pwd)/output"
mkdir -p "$output_dir"
rm -rf "${output_dir:?}/"*

tmpdir=$(mktemp -d)
trap 'rm -rf $tmpdir' EXIT

archs=(amd64 armhf arm64)

if [ "$*" != "" ]; then
    archs=("$@")
fi

for arch in "${archs[@]}"; do # TODO: armhf arm64
    package_name="raspi-updater_${VERSION}_${arch}"
    package_dir="$tmpdir/${package_name}"
    mkdir "$package_dir"

    cp -rf rootfs/* DEBIAN/ "$package_dir"

    output_bin="$package_dir/usr/share/raspi-updater/raspi-updater"
    case $arch in
    amd64)
        GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o "$output_bin" ../cmd/updater/*.go ;;
    armhf)
        GOOS=linux GOARCH=arm GOARM=6 go build -ldflags "-s -w" -o "$output_bin" ../cmd/updater/*.go ;;
    arm64)
        GOOS=linux GOARCH=arm64 go build -ldflags "-s -w" -o "$output_bin" ../cmd/updater/*.go ;; 
    esac
    
    cd "$tmpdir"
    chmod +x "$package_dir/usr/share/initramfs-tools/hooks"/* \
             "$package_dir/usr/share/initramfs-tools/scripts"/*/* \
             "$package_dir/usr/local/bin"/*
    sed -i "s/%version%/$VERSION/g" "$package_dir/DEBIAN/control"
    sed -i "s/%arch%/$arch/g" "$package_dir/DEBIAN/control"
    dpkg-deb -Znone --build --root-owner-group "$package_name"
    mv *.deb "$output_dir"
    rm -rf "${tmpdir:?}/"*
    cd -
done