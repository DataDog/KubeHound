# Queries

You can query KubeHound data stored in the JanusGraph database by using the [Gremlin query language](https://docs.janusgraph.org/getting-started/gremlin/).

## Basic queries

``` java title="Count the number of pods in the cluster"
g.V().hasLabel("Pod").count()
```

``` java title="View all possible container escapes in the cluster"
g.V().hasLabel("Container").outE().inV().hasLabel("Node").path()
```

``` java title="List the names of all possible attacks in the cluster"
g.E().groupCount().by(label)
```

``` java title="View all the mounted host path volumes in the cluster"
g.V().hasLabel("Volume").has("type", "HostPath").groupCount().by("sourcePath")
```

``` java title="View host path mounts that can be exploited to escape to a node"
g.E().hasLabel("EXPLOIT_HOST_READ", "EXPLOIT_HOST_WRITE").outV().groupCount().by("sourcePath")
```

``` java title="View all service endpoints by service name in the cluster"
// Leveraging the "EndpointExposureType" enum value to filter only on services
// c.f. https://github.com/DataDog/KubeHound/blob/main/pkg/kubehound/models/shared/constants.go
g.V().hasLabel("Endpoint").has("exposure", 3).groupCount().by("serviceEndpoint")
```

## Basic attack paths

``` java title="TODO-RELEVANT All paths between a volume and an identity"
g.V().hasLabel("Volume").repeat(out().simplePath()).until(hasLabel("Identity")).path()
```

``` java title="All paths (up to 5 hops) between a container and a node"
g.V().hasLabel("Container").repeat(out().simplePath()).until(hasLabel("Node").or().loops().is(5)).hasLabel("Node").path()
```

``` java title="All attack paths (up to 6 hops) from any compomised identity (e.g. service account) to a critical asset"
g.V().hasLabel("Identity").repeat(out().simplePath()).until(has("critical", true).or().loops().is(6)).has("critical", true).path().limit(5)
```

TODO more?

## Attack paths from compromised assets 

### Containers

``` java title="Attack paths (up to 10 hops) from a known breached container to any critical asset"
g.V().hasLabel("Container").has("name", "nsenter-pod").repeat(out().simplePath()).until(has("critical", true).or().loops().is(10)).has("critical", true).path()
```

``` java title="Attack paths (up to 10 hops) from a known backdoored container image to any critical asset"
g.V().hasLabel("Container").has("image", TextP.containing("malicious-image")).repeat(out().simplePath()).until(has("critical", true).or().loops().is(10)).has("critical", true).path()
```

### Credentials

``` java title="Attack paths (up to 10 hops) from a known breached identity to a critical asset"
g.V().hasLabel("Identity").has("name", "compromised-sa").repeat(out().simplePath()).until(has("critical", true).or().loops().is(10)).has("critical", true).path()
```

### Endpoints

TODO: how do we define an endpoint

``` java title="Attack paths (up to 6 hops) from any endpoint to a critical asset:"
g.V().hasLabel("Endpoint").repeat(out().simplePath()).until(has("critical", true).or().loops().is(6)).has("critical", true).path().limit(5)
```

``` java title="Attack paths (up to 10 hops) from a known risky endpoint (e.g JMX) to a critical asset"
g.V().hasLabel("Endpoint").has("portName", "jmx").repeat(out().simplePath()).until(has("critical", true).or().loops().is(6)).has("critical", true).path().limit(5)
```

## Risk assessment

``` java title="What is the shortest exploitable path between an exposed service and a critical asset?"
g.V().hasLabel("Endpoint").has("exposure", gte(3)).repeat(out().simplePath()).until(has("critical", true).or().loops().is(7)).has("critical", true).path().count(local).min()
```

``` java title="What percentage of internet facing services have an exploitable path to a critical asset?"
// Leveraging the "EndpointExposureType" enum value to filter only on services
// c.f. https://github.com/DataDog/KubeHound/blob/main/pkg/kubehound/models/shared/constants.go

// Base case
g.V().hasLabel("Endpoint").has("exposure", gte(3)).count()

// Has a critical path
g.V().hasLabel("Endpoint").has("exposure", gte(3)).where(repeat(out().simplePath()).until(has("critical", true).or().loops().is(10)).has("critical", true).limit(1)).count()
```

## CVE impact assessment

You can also use KubeHound to determine if workloads in your cluster may be vulnerable to a specific vulnerability.

First, evaluate if a known vulnerable image is running in the cluster:

```java
g.V().hasLabel("Container").has("image", TextP.containing("elasticsearch")).groupCount().by("image")
```

Then, check any exposed services that could be affected and have a path to a critical asset. This helps prioritizing patching and remediation.

```java
g.V().hasLabel("Container").has("image", "dockerhub.com/elasticsearch:7.1.4").where(inE("ENDPOINT_EXPOSE").outV().has("exposure", gte(3))).where(repeat(out().simplePath()).until(has("critical", true).or().loops().is(10)).has("critical", true).limit(1))
```

## Assessing the value of implementing new security controls

To verify concrete impact, this can be achieved by comparing the difference in the key risk metrics above, before and after the control change. To simulate the impact of introducing a control (e.g to evaluate ROI), we can add conditions to our path queries. For example if we wanted to evaluate the impact of adding a gatekeeper rule that would deny the use of `hostPID` we can use the following:

``` java title="What percentage level of attack path reduction was achieved by the introduction of a control?"
// Calculate the base case
g.V().hasLabel("Endpoint").has("exposure", gte(3)).repeat(out().simplePath()).until(has("critical", true).or().loops().is(6)).has("critical", true).path().count()

// Calculate the impact of preventing CE_NSENTER attack
g.V().hasLabel("Endpoint").has("exposure", gte(3)).repeat(outE().not(hasLabel("CE_NSENTER")).inV().simplePath()).emit().until(has("critical", true).or().loops().is(6)).has("critical", true).path().count()
```

``` java title="What type of control would cut off the largest number of attack paths to a specific asset in the cluster?"
// We count the number of instances of unique attack paths using
g.V().hasLabel("Container").repeat(outE().inV().simplePath()).emit()
.until(has("critical", true).or().loops().is(6)).has("critical", true)
.path().by(label).groupCount().order(local).by(select(values), desc)

/* Sample output:

  {
    "path[Container, IDENTITY_ASSUME, Identity, ROLE_GRANT, Role, TOKEN_LIST, Identity, ROLE_GRANT, Role, TOKEN_LIST, Identity, ROLE_GRANT, Role]" : 191,
    "path[Container, CE_SYS_PTRACE, Node, VOLUME_EXPOSE, Volume, TOKEN_STEAL, Identity, ROLE_GRANT, Role, TOKEN_LIST, Identity, ROLE_GRANT, Role]" : 48,
    "path[Container, IDENTITY_ASSUME, Identity, ROLE_GRANT, Role, TOKEN_BRUTEFORCE, Identity, ROLE_GRANT, Role, TOKEN_LIST, Identity, ROLE_GRANT, Role]" : 48,
    ...
  }
*/
```

## Threat modelling

``` java title="All unique attack paths by labels to a specific asset (here, the cluster-admin role)"
g.V().hasLabel("Container", "Identity")
.repeat(out().simplePath())
.until(has("name", "cluster-admin").or().loops().is(5))
.has("name", "cluster-admin").hasLabel("Role").path().as("p").by(label).dedup().select("p").path()
```

``` java title="All unique attack paths by labels to a any critical asset"
g.V().hasLabel("Container", "Identity")
.repeat(out().simplePath())
.until(has("critical", true).or().loops().is(5))
.has("critical", true).path().as("p").by(label).dedup().select("p").path()
```

## Tips for writing queries

To get started with Gremlin, have a look at the following tutorials:

- [Gremlin basics](https://dkuppitz.github.io/gremlin-cheat-sheet/101.html) by Daniel Kuppitz
- [Gremlin advanced](https://dkuppitz.github.io/gremlin-cheat-sheet/102.html) by Daniel Kuppitz

For large clusters it is recommended to add a `limit()` step to **all** queries where the graph output will be examined in the UI to prevent overloading it. An example looking for attack paths possible from a sample of 5 containers would look like:

```go
g.V().hasLabel("Container").limit(5).outE()
```

Additional tips:
- For queries to be displayed in the UI, try to limit the output to 1000 elements or less
- Enable [large cluster optimizations](TODO) if queries are returning too slowly
- Try to filter the initial element of queries by namespace/service/app to avoid generating too many results, for instance `g.V().hasLabel("Container").has("namespace", "your-namespace")`
