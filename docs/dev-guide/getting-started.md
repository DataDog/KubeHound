# Getting started

## Requirements Test

+ Kind: https://kind.sigs.k8s.io/docs/user/quick-start/#installing-with-a-package-manager
+ Kubectl: https://kubernetes.io/docs/tasks/tools/
+ go (v1.22): https://go.dev/doc/install

## Backend

The backend is built with Dockerfile `docker-compose.dev.[graph|ingestor|mongo|ui].yaml`. There are listed in [deployment directory](https://github.com/DataDog/KubeHound/tree/main/deployments/kubehound). To avoid running it manually, there is an hidden command `kubehound dev --help`. The backend stack will be flagged as `kubehound-dev-` in the name of each component.

### Building the minimum dev stack

To spin up the minimum stack is `mongo`, `graph` can spwaned by `kubehound dev` which is an equivalent of `docker compose -f docker-compose.yaml -f docker-compose.dev.graph.yaml -f docker-compose.dev.mongo.yaml`. By default it will always rebuild everything (no cache is being used).

### Building dev options

You can add components to the mininum stack (`ui` and `grpc endpoint`) by adding the following flag.

* `--ui` to add the Jupyter UI to the build.
* `--grpc` to add the ingestor endpoint (exposing the grpc server for KHaaS).

For instance, building locally the minimum stack with the `ui` component:

```bash
kubehound dev --ui
```

### Tearing down the dev stack 

To tear down the KubeHound dev stack, just add `--down` to your previous command. For instant to stop the dev stack with the `ui`:

```bash
kubehound dev --ui --down
```

## Build the binary

### Build from source

To build KubeHound locally from the sources, use the Makefile:

```bash
make build
```

All binaries will be output to the [bin](./bin/build) folder


### Releases

To release a new version of KubeHound, we are using `buildx` to have cross platform compatibility. We have to use as we are embedding docker compose library to enable KubeHound to spin up the KubeHound stack directly from the binary. It avoids user to take care of this part. The build relies on 2 files [docker-bake.hcl](https://github.com/DataDog/KubeHound/blob/main/docker-bake.hcl) and [Dockerfile](https://github.com/DataDog/KubeHound/blob/main/Dockerfile). The following bake targets are available:

* `validate` or `lint`: run the release CI linter
* `binary` (default option):   build kubehound just for the local architecture
* `binary-cross` or `release`: run the cross platform compilation 

!!! Note
    Those targets are made only for the CI and are not intented to be run run locally (except to test the CI locally).


##### Cross platform compilation

To test the cross platform compilation locally, use the buildx bake target `release`. This target is being run by the CI ([buildx](https://github.com/DataDog/KubeHound/blob/main/.github/workflows/buildx.yml#L77-L84 workflow). 

```bash
docker buildx bake release
```

!!! Warning
    The cross-binary compilation with `buildx` is not working in mac: `ERROR: Multi-platform build is not supported for the docker driver.`

## Push a new release

The CI will release a set of new images and binary when a tag is created. The goal is to set a new tag on the main branch:

```bash
git tag vX.X.X
git push origin vX.X.X
```

New tags will trigger the 2 following jobs:

* [docker](): pushing new images for `kubehound-graph`, `kubehound-ingestor` and `kubehound-ui` on ghcr.io. The images can be listed [here](https://github.com/orgs/DataDog/packages?repo_name=KubeHound).
* [buildx](https://github.com/DataDog/KubeHound/blob/main/.github/workflows/buildx.yml): compiling the binary for all platform. The platform supported can be listed using this `docker buildx bake binary-cross --print | jq -cr '.target."binary-cross".platforms'`.

The CI will draft a new release that **will need manual validation**. In order to get published, an admin has to to validate the new draft from the UI.

!!! Tip
    To resync all the tags from the main repo you can use `git tag -l | xargs git tag -d;git fetch --tags`.

## Testing

To ensure no regression in KubeHound, 2 kinds of tests are in place:

* classic unit test: can be identify with the `xxx_test.go` files in the source code
* system tests: end to end test where we run full ingestion from different scenario to simulate all use cases against a real cluster.

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

If you need to manually access the system test environment with kubectl and other commands, you'll need to set (assuming you are at the root dir):

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