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
| exposure | `string` | Enum value describing the level of exposure of the endpoint (see [EndpointExposureType](../../pkg/kubehound/models/shared/constants.go))  |


## Common Properties

+ [storeID](./COMMON.md#store-information)
+ [app](./COMMON.md#ownership-information)
+ [team](./COMMON.md#ownership-information)
+ [service](./COMMON.md#ownership-information)
+ [compromised](./COMMON.md#risk-information)
+ [namespace](./COMMON.md#namespace-information)
+ [isNamespaced](./COMMON.md#namespace-information)

## Definition

[vertex.Endpoint](../../pkg/kubehound/models/graph/endpoint.go)

## References

+ [Official Kubernetes documentation](https://kubernetes.io/docs/concepts/services-networking/endpoint-slices/)
+ [Exposing Kubernetes Applications](https://cylab.be/blog/154/exposing-a-kubernetes-application-service-hostport-nodeport-loadbalancer-or-ingresscontroller)
