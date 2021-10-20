#!/usr/bin/env bash

export APP_NAME=polarstack-daemon
export BUILD_VERSION=1.0.0
export BUILD_IMAGE=polardb/${APP_NAME}

./version_hack.sh

echo "docker build -t ${BUILD_IMAGE}:${BUILD_VERSION} ."
docker build -t "${BUILD_IMAGE}:${BUILD_VERSION}" .