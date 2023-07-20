# TOKEN_VAR_LOG_SYMLINK

Steal all K8s API tokens from a node via an exposed `/var/log` mount.

| Source                                    | Destination                           | MITRE                            |
| ----------------------------------------- | ------------------------------------- |----------------------------------|
| [Container](../vertices/CONTAINER.md) | [Node](../vertices/NODE.md) | [Escape to Host, T1611](https://attack.mitre.org/techniques/T1611/) |

## Details

A pod running as root and with a mount point to the node’s `/var/log` directory can expose the entire contents of its host filesystem to any user who has access to its logs, enabling an attacker to steal all K8s API tokens present on the K8s node. See [Kubernetes Pod Escape Using Log Mounts](https://blog.aquasec.com/kubernetes-security-pod-escape-log-mounts) for a more detailed explanation of the technique.

## Prerequisites

Execution as root within a container process with the host `/var/log/` (or any parent directory) mounted inside the container.

See the [example pod spec](../../test/setup/test-cluster/attacks/TOKEN_VAR_LOG_SYMLINK.yaml).

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

Listing secrets is a very powerful privilege and should not be required by the majority of users. Use an automated tool such a KubeHound to search for any risky permissions and users in the cluster and look to eliminate them.

## Calculation

+ [TokenList](../../pkg/kubehound/graph/edge/token_list.go)
+ [TokenListNamespace](../../pkg/kubehound/graph/edge/token_list_namespace.go)

## References:

+ [Official Kubernetes documentation: Secrets](https://kubernetes.io/docs/concepts/configuration/secret/#working-with-secrets)
+ [Securing Kubernetes Clusters by Eliminating Risky Permissions](https://www.cyberark.com/resources/threat-research-blog/securing-kubernetes-clusters-by-eliminating-risky-permissions)
