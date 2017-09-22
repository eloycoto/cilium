#!/bin/bash

export GOPATH="/go/"

mkdir -p $GOPATH/src/github.com/cilium/
rm -rf $GOPATH/src/github.com/cilium/cilium
cp -rf /src $GOPATH/src/github.com/cilium/cilium

cd $GOPATH/src/github.com/cilium/cilium

if echo $(hostname) | grep "k8s" -q;
then
    if [[ "$(hostname)" == "k8s1" ]]; then
        make docker-image-dev
        docker tag cilium 192.168.36.11:5000/cilium/cilium-dev
        docker push 192.168.36.11:5000/cilium/cilium-dev
    else
        echo "No master, no need to compile"
    fi
else
    make install
    #FIXME: Need to restart here?
fi
