---
hide:
  - toc
---

|   ID   | Name | MITRE ATT&CK Technique | MITRE ATT&CK Tactic |
| :----: | :--: | :-----------------: | :--------------------: |
| [CE_MODULE_LOAD](/attacks/CE_MODULE_LOAD) | Container escape: load a kernel module | Escape to host | Privilege escalation | 
| [CE_NSENTER](/attacks/CE_NSENTER) | Container escape: nsenter | Escape to host | Privilege escalation | 
| [CE_PRIV_MOUNT](/attacks/CE_PRIV_MOUNT) | Container escape: mount host filesystem | Escape to host | Privilege escalation | 
| [CE_SYS_PTRACE](/attacks/CE_SYS_PTRACE) | Container escape: attach to a host process through SYS_PTRACE | Escape to host | Privilege escalation | 
| [CE_UMH_CORE_PATTERN](/attacks/CE_UMH_CORE_PATTERN) | Container escape: through core_pattern usermode_helper | Escape to host | Privilege escalation | 
| [CONTAINER_ATTACH](/attacks/CONTAINER_ATTACH) | Attach to running container | N/A | Lateral Movement | 
| [EXPLOIT_CONTAINERD_SOCK](/attacks/EXPLOIT_CONTAINERD_SOCK) | Container escape: mounted containerd.sock | N/A | Lateral Movement | 
| [EXPLOIT_HOST_READ](/attacks/EXPLOIT_HOST_READ) | Read file from sensitive host mount | Escape to host | Privilege escalation | 
| [EXPLOIT_HOST_TRAVERSE](/attacks/EXPLOIT_HOST_TRAVERSE) | Steal service account token through kubelet host mount | Unsecured Credentials | Credential Access | 
| [EXPLOIT_HOST_WRITE](/attacks/EXPLOIT_HOST_WRITE) | Container escape: write to sensitive host mount | Escape to host | Privilege escalation | 
| [IDENTITY_IMPERSONATE](/attacks/IDENTITY_IMPERSONATE) | Impersonate user/group | Valid Accounts | Privilege escalation | 
| [POD_ATTACH](/attacks/POD_ATTACH) | Attach to running pod | N/A | Lateral Movement | 
| [POD_CREATE](/attacks/POD_CREATE) | Create privileged pod | Scheduled Task/Job: Container Orchestration Job | Privilege escalation | 
| [POD_EXEC](/attacks/POD_EXEC) | Exec into running pod | N/A | Lateral Movement | 
| [POD_PATCH](/attacks/POD_PATCH) | Patch running pod | N/A | Lateral Movement | 
