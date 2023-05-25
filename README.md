# KubeHound

A Kubernetes attack graph tool

Full documentation available on confluence: https://datadoghq.atlassian.net/wiki/spaces/ASE/pages/2871592134/KubeHound+1.0

## Build

Build the application via:

```
$ make build
```

All binaries will be output to the [bin](./bin/) folder

## Unit Testing

The full suite of unit tests can be run locally via:

```
$ make test
```

## System Testing

The repository includes a suite of system tests that will do the following:
+ create a local kubernetes cluster
+ collect kubernetes API data from the cluster
+ run KubeHound using the file collector to create a working graph database
+ query the graph database to ensure all expected vertices and edges have been created correctly

The cluster setup and running instances can be found under [test/setup](./test/setup/)

### Requirements

+ Kind: https://kind.sigs.k8s.io/docs/user/quick-start/#installing-with-a-package-manager
+ Kubectl: https://kubernetes.io/docs/tasks/tools/

### Setup

Setup the test kind cluster (you only need to do this once!) via:

```
$ make local-cluster-setup
```

Then run the system tests via:

```
$ make system-test
```

To cleanup the environment you can destroy the cluster via:

```
$ make local-cluster-destroy
```

### CI Testing

System tests will be run in CI via the [system-test](./.github/workflows/system-test.yml) github action 