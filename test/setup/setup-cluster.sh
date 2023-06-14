#!/bin/bash
set -e

CLUSTER_NAME=kubehound.test.local
CONFIG_DIR=./test-cluster
export KUBECONFIG=.kube/config

echo "[*] Creating test cluster "${CLUSTER_NAME}" via kind"
kind create cluster \
    --name "${CLUSTER_NAME}" \
    --config "${CONFIG_DIR}/cluster.yaml" \

kubectl cluster-info --context "kind-${CLUSTER_NAME}"

echo "[*] Cluster ${CLUSTER_NAME} configuration complete"