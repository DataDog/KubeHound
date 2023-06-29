# KubeHound Cheat Sheet

## General queries

Count the number of pods in the cluster:

```groovy
g.V().hasLabel("Pod").count()
```

Count the number of container escapes in the cluster:

```groovy
g.V().hasLabel("Container").out().TODO TODO
```

List the possible attacks in a cluster:

```groovy
TODO TODO
```

## Basic path queries

All paths between a volume and identity:

```groovygit s
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

### Credentials

All attack paths (up to 6 hops) from any compomised credential to a critical asset:

```groovy
g.V().hasLabel("Identity").repeat(out().simplePath()).until(has("critical", true).or().loops().is(6)).has("critical", true).path()
```

Attack paths (up to 10 hops) from a known breached credential (in this case the `pod-patch-sa` service account) to a critical asset:

```grovy
g.V().hasLabel("Identity").has("name", "pod-patch-sa").repeat(out().simplePath()).until(has("critical", true).or().loops().is(10)).has("critical", true).path()
```
## Critical asset exposurer

All attack paths (up to 5 hops) to a specific critical asset (in this case the system:auth-delegator) role from containers/identities/nodes:

```groovy
g.V().hasLabel("Container", "Identity", "Node").repeat(out().simplePath()).until(has("name", "system:auth-delegator").or().loops().is(5)).has("name", "system:auth-delegator").hasLabel("Role").path()
```


## Threat modelling

All unique attack paths by labels to a specific asset (in this case the `system:auth-delegator` role):

```groovy
g.V().hasLabel("Container", "Identity").repeat(out().simplePath()).until(has("name", "system:auth-delegator").or().loops().is(5)).has("name", "system:auth-delegator").hasLabel("Role").path().as("p").by(label).dedup().select("p").path()
```

All unique attack paths by labels to a ANY critical asset:

```groovy
g.V().hasLabel("Container", "Identity").repeat(out().simplePath()).until(has("critical", true).or().loops().is(5)).has("critical", true).path().as("p").by(label).dedup().select("p").path()
```


## Risk metrics

A) What is the shortest exploitable path between an internet facing service and a specific asset?

B) What percentage of internet facing services have an exploitable path to a specific asset?

C) What type of control would cut off the largest number of attack paths to a specific asset in our clusters?

D) What percentage level of attack path reduction was achieved by the introduction of a control?

(B(after) - B(before)) / B(before)
