# ROLE_GRANT

Represents the roles granted by an identity.

| Source                                    | Destination                           | MITRE                            |
| ----------------------------------------- | ------------------------------------- |----------------------------------|
| [Identity](../vertices/IDENTITY.md) | [Role](../vertices/ROLE.md) | [Valid Accounts, T1078](https://attack.mitre.org/techniques/T1078/) |

## Details

K8s RBAC aggregates sets of API permissions together under `Role` (namespaced) and `ClusterRole` (cluster-wide) objects. These are then assigned to specific users via a `RoleBinding` (namespaced) or `ClusterRoleBinding` (cluster-wide) objects. This edge represents this relationship granting one or more roles to an identity.

## Prerequisites

None

## Checks

A full list of identity → role mappings can be retrieved via:

```bash
kubectl get rolebindings,clusterrolebindings --all-namespaces -o wide  
```

## Exploitation

No exploitation is necessary. This edge simply indicates that an identity grants a specific role (effectively represents a `RoleBinding` or `ClusterRoleBinding` in K8s).

## Defences

None

## Calculation

+ [RoleGrant](../../pkg/kubehound/graph/edge/role_grant.go)

## References:

+ [Official Kubernetes Documentation:Using RBAC Authorization](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#rolebinding-and-clusterrolebinding)

