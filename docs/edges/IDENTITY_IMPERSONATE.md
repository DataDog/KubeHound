# IDENTITY_IMPERSONATE

With a [user impersonation privilege](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#user-impersonation) an attacker can impersonate a more privileged account.

| Source                                    | Destination                           | MITRE                            |
| ----------------------------------------- | ------------------------------------- |----------------------------------|
| [PermissionSet](../vertices/PERMISSIONSET.md)  | [Identity](../vertices/IDENTITY.md) | [Valid Accounts, T1078](https://attack.mitre.org/techniques/T1078/) |

## Details

Obtaining the `impersonate users/groups` permission will allow an attacker to execute K8s API actions on behalf of another user, including `cluster-admin` and other highly privileged users.

## Prerequisites

Ability to interrogate the K8s API with a role allowing impersonate access to users and/or groups.

See the [example pod spec](../../test/setup/test-cluster/attacks/IDENTITY_IMPERSONATE.yaml).

## Checks

Simply ask kubectl:

```bash
kubectl auth can-i impersonate users
kubectl auth can-i impersonate groups
```

## Exploitation

Execute any action in the K8s API impersonating a privileged group (e.g `system:masters`) or user using the syntax:

```bash
$ kubectl <verb> <noun> –as=null –as-group=system:masters -o json | jq
```

## Defences

### Monitoring

+ Monitoring the follow-on activity from user impersonation may be a more fruitful endeavour.

### Implement least privilege access

Impersonating users is a very powerful privilege and should not be required by the majority of users. Use an automated tool such a KubeHound to search for any risky permissions and users in the cluster and look to eliminate them.

## Calculation

+ [IdentityImpersonate](../../pkg/kubehound/graph/edge/identity_impersonate.go)
+ [IdentityImpersonateNamespace](../../pkg/kubehound/graph/edge/identity_impersonate_namespace.go)

## References:

+ [Official Kubernetes Documentation: Authenticating](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#user-impersonation)
+ [Securing Kubernetes Clusters by Eliminating Risky Permissions](https://www.cyberark.com/resources/threat-research-blog/securing-kubernetes-clusters-by-eliminating-risky-permissions)
