# KubeHound

A Kubernetes attack graph tool

Full documentation available on confluence: https://datadoghq.atlassian.net/wiki/spaces/ASE/pages/2871592134/KubeHound+1.0

## Run

To run the application, you can use docker image with the compose:
* Copy `deployments/kubehound/.env.tpl` to `deployments/kubehound/.env`
* Edit the variables (datadog env `DD_*` related and `KUBEHOUND_ENV`)
* Run `make backend-up`

Note:
* KUBEHOUND_ENV=prod will use prebuilt image from ghcr.io
* KUBEHOUND_ENV=dev will build the images locally

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
cd test/setup/ && export KUBECONFIG=$(pwd)/.kube-config
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

Note: if you are running on Linux but you dont want to run `sudo` for `kind` and `docker` command, you can overwrite this behavior by editing the following var in `test/setup/.config`:
* `DOCKER_CMD="docker"` for docker command
* `KIND_CMD="kind"` for kind command 

### CI Testing

System tests will be run in CI via the [system-test](./.github/workflows/system-test.yml) github action 

### Sample Graph

To view a sample graph demonstrating attacks in a very, very vulnerable cluster you can generate data via the system tests:

```bash
make local-cluster-deploy && make system-test
```

Then use a graph visualizer of choice (we recommend [gdotv](https://gdotv.com/)) to connect to localhost and view and query the sample data.

### Querying Kubehound

To query the KubeHound graph data requires using the [Gremlin](https://tinkerpop.apache.org/gremlin.html) query language. See the provided [cheatsheet](./pkg/kubehound/graph/CHEATSHEET.md) for examples of useful queries for various use cases.