# CE_UMH_CORE_PATTERN

Container escape via the core_pattern usermode_helper (UMH) in the case of an exposed `/proc` mount.

| Source                                    | Destination                           | MITRE                            |
| ----------------------------------------- | ------------------------------------- |----------------------------------|
| [Container](../vertices/CONTAINER.md) | [Node](../vertices/NODE.md) | [Escape to Host, T1611](https://attack.mitre.org/techniques/T1611/) |

## Details

[/proc/sys/kernel/core_pattern](https://man7.org/linux/man-pages/man5/core.5.html) defines a program which is executed on core-file generation (typically a program crash) and is passed the core file as standard input if the first character of this file is a pipe symbol `|`. This program is run by the root user and will allow up to 128 bytes of command line arguments. Attacker control of this progam would allow trivial code execution within the container host given any crash and core file generation (which can be simply discarded during a myriad of malicious actions). With write access to the host `/proc` directory and no additional privileges, an attacker can abuse this to escape a container and gain root on the containing K8s node.

## Prerequisites

Execution within a container process with the host `/proc/sys/kernel` (or any parent directory) mounted inside the container.

See the [example pod spec](../../test/setup/test-cluster/attacks/CE_UMH_CORE_PATTERN.yaml).

## Checks

TODO

## Exploitation

TODO

## Defences

### Monitoring

+ Use the Datadog agent to monitor for creation of new `usermode_helper` programs via writes to known locations, in this case `/proc/sys/kernel_core_pattern`.

### Implement security policies

Use a pod security policy or admission controller to prevent or limit the creation of pods with a `hostPath` mount of `/proc` or other sensitive locations.

### Least Privilege

Avoid running containers as the root user. Specify an unprivileged user account using the `securityContext`.

## Calculation

[EscapeUmhCorePattern](../../pkg/kubehound/graph/edge/escape_umh_core_pattern.go)

## References:

+ [Compendium Of Container Escapes](https://i.blackhat.com/USA-19/Thursday/us-19-Edwards-Compendium-Of-Container-Escapes-up.pdf)
+ [Sensitive Mounts](https://0xn3va.gitbook.io/cheat-sheets/container/escaping/sensitive-mounts)
+ [Escaping privileged containers for fun](https://pwning.systems/posts/escaping-containers-for-fun/)

