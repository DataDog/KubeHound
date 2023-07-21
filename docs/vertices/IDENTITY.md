# Identity

Identity represents a Kubernetes user or service account.

## Properties

| Property            | Type      | Description |
| ----------------| --------- |----------------------------------------|
| name | `string` |  Name of the identity principal in Kubernetes |  
| type | `string` |  Type of identity (user, serviceaccount, etc) |  

## Common Properties

+ [storeID](./COMMON.md#store-information)
+ [app](./COMMON.md#ownership-information)
+ [team](./COMMON.md#ownership-information)
+ [service](./COMMON.md#ownership-information)
+ [critical](./COMMON.md#risk-information)
+ [namespace](./COMMON.md#namespace-information)
+ [isNamespaced](./COMMON.md#namespace-information)

## Definition

[vertex.Identity](../../pkg/kubehound/models/graph/identity.go)

## References

+ [Official Kubernetes documentation I: Authorization Overview](https://kubernetes.io/docs/reference/access-authn-authz/authorization/) 
+ [Official Kubernetes documentation II: RBAC](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#subject-v1-rbac-authorization-k8s-io)
