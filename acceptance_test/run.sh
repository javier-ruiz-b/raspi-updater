#!/bin/bash
set -euo pipefail

# Define the image name and tag
IMAGE_NAME="raspi-updater-test"
IMAGE_TAG="latest"

# Change the working directory to the directory where the script is located
cd "$(dirname "$0")"

# Build the Dockerfile in the current directory
docker build \
    -t "${IMAGE_NAME}:${IMAGE_TAG}" \
    --build-arg USER_ID="$(id -u)" \
    --build-arg GROUP_ID="$(id -g)" \
    .

cd ..

# Run the container with the current directory mounted as a volume
# SYS_ADMIN required to mount /dev
    # --cap-add SYS_ADMIN \
    # --cap-add SYSLOG \
docker run --rm -it \
    -u "$(id -u):$(id -g)" \
    -v "$(pwd):$(pwd)" \
    -w "$(pwd)" \
    -v ${IMAGE_NAME}-cache:/cache \
    "${IMAGE_NAME}:${IMAGE_TAG}"
