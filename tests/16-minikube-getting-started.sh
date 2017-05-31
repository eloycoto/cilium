#!/bin/bash

set -xv

JOB_NAME=$1
BUILD_NUMBER=$2

ID="${JOB_NAME}${BUILD_NUMBER}"

source "./helpers.bash"

function cleanup {
	minikube delete -p $ID
}

trap cleanup exit

cleanup

echo "------ booting Minikube cluster with Container Network Interface network plugin enabled"
minikube start -p $ID --network-plugin=cni --iso-url https://github.com/cilium/minikube-iso/raw/master/minikube.iso

until [ "$(kubectl get cs --context $ID | grep -v "STATUS" | grep -c "Healthy")" -eq "3" ]; do 
	echo "---- Waiting for cluster to get into a good state ----"
	kubectl get cs --context $ID
	sleep 5
done

echo "----- deploying Cilium Daemon Set onto cluster -----"
kubectl create -f https://raw.githubusercontent.com/cilium/cilium/master/examples/minikube/cilium-ds.yaml

until [ "$(kubectl get ds --namespace kube-system --context $ID | grep -v 'READY' | awk '{ print $4}' | grep -c '1')" -eq "2" ]; do
	echo "----- Waiting for Cilium to get into 'ready' state in Minikube cluster -----"
	kubectl get ds --namespace kube-system --context $ID
	sleep 5
done

# Let Cilium spin up.
sleep 15

echo "----- deploying demo application onto cluster -----"
kubectl create -f https://raw.githubusercontent.com/cilium/cilium/master/examples/minikube/demo.yaml --context $ID

until [ "$(kubectl get pods --context $ID | grep -v STATUS | grep -c "Running")" -eq "4" ]; do
	echo "----- Waiting for demo apps to get into 'Running' state -----"
	sleep 5
done

echo "----- adding L3 L4 policy via kubectl -----"
kubectl create -f https://raw.githubusercontent.com/cilium/cilium/master/examples/minikube/l3_l4_policy.yaml --context $ID

CILIUM_POD=$(kubectl -n kube-system get pods -l k8s-app=cilium --context $ID | grep -v 'AGE' | awk '{ print $1 }')
kubectl --context $ID -n kube-system exec ${CILIUM_POD} cilium endpoint list | grep -c 'ready'
until [ "$(kubectl -n kube-system exec ${CILIUM_POD} cilium endpoint list --context $ID | grep -c 'ready')" -eq "4" ]; do
	echo "----- Waiting for endpoints to get into 'ready' state -----"
	sleep 5	      
done

echo "----- testing L3/L4 policy -----"
APP2_POD=$(kubectl get pods --context $ID -l id=app2 -o jsonpath='{.items[0].metadata.name}')
SVC_IP=$(kubectl get svc app1-service -o jsonpath='{.spec.clusterIP}' --context $ID )

echo "----- testing app2 can reach app1 (expected behavior: can reach) -----"
RETURN=$(kubectl --context $ID exec $APP2_POD -- curl -s --output /dev/stderr -w '%{http_code}' --connect-timeout 10 -XGET $SVC_IP)
if [[ "${RETURN//$'\n'}" != "200" ]]; then
	abort "Error: could not reach pod allowed by L3 L4 policy"
fi

echo "----- testing that app3 cannot reach app 1 (expected behavior: cannot reach)"
APP3_POD=$(kubectl get pods -l id=app3 -o jsonpath='{.items[0].metadata.name}' --context $ID)
RETURN=$(kubectl --context $ID exec $APP3_POD -- curl -s --output /dev/stderr -w '%{http_code}' --connect-timeout 10 -XGET $SVC_IP)
if [[ "${RETURN//$'\n'}" != "000" ]]; then
	abort "Error: unexpectedly reached pod allowed by L3 L4 Policy, received return code ${RETURN}"
fi

echo "------ performing HTTP GET on ${SVC_IP}/public from service2 ------"
RETURN=$(kubectl --context $ID exec $APP2_POD -- curl -s --output /dev/stderr -w '%{http_code}' --connect-timeout 10 http://${SVC_IP}/public)
if [[ "${RETURN//$'\n'}" != "200" ]]; then
	  abort "Error: Could not reach ${SVC_IP}/public on port 80"
fi

echo "------ performing HTTP GET on ${SVC_IP}/private from service2 ------"
RETURN=$(kubectl --context $ID exec $APP2_POD -- curl -s --output /dev/stderr -w '%{http_code}' --connect-timeout 10 http://${SVC_IP}/private)
if [[ "${RETURN//$'\n'}" != "200" ]]; then
	abort "Error: Could not reach ${SVC_IP}/public on port 80"
fi

echo "----- creating L7-aware policy -----"
kubectl create -f https://raw.githubusercontent.com/cilium/cilium/master/examples/minikube/l3_l4_l7_policy.yaml

#TODO - add check for all endpoints to be in ready state?
CILIUM_POD=$(kubectl --context $ID -n kube-system get pods -l k8s-app=cilium | grep -v 'AGE' | awk '{ print $1 }')
kubectl -n kube-system exec ${CILIUM_POD} cilium endpoint list | grep -c 'ready'
until [ "$(kubectl --context $ID -n kube-system exec ${CILIUM_POD} cilium endpoint list | grep -c 'ready')" -eq "4" ]; do
       echo "----- Waiting for endpoints to get into 'ready' state -----"
       sleep 5
done

echo "------ performing HTTP GET on ${SVC_IP}/public from service2 ------"
RETURN=$(kubectl --context $ID exec $APP2_POD -- curl -s --output /dev/stderr -w '%{http_code}' --connect-timeout 10 http://${SVC_IP}/public)
if [[ "${RETURN//$'\n'}" != "200" ]]; then
	abort "Error: Could not reach ${SVC_IP}/public on port 80"
fi

echo "------ performing HTTP GET on ${SVC_IP}/private from service2 ------"
RETURN=$(kubectl --context $ID exec $APP2_POD -- curl -s --output /dev/stderr -w '%{http_code}' --connect-timeout 10 http://${SVC_IP}/private)
if [[ "${RETURN//$'\n'}" != "403" ]]; then
	abort "Error: Unexpected success reaching  ${SVC_IP}/public on port 80"
fi

echo "------ L7 policy success ! ------"
