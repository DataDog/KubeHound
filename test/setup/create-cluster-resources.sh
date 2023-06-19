#!/bin/bash
set -e

CLUSTER_NAME=kubehound.test.local
CONFIG_DIR=./test-cluster
# export KUBECONFIG=.kube/config

echo "[*] Deploying test resources via kubectl apply"
for attack in ${CONFIG_DIR}/attacks/*.yaml; do
    [ -e "$attack" ] || continue
    
    kubectl apply -f "$attack" --context "kind-${CLUSTER_NAME}"
done

echo "[*] Deployments complete"