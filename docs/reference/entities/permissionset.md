# PermissionSet

A permission set represents a Kubernetes RBAC `Role` or `ClusterRole`, which contain rules that represent a set of permissions that has been bound to an identity via a `RoleBinding` or `ClusterRoleBinding`. Permissions are purely additive (there are no "deny" rules).

## Properties

| Property | Type       | Description                                                                                                                                                                                       |
| -------- | ---------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| name     | `string`   | Name of the underlying role in Kubernetes                                                                                                                                                         |
| rules    | `[]string` | List of strings representing the access granted by the role (see generator function [flattenPolicyRules](https://github.com/DataDog/KubeHound/tree/main/pkg/kubehound/models/converter/graph.go)) |

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

[vertex.PermissionSet](https://github.com/DataDog/KubeHound/tree/main/pkg/kubehound/models/graph/permission_set.go)

## References

+ [Official Kubernetes documentation](https://kubernetes.io/docs/reference/access-authn-authz/rbac/) 
