# KubeHound  <!-- omit in toc -->

<p align="center">
  <img src="./docs/logo.png" alt="KubeHound" width="300" />
</p>
A Kubernetes attack graph tool allowing automated calculation of attack paths between assets in a cluster

## Quick Links <!-- omit in toc -->

+ For an overview of the application architecture see the [design canvas](./docs/application/Architecture.excalidraw)
+ To see the attacks covered see the [edge definitions](./docs/edges/)
+ To contribute a new attack to the project follow the [contribution guidelines](./CONTRIBUTING.md)

## Sample Attack Path  <!-- omit in toc -->

![Example Path](./docs/images/example-graph.png)

## Contents  <!-- omit in toc -->

- [Requirements](#requirements)
  - [Application](#application)
  - [Test (Development only)](#test-development-only)
- [Quick Start](#quick-start)
  - [Prebuilt Releases](#prebuilt-releases)
  - [From Source](#from-source)
  - [Sample Data](#sample-data)
- [Advanced Usage](#advanced-usage)
  - [Infrastructure Setup](#infrastructure-setup)
  - [Running Kubehound](#running-kubehound)
- [Using KubeHound Data](#using-kubehound-data)
- [Development](#development)
  - [Build](#build)
  - [Release build](#release-build)
  - [Unit Testing](#unit-testing)
  - [System Testing](#system-testing)
    - [Environment variable:](#environment-variable)
    - [Setup](#setup)
    - [CI Testing](#ci-testing)
- [Acknowledgements](#acknowledgements)

## Requirements

### Application

+ Golang `>= 1.20`: https://go.dev/doc/install
+ Docker `>= 19.03`: https://docs.docker.com/engine/install/
+ Docker Compose `V2`: https://docs.docker.com/compose/compose-file/compose-versioning/

### Test (Development only)

+ Kind: https://kind.sigs.k8s.io/docs/user/quick-start/#installing-with-a-package-manager
+ Kubectl: https://kubernetes.io/docs/tasks/tools/

## Quick Start

### Prebuilt Releases

Release binaries are available for Linux / Windows / Mac OS via the [releases](https://github.com/DataDog/KubeHound/releases) page. These provide access to core KubeHound functionality but lack support for the `make` commands detailed in subsequent sections. Once the release archive is downloaded and extracted start the backend via:

```bash
./kubehound.sh backend-up
```

*NOTE*:
+ If downloading the releases via a browser you must run e.g `xattr -d com.apple.quarantine KubeHound_Darwin_arm64.tar.gz` before running to prevent [MacOS blocking execution](https://support.apple.com/en-gb/guide/mac-help/mchleab3a043/mac)

Next choose a target Kubernetes cluster, either:

* Select the targeted cluster via `kubectx` (need to be installed separately)     
* Use a specific kubeconfig file by exporting the env variable: `export KUBECONFIG=/your/path/to/.kube/config`

Finally run the compiled binary with packaged configuration (`config.yaml`):

```bash
./kubehound.sh run
```

### From Source

Clone this repository via git:

```bash
git clone https://github.com/DataDog/KubeHound.git
```

KubeHound ships with a sensible default configuration designed to get new users up and running quickly. First step is to prepare the application:

```bash
make kubehound
```

This will do the following:
* Start the backend services via docker compose (wiping any existing data)
* Compile the kubehound binary from source

Next choose a target Kubernetes cluster, either:

* Select the targeted cluster via `kubectx` (need to be installed separately)     
* Use a specific kubeconfig file by exporting the env variable: `export KUBECONFIG=/your/path/to/.kube/config`

Finally run the compiled binary with default configuration:

```bash
bin/kubehound
```

To view the generated graph see the [Using KubeHound Data](#using-kubehound-data) section.

### Sample Data

To view a sample graph demonstrating attacks in a very, very vulnerable cluster you can generate data via running the app against the provided kind cluster:

```bash
make sample-graph
```

To view the generated graph see the [Using KubeHound Data](#using-kubehound-data) section. 

## Advanced Usage

### Infrastructure Setup

First create and populate a .env file with the required variables:

```bash
cp deployments/kubehound/.env.tpl deployments/kubehound/.env
```

Edit the variables (datadog env `DD_*` related and `KUBEHOUND_ENV`):

* `KUBEHOUND_ENV`: `dev` or `release` 
* `DD_API_KEY`: api key you created from https://app.datadoghq.com/ website

Note:
* `KUBEHOUND_ENV=dev` will build the images locally (and provide some local debugging containers e.g `mongo-express`)
* `KUBEHOUND_ENV=release` will use prebuilt images from ghcr.io 

### Running Kubehound

To replicate the automated command and run KubeHound step-by-step. First build the application:

```bash
make build
```

Next spawn the backend infrastructure

```bash
make backend-up
```

Next create a configuration file:

```yaml
collector:
  type: live-k8s-api-collector
telemetry:
  enabled: true
```

A tailored sample configuration file can be found [here](./configs/etc/kubehound.yaml), a full configuration reference containing all possible parameters [here](./configs/etc/kubehound-reference.yaml). 

Finally run the KubeHound binary, passing in the desired configuration:

```bash
bin/kubehound -c <config path>
```

Remember the targeted cluster must be set via `kubectx` or setting the `KUBECONFIG` environment variable. Additional functionality for managing the application can be found via:

```bash
make help
```

## Using KubeHound Data

To query the KubeHound graph data requires using the [Gremlin](https://tinkerpop.apache.org/gremlin.html) query language via an API call or dedicated graph query UI. A number of graph query UIs are availble, but we recommend [gdotv](https://gdotv.com/). To access the KubeHound graph using `gdotv`:

+ Download and install the application from https://gdotv.com/
+ Create a connection to the local janusgraph instance by following the steps here https://docs.gdotv.com/connection-management/ and using `hostname=localhost`
+ Navigate to the query editor and enter a sample query e.g `g.V().count()`. See detailed instructions here: https://docs.gdotv.com/query-editor/#run-your-query

## Development

### Build

Build the application via:

```bash
make build
```

All binaries will be output to the [bin](./bin/) folder

### Release build

Build the release packages locally using [goreleaser](https://goreleaser.com/install):

```bash
make local-release
```

### Unit Testing

The full suite of unit tests can be run locally via:

```bash
make test
```

### System Testing

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

#### Environment variable:
- `DD_API_KEY` (optional): set to the datadog API key used to submit metrics and other observability data.

#### Setup

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

#### CI Testing

System tests will be run in CI via the [system-test](./.github/workflows/system-test.yml) github action 


## Acknowledgements

KubeHound was created by the Adversary Simulation Engineering (ASE) team at Datadog:

+ Jeremy Fox [@0xff6a](https://www.twitter.com/0xff6a)
+ Julien Terriac
+ Edouard Schweisguth [@edznux](https://www.twitter.com/edznux)

With additional support from:

+ Christophe Tafani-Dereeper [@christophetd](https://twitter.com/christophetd)

We would also like to acknowledge the [BloodHound](https://github.com/BloodHoundAD/BloodHound) team for pioneering the use of graph theory in offensive security and inspiring us to create this project. 