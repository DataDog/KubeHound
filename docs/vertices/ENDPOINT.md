# Endpoint

TODO

## Properties

| Property            | Type      | Description |
| ----------------| --------- |----------------------------------------|
| name | `string` | TODO | 
| addressType | `string` |  Type of the addresses array (IPv4, IPv6, etc) | 
| addresses | `string` |  Array of addresses exposing the endpoint | 
| port | `int` | Exposed port of the endpoint |
| portName | `string` | Name of the exposed port  |
| protocol | `string` | Endpoint protocol (TCP, UDP, etc) |
| exposure | `string` | Enum value describing the level of exposure of the endpoint (see [EndpointExposureType](../../pkg/kubehound/models/shared/constants.go))  |


TODO TODO

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

+ Official Kubernetes documentation: [Containers](https://kubernetes.io/docs/concepts/containers/)
