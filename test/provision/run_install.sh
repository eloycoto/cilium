#!/bin/bash

HOST=$(hostname)
CILIUM_CONFIG_DIR="/opt/cilium"
sudo apt-get -y install llvm

/tmp/provision/compile.sh
