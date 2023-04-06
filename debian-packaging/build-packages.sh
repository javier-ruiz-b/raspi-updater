#!/bin/bash
set -euxo pipefail
cd "$(dirname "$(realpath "$0")")"

# https://www.internalpointers.com/post/build-binary-deb-package-practical-guide
VERSION=$(cat ./version.txt)
output_dir="$(pwd)/output"
mkdir -p "$output_dir" "$GOPATH"
rm -rf "${output_dir:?}/"*

tmpdir=$(mktemp -d)
trap 'rm -rf $tmpdir' EXIT

archs=(amd64 armhf arm64)

if [ "$*" != "" ]; then
    archs=("$@")
fi

for arch in "${archs[@]}"; do
    package_name="raspi-updater_${VERSION}_${arch}"
    package_dir="$tmpdir/${package_name}"
    mkdir "$package_dir"

    cp -rf rootfs/* DEBIAN/ "$package_dir"

    output_bin="$package_dir/usr/share/raspi-updater/raspi-updater"
    export GOOS=linux
    export CGO_ENABLED=0
    case $arch in
    amd64)  
        export GOARCH=amd64
        export CC="" 
        ;;
    armhf)  
        export GOARCH=arm
        export GOARM=6
        export CC=arm-linux-gnueabihf-gcc
        ;;
    arm64)  
        export GOARCH=arm64
        export CC=aarch64-linux-gnu-gcc 
        ;; 
    *)      echo "Architecture $arch unknown"; exit 1 ;;
    esac

    go build -ldflags "-s -w -linkmode external -extldflags '-static'" -o "$output_bin" ../cmd/updater/*.go 

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