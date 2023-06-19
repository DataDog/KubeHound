# KubeHound

A Kubernetes attack graph tool

Full documentation available on confluence: https://datadoghq.atlassian.net/wiki/spaces/ASE/pages/2871592134/KubeHound+1.0

## Build

Build the application via:

```bash
make build
```

All binaries will be output to the [bin](./bin/) folder

## Unit Testing

The full suite of unit tests can be run locally via:

```bash
make test
```

## System Testing

The repository includes a suite of system tests that will do the following:
+ create a local kubernetes cluster
+ collect kubernetes API data from the cluster
+ run KubeHound using the file collector to create a working graph database
+ query the graph database to ensure all expected vertices and edges have been created correctly

The cluster setup and running instances can be found under [test/setup](./test/setup/)

If you need to manually access the system test environement with kubectl and other commands, you'll need to set (assuming you are at the root dir):
```bash
cd test/setup/ && export KUBECONFIG=$(pwd)/.kube/config
```

### Requirements

+ Kind: https://kind.sigs.k8s.io/docs/user/quick-start/#installing-with-a-package-manager
+ Kubectl: https://kubernetes.io/docs/tasks/tools/

#### Environment variable:
- `DD_API_KEY` (optional): set to the datadog API key used to submit metrics and other observability data.

### Setup

Setup the test kind cluster (you only need to do this once!) via:

```bash
make local-cluster-deploy
```

Then run the system tests via:

```bash
make system-test
```

To cleanup the environment you can destroy the cluster via:

```bash
make local-cluster-destroy
```

To list all the available commands, run:

```bash
make help
```

### CI Testing

System tests will be run in CI via the [system-test](./.github/workflows/system-test.yml) github action 