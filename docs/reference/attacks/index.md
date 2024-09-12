---
hide:
  - toc
---

# Attack Reference

All edges in the KubeHound graph represent attacks with a net "improvement" in an attacker's position or a lateral movement opportunity.

!!! note

    For instance, an assume role or ([IDENTITY_ASSUME](./IDENTITY_ASSUME.md)) is considered as an attack.

|                           ID                            |                            Name                            |             MITRE ATT&CK Technique              | MITRE ATT&CK Tactic  | Coverage |
| :-----------------------------------------------------: | :--------------------------------------------------------: | :---------------------------------------------: | :------------------: | :------: |
|          [CE_MODULE_LOAD](./CE_MODULE_LOAD.md)          |            Container escape: Load kernel module            |                 Escape to host                  | Privilege escalation |   Full   |
|              [CE_NSENTER](./CE_NSENTER.md)              |                 Container escape: nsenter                  |                 Escape to host                  | Privilege escalation |   Full   |
|           [CE_PRIV_MOUNT](./CE_PRIV_MOUNT.md)           |          Container escape: Mount host filesystem           |                 Escape to host                  | Privilege escalation |   Full   |
|           [CE_SYS_PTRACE](./CE_SYS_PTRACE.md)           |  Container escape: Attach to host process via SYS_PTRACE   |                 Escape to host                  | Privilege escalation |   Full   |
|     [CE_UMH_CORE_PATTERN](./CE_UMH_CORE_PATTERN.md)     |   Container escape: through core_pattern usermode_helper   |                 Escape to host                  | Privilege escalation |   None   |
|      [CE_VAR_LOG_SYMLINK](./CE_VAR_LOG_SYMLINK.md)      |            Read file from sensitive host mount             |                 Escape to host                  | Privilege escalation |   Full   |
|        [CONTAINER_ATTACH](./CONTAINER_ATTACH.md)        |                Attach to running container                 |                       N/A                       |   Lateral Movement   |   Full   |
|        [ENDPOINT_EXPLOIT](./ENDPOINT_EXPLOIT.md)        |                  Exploit exposed endpoint                  |         Exploitation of Remote Services         |   Lateral Movement   |   Full   |
| [EXPLOIT_CONTAINERD_SOCK](./EXPLOIT_CONTAINERD_SOCK.md) | Container escape: Through mounted container runtime socket |                       N/A                       |   Lateral Movement   |   None   |
|       [EXPLOIT_HOST_READ](./EXPLOIT_HOST_READ.md)       |            Read file from sensitive host mount             |                 Escape to host                  | Privilege escalation |   Full   |
|   [EXPLOIT_HOST_TRAVERSE](./EXPLOIT_HOST_TRAVERSE.md)   |   Steal service account token through kubelet host mount   |              Unsecured Credentials              |  Credential Access   |   Full   |
|      [EXPLOIT_HOST_WRITE](./EXPLOIT_HOST_WRITE.md)      |      Container escape: Write to sensitive host mount       |                 Escape to host                  | Privilege escalation |   Full   |
|         [IDENTITY_ASSUME](./IDENTITY_ASSUME.md)         |                      Act as identity                       |                 Valid Accounts                  | Privilege escalation |   Full   |
|    [IDENTITY_IMPERSONATE](./IDENTITY_IMPERSONATE.md)    |                   Impersonate user/group                   |                 Valid Accounts                  | Privilege escalation |   Full   |
|     [PERMISSION_DISCOVER](./PERMISSION_DISCOVER.md)     |                   Enumerate permissions                    |           Permission Groups Discovery           |      Discovery       |   Full   |
|              [POD_ATTACH](./POD_ATTACH.md)              |                   Attach to running pod                    |                       N/A                       |   Lateral Movement   |   Full   |
|              [POD_CREATE](./POD_CREATE.md)              |                   Create privileged pod                    | Scheduled Task/Job: Container Orchestration Job | Privilege escalation |   Full   |
|                [POD_EXEC](./POD_EXEC.md)                |                   Exec into running pod                    |                       N/A                       |   Lateral Movement   |   Full   |
|               [POD_PATCH](./POD_PATCH.md)               |                     Patch running pod                      |                       N/A                       |   Lateral Movement   |   Full   |
|               [ROLE_BIND](./ROLE_BIND.md)               |                    Create role binding                     |                 Valid Accounts                  | Privilege Escalation | Partial  |
|      [SHARE_PS_NAMESPACE](./SHARE_PS_NAMESPACE.md)      |        Access container in shared process namespace        |                       N/A                       |   Lateral Movement   |   Full   |
|        [TOKEN_BRUTEFORCE](./TOKEN_BRUTEFORCE.md)        |      Brute-force secret name of service account token      |         Steal Application Access Token          |  Credential Access   |   Full   |
|              [TOKEN_LIST](./TOKEN_LIST.md)              |            Access service account token secrets            |         Steal Application Access Token          |  Credential Access   |   Full   |
|             [TOKEN_STEAL](./TOKEN_STEAL.md)             |          Steal service account token from volume           |              Unsecured Credentials              |  Credential Access   |   Full   |
|           [VOLUME_ACCESS](./VOLUME_ACCESS.md)           |                     Access host volume                     |        Container and Resource Discovery         |      Discovery       |   Full   |
|         [VOLUME_DISCOVER](./VOLUME_DISCOVER.md)         |                 Enumerate mounted volumes                  |        Container and Resource Discovery         |      Discovery       |   Full   |
