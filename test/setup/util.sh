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
    _printf_warn "Loading env vars from $SCRIPT_DIR/.config ..."
    if [ -f $SCRIPT_DIR/.config ]; then
        set -a
        source $SCRIPT_DIR/.config 
        set +a
    fi
}

load_env

# post load env
# Set configuration for linux - https://docs.github.com/en/actions/learn-github-actions/variables
if [ -z $KIND_CMD ]; then 
    if [[ "$OSTYPE" == "linux-gnu"* && "$CI" != "true" ]]; then
        _printf_warn "sudo mode activated"
        KIND_CMD="sudo kind"
    else
        KIND_CMD="kind"
    fi
fi
