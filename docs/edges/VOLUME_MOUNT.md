# VOLUME_MOUNT

Represents a container's access to a mounted volume.

| Source                                    | Destination                           | MITRE                            |
| ----------------------------------------- | ------------------------------------- |----------------------------------|
| [Container](../vertices/CONTAINER.md) | [Volume](../vertices/VOLUME.md) | [Container and Resource Discovery, T1613](https://attack.mitre.org/techniques/T1613/) |

## Details

Volumes can contains K8s API tokens or other resources useful to an attacker in building an attack path.

## Prerequisites

A volume mounted into a container. Currently supports `HostPath` and `Projected` volume types.

## Checks

Check for mounted volumes using procfs (format definition of output [here](https://stackoverflow.com/questions/18122123/how-to-interpret-proc-mounts)):

```bash
cat /proc/self/mounts
```

Or use the mount command:

```bash
mount
```

Or use [quarkslab/kdigger](https://github.com/quarkslab/kdigger). Output here shows a mounted `proc` filesystem that is worth investigating:

```bash
kdigger dig mounts

### MOUNT ###
Comments:
- 23 devices are mounted.
+-----------+---------------------------------+------------+---------------------------------+
|   DEVICE  |               PATH              | FILESYSTEM |              FLAGS              |
+-----------+---------------------------------+------------+---------------------------------+
| overlay   | /                               | overlay    | rw,relatime,lowerdir=/var/lib/c |
|           |                                 |            | ontainerd/io.containerd.snapsho |
|           |                                 |            | tter.v1.overlayfs/snapshots/27/ |
|           |                                 |            | fs,upperdir=/var/lib/containerd |
|           |                                 |            | /io.containerd.snapshotter.v1.o |
|           |                                 |            | verlayfs/snapshots/71/fs,workdi |
|           |                                 |            | r=/var/lib/containerd/io.contai |
|           |                                 |            | nerd.snapshotter.v1.overlayfs/s |
|           |                                 |            | napshots/71/work                |
| proc      | /proc                           | proc       | rw,nosuid,nodev,noexec,relatime |
| tmpfs     | /dev                            | tmpfs      | rw,nosuid,size=65536k,mode=755  |
| devpts    | /dev/pts                        | devpts     | rw,nosuid,noexec,relatime,gid=5 |
|           |                                 |            | ,mode=620,ptmxmode=666          |
| mqueue    | /dev/mqueue                     | mqueue     | rw,nosuid,nodev,noexec,relatime |
| sysfs     | /sys                            | sysfs      | ro,nosuid,nodev,noexec,relatime |
| cgroup    | /sys/fs/cgroup                  | cgroup2    | ro,nosuid,nodev,noexec,relatime |
| proc      | /hostproc                       | proc       | rw,nosuid,nodev,noexec,relatime |
| /dev/vda1 | /etc/hosts                      | ext4       | rw,relatime                     |
| /dev/vda1 | /dev/termination-log            | ext4       | rw,relatime                     |
| /dev/vda1 | /etc/hostname                   | ext4       | rw,relatime                     |
| /dev/vda1 | /etc/resolv.conf                | ext4       | rw,relatime                     |
| shm       | /dev/shm                        | tmpfs      | rw,nosuid,nodev,noexec,relatime |
|           |                                 |            | ,size=65536k                    |
| tmpfs     | /run/secrets/kubernetes.io/serv | tmpfs      | ro,relatime,size=8039936k       |
|           | iceaccount                      |            |                                 |
| proc      | /proc/bus                       | proc       | ro,nosuid,nodev,noexec,relatime |
| proc      | /proc/fs                        | proc       | ro,nosuid,nodev,noexec,relatime |
| proc      | /proc/irq                       | proc       | ro,nosuid,nodev,noexec,relatime |
| proc      | /proc/sys                       | proc       | ro,nosuid,nodev,noexec,relatime |
| proc      | /proc/sysrq-trigger             | proc       | ro,nosuid,nodev,noexec,relatime |
| tmpfs     | /proc/kcore                     | tmpfs      | rw,nosuid,size=65536k,mode=755  |
| tmpfs     | /proc/keys                      | tmpfs      | rw,nosuid,size=65536k,mode=755  |
| tmpfs     | /proc/timer_list                | tmpfs      | rw,nosuid,size=65536k,mode=755  |
| tmpfs     | /sys/firmware                   | tmpfs      | ro,relatime                     |
+-----------+---------------------------------+------------+---------------------------------+
```

## Exploitation

No exploitation is necessary. This edge simply indicates that a volume is mounted within a container.

## Defences

None

## Calculation

+ [VolumeMount](../../pkg/kubehound/graph/edge/volume_mount.go)

## References:

+ [Official Kubernetes documentation: Volumes ](https://kubernetes.io/docs/concepts/storage/volumes/)

