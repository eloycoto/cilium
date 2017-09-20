#!/bin/bash

HOST=$(hostname)
TOKEN="258062.5d84c017c9b2796c"
CILIUM_CONFIG_DIR="/opt/cilium"
ETCD_VERSION="v3.1.0"
NODE=$1
IP=$2
K8S_VERSION=$3
DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

apt-get update
apt-get install -y curl jq apt-transport-https htop bmon

cat <<EOF > /etc/hosts
127.0.0.1       localhost
::1     localhost ip6-localhost ip6-loopback
ff02::1 ip6-allnodes
ff02::2 ip6-allrouters
$IP     $NODE
EOF


cat <<EOF > /etc/apt/sources.list.d/kubernetes.list
deb http://apt.kubernetes.io/ kubernetes-xenial main
EOF

curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add -

curl -sSL https://get.docker.com/ | sh
systemctl start docker

apt-get install --allow-downgrades -y \
    kubelet="${K8S_VERSION}*" \
    kubeadm="${K8S_VERSION}*" \
    kubectl="${K8S_VERSION}*" \
    kubernetes-cni htop bmon

sudo mkdir -p ${CILIUM_CONFIG_DIR}

function install_etcd(){
    wget -nv https://github.com/coreos/etcd/releases/download/${ETCD_VERSION}/etcd-${ETCD_VERSION}-linux-amd64.tar.gz
    tar -xvf etcd-${ETCD_VERSION}-linux-amd64.tar.gz
    sudo mv etcd-${ETCD_VERSION}-linux-amd64/etcd* /usr/bin/

    sudo tee /etc/systemd/system/etcd.service <<EOF
[Unit]
Description=etcd
Documentation=https://github.com/coreos

[Service]
ExecStart=/usr/bin/etcd --name=cilium --data-dir=/var/etcd/cilium --advertise-client-urls=http://192.168.36.11:9732 --listen-client-urls=http://0.0.0.0:9732 --listen-peer-urls=http://0.0.0.0:9733
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF
    sudo systemctl enable etcd
    sudo systemctl start etcd
}

sudo mount bpffs /sys/fs/bpf -t bpf

if [[ "${HOST}" == "k8s1" ]]; then
    # FIXME: IP needs to be dynamic
    kubeadm init --token=$TOKEN --apiserver-advertise-address="192.168.36.11" --pod-network-cidr=10.10.0.0/16

    mkdir -p /root/.kube
    sudo cp -i /etc/kubernetes/admin.conf /root/.kube/config
    sudo chown root:root /root/.kube/config

    sudo -u vagrant mkdir -p /home/vagrant/.kube
    sudo cp -i /etc/kubernetes/admin.conf /home/vagrant/.kube/config
    sudo chown vagrant:vagrant /home/vagrant/.kube/config

    sudo cp /etc/kubernetes/admin.conf ${CILIUM_CONFIG_DIR}/kubeconfig
    kubectl taint nodes --all node-role.kubernetes.io/master-

    install_etcd
else
    kubeadm join --token=$TOKEN 192.168.36.11:6443
    cp /etc/kubernetes/kubelet.conf ${CILIUM_CONFIG_DIR}/kubeconfig
fi

# install_etcd
