# Testing

To ensure no regression in KubeHound, 2 kinds of tests are in place:

- classic unit test: can be identify with the `xxx_test.go` files in the source code
- system tests: end to end test where we run full ingestion from different scenario to simulate all use cases against a real cluster.

## Requirements test

- [Golang](https://go.dev/doc/install) `>= 1.23`
- [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installing-with-a-package-manager)
- [Kubectl](https://kubernetes.io/docs/tasks/tools/)

## Unit Testing

The full suite of unit tests can be run locally via:

```bash
make test
```

## System Testing

The repository includes a suite of system tests that will do the following:

- create a local kubernetes cluster
- collect kubernetes API data from the cluster
- run KubeHound using the file collector to create a working graph database
- query the graph database to ensure all expected vertices and edges have been created correctly

The cluster setup and running instances can be found under [test/setup](https://github.com/DataDog/KubeHound/tree/main/test/setup)

If you need to manually access the system test environment with kubectl and other commands, you'll need to set (assuming you are at the root dir):

```bash
cd test/setup/ && export KUBECONFIG=$(pwd)/.kube-config
```

### Environment variable:

- `DD_API_KEY` (optional): set to the datadog API key used to submit metrics and other observability data (see [datadog](https://kubehound.io/dev-guide/datadog/) section)

### Setup

Setup the test kind cluster (you only need to do this once!) via:

```bash
make local-cluster-deploy
```

### Running the system tests

Then run the system tests via:

```bash
make system-test
```

### Cleanup

To cleanup the environment you can destroy the cluster via:

```bash
make local-cluster-destroy
```

!!! note

    if you are running on Linux but you dont want to run `sudo` for `kind` and `docker` command, you can overwrite this behavior by editing the following var in `test/setup/.config`:

        * `DOCKER_CMD="docker"` for docker command
        * `KIND_CMD="kind"` for kind command

### CI Testing

System tests will be run in CI via the [system-test](https://github.com/DataDog/KubeHound/blob/main/.github/workflows/system-test.yml) github action
