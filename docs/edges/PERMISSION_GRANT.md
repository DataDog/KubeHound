# PERMISSION_GRANT

Represents the permissions granted to an identity.

| Source                                    | Destination                           | MITRE                            |
| ----------------------------------------- | ------------------------------------- |----------------------------------|
| [Identity](../vertices/IDENTITY.md) | [PermissionSet](../vertices/PERMISSIONSET.md) | [Valid Accounts, T1078](https://attack.mitre.org/techniques/T1078/) |

## Details

K8s RBAC aggregates sets of API permissions together under `Role` (namespaced) and `ClusterRole` (cluster-wide) objects. These are then assigned to specific users via a `RoleBinding` (namespaced) or `ClusterRoleBinding` (cluster-wide) objects. This edge represents this relationship granting one or more permissions to an identity.

## Prerequisites

None

## Checks

A full list of identity → role mappings can be retrieved via:

```bash
kubectl get rolebindings,clusterrolebindings --all-namespaces -o wide  
```

## Exploitation

No exploitation is necessary. This edge simply indicates that an identity grants a specific set of permissions (effectively represents a `RoleBinding` or `ClusterRoleBinding` in K8s).

## Defences

None

## Calculation

+ [PermissionGrant](../../pkg/kubehound/graph/edge/permission_grant.go)

## References:

+ [Official Kubernetes Documentation:Using RBAC Authorization](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#rolebinding-and-clusterrolebinding)
+ [Kubernetes RBAC Details](https://octopus.com/blog/k8s-rbac-roles-and-bindings)

