#!/usr/bin/env bash

source .gitlab-ci-variables

./version_hack.sh

echo "docker build -t ${BUILD_IMAGE}:${BUILD_VERSION}-SNAPSHOT ."
docker build -t "${BUILD_IMAGE}:${BUILD_VERSION}-SNAPSHOT" \
    --build-arg ssh_prv_key="$(cat ~/.ssh/id_rsa)" \
    --build-arg ssh_pub_key="$(cat ~/.ssh/id_rsa.pub)" .

echo "docker push ${BUILD_IMAGE}:${BUILD_VERSION}-SNAPSHOT"
docker push "${BUILD_IMAGE}:${BUILD_VERSION}-SNAPSHOT"