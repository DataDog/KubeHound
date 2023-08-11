# Node

A Kubernetes node. Kubernetes runs workloads by placing containers into Pods to run on Nodes. A node may be a virtual or physical machine, depending on the cluster.

## Properties

| Property            | Type      | Description |
| ----------------| --------- |----------------------------------------|
| name | `string` |  Name of the node in Kubernetes |  

## Common Properties

+ [storeID](./common.md#store-information)
+ [app](./common.md#ownership-information)
+ [team](./common.md#ownership-information)
+ [service](./common.md#ownership-information)
+ [compromised](./common.md#risk-information)
+ [critical](./common.md#risk-information)
+ [namespace](./common.md#namespace-information)
+ [isNamespaced](./common.md#namespace-information)

## Definition

[vertex.Node](https://github.com/DataDog/KubeHound/tree/main/pkg/kubehound/models/graph/node.go)

## References

+ [Official Kubernetes documentation](https://kubernetes.io/docs/concepts/architecture/nodes/) 

