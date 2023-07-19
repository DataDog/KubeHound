# IDENTITY_ASSUME

Represents the capacity to act as an [Identity](../vertices/IDENTITY.md) via execution within a [Container](../vertices/POD.md), ownership of a token, etc.

| Source                                    | Destination                           | MITRE                            |
| ----------------------------------------- | ------------------------------------- |----------------------------------|
| [Container](../vertices/POD.md) | [Identity](../vertices/CONTAINER.md)  | [Valid Accounts, T1078](https://attack.mitre.org/techniques/T1078/) |

## Details

Authentication to the K8s API is performed via passing certificates, static tokens or OIDC tokens in the API request. This edge represents the ability to assume an identity using either acquired credentials (such as a service account token) or the intrinsic identity of a resource in K8s such as executing commands from inside a pod with a bound serviceaccount.

## Prerequisites

Control of execution within a container with a bound serviceaccount.

## Checks

Check for a mounted service account tokens or secrets:

```bash
ls -la /var/run/secrets/kubernetes.io/serviceaccount/
```

## Exploitation

Assuming a valid token is recovered from e.g a mounted service acc a token is found it can be used to interact with the K8s API, to potentially access new resources:

```bash
KUBE_TOKEN=$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)
curl -sSk -H "Authorization: Bearer $KUBE_TOKEN" \
      https://$KUBERNETES_SERVICE_HOST:$KUBERNETES_PORT_443_TCP_PORT/api/v1/namespaces/kube-system/secrets
```

## Defences

### Monitoring

+ Monitor for installation and/or execution of kubectl within pods. This is anomalous activity but may be triggered by legitimate SRE or developer activities.
+ Interacting directly with the K8s API via non-standard tools e.g curl could be observed via the User-Agent field in audit logs. However, this is attacker controlled so should not be relied on.

### Implement security policies

Use a pod security policy or admission controller to prevent or limit the identities under which new pods can run.

## Calculation

[IdentityAssume](../../pkg/kubehound/graph/edge/identity_assume.go)

## References:  

+ [Official Kubernetes Documentation](https://kubernetes.io/docs/reference/access-authn-authz/authentication/#authentication-strategies)
+ [CURLing the Kubernetes API](https://nieldw.medium.com/curling-the-kubernetes-api-server-d7675cfc398c)