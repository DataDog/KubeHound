# Node

A Kubernetes node. Kubernetes runs workloads by placing containers into Pods to run on Nodes. A node may be a virtual or physical machine, depending on the cluster.

## Properties

| Property            | Type      | Description |
| ----------------| --------- |----------------------------------------|
| name | `string` |  Name of the node in Kubernetes |  

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

[vertex.Node](../../pkg/kubehound/models/graph/node.go)

## References

+ [Official Kubernetes documentation](https://kubernetes.io/docs/concepts/architecture/nodes/) 

