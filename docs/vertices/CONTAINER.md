# Container

A container image running on a Kubernetes pod. Containers in a Pod are co-located and co-scheduled to run on the same node.

NOTE: properties that are interesting to attackers can be set at a Pod level such as hostPid, or container level such a capabilities. To simplify the graph model, the container node is chosen as the single source of truth for all host security related information. Any capabilities derived from the containing Pod are set ONLY on the container (and inheritance/override rules applied)

## Properties

| Property            | Type      | Description |
| ----------------| --------- |----------------------------------------|
| name | `string` |  Name of the container in Kubernetes | 
| image | `string` |  Docker the image run by the container | 
| command | `[]string` |  The container entrypoint| 
| args | `[]string` |  list of arguments passed to the container | 
| capabilities | `[]string` |  list of additional [capabilities](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/#set-capabilities-for-a-container) added to the container via k8s securityContext | 
| privileged | `bool` |  Whether the container is run in [privileged](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podsecuritycontext-v1-core) mode | 
| privesc | `bool` | Whether the container can gain more privileges than its parent process [details here](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podsecuritycontext-v1-core) | 
| hostPid | `bool` |  Whether the container can access the host’s PID namespace | 
| hostIpc | `bool` |  Whether the container can access the host’s IPC namespace | 
| hostNetwork | `bool` |  Whether the container can access the host’s network namespace| 
| runAsUser | `int64` |  The user account the container is running under e.g 0 for root | 
| ports | `[]string` |  List of ports exposed by the container | 
| pod | `string` |  The name of the pod running the container | 
| node | `string` |  The name of the node running the container | 

## Common Properties

+ [storeID](./COMMON.md#store-information)
+ [app](./COMMON.md#ownership-information)
+ [team](./COMMON.md#ownership-information)
+ [service](./COMMON.md#ownership-information)
+ [compromised](./COMMON.md#risk-information)
+ [namespace](./COMMON.md#namespace-information)
+ [isNamespaced](./COMMON.md#namespace-information)

## Definition

[vertex.Container](../../pkg/kubehound/models/graph/container.go)

## References

+ Official Kubernetes documentation: [Containers](https://kubernetes.io/docs/concepts/containers/)
