# Endpoint

A network endpoint exposed by a container accessible via a Kubernetes service, external node port or cluster IP/port tuple.

## Properties

| Property            | Type      | Description |
| ----------------| --------- |----------------------------------------|
| name | `string` | Unique endpoint name | 
| serviceEndpoint | `string` | Name of the service if the endpoint is exposed outside the cluster via an endpoint slice | 
| serviceDns | `string` | FQDN of the service if the endpoint is exposed outside the cluster via an endpoint slice | 
| addressType | `string` |  Type of the addresses array (IPv4, IPv6, etc) | 
| addresses | `string` |  Array of addresses exposing the endpoint | 
| port | `int` | Exposed port of the endpoint |
| portName | `string` | Name of the exposed port  |
| protocol | `string` | Endpoint protocol (TCP, UDP, etc) |
| exposure | `string` | Enum value describing the level of exposure of the endpoint (see [EndpointExposureType](https://github.com/DataDog/KubeHound/tree/main/pkg/kubehound/models/shared/constants.go))  |


## Common Properties

+ [storeID](./common.md#store-information)
+ [app](./common.md#ownership-information)
+ [team](./common.md#ownership-information)
+ [service](./common.md#ownership-information)
+ [compromised](./common.md#risk-information)
+ [namespace](./common.md#namespace-information)
+ [isNamespaced](./common.md#namespace-information)

## Definition

[vertex.Endpoint](https://github.com/DataDog/KubeHound/tree/main/pkg/kubehound/models/graph/endpoint.go)

## References

+ [Official Kubernetes documentation](https://kubernetes.io/docs/concepts/services-networking/endpoint-slices/)
+ [Exposing Kubernetes Applications](https://cylab.be/blog/154/exposing-a-kubernetes-application-service-hostport-nodeport-loadbalancer-or-ingresscontroller)
