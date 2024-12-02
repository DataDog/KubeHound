# Identity

Identity represents a Kubernetes user or service account.

## Properties

| Property | Type     | Description                                    |
| -------- | -------- | ---------------------------------------------- |
| name     | `string` | Name of the identity principal in Kubernetes   |
| type     | `string` | Type of identity (user, serviceaccount, group) |

## Common Properties

+ [app](./common.md#ownership-information)
+ [cluster](./common.md#run-information)
+ [critical](./common.md#risk-information)
+ [isNamespaced](./common.md#namespace-information)
+ [namespace](./common.md#namespace-information)
+ [runID](./common.md#run-information)
+ [service](./common.md#ownership-information)
+ [storeID](./common.md#store-information)
+ [team](./common.md#ownership-information)

## Definition

[vertex.Identity](https://github.com/DataDog/KubeHound/tree/main/pkg/kubehound/models/graph/identity.go)

## References

+ [Official Kubernetes documentation I: Authorization Overview](https://kubernetes.io/docs/reference/access-authn-authz/authorization/) 
+ [Official Kubernetes documentation II: RBAC](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#subject-v1-rbac-authorization-k8s-io)
