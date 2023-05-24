# System Testing

## Local Testing

### Requirements

+ Kind: https://kind.sigs.k8s.io/docs/user/quick-start/#installing-with-a-package-manager
+ Kubectl

### Setup

Run setup-cluster
Run create-cluster-resources

Run make system-test

TO cleanup / recreate the enviornemtn run destroy clister and start again from 1)
ONLY needs to be done if changing cluster config

## CI Testing