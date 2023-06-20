#!/bin/bash
set -e

# Internal vars
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
SCRIPT_ACTION="$1"
source $SCRIPT_DIR/util.sh

# Project vars
PROJECT_MAN="options: [delete | deploy]"

function handle_resources(){
    _printf_warn "$2 test resources via kubectl apply"
    for attack in ${SCRIPT_DIR}/${CONFIG_DIR}/attacks/*.yaml; do
        [ -e "$attack" ] || continue
        _printf_ok "$attack"
        # since deletion can take some times, || true to be able to retry in case of C-C
        kubectl $1 -f "$attack" --context "kind-${CLUSTER_NAME}" || true
    done

    _printf_ok "Action complete"
}

case $SCRIPT_ACTION in
deploy)
    handle_resources "apply" "Deploying"
;;
delete)
    handle_resources  "delete" "Deleting"
;;
*)
	echo "$PROJECT_MAN"
	exit
;;
esac