#!/bin/bash
set -e

source /tmp/venv/bin/activate

init_setup_path="${find . -iname "initial_setup.ipynb}"

for i in $(find . -iname "*.ipynb" -maxdepth 1); do nbmerge $init_setup_path "$i";done