#!/bin/bash
set -e

CLUSTER_NAME=kubehound.test.local
export KUBECONFIG=.kube/config

echo "[*] Destroying test cluster "${CLUSTER_NAME}" via kind"
kind delete cluster --name "${CLUSTER_NAME}" 