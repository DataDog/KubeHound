# Metrics

Aside from the obvious offensive use cases, KubeHound is most useful in calculating quantitative risk metrics representing the security posture of a cluster. The original goal of the project was to enable answering the following questions:

+ What is the shortest exploitable path between an external service and a critical asset?
+ What percentage of external services have an exploitable path to a critical asset?
+ What type of control would cut off the largest number of attack paths to a critical asset in our clusters?
+ What percentage level of attack path reduction could be achieved by the introduction of a control?

This section provides a short cheatsheet of gremlin queries to answer each of these in turn

## What is the shortest exploitable path between an external service and a critical asset?

```groovy
kh.services().minHopsToCritical()
```

## What percentage of external services have an exploitable path to a critical asset?

```groovy
// number of services with a path to a critical asset = N
kh.services().hasCriticalPath().count()

// total number of services = D
kh.services().count()

// Answer = N / D
```

## What type of control would cut off the largest number of attack paths to a critical asset in our clusters?

In this example an infrastructure team is prioritising new security mitigations for a cluster. Based on a shortest path analysis using `minHopsToCritical` they are looking to prioritise attack paths of 4 hops. What are the most common paths here? Are there any common patterns that emerge? This can be surfaced via the `criticalPathsFreq` DSL method:

```groovy
kh.services().criticalPathsFreq(4)
```

Running against the cluster generates the following attack path grouping:

```json
{
  "path[Endpoint, ENDPOINT_EXPLOIT, Container, IDENTITY_ASSUME, Identity, PERMISSION_DISCOVER, PermissionSet]" : 6,
  "path[Endpoint, ENDPOINT_EXPLOIT, Container, VOLUME_DISCOVER, Volume, TOKEN_STEAL, Identity, PERMISSION_DISCOVER, PermissionSet]" : 6,
  "path[Endpoint, ENDPOINT_EXPLOIT, Container, CE_NSENTER, Node, IDENTITY_ASSUME, Identity, PERMISSION_DISCOVER, PermissionSet]" : 1,
  "path[Endpoint, ENDPOINT_EXPLOIT, Container, CE_MODULE_LOAD, Node, IDENTITY_ASSUME, Identity, PERMISSION_DISCOVER, PermissionSet]" : 1,
  "path[Endpoint, ENDPOINT_EXPLOIT, Container, CE_PRIV_MOUNT, Node, IDENTITY_ASSUME, Identity, PERMISSION_DISCOVER, PermissionSet]" : 1
}
```

From this analysis it appears that the top attack paths are coming from containers with external services running with service accounts with `critical` permissions which look like strong candidates for immediate remediation!


## What percentage level of attack path reduction could be achieved by the introduction of a control?

In this example a threat detection team is considering implementing detections and auto remediation on secret access that would completely mitigate the `TOKEN_BRUTEFORCE` and `TOKEN_LIST` attacks, but the work is resource intensive. Is it worth the effort? A good measure of the impact would be to evaluate the attack path reduction as a result of this change:

```groovy
// number of attack paths from service endpoints excluding the mitigated attack = A
kh.services().criticalPathsFilter(10, "TOKEN_BRUTEFORCE", "TOKEN_LIST").count()

// total number of attack paths from service endpoints = B
kh.V().criticalPaths().count()

// Answer = (A-B)/A
```