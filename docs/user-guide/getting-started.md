# Getting started

## Prerequisites

To get started with KubeHound, you'll need the following pre-requirements on your system:

- [Golang](https://go.dev/doc/install) >= 1.20 (`go version`)
- [Docker](https://docs.docker.com/engine/install/) >= 19.03 (`docker version`)
- [Docker Compose](https://docs.docker.com/compose/compose-file/compose-versioning/) >= v2.0 (`docker compose version`)

## Running KubeHound

??? info "tl;dr"

    ```bash
    make kubehound && bin/kubehound
    ```

KubeHound ships with a sensible default configuration designed to get new users up and running quickly. First, prepare the application by running:

```bash
make kubehound
```

This will start [backend services](../architecture.md) via docker compose (wiping any existing data), and compile the kubehound binary from source

Next, make sure your current kubectl context points at the target cluster:

```bash
# View the current context
kubectl config current-context

# Set your context
kubectl config set-context <name>

# alternatively, use https://github.com/ahmetb/kubectx
```

Finally, run KubeHound with the default [configuration](TODO):

```
bin/kubehound
```

Sample output:

```text
INFO[0000] Starting KubeHound (run_id: aff49337-5e36-46ea-ac1f-ed224bf215ba)  component=kubehound run_id=aff49337-5e36-46ea-ac1f-ed224bf215ba service=kubehound
INFO[0000] Initializing launch options                   component=kubehound run_id=aff49337-5e36-46ea-ac1f-ed224bf215ba service=kubehound
INFO[0000] Loading application configuration from default embedded  component=kubehound run_id=aff49337-5e36-46ea-ac1f-ed224bf215ba service=kubehound
INFO[0000] Initializing application telemetry            component=kubehound run_id=aff49337-5e36-46ea-ac1f-ed224bf215ba service=kubehound
INFO[0000] Loading cache provider                        component=kubehound run_id=aff49337-5e36-46ea-ac1f-ed224bf215ba service=kubehound
INFO[0000] Loaded MemCacheProvider cache provider        component=kubehound run_id=aff49337-5e36-46ea-ac1f-ed224bf215ba service=kubehound
INFO[0000] Loading store database provider               component=kubehound run_id=aff49337-5e36-46ea-ac1f-ed224bf215ba service=kubehound
INFO[0000] Loaded MongoProvider store provider           component=kubehound run_id=aff49337-5e36-46ea-ac1f-ed224bf215ba service=kubehound
INFO[0000] Loading graph database provider               component=kubehound run_id=aff49337-5e36-46ea-ac1f-ed224bf215ba service=kubehound
INFO[0000] Loaded JanusGraphProvider graph provider      component=kubehound run_id=aff49337-5e36-46ea-ac1f-ed224bf215ba service=kubehound
INFO[0001] Starting Kubernetes raw data ingest           component=kubehound run_id=aff49337-5e36-46ea-ac1f-ed224bf215ba service=kubehound
INFO[0001] Loading Kubernetes data collector client      component=kubehound run_id=aff49337-5e36-46ea-ac1f-ed224bf215ba service=kubehound
INFO[0001] Loaded k8s-api-collector collector client     component=kubehound run_id=aff49337-5e36-46ea-ac1f-ed224bf215ba service=kubehound
...
INFO[0028] Building edge ExploitHostWrite                component=kubehound run_id=aff49337-5e36-46ea-ac1f-ed224bf215ba service=kubehound
INFO[0028] Edge writer 22 ContainerAttach::CONTAINER_ATTACH written  component=kubehound run_id=aff49337-5e36-46ea-ac1f-ed224bf215ba service=kubehound
INFO[0028] Building edge IdentityAssumeNode              component=kubehound run_id=aff49337-5e36-46ea-ac1f-ed224bf215ba service=kubehound
INFO[0029] Edge writer 8 ExploitHostWrite::EXPLOIT_HOST_WRITE written  component=kubehound run_id=aff49337-5e36-46ea-ac1f-ed224bf215ba service=kubehound
...
INFO[0039] Completed edge construction                   component=kubehound run_id=aff49337-5e36-46ea-ac1f-ed224bf215ba service=kubehound
INFO[0039] Completed graph construction                  component=kubehound run_id=aff49337-5e36-46ea-ac1f-ed224bf215ba service=kubehound
INFO[0039] Attack graph generation complete in 39.108174109s  component=kubehound run_id=aff49337-5e36-46ea-ac1f-ed224bf215ba service=kubehound
```


## Access the KubeHound data

At this point, the KubeHound data has been ingested in KubeHound's [graph database](../architecture.md). 
You can use any client that supports accessing JanusGraph - we recommend using [gdotv](https://gdotv.com/):

- [Download and install] gdotv from the [official website](https://gdotv.com/)
- [Create a connection] to the local KubeHound JanusGraph instance
  1. Click on the `New database connection` button
  2. Enter `localhost` as an hostname, and click on the `Test connection` button
  3. Once the connection is successful, click `Submit` - you're good to go!

## Visualize and query the KubeHound data

Once the data is loaded in the graph database, it's time to visualize and query it! 

You can explore it interactively in your graph client. Then, refer to KubeHound's [query library](/queries/) to start asking questions to your data.

## Generating sample data

If you don't have a cluster at your disposal, install [kind](https://kind.sigs.k8s.io/#installation-and-usage) and run the following command:

```bash
make sample-graph
```

This will spin up a temporary local kind cluster, run KubeHound on it, and destroy the cluster.

