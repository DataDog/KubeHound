---
title: PERMISSION_DISCOVER
---

<!--
id: PERMISSION_DISCOVER
name: "Enumerate permissions"
mitreAttackTechnique: T1069 - Permission Groups Discovery
mitreAttackTactic: TA0007 - Discovery
-->

# PERMISSION_DISCOVER

Represents the permissions granted to an identity that can be discovered by an attacker.

| Source                                    | Destination                           | MITRE                            |
| ----------------------------------------- | ------------------------------------- |----------------------------------|
| [Identity](../entities/identity.md) | [PermissionSet](../entities/permissionset.md) | [Permission Groups Discovery, T1069](https://attack.mitre.org/techniques/T1078/) |

## Details

K8s RBAC aggregates sets of API permissions together under `Role` (namespaced) and `ClusterRole` (cluster-wide) objects. These are then assigned to specific users via a `RoleBinding` (namespaced) or `ClusterRoleBinding` (cluster-wide) objects. This edge represents this relationship granting one or more permissions to an identity, which can be discovered by an attacker.

## Prerequisites

None

## Checks

A full list of identity → role mappings can be retrieved via:

```bash
kubectl get rolebindings,clusterrolebindings --all-namespaces -o wide  
```

To discover the permissions of the current identity use:

```bash
kubectl auth can-i --list
```

## Exploitation

No exploitation is necessary. This edge simply indicates that an identity grants a specific set of permissions (effectively represents a `RoleBinding` or `ClusterRoleBinding` in K8s).

## Defences

None

## Calculation

+ [PermissionDiscover](https://github.com/DataDog/KubeHound/tree/main/pkg/kubehound/graph/edge/permission_discover.go)

## References:

+ [Official Kubernetes Documentation:Using RBAC Authorization](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#rolebinding-and-clusterrolebinding)
+ [Kubernetes RBAC Details](https://octopus.com/blog/k8s-rbac-roles-and-bindings)

