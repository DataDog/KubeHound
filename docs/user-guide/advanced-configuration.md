# Advanced configuration

## Running KubeHound from source

Clone the KubeHound repository and build KubeHound using the makefile:

```bash
git clone https://github.com/DataDog/KubeHound.git
cd KubeHound
make build
```

The built binary is now available at:

```bash
bin/build/kubehound
```

!!! Warning
    We do not advise to build KubeHound from the sources as the docker images will use the latest flag instead of a specific release version. This mainly used by the developers/maintainers of KubeHound. 

## Configuration

When using KubeHound you can setup different options through a config file with `-c` flags. You can use [kubehound-reference.yaml](https://github.com/DataDog/KubeHound/blob/main/configs/etc/kubehound-reference.yaml) as an example which list every options possible.

### Collector configuration

KubeHound is supporting 2 type of collector:

* `file-collector`: The file collector which can process an offline dump (made by KubeHound - see [common operation](https://kubehound.io/) for the dump command).
* `live-k8s-api-collector` (by default): The live k8s collector which will retrieve all kubernetes objects from the k8s API. 

#### File Collector

To use the file collector, you just have to specify:

* `directory`:  directory holding the K8s json data files
* `cluster`: the name of the targeted cluster

!!! Tip
    If you want to ingest data from a previous dump, we advise you to use `ingest local` command - [more detail here](https://kubehound.io/user-guide/common-operations/#ingest).

#### Live Collector

When retrieving the kubernetes resources form the k8s API, KubeHound setup limitation to avoid resources exhaustion on the k8s API:

* `rate_limit_per_second` (by default `50`): Rate limit of requests/second to the Kubernetes API. 
* `page_size` (by default `500`): Number of entries retrieved by each call on the API (same for all Kubernetes entry types)
* `page_buffer_size` (by default `10`): Number of pages to buffer

!!! Note
    Most (>90%) of the current runtime of KubeHound is spent in the transfer of data from the remote K8s API server, and the bulk of that is spent waiting on rate limit. As such increasing `rate_limit_per_second` will improve performance roughly linearly.

!!! Tip
    You can disable the interactive mod with `non_interactive` set to true. This will automatically dump all k8s resources from the k8s API without any user interaction.

### Builder 

The `builder` section allows you to customize how you want to chunk the data during the ingestion process. It is being splitted in 2 sections `vertices` and `edges`. For both graph entities, KubeHound uses a `batch_size` of `500` element by default.

!!! Warning
    Increasing batch sizes can have some performance improvements by reducing network latency in transferring data between KubeGraph and the application. However, increasing it past a certain level can overload the backend leading to instability and eventually exceed the size limits of the websocket buffer used to transfer the data. **Changing the default following setting is not recommended.**

#### Vertices builder

For the vertices builder, there is 2 options:

* `batch_size_small` (by default `500`): to control the batch size of vertices you want to insert through 
* `batch_size_small` (by default `100`): handle only the PermissionSet resouces. This resource is quite intensive because it is the only requirering  aggregation between multiples k8s resources (from `roles` and `rolebindings`).

!!! Note
    Since there is expensive insert on vertices the `batch_size_small` is currently not used.

#### Edges builder

By default, KubeHound will optimize the attack paths for large cluster by using `large_cluster_optimizations` (by default `true`). This will limit the number of attack paths being build in the targetted cluster. Using this optimisation will remove some attack paths. For instance, for the token based attacks (i.e. `TOKEN_BRUTEFORCE`), the optimisation will build only edges (between permissionSet and Identity) only if the targetted identity is `system:masters` group. This will reduce redundant attack paths:

* If the `large_cluster_optimizations` is activated, KubeHound will use the default `batch_size` (by default `500).
* If the `large_cluster_optimizations` is deactivated, KubeHound will use a specific batch size configured through `batch_size_cluster_impact` for all attacks that make the graph grow exponentially.

Lastly, the graph builder is using [pond](https://github.com/alitto/pond) library under the hood to handle the asynchronous tasks of inserting edges: 

* `worker_pool_size` (by default `5`): parallels ingestion process running at the same time (number of workers).
* `worker_pool_capacity` (by default `100`): number of cached elements in the worker pool.
