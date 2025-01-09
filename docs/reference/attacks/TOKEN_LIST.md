---
title: TOKEN_LIST
---

<!--
id: TOKEN_LIST
name: "Access service account token secrets"
mitreAttackTechnique: T1528 - Steal Application Access Token
mitreAttackTactic: TA0006 - Credential Access
-->

# TOKEN_LIST

| Source                                        | Destination                         | MITRE ATT&CK                                                                        |
| --------------------------------------------- | ----------------------------------- | ----------------------------------------------------------------------------------- |
| [PermissionSet](../entities/permissionset.md) | [Identity](../entities/identity.md) | [Steal Application Access Token, T1528](https://attack.mitre.org/techniques/T1528/) |

An identity with a role that allows listing secrets can potentially view all the secrets in a specific namespace or in the whole cluster (with ClusterRole).

## Details

Obtaining the list secrets permission will be a significant advantage to an attacker. It may lead to disclosure of application credentials, SSH keys, other more privileged user’s tokens and more.  All of these can be used in different ways depending on their capabilities. For our graph model we focus on the latter case of extracting K8s tokens only.

## Prerequisites

Ability to interrogate the K8s API with a role allowing list access to secrets.

See the [example pod spec](https://github.com/DataDog/KubeHound/tree/main/test/setup/test-cluster/attacks/TOKEN_LIST.yaml).

## Checks

Simply ask kubectl:

```bash
kubectl auth can-i list secrets
```

## Exploitation

Simply dump all secrets using kubectl:

```bash
kubectl get secrets -o json | jq
``` 

## Defences

### Monitoring

+ Monitor anomalous access to the secrets API including listing all secrets, unusual User-Agent headers and other outliers.

### Implement least privilege access

Listing secrets is a very powerful privilege and should not be required by the majority of users. Use an automated tool such as KubeHound to search for any risky permissions and users in the cluster and look to eliminate them.

## Calculation

+ [TokenList](https://github.com/DataDog/KubeHound/tree/main/pkg/kubehound/graph/edge/token_list.go)
+ [TokenListNamespace](https://github.com/DataDog/KubeHound/tree/main/pkg/kubehound/graph/edge/token_list_namespace.go)

## References:

+ [Official Kubernetes documentation: Secrets](https://kubernetes.io/docs/concepts/configuration/secret/#working-with-secrets)
+ [Securing Kubernetes Clusters by Eliminating Risky Permissions](https://www.cyberark.com/resources/threat-research-blog/securing-kubernetes-clusters-by-eliminating-risky-permissions)
+ [Official Kubernetes documentation: List Secret Risks](https://kubernetes.io/docs/concepts/security/rbac-good-practices/#listing-secrets)
