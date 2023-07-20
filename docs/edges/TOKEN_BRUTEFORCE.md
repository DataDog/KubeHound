# TOKEN_BRUTEFORCE

An identity with a role that allows *get* on secrets (vs list) can potentially view all the serviceaccount tokens in a specific namespace or in the whole cluster (with ClusterRole).

| Source                                    | Destination                           | MITRE                            |
| ----------------------------------------- | ------------------------------------- |----------------------------------|
| [Role](../vertices/ROLE.md) | [Identity](../vertices/IDENTITY.md) | [Steal Application Access Token, T1528](https://attack.mitre.org/techniques/T1528/) |

## Details

An attacker in possession of a token with permission to read a secret cannot use this permission without knowing the secret’s full name. This permission is different from the listing secrets permission described in [TOKEN_LIST](./TOKEN_LIST.md). However it may be possible to extract secrets via bruteforce for all K8s serviceaccounts due to their predictable naming convention.

## Prerequisites

Ability to interrogate the K8s API with a role allowing get access to secrets.

See the [example pod spec](../../test/setup/test-cluster/attacks/TOKEN_BRUTEFORCE.yaml).

## Checks

Simply ask kubectl:

```bash
kubectl auth can-i get secrets
```

## Exploitation

Exploitation of this vulnerability can be complex and time-consuming. See the [original research](https://www.cyberark.com/resources/threat-research-blog/securing-kubernetes-clusters-by-eliminating-risky-permissions) for a detailed description for the steps required.

## Defences

### Monitoring

+ Monitor anomalous access to the secrets API including listing all secrets, unusual User-Agent headers and other outliers.
+ Alert on anomalous volume of requests to the secrets API in a short time period.

### Implement least privilege access

Even *get* on secrets is a very powerful privilege and should not be required by the majority of users. Use an automated tool such a KubeHound to search for any risky permissions and users in the cluster and look to eliminate them.

## Calculation

[TokenBruteforce](../../pkg/kubehound/graph/edge/token_bruteforce.go)
[TokenBruteforceNamespace](../../pkg/kubehound/graph/edge/token_bruteforce_namespace.go)

## References:

+ [Official Kubernetes documentation: Secrets](https://kubernetes.io/docs/concepts/configuration/secret/#working-with-secrets)
+ [Securing Kubernetes Clusters by Eliminating Risky Permissions](https://www.cyberark.com/resources/threat-research-blog/securing-kubernetes-clusters-by-eliminating-risky-permissions)
