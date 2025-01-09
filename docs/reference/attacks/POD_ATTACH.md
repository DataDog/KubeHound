---
title: POD_ATTACH
---

<!--
id: POD_ATTACH
name: "Attach to running pod"
mitreAttackTechnique: T1609 - Container Administration Command
mitreAttackTactic: TA0002 - Execution
-->

# POD_ATTACH

| Source                      | Destination               | MITRE ATT&CK                                                         |
| --------------------------- | ------------------------- | -------------------------------------------------------------------- |
| [Node](../entities/node.md) | [Pod](../entities/pod.md) | [Container Administration Command, T1609](https://attack.mitre.org/tactics/T1609/) |

Attach to a running K8s pod from a K8s node.

## Details

A node in K8s is the host of a number of pods. A node has a supervisory function and thus access to the node grants full access to the pods. The only obstacle is the right tooling to interact easily with `containerd`. 

## Prerequisites

Full access to a node.

## Checks

Ensure that the node is running containers using containerd via a ps -ef command. You should see an output similar to the below:

```bash
root@k8s-node:~# ps -ef
UID          PID    PPID  C STIME TTY          TIME CMD
root           1       0  0 09:48 ?        00:00:02 /sbin/init
root          86       1  0 09:48 ?        00:00:01 /lib/systemd/systemd-journald
root          99       1  1 09:48 ?        00:05:31 /usr/local/bin/containerd
root         220       1  2 09:48 ?        00:09:13 /usr/bin/kubelet --bootstrap-kubeconfig=/etc/kubernetes/bootstrap-kubelet.conf --kubeconfig=/etc/kubernetes/kubelet.conf --config=/var/lib/kubelet/config.yaml -
root         304       1  0 09:48 ?        00:00:11 /usr/local/bin/containerd-shim-runc-v2 -namespace k8s.io -id ae35055acaccc10634d9160f08937d71d856be6b0d954cc772e72c83ee9bffee -address /run/containerd/container
root         320       1  0 09:48 ?        00:00:11 /usr/local/bin/containerd-shim-runc-v2 -namespace k8s.io -id ef5d38bf4bf263cbb80f0ea8cb75059e2fc997b9885d649afe0cc7d48e4f8b3f -address /run/containerd/container
65535        351     304  0 09:48 ?        00:00:00 /pause
65535        357     320  0 09:48 ?        00:00:00 /pause
root         401     320  0 09:48 ?        00:00:09 /usr/local/bin/kube-proxy --config=/var/lib/kube-proxy/config.conf --hostname-override=k8s-gameday-worker
root         511     304  0 09:48 ?        00:00:07 /bin/kindnetd
root         889       1  0 09:49 ?        00:00:12 /usr/local/bin/containerd-shim-runc-v2 -namespace k8s.io -id 2446353ed4380ef6e81172b5c09a9191e9e300c93ded0febea42967177aae3d1 -address /run/containerd/container
65535        909     889  0 09:49 ?        00:00:00 /pause
root        1741     889  0 09:49 ?        00:03:47 datadog-cluster-agent start
root       12899       1  0 10:07 ?        00:00:11 /usr/local/bin/containerd-shim-runc-v2 -namespace k8s.io -id 61250b69765dbf0fadb9cdcfa87db7c0b8c4c85ff93051f83e52574a3b4b7bd4 -address /run/containerd/container
65535      12922   12899  0 10:07 ?        00:00:00 /pause
root       12955   12899  0 10:07 ?        00:00:00 /bin/sh -c -- while true; do sleep 30; done;
root       50503       1  0 11:13 ?        00:00:37 /usr/local/bin/containerd-shim-runc-v2 -namespace k8s.io -id 026430a39a6452201524dfbe9872b2eb6027a9bcc2085fa8f7d95096150165c0 -address /run/containerd/container
root       50523   50503  0 11:13 ?        00:00:00 /pause
root       50854   50503  2 11:13 ?        00:08:06 agent run
root       50895   50503  0 11:13 ?        00:01:20 trace-agent -config=/etc/datadog-agent/datadog.yaml
root       50940   50503  0 11:13 ?        00:01:58 process-agent --cfgpath=/etc/datadog-agent/datadog.yaml
root       51064   50503  0 11:13 ?        00:02:07 security-agent start -c=/etc/datadog-agent/datadog.yaml
root      233872       1  0 16:12 ?        00:00:00 /usr/local/bin/containerd-shim-runc-v2 -namespace k8s.io -id ba0483c7235ccdbc9baf9f2ac0e39bea1cf54c3a16cab068c0c51954bbde99d1 -address /run/containerd/container
65535     233893  233872  0 16:12 ?        00:00:00 /pause
root      233924  233872  0 16:12 pts/0    00:00:00 nsenter --all --target=1 -- su -
root      233939  233924  0 16:12 pts/0    00:00:00 su -
root      233957  233939  0 16:12 pts/0    00:00:00 -bash
root      237272  233957  0 16:17 pts/0    00:00:00 ctr -n k8s.io task exec -t --exec-id full-control 0f36d12d60d12d041df894132882380a1175d462b654d62cc2907994cbf6c238 /bin/sh
root      237292  233872  0 16:17 pts/1    00:00:00 /bin/sh
root      237719   12955  0 16:18 ?        00:00:00 sleep 30
root      237720  237292  0 16:18 pts/1    00:00:00 ps -ef
```

Take note of the containerd-shim-runc-v2 processes, in particular the namespace which we will need later! Ensure the ctr utility is installed so we can interact directly with containerd:

```bash
which ctr
# /usr/local/bin/ctr
```
If ctr is not installed, we can install it ourselves:

```bash
curl -L https://github.com/containerd/containerd/releases/download/v1.6.19/containerd-1.6.19-linux-arm64.tar.gz > containerd.tar.gz
tar -xzf containerd.tar.gz
```
Now we can use ctr to examine the running containers:

```bash
$ ctr -n k8s.io containers ls
CONTAINER                                                           IMAGE                                            RUNTIME
07c72331fc4f2ffb7dc385beda170862f93993bd53d69232652fe1d6af83c8a8    docker.io/kindest/kindnetd:v20221004-44d545d1    io.containerd.runc.v2
33ba7bd90621cca281c99e58a83fe8fe974221085bebedc2de69ccf5988daa43    gcr.io/datadoghq/agent:7.43.0                    io.containerd.runc.v2
3bc41387a69a36fd031cdedb17aaf41a72faebce15fbad94b5e694600bde7a20    registry.k8s.io/pause:3.7                        io.containerd.runc.v2
43d816d3044000b37eed5710c2ad3da5f8f4a2e6ede47132238c0034de708612    gcr.io/datadoghq/agent:7.43.0                    io.containerd.runc.v2
44009453f6cb08fa4f617ef9b8ef5838f3dfc20c10693318b16119af4b37f7d4    registry.k8s.io/pause:3.7                        io.containerd.runc.v2
53384fe00e0a9846bfe406b4cc7ba84ba2fb34a05a86ea106064d92e648b2715    gcr.io/datadoghq/agent:7.43.0                    io.containerd.runc.v2
62ff4b5804ebb11a88af9d496afe09d2aa47fa3c0152f3ae90cf347899dcce71    docker.io/library/ubuntu:22.04                   io.containerd.runc.v2
876ef7a4f348010ce85bc6fd96ce8760c6357d7d24201b1459d8344155f4092f    docker.io/library/ubuntu:22.04                   io.containerd.runc.v2
8a8f35d4ece179d6cb968d66a1cca3dfb7dfe10b72b221cd5ff3cc42abd60c7f    gcr.io/datadoghq/agent:7.43.0                    io.containerd.runc.v2
a2036c90221e0993fb87a378b002bd43e15c1be2a70be43e7fc6416e99be4446    gcr.io/datadoghq/agent:7.43.0                    io.containerd.runc.v2
a2d0bb15f6e4b7a52d42e79685a2f57b06aa8f6ffbc15a38a1d303289a164e1d    gcr.io/datadoghq/agent:7.43.0                    io.containerd.runc.v2
bb9790ddf282fdb15c768013da80486989d4fda8f45ab94f70646cc9cce947e7    gcr.io/datadoghq/agent:7.43.0                    io.containerd.runc.v2
bdcc01c28b288e5c78e1f358716da80556b4a073c24549bc565180979c351891    registry.k8s.io/pause:3.7                        io.containerd.runc.v2
c5e3934d2e533b8580242f360f8a2e4f99a38d5cb50c9a579bf471869c38f043    gcr.io/datadoghq/agent:7.43.0                    io.containerd.runc.v2
c952cc4329ba0c822dacbade96153dea12d415c7c80ecb3f4c4a72d895629971    registry.k8s.io/kube-proxy:v1.25.3               io.containerd.runc.v2
ec74a2a48f514861096cf4b7b12550d8ce17d5b0803ec0aaf6a26dda4876e100    registry.k8s.io/pause:3.7                        io.containerd.runc.v2
f681a2a6f64b11bf53ec8d1c13de0d8715ac30819f21f94f316e0bc26a844770    registry.k8s.io/pause:3.7                        io.containerd.runc.v2
```

and the running tasks:

```bash
$ ctr -n k8s.io tasks ls
TASK                                                                PID       STATUS
839a74e97f50a275d306252842f3e79fca684a3b8b5ed3175d154fb990c80f31    51064     RUNNING
ae35055acaccc10634d9160f08937d71d856be6b0d954cc772e72c83ee9bffee    351       RUNNING
d2ff37e3f64e4032c34f635aa5d494e3d52e481772e2e8a71c5677f58938c35d    511       RUNNING
2446353ed4380ef6e81172b5c09a9191e9e300c93ded0febea42967177aae3d1    909       RUNNING
5bebdfa7f163311e44997625d4c1ac958f46ff1398925e9dbc7fb8a2ac2fc25a    12955     RUNNING
026430a39a6452201524dfbe9872b2eb6027a9bcc2085fa8f7d95096150165c0    50523     RUNNING
332f857ec069937342b07ed526e1d714b552e4656ed3e7c938b2b69c3790eb18    50895     RUNNING
24ab75252feedcf5232af1fa02a215bb0851599bc762933f7b7442943fcd58f9    50940     RUNNING
48e7c28fdc68c9473a31dea3194b922e064ad9e37499da7b7cf2481722d97dc6    401       RUNNING
80eeba17282348b418d786c6fa76338df7a7f6a6734767492825c1d6674a376a    50854     RUNNING
61250b69765dbf0fadb9cdcfa87db7c0b8c4c85ff93051f83e52574a3b4b7bd4    12922     RUNNING
ef5d38bf4bf263cbb80f0ea8cb75059e2fc997b9885d649afe0cc7d48e4f8b3f    357       RUNNING
0f36d12d60d12d041df894132882380a1175d462b654d62cc2907994cbf6c238    233924    RUNNING
ba0483c7235ccdbc9baf9f2ac0e39bea1cf54c3a16cab068c0c51954bbde99d1    233893    RUNNING
f4efb164afc51bf45cb09ed99058e72ea2e70cf887b8f8cfd5a058e69e65aa2e    1741      RUNNING
```

## Exploitation

We have full control of all the running containers. The simplest way to make us of this is to execute a new task inside one of the running containers to get a shell:

```bash
ctr -n k8s.io task exec -t --exec-id full-control 0f36d12d60d12d041df8941
```

## Defences

### Monitoring

+ Monitor for use of the CTR binary (or equivalents such as [crictl](https://kubernetes.io/docs/tasks/debug/debug-cluster/crictl/) or [nerdctl](https://github.com/containerd/nerdctl)) within nodes via the Datadog agent. This activity should be very unusual.

## Calculation

+ [PodAttach](https://github.com/DataDog/KubeHound/tree/main/pkg/kubehound/graph/edge/pod_attach.go)

## References:

+ [Kubernetes API Reference Docs](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#pod-v1-core)
+ https://iximiuz.com/en/posts/containerd-command-line-clients/
+ https://nanikgolang.netlify.app/post/containers/
+ https://www.mankier.com/8/ctr
