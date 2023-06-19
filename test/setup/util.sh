#!/bin/bash
set -e

###### SCRIPT CONF
BLUE='\033[0;34m'
RED='\033[0;31m'
GREEN='\033[0;32m'
ORANGE='\033[0;33m'
NC='\033[0m' # No Color

function _printf_question(){
	printf "$ORANGE[?] $1$NC"
}

function _printf_err(){
	printf "$RED[!] $1$NC\n"
}

function _printf_ok(){
	printf "$GREEN[+] $1$NC\n"
}

function _printf_warn(){
	printf "$ORANGE[-] $1$NC\n"
}

function load_env(){
    _printf_warn "Loading env vars from $SCRIPT_DIR/.env.local ..."
    if [ -f $SCRIPT_DIR/.env.local ]; then
        set -a
        source $SCRIPT_DIR/.env 
        set +a
    fi
}

load_env

# post load env
KIND=kind
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    KIND="sudo kind"
fi

KIND="$KIND --kubeconfig $KUBECONFIG"
if [ -f $KUBECONFIG ]; then
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        sudo chown $USER:$USER $KUBECONFIG
    fi
fi
echo "Using KUBECONFIG: $(printenv KUBECONFIG)"