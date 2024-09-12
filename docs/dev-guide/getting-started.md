# Getting started

To list all the available developpers commands from the makefile, run:

```bash
make help
```

## Requirements build

+ go (v1.22): https://go.dev/doc/install
+ [Docker](https://docs.docker.com/engine/install/) >= 19.03 (`docker version`)
+ [Docker Compose](https://docs.docker.com/compose/compose-file/compose-versioning/) >= v2.0 (`docker compose version`)

## Backend

The backend images are built with the Dockerfiles `docker-compose.dev.[graph|ingestor|mongo|ui].yaml`. There are listed in [deployment directory](https://github.com/DataDog/KubeHound/tree/main/deployments/kubehound). To avoid running docker-compose it manually, there is an hidden command `kubehound dev --help`. The backend stack will be flagged as `kubehound-dev-` in the name of each component.

### Building the minimum dev stack

The minimum stack (`mongo` & `graph`) can be spawned with

* `kubehound dev` which is an equivalent of 
* `docker compose -f docker-compose.yaml -f docker-compose.dev.graph.yaml -f docker-compose.dev.mongo.yaml`. By default it will always rebuild everything (no cache is being used).

### Building dev options

You can add components to the mininum stack (`ui` and `grpc endpoint`) by adding the following flag.

* `--ui` to add the Jupyter UI to the build.
* `--grpc` to add the ingestor endpoint (exposing the grpc server for KHaaS).

For instance, building locally the minimum stack with the `ui` component:

```bash
kubehound dev --ui
```

### Tearing down the dev stack 

To tear down the KubeHound dev stack, just use `--down` flag:

```bash
kubehound dev --down
```

!!! Note
    It will stop all the component from the dev stack (including the `ui` and `grpc endpoint` if started)

## Build the binary

### Build from source

To build KubeHound locally from the sources, use the Makefile:

```bash
make build
```

KubeHound binary will be output to `./bin/build/kubehound`.


### Releases

We use `buildx` to release new versions of KubeHound, for cross platform compatibility and because we are embedding the docker compose library (to enable KubeHound to spin up the KubeHound stack directly from the binary). This saves the user from having to take care of this part. The build relies on 2 files [docker-bake.hcl](https://github.com/DataDog/KubeHound/blob/main/docker-bake.hcl) and [Dockerfile](https://github.com/DataDog/KubeHound/blob/main/Dockerfile). The following bake targets are available:

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

The CI releases a set of new images and binaries when a tag is created. To set a new tag on the main branch:

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
