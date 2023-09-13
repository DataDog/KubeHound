# Metrics

TODO

##

``` java title="What is the shortest exploitable path between an exposed service and a critical asset?"
g.V().hasLabel("Endpoint").has("exposure", gte(3)).repeat(out().simplePath()).until(has("critical", true).or().loops().is(7)).has("critical", true).path().count(local).min()
```

``` java title="What percentage of external facing services have an exploitable path to a critical asset?"
// Leveraging the "EndpointExposureType" enum value to filter only on services
// c.f. https://github.com/DataDog/KubeHound/blob/main/pkg/kubehound/models/shared/constants.go

// Base case
g.V().hasLabel("Endpoint").has("exposure", gte(3)).count()

// Has a critical path
g.V().hasLabel("Endpoint").has("exposure", gte(3)).where(repeat(out().simplePath()).until(has("critical", true).or().loops().is(10)).has("critical", true).limit(1)).count()
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
    "path[Container, IDENTITY_ASSUME, Identity, PERMISSION_DISCOVER, PermissionSet, TOKEN_LIST, Identity, PERMISSION_DISCOVER, PermissionSet, TOKEN_LIST, Identity, PERMISSION_DISCOVER, PermissionSet]" : 191,
    "path[Container, CE_SYS_PTRACE, Node, VOLUME_EXPOSE, Volume, TOKEN_STEAL, Identity, PERMISSION_DISCOVER, PermissionSet, TOKEN_LIST, Identity, PERMISSION_DISCOVER, PermissionSet]" : 48,
    "path[Container, IDENTITY_ASSUME, Identity, PERMISSION_DISCOVER, PermissionSet, TOKEN_BRUTEFORCE, Identity, PERMISSION_DISCOVER, PermissionSet, TOKEN_LIST, Identity, PERMISSION_DISCOVER, PermissionSet]" : 48,
    ...
  }
*/
```