#!/bin/bash
set -euo pipefail

# Define the image name and tag
IMAGE_NAME="raspi-updater-test"
IMAGE_TAG="latest"

# Build the Dockerfile in the current directory
docker build \
    -t "${IMAGE_NAME}:${IMAGE_TAG}" \
    --build-arg USER_ID="$(id -u)" \
    --build-arg GROUP_ID="$(id -g)" \
    "$(dirname "$0")"

# Run the container with the current directory mounted as a volume
docker run --rm -it \
    -u "$(id -u):$(id -g)" \
    -v "$(pwd):$(pwd)" \
    -w "$(pwd)" \
    -v ${IMAGE_NAME}-cache:/cache \
    -e GOPATH=/cache/go \
    "${IMAGE_NAME}:${IMAGE_TAG}" "$@"
