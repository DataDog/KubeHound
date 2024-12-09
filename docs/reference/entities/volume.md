# Volume

Volume represents a volume mounted in a container and exposed by a node.

## Properties

| Property   | Type     | Description                                                                                                                                                                |
| ---------- | -------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| name       | `string` | Name of the volume mount in the container spec                                                                                                                             |
| type       | `string` | Type of volume mount (host/projected/etc). See [Kubernetes documentation](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#volume-v1-core) for details |
| sourcePath | `string` | The path of the volume in the host (i.e node) filesystem                                                                                                                   |
| mountPath  | `string` | The path of the volume in the container filesystem                                                                                                                         |
| readonly   | `bool`   | Whether the volume has been mounted with `readonly` access                                                                                                                 |

## Common Properties

+ [app](./common.md#ownership-information)
+ [cluster](./common.md#run-information)
+ [isNamespaced](./common.md#namespace-information)
+ [namespace](./common.md#namespace-information)
+ [runID](./common.md#run-information)
+ [service](./common.md#ownership-information)
+ [storeID](./common.md#store-information)
+ [team](./common.md#ownership-information)

## Definition

[vertex.Volume](https://github.com/DataDog/KubeHound/tree/main/pkg/kubehound/models/graph/volume.go)

## References

+ [Official Kubernetes documentation](https://kubernetes.io/docs/concepts/storage/volumes/) 
