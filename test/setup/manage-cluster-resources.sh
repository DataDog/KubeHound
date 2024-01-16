#!/bin/bash
set -e

# Internal vars
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
# This is needed to ensure that we are targetting the kind cluster and not another Kub cluster
export KUBECONFIG=${SCRIPT_DIR}/${KIND_KUBECONFIG}
SCRIPT_ACTION="$1"
source $SCRIPT_DIR/util.sh

# user | groups | namespaces
readonly users=(       
    'user1|alpha,beta,group-rb-r-rb-r|default'
    'user2|alpha,beta,group-rb-r-crb-cr-fail|default,dev'
    'user3|alpha,group-rb-r-rb-crb|default'
    'user-rb-r-crb-cr-fail|alpha|default'
    'user-rb-r-rb-crb|alpha|default'
    'user-rb-r-rb-r|alpha|default'
)

namespaces=(
    'default'
    'vault'
    'dev'
)

# Project vars
PROJECT_MAN="options: [delete | deploy]"

function handle_resources(){
    for namespace in ${namespaces[@]}
    do
         create_namespace ${namespace}
    done

    _printf_warn "$2 test resources via kubectl apply"
    for attack in ${SCRIPT_DIR}/${CONFIG_DIR}/attacks/*.yaml; do
        [ -e "$attack" ] || continue
        _printf_ok "$attack"

        # since deletion can take some times, || true to be able to retry in case of C-C
        kubectl $1 -f "$attack" --context "kind-${CLUSTER_NAME}" || true
    done
}

function create_namespace(){
    local namespace=$1
    # Creating namespace only if not defined
    kubectl create namespace ${namespace} --dry-run=client -o yaml | kubectl apply -f -
}

function create_users(){
    local username groups namespaces
    _printf_warn "$2 test resources via kubectl apply"
    for fields in ${users[@]}
    do
            IFS=$'|' read -r username groups namespaces <<< "$fields"
            IFS=','; for namespace in $(echo "${namespaces}"); do
                create_namespace ${namespace}
            done
            create_user ${username} ${groups} ${namespaces}
    done
     _printf_ok "Users created"
}

function create_user(){
    local username=$1
    local groups=$2
    local namespace=$3
    
    local rbac_dir=${SCRIPT_DIR}/${CONFIG_DIR}/RBAC
    local rbac_user_dir=${rbac_dir}/${username}

    _printf_warn "Creating ${username} details in ${rbac_user_dir} ..."

    mkdir -p ${rbac_user_dir}
    
    # Generate key for user
    openssl genrsa -out ${rbac_user_dir}/info.key 2048
    
    #Generate csr for user
    IFS=',';csr_groups=$(for i in $(echo "${groups}"); do echo -n "/O=$i";done)
    openssl req -new -key ${rbac_user_dir}/info.key -subj "/CN=${username}${csr_groups}" -out ${rbac_user_dir}/info.csr
    
    # Dump CSR to base64 for certificate signing request 
    export CSR_CLIENT=$(cat ${rbac_user_dir}/info.csr | base64 -w 0)
    
    # Create CSR object file using a template
    cat ${SCRIPT_DIR}/${CONFIG_DIR}/CertificateSigningRequest.tpl.yaml | sed "s/<name>/${username}/ ; s/<csr-base64>/${CSR_CLIENT}/" > ${rbac_user_dir}/info_csr.yaml

    # Create CSR object
    kubectl create -f ${rbac_user_dir}/info_csr.yaml

    # Approve CSR 
    kubectl certificate approve ${username}
    
    #extracting client certificate
    kubectl get csr ${username} -o jsonpath='{.status.certificate}' | base64 --decode > ${rbac_user_dir}/info.crt
    
    # #CA extraction 
    kubectl config view --raw -o jsonpath='{..cluster.certificate-authority-data}' | base64 --decode >  ${rbac_user_dir}/info-ca.crt

    # Configure kubeconfig file for user
    local ca_crt=$(cat  ${rbac_user_dir}/info-ca.crt | base64 -w 0)
    local context=$(kubectl config current-context)
    local cluster_endpoint=$(kubectl config view -o jsonpath='{.clusters[?(@.name=="'"$context"'")].cluster.server}')
    local crt=$(cat  ${rbac_user_dir}/info.crt | base64 -w 0)
    local key=$(cat ${rbac_user_dir}/info.key | base64 -w 0)
    
    cat ${SCRIPT_DIR}/${CONFIG_DIR}/kubeconfig-template.yaml | sed "s#<context>#${CONTEXT}# ;
    s#<cluster-name>#${context}# ;
    s#<ca.crt>#${ca_crt}# ;
    s#<cluster-endpoint>#${cluster_endpoint}# ;
    s#<user-name>#${username}# ;
    s#<namespace>#${namespace}# ;
    s#<user.crt>#${crt}# ; 
    s#<user.key>#${key}#" > ${rbac_user_dir}/kubeconfig

}

case $SCRIPT_ACTION in
    deploy)
        handle_resources "apply" "Deploying"
        create_users "create" "Creating"
        ;;
    delete)
        handle_resources  "delete" "Deleting"
        create_users "delete" "Creating"
        ;;
    *)
        echo "$PROJECT_MAN"
        exit
        ;;
esac
_printf_ok "Action complete"