# Container

A container image running on a Kubernetes pod. Containers in a Pod are co-located and co-scheduled to run on the same node.

Properties that are interesting to attackers can be set at a Pod level such as hostPid, or container level such a capabilities. To simplify the graph model, the container node is chosen as the single source of truth for all host security related information. Any capabilities derived from the containing Pod are set ONLY on the container (and inheritance/override rules applied)

## Properties

| Property            | Type      | Description |
| ----------------| --------- |----------------------------------------|
| name | `string` |  Name of the container in Kubernetes | 
| image | `string` |  Docker the image run by the container | 
| command | `[]string` |  The container entrypoint| 
| args | `[]string` |  List of arguments passed to the container | 
| capabilities | `[]string` |  List of additional [capabilities](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/#set-capabilities-for-a-container) added to the container via k8s securityContext | 
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

+ [storeID](./common.md#store-information)
+ [app](./common.md#ownership-information)
+ [team](./common.md#ownership-information)
+ [service](./common.md#ownership-information)
+ [compromised](./common.md#risk-information)
+ [namespace](./common.md#namespace-information)
+ [isNamespaced](./common.md#namespace-information)

## Definition

[vertex.Container](https://github.com/DataDog/KubeHound/tree/main/pkg/kubehound/models/graph/container.go)

## References

+ Official Kubernetes documentation: [Containers](https://kubernetes.io/docs/concepts/containers/)
