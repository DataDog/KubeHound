#!/usr/bin/env bash

echo "cleaning previous artefacts"
rm -rf test-cluster archive.tar.gz
mkdir -p test-cluster

echo "copying testdata from the ingestor package"
cp ../../../kubehound/ingestor/pipeline/testdata/{pod,node,role,rolebinding}.json ./test-cluster

echo "Bunlding the testdata into a tarball compressed by gzip"
tar -czvf archive.tar.gz test-cluster/
rm -rf test-cluster