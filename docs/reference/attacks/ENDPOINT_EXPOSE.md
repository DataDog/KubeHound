<<<<<<<< HEAD:docs/reference/attacks/ENDPOINT_EXPLOIT.md
# ENDPOINT_EXPLOIT
========

# ENDPOINT_EXPOSE
>>>>>>>> a625a6a (First version of docs website):docs/reference/attacks/ENDPOINT_EXPOSE.md

Represents a network endpoint exposed by a container that could be exploited by an attacker (via means known or unknown). This can correspond to a Kubernetes service, node service, node port, or container port.

| Source                                    | Destination                           | MITRE                            |
| ----------------------------------------- | ------------------------------------- |----------------------------------|
| [Endpoint](../entities/endpoint.md) | [Container](../entities/container.md) | [Network Service Discovery, T1046](https://attack.mitre.org/techniques/T1046/) |

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

+ [EndpointExploitInternal](https://github.com/DataDog/KubeHound/tree/main/pkg/kubehound/graph/edge/endpoint_exploit_internal.go)
+ [EndpointExploitExternal](https://github.com/DataDog/KubeHound/tree/main/pkg/kubehound/graph/edge/endpoint_exploit_external.go)


## References:

+ [Official Kubernetes documentation: EndpointSlices ](https://kubernetes.io/docs/concepts/storage/volumes/)

