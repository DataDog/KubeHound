# Getting started

## Prerequisites

To get started with KubeHound, you'll need the following pre-requirements on your system:

- [Docker](https://docs.docker.com/engine/install/) >= 19.03 (`docker version`)
- [Docker Compose](https://docs.docker.com/compose/compose-file/compose-versioning/) >= v2.0 (`docker compose version`)

These two are used to start the backend infrastructure required to run KubeHound. It also provides a default user interface via Jupyter notebooks.

## Running KubeHound

KubeHound ships with a sensible default configuration as well as a pre-built binary, designed to get new users up and running quickly. 

Download the latest KubeHound binary for you platform:

```bash
wget https://github.com/DataDog/KubeHound/releases/download/latest/kubehound-$(uname -o | sed 's/GNU\///g')-$(uname -m) -O kubehound
chmod +x kubehound
```

This will start [backend services](../architecture.md) via docker compose (wiping any existing data), and compile the kubehound binary from source.

Next, make sure your current kubectl context points at the target cluster:

```bash
# View the current context
kubectl config current-context

# Set your context
kubectl config set-context <name>

# alternatively, use https://github.com/ahmetb/kubectx
```

Finally, run KubeHound with the default [configuration](https://github.com/DataDog/KubeHound/blob/main/configs/etc/kubehound.yaml):

```
kubehound
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
You can use any client that supports accessing JanusGraph - a comprehensive list is available on the [JanusGraph home page](https://janusgraph.org/). We also provide a showcase [Jupyter Notebook](../../deployments/kubehound/notebook/KubeHound.ipynb) to get you started. This is accessible on [http://locahost:8888](http://locahost:8888) after starting KubeHound backend. The default password is `admin` but you can change this by setting the `NOTEBOOK_PASSWORD` environment variable in your `.env file`.

## Visualize and query the KubeHound data

Once the data is loaded in the graph database, it's time to visualize and query it! 

You can explore it interactively in your graph client. Then, refer to KubeHound's [query library](../queries/index.md) to start asking questions to your data.

## Generating sample data

If you don't have a cluster at your disposal: clone the KubeHound repository, install [kind](https://kind.sigs.k8s.io/#installation-and-usage) and run the following command:

```bash
make sample-graph
```

This will spin up a temporary local kind cluster, run KubeHound on it, and destroy the cluster.

