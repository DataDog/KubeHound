---
title: POD_EXEC
---

<!--
id: POD_EXEC
name: "Exec into running pod"
mitreAttackTechnique: N/A - N/A
mitreAttackTactic: TA0008 - Lateral Movement
-->

# POD_EXEC

With the correct privileges an attacker can use the Kubernetes API to obtain a shell on a running pod.

| Source                                    | Destination                           | MITRE                            |
| ----------------------------------------- | ------------------------------------- |----------------------------------|
| [PermissionSet](../entities/permissionset.md)  | [Pod](../entities/pod.md) | [Lateral Movement, TA0008](https://attack.mitre.org/tactics/TA0008/)  |

## Details

An attacker with sufficient permissions can execute arbitrary commands inside the container using the `kubectl exec` command.

## Prerequisites

Ability to interrogate the K8s API with a role allowing exec access to pods which have the binary you want to execute (e.g. `/bin/bash`) available.

See the [example pod spec](https://github.com/DataDog/KubeHound/tree/main/test/setup/test-cluster/attacks/POD_EXEC.yaml).

## Checks

Simply ask kubectl:

```bash
k auth can-i create pod/exec
```

## Exploitation

Spawn a new interactive shell on the target pod:

```bash
k exec  --stdin --tty <POD NAME> -- /bin/bash
```

## Defences

### Monitoring

+ Monitor for pod exec from within an existing pod 
+ This activity will be BAU for SREs and as such monitoring for follow on actions may be more fruitful

### Implement least privilege access

Pod interactive execution is a very powerful privilege and should not be required by the majority of users. Use an automated tool such a KubeHound to search for any risky permissions and users in the cluster and look to eliminate them.

## Calculation

+ [PodExec](https://github.com/DataDog/KubeHound/tree/main/pkg/kubehound/graph/edge/pod_exec.go)
+ [PodExecNamespace](https://github.com/DataDog/KubeHound/tree/main/pkg/kubehound/graph/edge/pod_exec_namespace.go)

## References:

+ [Official Kubernetes Documentation](https://kubernetes.io/docs/tasks/debug/debug-application/get-shell-running-container/)
