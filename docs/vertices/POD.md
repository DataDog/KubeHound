# Pod

A Kubernetes pod - the smallest deployable units of computing that you can create and manage in Kubernetes.

## Properties

| Property            | Type      | Description |
| ----------------| --------- |----------------------------------------|
| name | `string` |  Name of the pod in Kubernetes |  
| sharedProcessNamespace | `bool` |  whether all the containers in the pod share a process namespace (details [here](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#pod-v1-core)) |  
| serviceAccount | `string` |  The name of the `serviceaccount` used to run this pod. See [Kubernetes documentation](https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/) for further details |  
| node | `string` |  The name of the node running the pod |  

## Common Properties

+ [storeID](./COMMON.md#store-information)
+ [app](./COMMON.md#ownership-information)
+ [team](./COMMON.md#ownership-information)
+ [service](./COMMON.md#ownership-information)
+ [compromised](./COMMON.md#risk-information)
+ [critical](./COMMON.md#risk-information)
+ [namespace](./COMMON.md#namespace-information)
+ [isNamespaced](./COMMON.md#namespace-information)

## Definition

[vertex.Pod](../../pkg/kubehound/models/graph/pod.go)

## References

+ [Official Kubernetes documentation](https://kubernetes.io/docs/concepts/workloads/pods/) 

