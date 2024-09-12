# KubeHound DSL

The KubeHound graph ships with a custom DSL that simplifies queries for the most common use cases

## Using the KubeHound graph

The KubeHound DSL can be used by starting a traversal with `kh` instead of the traditional `g`. All gremlin queries will work exactly as normal, but a number of additional methods, specific to KubeHound, will be available.

```groovy
// First 100 vertices in the kubehound graph
kh.V().limit(100)
```

## List of available methods

_DSL definition code available [here](https://github.com/DataDog/KubeHound/blob/main/deployments/kubehound/kubegraph/dsl/kubehound/src/main/java/com/datadog/ase/kubehound/)._

### Retrieve cluster data

| Method                      | Gremlin equivalent                                    |
| --------------------------- | ----------------------------------------------------- |
| `.cluster([string...])`     | `.hasLabel("Cluster")`                                |
| `.containers([string...])`  | `.hasLabel("Container")`                              |
| `.endpoints([int])`         | `.hasLabel("Endpoint")`                               |
| `.groups([string...])`      | `.hasLabel("Group")`                                  |
| `.hostMounts([string...])`  | `.hasLabel("Volume").has("type", "HostPath")`         |
| `.nodes([string...])`       | `.hasLabel("Node")`                                   |
| `.permissions([string...])` | `.hasLabel("PermissionSet")`                          |
| `.pods([string...])`        | `.hasLabel("Pod")`                                    |
| `.run([string...])`         | `.has("runID", P.within(ids)`                         |
| `.sas([string...])`         | `.hasLabel("Identity").has("type", "ServiceAccount")` |
| `.services([string...])`    | `.hasLabel("Endpoint").has("exposure", EXTERNAL)`     |
| `.users([string...])`       | `.hasLabel("Identity").has("type", "User")`           |
| `.volumes([string...])`     | `.hasLabel("Volume")`                                 |

### Retrieving attack oriented data

| Method                                 | Gremlin equivalent                                                                                                                                                                                |
| -------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `.attacks()`                           | `.outE().inV().path()`                                                                                                                                                                            |
| `.critical()`                          | `.has("critical", true)`                                                                                                                                                                          |
| `.criticalPaths(int)`                  | see [KubeHoundTraversalDsl.java](https://github.com/DataDog/KubeHound/blob/main/deployments/kubehound/kubegraph/dsl/kubehound/src/main/java/com/datadog/ase/kubehound/KubeHoundTraversalDsl.java) |
| `.criticalPathsFilter(int, string...)` | see [KubeHoundTraversalDsl.java](https://github.com/DataDog/KubeHound/blob/main/deployments/kubehound/kubegraph/dsl/kubehound/src/main/java/com/datadog/ase/kubehound/KubeHoundTraversalDsl.java) |
| `.criticalPathsFreq([maxHops])`        | see [KubeHoundTraversalDsl.java](https://github.com/DataDog/KubeHound/blob/main/deployments/kubehound/kubegraph/dsl/kubehound/src/main/java/com/datadog/ase/kubehound/KubeHoundTraversalDsl.java) |
| `.hasCriticalPath()`                   | `.where(__.criticalPaths().limit(1))`                                                                                                                                                             |
| `.minHopsToCritical([maxHops])`        | see [KubeHoundTraversalDsl.java](https://github.com/DataDog/KubeHound/blob/main/deployments/kubehound/kubegraph/dsl/kubehound/src/main/java/com/datadog/ase/kubehound/KubeHoundTraversalDsl.java) |

For more detailed explanation, please see below.

Example of a kubehound DSL capabilities:

```groovy
// Example returning all attacks from containers running the cilium 1.11.18 image
kh.containers().has("image", "eu.gcr.io/internal/cilium:1.11.18").attacks()
```

## KubeHound Constants

### Endpoint Exposure

Represents the exposure level of endpoints in the KubeHound graph

```java
// Defines the exposure of an endpoint within the KubeHound model
public enum EndpointExposure {
  None,
	ClusterIP,                      // Container port exposed to cluster
	NodeIP,                         // Kubernetes endpoint exposed outside the cluster
	External,                       // Kubernetes endpoint exposed outside the cluster
}
```

## Traversal Source Reference

### Run Step

Starts a traversal that finds all vertices from the specified KubeHound run(s).

```java
GraphTraversal<Vertex, Vertex> run(String... ids)
```

Example usage:

```groovy
// All vertices in the graph from a single run
kh.run("01he5ebh73tah762qgdd5k4wqp")

// All vertices in the graph from a multiple runs
kh.run("01he5ebh73tah762qgdd5k4wqp", "01he5eagzbnhtfnwzg7xxbyfz4")

// All containers in the graph from a single run
kh.run("01he5ebh73tah762qgdd5k4wqp").containers()
```

### Cluster Step

Starts a traversal that finds all vertices from the specified cluster(s).

```java
GraphTraversal<Vertex, Vertex> cluster(String... names)
```

Example usage:

```groovy
// All vertices in the graph from the kind-kubehound.local cluster
kh.cluster("kind-kubehound.local")

// All containers in the graph from the kind-kubehound.local cluster
kh.cluster("kind-kubehound.local").containers()
```

### Containers Step

Starts a traversal that finds all vertices with a "Container" label and optionally allows filtering of those vertices on the "name" property.

```java
GraphTraversal<Vertex, Vertex> containers(String... names)
```

Example usage:

```groovy
// All containers in the graph
kh.containers()

// All containers in the graph with name filter
kh.containers("elasticsearch", "mongo")

// All containers in the graph with additional filters
kh.containers().has("namespace", "ns1").limit(10)
```

### Pods Step

Starts a traversal that finds all vertices with a "Pod" label and optionally allows filtering of those vertices on the "name" property.

```java
GraphTraversal<Vertex, Vertex> pods(String... names)
```

Example usage:

```groovy
// All pods in the graph
kh.pods()

// All pod in the graph with name filter
kh.pods("app-pod", "sidecar-pod")

// All pods in the graph with additional filters
kh.pods().has("namespace", "ns1").limit(10)
```

### Nodes Step

Starts a traversal that finds all vertices with a "Node" label and optionally allows filtering of those vertices on the "name" property.

```java
GraphTraversal<Vertex, Vertex> nodes(String... names)
```

Example usage:

```groovy
// All nodes in the graph
kh.nodes()

// All nodes in the graph with name filter
kh.nodes("control-plane")

// All nodes in the graph with additional filters
kh.nodes().has("team", "sre").limit(10)
```

### Escapes Step

Starts a traversal that finds all container escape edges from a Container vertex to a Node vertex and optionally allows filtering of those vertices on the "nodeNames" property.

```java
GraphTraversal<Vertex, Path> escapes(String... nodeNames)
```

Example usage:

```groovy
// All container escapes in the graph
kh.escapes()

// All container escapes in the graph with node name filter
kh.escapes("control-plane")
```

### Endpoints Step

Starts a traversal that finds all vertices with a "Endpoint" label.

```java
GraphTraversal<Vertex, Vertex> endpoints()
GraphTraversal<Vertex, Vertex> endpoints(EndpointExposure exposure)
```

Example usage:

```groovy
// All endpoints in the graph
kh.endpoints()

// All endpoints in the graph with additional filters
kh.endpoints().has("port", 3000).limit(10)

// All endpoints with K8s service exposure
kh.endpoints(EndpointExposure.External)
```

### Services Step

Starts a traversal that finds all vertices with a "Endpoint" label representing K8s services.

```java
GraphTraversal<Vertex, Vertex> services(String... portNames)
```

Example usage:

```groovy
// All services in the graph
kh.services()

// All services in the graph with name filter
kh.services("jmx", "redis")

// All services in the graph with additional filters
kh.services().has("port", 9999).limit(10)
```

### Volumes Step

Starts a traversal that finds all vertices with a "Volume" label and optionally allows filtering of those vertices on the "name" property.

```java
GraphTraversal<Vertex, Vertex> volumes(String... names)
```

Example usage:

```groovy
// All volumes in the graph
kh.volumes()

// All volumes in the graph with name filter
kh.volumes("db-data", "proc-mount")

// All volumes in the graph with additional filters
kh.volumes().has("sourcePath", "/").has("app", "web-app")
```

### HostMounts Step

Starts a traversal that finds all vertices representing volume host mounts and optionally allows filtering of those vertices on the "sourcePath" property.

```java
GraphTraversal<Vertex, Vertex> hostMounts(String... sourcePaths)
```

Example usage:

```groovy
// All host mounted volumes in the graph
kh.hostMounts()

// All host mount volumes in the graph with source path filter
kh.hostMounts("/", "/proc")

// All host mount volumes in the graph with additional filters
kh.hostMounts().has("app", "web-app").limit(10)
```

### Identities Step

Starts a traversal that finds all vertices with a "Identity" label and optionally allows filtering of those vertices on the "name" property.

```java
GraphTraversal<Vertex, Vertex> identities(String... names)
```

Example usage:

```groovy
// All identities in the graph
kh.identities()

// All identities in the graph with name filter
kh.identities("postgres-admin", "db-reader")

// All identities in the graph with additional filters
kh.identities().has("app", "web-app").limit(10)
```

### SAS Step

Starts a traversal that finds all vertices representing service accounts and optionally allows filtering of those vertices on the "name" property.

```java
GraphTraversal<Vertex, Vertex> sas(String... names)
```

Example usage:

```groovy
// All service accounts in the graph
kh.sas()

// All service accounts in the graph with name filter
kh.sas("postgres-admin", "db-reader")

// All service accounts in the graph with additional filters
kh.sas().has("app", "web-app").limit(10)
```

### Users Step

Starts a traversal that finds all vertices representing users and optionally allows filtering of those vertices on the "name" property.

```java
GraphTraversal<Vertex, Vertex> users(String... names)
```

Example usage:

```groovy
// All users in the graph
kh.users()

// All users in the graph with name filter
kh.users("postgres-admin", "db-reader")

// All users in the graph with additional filters
kh.users().has("app", "web-app").limit(10)
```

### Groups Step

Starts a traversal that finds all vertices representing groups and optionally allows filtering of those vertices on the "name" property.

```java
GraphTraversal<Vertex, Vertex> groups(String... names)
```

Example usage:

```groovy
// All groups in the graph
kh.groups()

// All groups in the graph with name filter
kh.groups("postgres-admin", "db-reader")

// All groups in the graph with additional filters
kh.groups().has("app", "web-app").limit(10)
```

### Permissions Step

Starts a traversal that finds all vertices with a "PermissionSet" label and optionally allows filtering of those vertices on the "role" property.

```java
GraphTraversal<Vertex, Vertex> permissions(String... roles)
```

Example usage:

```groovy
// All permissions sets in the graph
kh.permissions()

// All permissions sets in the graph with role filter
kh.permissions("postgres-admin", "db-reader")

// All permissions sets in the graph with additional filters
kh.permissions().has("app", "web-app").limit(10)
```

## Traversal Reference

### Attacks Step

From a Vertex traverse immediate edges to display the next set of possible attacks and targets

```java
GraphTraversal<S, Path> attacks()
```

Example usage:

```groovy
// All attacks possible from a specific container in the graph
kh.containers("pwned-container").attacks()
```

### Critical Step

From a Vertex filter on whether incoming vertices are critical assets

```java
GraphTraversal<S, E> critical()
```

Example usage:

```groovy
// All critical assets in the graph
kh.V().critical()

// Check whether a specific permission set is marked as critical
kh.permissions("system::kube-controller").critical()
```

### CriticalPaths Step

From a Vertex traverse edges until {@code maxHops} is exceeded or a critical asset is reached and return all paths.

```java
GraphTraversal<S, Path> criticalPaths()
GraphTraversal<S, Path> criticalPaths(int maxHops)
```

Example usage:

```groovy
// All attack paths from services to a critical asset
kh.services().criticalPaths()

// All attack paths (up to 5 hops) from a compromised credential to a critical asset
kh.group("engineering").criticalPaths(5)
```

### CriticalPathsFilter Step

From a Vertex traverse edges EXCLUDING labels provided in `exclusions` until `maxHops` is exceeded or a critical asset is reached and return all paths.

```java
GraphTraversal<S, Path> criticalPathsFilter(int maxHops, String... exclusions)
```

Example usage:

```groovy
// All attack paths (up to 10 hops) from services to a critical asset excluding the TOKEN_BRUTEFORCE and TOKEN_LIST attacks
kh.services().criticalPathsFilter(10, "TOKEN_BRUTEFORCE", "TOKEN_LIST")
```

### HasCriticalPath Step

From a Vertex filter on whether incoming vertices have at least one path to a critical asset

```java
GraphTraversal<S, E> hasCriticalPath()
```

Example usage:

```groovy
// All services with an attack path to a critical asset
kh.services().hasCriticalPath()
```

### MinHopsToCritical Step

From a Vertex returns the hop count of the shortest path to a critical asset.

```java
<E2 extends Comparable> GraphTraversal<S, E2> minHopsToCritical()
<E2 extends Comparable> GraphTraversal<S, E2> minHopsToCritical(int maxHops)
```

Example usage:

```groovy
// Shortest hops from a service to a critical asset
kh.services().minHopsToCritical()

// Shortest hops from a compromised engineer credential to a critical asset (up to 6)
kh.group("engineering").minHopsToCritical(6)
```

### CriticalPathsFreq Step

From a Vertex returns a group count (by label) of paths to a critical asset.

```java
<K> GraphTraversal<S, Map<K, Long>> criticalPathsFreq()
<K> GraphTraversal<S, Map<K, Long>> criticalPathsFreq(int maxHops)
```

Example usage:

```groovy
// Most common critical paths from services
kh.services().criticalPathsFreq()

// Most common critical paths from a compromised engineer credential of up to 4 hops
kh.group("engineering").criticalPathsFreq(4)
```

Sample output:

```json
{
  "path[Endpoint, ENDPOINT_EXPLOIT, Container, IDENTITY_ASSUME, Identity, PERMISSION_DISCOVER, PermissionSet]": 6,
  "path[Endpoint, ENDPOINT_EXPLOIT, Container, VOLUME_DISCOVER, Volume, TOKEN_STEAL, Identity, PERMISSION_DISCOVER, PermissionSet]": 6,
  "path[Endpoint, ENDPOINT_EXPLOIT, Container, CE_NSENTER, Node, IDENTITY_ASSUME, Identity, PERMISSION_DISCOVER, PermissionSet]": 1,
  "path[Endpoint, ENDPOINT_EXPLOIT, Container, CE_MODULE_LOAD, Node, IDENTITY_ASSUME, Identity, PERMISSION_DISCOVER, PermissionSet]": 1,
  "path[Endpoint, ENDPOINT_EXPLOIT, Container, CE_PRIV_MOUNT, Node, IDENTITY_ASSUME, Identity, PERMISSION_DISCOVER, PermissionSet]": 1
}
```
