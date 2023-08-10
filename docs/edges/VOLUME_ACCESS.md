# VOLUME_ACCESS

Represents an attacker with access to a node filesystem gaining access to any volumes mounted inside a container (by definition).

| Source                                    | Destination                           | MITRE                            |
| ----------------------------------------- | ------------------------------------- |----------------------------------|
| [Node](../vertices/CONTAINER.md) | [Volume](../vertices/VOLUME.md) | [Container and Resource Discovery, T1613](https://attack.mitre.org/techniques/T1613/) |

## Details

Volumes can contains K8s API tokens or other resources useful to an attacker in building an attack path. This edge represents the link between a node and a mounted volume such that access to a node yields access to all exposed volumes for attack path calculations.

## Prerequisites

A volume exposed to a container. Currently supports `HostPath` and `Projected` volume types.

## Checks

None.

## Exploitation

No exploitation is necessary. This edge simply indicates that a volume is exposed to a container and accessible from the node filesystem

## Defences

None

## Calculation

+ [VolumeAccess](../../pkg/kubehound/graph/edge/volume_access.go)

## References:

+ [Official Kubernetes documentation: Volumes ](https://kubernetes.io/docs/concepts/storage/volumes/)
