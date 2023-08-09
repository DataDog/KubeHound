# ENDPOINT_EXPOSE

Represents a network endpoint exposed by a container. This can correspond to a Kubernetes service, node service, node port, or container port.

| Source                                    | Destination                           | MITRE                            |
| ----------------------------------------- | ------------------------------------- |----------------------------------|
| [Endpoint](../vertices/ENDPOINT.md) | [Container](../vertices/CONTAINER.md) | [Network Service Discovery, T1046](https://attack.mitre.org/techniques/T1046/) |

## Details

Exposed endpoints represent the most common entry point for attackers into a cluster.

## Prerequisites

A network endpoint exposed by a container.

## Checks

Endpoints exposed outside the cluster can be queried via `kubectl`:

```bash
kubectl get endpointslices
```

Alternatively open ports can be discovered by traditional port scanning techniques or a tool like [KubeHunter](https://github.com/aquasecurity/kube-hunter#scanning-options)

## Exploitation

This edge simply indicates that an endpoint is exposed by a container. It does not signal that the endpoint is exploitable but serves as a useful starting point for path traversal queries.

## Defences

None

## Calculation

+ [EndpointExposeContainer](../../pkg/kubehound/graph/edge/endpoint_expose_container.go)
+ [EndpointExposeSlice](../../pkg/kubehound/graph/edge/endpoint_expose_slice.go)


## References:

+ [Official Kubernetes documentation: EndpointSlices ](https://kubernetes.io/docs/concepts/storage/volumes/)

