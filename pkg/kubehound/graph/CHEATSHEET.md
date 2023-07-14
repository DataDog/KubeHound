# KubeHound Cheat Sheet

## General queries

Count the number of pods in the cluster:

```groovy
g.V().hasLabel("Pod").count()
```

View all the possible container escapes in the cluster:

```groovy
g.V().hasLabel("Container").outE().inV().hasLabel("Node").path()
```

List the names of all possible attacks in a cluster with total count:

```groovy
g.E().groupCount().by(label)
```

## Basic path queries

All paths between a volume and identity:

```groovy
g.V().hasLabel("Volume").repeat(out().simplePath()).until(hasLabel("Identity")).path()
```

All paths (up to 5 hops) between a container and a node:

```groovy
g.V().hasLabel("Container").repeat(out().simplePath()).until(hasLabel("Node").or().loops().is(5)).hasLabel("Node").path()
```

## Attack paths from compromised assets

### Containers

All attack paths (up to 6 hops) from any container to a critical asset:

```groovy
g.V().hasLabel("Container").repeat(out().simplePath()).until(has("critical", true).or().loops().is(6)).has("critical", true).path()
```
Attack paths (up to 10 hops) from a known breached container (in this case the `nsenter-pod` container) to any critical asset:

```groovy
g.V().hasLabel("Container").has("name", "nsenter-pod").repeat(out().simplePath()).until(has("critical", true).or().loops().is(10)).has("critical", true).path()
```

Attack paths (up to 10 hops) from a known backdoored container image (in this case the `config-file-writer-go:` container) to any critical asset:

```groovy
g.V().hasLabel("Container").has("image", TextP.containing("eu.gcr.io/datadog-staging/config-file-writer-go")).repeat(out().simplePath()).until(has("critical", true).or().loops().is(10)).has("critical", true).path()
```


### Credentials

All attack paths (up to 6 hops) from any compomised credential to a critical asset:

```groovy
g.V().hasLabel("Identity").repeat(out().simplePath()).until(has("critical", true).or().loops().is(6)).has("critical", true).path()
```

Attack paths (up to 10 hops) from a known breached credential (in this case the `pod-patch-sa` service account) to a critical asset:

```groovy
g.V().hasLabel("Identity").has("name", "pod-patch-sa").repeat(out().simplePath()).until(has("critical", true).or().loops().is(10)).has("critical", true).path()
```
## Critical asset exposure

All attack paths (up to 5 hops) to a specific critical asset (in this case the `system:auth-delegator`) role from containers/identities/nodes:

```groovy
g.V().hasLabel("Container", "Identity", "Node").repeat(out().simplePath()).until(has("name", "system:auth-delegator").or().loops().is(5)).has("name", "system:auth-delegator").hasLabel("Role").path()
```

## Threat modelling

All unique attack paths by labels to a specific asset (in this case the `cluster-admin` role):

```groovy
g.V().hasLabel("Container", "Identity").repeat(out().simplePath()).until(has("name", "cluster-admin").or().loops().is(5)).has("name", "cluster-admin").hasLabel("Role").path().as("p").by(label).dedup().select("p").path()
```

All unique attack paths by labels to a ANY critical asset:

```groovy
g.V().hasLabel("Container", "Identity").repeat(out().simplePath()).until(has("critical", true).or().loops().is(5)).has("critical", true).path().as("p").by(label).dedup().select("p").path()
```

## Risk metrics

**What is the shortest exploitable path between an exposed service and a critical asset?**

In this case we can look for containers with specific properties e.g image/tag etc and query the minimum path size to reach a critical assets. In this example we use exposed ports as a proxy for a the container offering a service that can be exploited, in production consider filtering on tags/image/etc:

```groovy
g.V().hasLabel("Container").has("ports", neq([])).repeat(out().simplePath()).until(has("critical", true).or().loops().is(7)).has("critical", true).path().count(local).min()
```

**What percentage of internet facing services have an exploitable path to a critical asset?**

Again using exposed ports as a proxy for a the container offering a service that can be exploited:

```groovy
// Base case
g.V().hasLabel("Container").has("ports", neq([])).count()

// Has a critical path
g.V().hasLabel("Container").has("ports", neq([])).where(repeat(out().simplePath()).until(has("critical", true).or().loops().is(10)).has("critical", true)).count()
```

**What percentage level of attack path reduction was achieved by the introduction of a control?**

To verify concrete impact, this can be achieved by comparing the difference in the key risk metrics above, before and after the control change. To simulate the impact of introducing a control (e.g to evaluate ROI), we can add conditions to our path queries. For example if we wanted to evaluate the impact of adding a gatekeeper rule that would deny the use of `hostPid=true` we could do the following:

```groovy
// Calculate the base case
g.V().hasLabel("Container").repeat(out().simplePath()).until(has("critical", true).or().loops().is(6)).has("critical", true).path().count()

// Calculate the impact of preventing CE_NSENTER attack
g.V().hasLabel("Container").repeat(outE().not(hasLabel("CE_NSENTER")).inV().simplePath()).emit().until(has("critical", true).or().loops().is(6)).has("critical", true).path().count()
```

**What type of control would cut off the largest number of attack paths to a specific asset in our clusters?**

We can count the number of instances of unique attack paths using:

```groovy
g.V().hasLabel("Container").repeat(outE().inV().simplePath()).emit().until(has("critical", true).or().loops().is(6)).has("critical", true).path().by(label).groupCount()
```

This gives an output of the form:

```groovy
{
  "path[Container, CE_MODULE_LOAD, Node, POD_ATTACH, Pod, CONTAINER_ATTACH, Container, IDENTITY_ASSUME, Identity, ROLE_GRANT, Role]" : 18,
  "path[Container, IDENTITY_ASSUME, Identity, ROLE_GRANT, Role, TOKEN_BRUTEFORCE, Identity, ROLE_GRANT, Role, TOKEN_BRUTEFORCE, Identity, ROLE_GRANT, Role]" : 1824,
}
```

We can further reduce this to group by attacks, rather than full paths in post-processing or modifying the query.

## Tips

+ Always put a max hop count on path queries or runtime can get very long
+ For queries to be displayed in the UI, try to limit the output to 1000 elements or less
+ Enable large cluster optimizations if queries are returning too slowly