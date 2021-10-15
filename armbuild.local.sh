#!/usr/bin/env bash

source .gitlab-ci-variables

./version_hack.sh


# docker build multi-arch images and push to registry (--push)
echo "docker buildx build  --platform linux/amd64,linux/arm64 --pull --build-arg CodeSource=${CodeSource} --build-arg CodeBranches=${CodeBranches} --build-arg CodeVersion=${CodeVersion} -t ${BUILD_IMAGE}:${BUILD_VERSION}${BUILD_SUFFIX} --build-arg ssh_prv_key="$(cat ~/.ssh/id_rsa)" --build-arg ssh_pub_key="$(cat ~/.ssh/id_rsa.pub)" ./ --push"
docker buildx build  --platform linux/amd64,linux/arm64 --pull   -t ${BUILD_IMAGE}:${BUILD_VERSION}${BUILD_SUFFIX}  ./  --push
echo "docker buildx imagetools inspect ${BUILD_IMAGE}:${BUILD_VERSION}${BUILD_SUFFIX}"
docker buildx imagetools inspect ${BUILD_IMAGE}:${BUILD_VERSION}${BUILD_SUFFIX}

echo "echo docker push ${BUILD_IMAGE}:${BUILD_VERSION}${BUILD_SUFFIX}"
