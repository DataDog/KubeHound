# Local Common Operations

When running `./kubehound`, it will execute the 3 following action:

- run the `backend` (graphdb, storedb and UI)
- `dump` the kubernetes resources needed to build the graph
- `ingest` the dumped data and generate the attack path for the targeted Kubernetes cluster.

All those 3 steps can be run separately.

[![](../images/kubehound-local-commands.png)](../images/kubehound-local-commands.png)

!!! note

    if you want to skip the interactive mode, you can provide `-y` or `--non-interactive` to skip the cluster confirmation.

## Backend

In order to run, KubeHound needs some docker containers to be running. Every commands has been embedded into KubeHound to simplify the user experience. You can find all the docker-compose used [here](https://github.com/DataDog/KubeHound/tree/main/deployments/kubehound).

### Starting the backend

#### Starting the backend with default images

The backend stack can be started by using:

```bash
kubehound backend up
```

It will use the latest [kubehound images releases](https://github.com/orgs/DataDog/packages?repo_name=KubeHound)

#### Starting the backend with overrides

For various reasons, you might want to use a specific version or pull the image from a specific registry. You can override the [default behaviour](https://github.com/DataDog/KubeHound/blob/main/deployments/kubehound/docker-compose.yaml) by using the following `docker-compose.overrides.yaml` file:

```yaml
name: kubehound-release
services:
  mongodb:
    image: your.registry.tld/mongo/mongo:6.0.6
    ports:
      - "127.0.0.1:27017:27017"

  kubegraph:
    image: your.registry.tld/datadog/kubehound-graph:my-specific-tag
    ports:
      - "127.0.0.1:8182:8182"
      - "127.0.0.1:8099:8099"
  
  ui-jupyter:
    image: your.registry.tld/datadog/kubehound-ui:my-specific-tag

  ui-invana-engine:
    image: your.registry.tld/invanalabs/invana-engine:latest

  ui-invana-studio:
    image: your.registry.tld/invanalabs/invana-studio:latest
```

Then you can start the backend with the following command:

```bash
kubehound backend up -f docker-compose.overrides.yml
```

### Restarting/stopping the backend

The backend stack can be restarted by using:

```bash
kubehound backend reset
```

or just stopped:

```bash
kubehound backend down
```

These commands will simply reboot backend services, but persist the data via docker volumes.

### Wiping the database

The backend data can be wiped by using:

```bash
kubehound backend wipe
```

!!! warning

    This command will **wipe ALL docker DATA (docker volume and containers) and will not be recoverable**.

## Dump

### Create a dump localy with all needed k8s resources

For instance, if you want to dump a configuration to analyse it later or just on another computer, KubeHound can create a self sufficient dump with the Kubernetes resources needed. By default it will create a `.tar.gz` file with all the dumper k8s resources needed.

```bash
kubehound dump local [directory to dump the data]
```

If for some reasons you need to have the raw data, you can add `--no-compress` flag to have a raw extract.

!!! note

    This step does not require any backend as it only automate grabbing k8s resources from the k8s api.

## Ingest

### Ingest a local dump

To ingest manually an extraction made by KubeHound, just specify where the dump is being located and the associated cluster name.

```bash
kubehound ingest local [directory or tar.gz path]
```

!!! warning

    This step requires the backend to be started, it will not start it for you.

!!! warning "deprecated"

    The `--cluster` is deprecated since v1.5.0. Now a metadata.json is being embeded with the cluster name. If you are using old dump you can either still use the `--cluster` flag or auto detect it from the path.
