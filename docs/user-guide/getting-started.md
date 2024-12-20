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
wget https://github.com/DataDog/KubeHound/releases/latest/download/kubehound-$(uname -o | sed 's/GNU\///g')-$(uname -m) -O kubehound
chmod +x kubehound
```

Then just run `./kubehound`, it will start [backend services](../architecture.md) via docker compose v2 API.

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
./kubehound
INFO[01:42:19] Loading application configuration from default embedded
WARN[01:42:19] No local config file was found (kubehound.yaml)
INFO[01:42:19] Using /home/datadog/kubehound for default config
INFO[01:42:19] Initializing application telemetry
WARN[01:42:19] Telemetry disabled via configuration
INFO[01:42:19] Loading backend from default embedded
WARN[01:42:19] Loading the kubehound images with tag latest - dev branch detected
INFO[01:42:19] Spawning the kubehound stack
[+] Running 3/3
 ✔ Container kubehound-release-kubegraph-1 Healthy                                                                                                                                                                              50.3s
 ✔ Container kubehound-release-ui-jupyter-1  Healthy                                                                                                                                                                              50.3s
 ✔ Container kubehound-release-mongodb-1     Healthy                                                                                                                                                                              58.4s
INFO[01:43:20] Starting KubeHound (run_id: 01j4fwbg88j6eptasgegdh2sgs)
INFO[01:43:20] Initializing providers (graph, cache, store)
INFO[01:43:20] Loading cache provider
INFO[01:43:20] Loaded memcache cache provider
INFO[01:43:20] Loading store database provider
INFO[01:43:20] Loaded mongodb store provider
INFO[01:43:21] Loading graph database provider
INFO[01:43:21] Loaded janusgraph graph provider
INFO[01:43:21] Running the ingestion pipeline
INFO[01:43:21] Loading Kubernetes data collector client
WARN[01:43:21] About to dump k8s cluster: "kind-kubehound.test.local" - Do you want to continue ? [Yes/No]
yes
INFO[01:43:30] Loaded k8s-api-collector collector client
INFO[01:43:30] Starting Kubernetes raw data ingest
INFO[01:43:30] Loading data ingestor
INFO[01:43:30] Running dependency health checks
INFO[01:43:30] Running data ingest and normalization
INFO[01:43:30] Starting ingest sequences
INFO[01:43:30] Waiting for ingest sequences to complete
INFO[01:43:30] Running ingestor sequence core-pipeline
INFO[01:43:30] Starting ingest sequence core-pipeline
INFO[01:43:30] Running ingest group k8s-role-group
INFO[01:43:30] Starting k8s-role-group ingests
INFO[01:43:30] Waiting for k8s-role-group ingests to complete
INFO[01:43:30] Running ingest k8s-role-ingest
INFO[01:43:30] Running ingest k8s-cluster-role-ingest
INFO[01:43:30] Streaming data from the K8s API
INFO[01:43:32] Completed k8s-role-group ingest
INFO[01:43:32] Finished running ingest group k8s-role-group
...
INFO[01:43:35] Completed k8s-pod-group ingest
INFO[01:43:35] Finished running ingest group k8s-pod-group
INFO[01:43:35] Completed ingest sequence core-pipeline
INFO[01:43:35] Completed pipeline ingest
INFO[01:43:35] Completed data ingest and normalization in 5.065238542s
INFO[01:43:35] Loading graph edge definitions
INFO[01:43:35] Loading graph builder
INFO[01:43:35] Running dependency health checks
INFO[01:43:35] Constructing graph
WARN[01:43:35] Using large cluster optimizations in graph construction
INFO[01:43:35] Starting mutating edge construction
INFO[01:43:35] Building edge PodCreate
INFO[01:43:36] Edge writer 10 PodCreate::POD_CREATE written
INFO[01:43:36] Building edge PodExec
...
INFO[01:43:36] Starting dependent edge construction
INFO[01:43:36] Building edge ContainerEscapeVarLogSymlink
INFO[01:43:36] Edge writer 5 ContainerEscapeVarLogSymlink::CE_VAR_LOG_SYMLINK written
INFO[01:43:36] Completed edge construction
INFO[01:43:36] Completed graph construction in 773.2935ms
INFO[01:43:36] Stats for the run time duration: 5.838839708s / wait: 5.926496s / throttling: 101.501262%
INFO[01:43:36] KubeHound run (id=01j4fwbg88j6eptasgegdh2sgs) complete in 15.910406167s
WARN[01:43:36] KubeHound as finished ingesting and building the graph successfully.
WARN[01:43:36] Please visit the UI to view the graph by clicking the link below:
WARN[01:43:36] http://localhost:8888
WARN[01:43:36] Password being 'admin'
```

## Access the KubeHound data

At this point, the KubeHound data has been ingested in KubeHound's [graph database](../architecture.md).
You can use any client that supports accessing JanusGraph - a comprehensive list is available on the [JanusGraph home page](https://janusgraph.org/).
We also provide a showcase [Jupyter Notebook](https://github.com/DataDog/KubeHound/blob/main/deployments/kubehound/ui/KubeHound.ipynb) to get you started. This is accessible on [http://locahost:8888](http://locahost:8888) after starting KubeHound backend. The default password is `admin` but you can change this by setting the `NOTEBOOK_PASSWORD` environment variable in your `.env file`.

## Visualize and query the KubeHound data

!!! note

    You can find the visual representation of the KubeHound graph model [here](/reference/).

Once the data is loaded in the graph database, it's time to visualize and query it!

You can explore it interactively in your graph client. Then, refer to KubeHound's [query library](../queries/index.md) to start asking questions to your data.

## Generating sample data

If you don't have a cluster at your disposal: clone the KubeHound repository, install [kind](https://kind.sigs.k8s.io/#installation-and-usage) and run the following command:

```bash
make sample-graph
```

This will spin up a temporary local kind cluster, run KubeHound on it, and destroy the cluster.
