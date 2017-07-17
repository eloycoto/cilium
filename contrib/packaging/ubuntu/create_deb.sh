#!/bin/bash

set -e
set -x

BUILD_DIR="/go/src/github.com/cilium/cilium"
BASE_DIR="/opt/cilium/"
SYSCONFIG_DIR="${BUILD_DIR}/contrib/systemd/"
export VERSION=$(cat ${BASE_DIR}/cilium/VERSION)
echo $VERSION

mkdir -p ${BUILD_DIR}
mv ${BASE_DIR}/cilium/ $(dirname ${BUILD_DIR})
cp -R ${BASE_DIR}/debian ${BUILD_DIR}
#envsubst \\\$VERSION < "${BASE_DIR}/debian/control" > "${BUILD_DIR}/debian/control"
cd ${BUILD_DIR}

git checkout -b master
gbp dch  --auto --release --git-author --new-version=$VERSION

debuild -e GOPATH -e GOROOT -e PATH -us -uc -b
cp ../cilium_* /output/
