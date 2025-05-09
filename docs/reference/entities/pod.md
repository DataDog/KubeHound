# Pod

A Kubernetes pod - the smallest deployable units of computing that you can create and manage in Kubernetes.

## Properties

| Property              | Type     | Description                                                                                                                                                                                    |
| --------------------- | -------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| name                  | `string` | Name of the pod in Kubernetes                                                                                                                                                                  |
| shareProcessNamespace | `bool`   | whether all the containers in the pod share a process namespace (details [here](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#pod-v1-core))                             |
| serviceAccount        | `string` | The name of the `serviceaccount` used to run this pod. See [Kubernetes documentation](https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/) for further details |
| node                  | `string` | The name of the node running the pod                                                                                                                                                           |

## Common Properties

+ [app](./common.md#ownership-information)
+ [cluster](./common.md#run-information)
+ [compromised](./common.md#risk-information)
+ [critical](./common.md#risk-information)
+ [isNamespaced](./common.md#namespace-information)
+ [namespace](./common.md#namespace-information)
+ [runID](./common.md#run-information)
+ [service](./common.md#ownership-information)
+ [storeID](./common.md#store-information)
+ [team](./common.md#ownership-information)

## Definition

[vertex.Pod](https://github.com/DataDog/KubeHound/tree/main/pkg/kubehound/models/graph/pod.go)

## References

+ [Official Kubernetes documentation](https://kubernetes.io/docs/concepts/workloads/pods/) 
