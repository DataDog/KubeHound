#!/bin/bash
set -e

CLUSTER_NAME=kubehound.test.local
CONFIG_DIR=./test-cluster

echo "[*] Deploying test resources via kubectl apply"
kubectl apply -f "${CONFIG_DIR}/priv-pod.yaml" --context "kind-${CLUSTER_NAME}"
kubectl apply -f "${CONFIG_DIR}/priv-pid-pod.yaml" --context "kind-${CLUSTER_NAME}"
kubectl apply -f "${CONFIG_DIR}/hostpath-pod.yaml" --context "kind-${CLUSTER_NAME}"
