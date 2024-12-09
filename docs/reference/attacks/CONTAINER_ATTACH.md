---
title: CONTAINER_ATTACH
---

<!--
id: CONTAINER_ATTACH
name: "Attach to running container"
mitreAttackTechnique: N/A - N/A
mitreAttackTactic: TA0008 - Lateral Movement
-->

# CONTAINER_ATTACH

| Source                    | Destination                           | MITRE                                                                |
| ------------------------- | ------------------------------------- | -------------------------------------------------------------------- |
| [Pod](../entities/pod.md) | [Container](../entities/container.md) | [Lateral Movement, TA0008](https://attack.mitre.org/tactics/TA0008/) |

Attach to a container running within a pod given access to the pod.

## Details

In order to attach a container running in a pod, you can create a debugging container with the `kubectl debug` command. It will spawn an ephemeral container that will attach to the console. To do so you need:
+ The target pod
+ The image to spawn as an ephemeral container

In order to access the target process, you need the id of the targeted container. Then by using the  `--target` flag, the ephemeral container will share the linux process namespace with the target By default, the process namespace is not shared between containers in a pod.

## Prerequisites

Permissions to debug the pod

## Checks

Check if sufficient permissions to attach to pods in the namespace of the target. First find the pod's namespace and id:

```bash
 kubectl get pods  --all-namespaces | grep <pod name>
```

Then check permissions:

```bash
kubectl auth can-i get pods/debug -n <namespace>
```

## Exploitation

Create and attach an ephemeral debugging container to the target pod via:

```bash
kubectl debug -it <pod name> --image=busybox:1.28 --target=<target container>
```

To determine the containers running in the pod (required to set a target above), you can use:

```bash
kubectl describe pod <pod name>
```

## Defences

### Monitoring

+ Monitor K8s audit logs for pod debug events as these should be fairly unusual, but may be triggered by legitimate SRE or developer activities.

## Calculation

+ [ContainerAttach](https://github.com/DataDog/KubeHound/tree/main/pkg/kubehound/graph/edge/container_attach.go)

## References:

+ [Official Kubernetes Documentation](https://kubernetes.io/docs/tasks/debug/debug-application/debug-running-pod/)
