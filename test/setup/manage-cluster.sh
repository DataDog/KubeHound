#!/bin/bash
set -e

# Internal vars
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
SCRIPT_ACTION="$1"
source $SCRIPT_DIR/util.sh

# Project vars
PROJECT_MAN="options: [create | destroy]"

function create_cluster(){
    _printf_ok "Creating test cluster "${CLUSTER_NAME}" via kind"
    $KIND_CMD create cluster \
        --name "${CLUSTER_NAME}" \
        --config "${SCRIPT_DIR}/${CONFIG_DIR}/cluster.yaml" \

    _printf_warn "Using KUBECONFIG: $(printenv KUBECONFIG)"
    kubectl cluster-info --context "kind-${CLUSTER_NAME}"

    dump_config_file
    echo "[*] Cluster ${CLUSTER_NAME} configuration complete"
}

function destroy_cluster(){
    _printf_ok "Destroying test cluster "${CLUSTER_NAME}" via kind"
    $KIND_CMD delete cluster --name "${CLUSTER_NAME}" 
}

function remove_config_files(){
    _printf_ok "Removing config files for kind cluster "${CLUSTER_NAME}""
    rm -f ${SCRIPT_DIR}/${KIND_KUBECONFIG}
    rm -f  ${SCRIPT_DIR}/${KIND_KUBECONFIG_INTERNAL}
}

function dump_config_file(){
    _printf_ok "Dump kind cluster "${CLUSTER_NAME}" via kind for Docker env"
    $KIND_CMD get kubeconfig --name "${CLUSTER_NAME}" > ${SCRIPT_DIR}/${KIND_KUBECONFIG}
    $KIND_CMD get kubeconfig --internal --name "${CLUSTER_NAME}" > ${SCRIPT_DIR}/${KIND_KUBECONFIG_INTERNAL}
}

case $SCRIPT_ACTION in
create)
    create_cluster
;;
destroy)
    destroy_cluster
    remove_config_files
;;
*)
	echo "$PROJECT_MAN"
	exit
;;
esac