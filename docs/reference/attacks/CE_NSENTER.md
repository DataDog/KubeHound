---
title: CE_NSENTER
---

<!--
id: CE_NSENTER
name: "Container escape: nsenter"
mitreAttackTechnique: T1611 - Escape to host
mitreAttackTactic: TA0004 - Privilege escalation
-->

# CE_NSENTER

| Source                                    | Destination                           | MITRE                            |
| ----------------------------------------- | ------------------------------------- |----------------------------------|
| [Container](../entities/container.md) | [Node](../entities/node.md) | [Escape to Host, T1611](https://attack.mitre.org/techniques/T1611/) |

Container escape via the nsenter built-in linux program that allows executing a binary into another namespace.

## Details

When both `hostPID: true` and `privileged: true` are set, the pod can see all of the processes on the host, and you can enter the init system (PID 1) on the host, and execute your shell on the node 

## Prerequisites

Execution within a container process created with `--privileged` AND the `--pid=host` enabled.

See the [example pod spec](https://github.com/DataDog/KubeHound/tree/main/test/setuptest-cluster/attacks/CE_NSENTER.yaml).

## Checks

There is no straightforward way to detect if `hostPID` is activated from a container. The only way is to detect host program running from a pod. The most common way is to look for the `kubelet` binary running:

```bash
ps -ef | grep kubelet
```

## Exploitation

`nsenter` is a tool that allows us to enter the namespaces of one or more other processes and then executes a specified program. When you exec a binary into a container using the docker exec command:

```bash
docker exec -it --user root <container name> sh
```

You could do the same with nsenter:
+ Target a specific container
+ Look for PID of the targetted container
+ Execute `nsenter` of this specific PID and ask for all namespaces

```bash
docker ps | grep <container name>
CONTAINER_PID=$(docker inspect <container name> --format='{{ .State.Pid }}')
sudo nsenter -t $CONTAINER_PID -m -u -n -i -p sh
```

So to escape from a container and access the pod you just run, you need to target running on the host as root (PID of 1 is running the init for the host) ask for all the namespaces:

```bash
nsenter --target 1 --mount --uts --ipc --net --pid -- bash
```

The options `-m -u -n -i -p` are referring to the various namespaces that you want to access (e.g mount, UTS, IPC, net, pid).

## Defences

### Monitoring

+ Monitor for the use of the nsenter binary.
+ Monitor the `setns` syscall

### Implement security policies

Use a pod security policy or admission controller to prevent or limit the creation of pods with `privileged` or `hostPid` enabled.

### Least Privilege

Avoid running containers as the `root` user. Enforce runnin as an unprivileged user account using the `runAsNonRoot` setting inside `securityContext` (or explicitly setting `runAsUser` to an unprivileged user). Additionally, ensure that `allowPrivilegeEscalation: false` is set in `securityContext` to prevent a container running as an unprivileged user from being able to escalate to running as the `root` user.

## Calculation

+ [EscapeNsenter](https://github.com/DataDog/KubeHound/tree/main/pkg/kubehound/graph/edge/escape_nsenter.go)

## References:

+ [nsenter(1) - Linux manual page](https://man7.org/linux/man-pages/man1/nsenter.1.html)
+ [Docker Breakout / Privilege Escalation](https://book.hacktricks.xyz/linux-hardening/privilege-escalation/docker-breakout/docker-breakout-privilege-escalation#privileged-+-hostpid)
+ [Debugging containers using nsenter](https://jaanhio.me/blog/nsenter-debug/)

