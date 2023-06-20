#!/bin/bash
set -e

# Internal vars
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
SCRIPT_ACTION="$1"
source $SCRIPT_DIR/util.sh

# Project vars
PROJECT_MAN="options: [create | destroy]"

function create_cluster(){
    echo "[*] Creating test cluster "${CLUSTER_NAME}" via kind"
    $KIND create cluster \
        --name "${CLUSTER_NAME}" \
        --config "${CONFIG_DIR}/cluster.yaml" \

    kubectl cluster-info --context "kind-${CLUSTER_NAME}"

    echo "[*] Cluster ${CLUSTER_NAME} configuration complete"
}

function destroy_cluster(){
    echo "[*] Destroying test cluster "${CLUSTER_NAME}" via kind"
    $KIND delete cluster --name "${CLUSTER_NAME}" 
}

case $SCRIPT_ACTION in
create)
    create_cluster
;;
destroy)
    destroy_cluster
;;
*)
	echo "$PROJECT_MAN"
	exit
;;
esac