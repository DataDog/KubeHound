---
hide:
  - toc
---

# Entities

Tne entities represents all the vertices in KubeHound graph model. Those are an abstract representation of a Kubernetes component that form the vertices of the graph.

## Entities

!!! note

    For instance: [PERMISSION_SET](./permissionset.md) is an abstract of Role and RoleBinding.

|                  ID                  |                                                                                                                                 Description                                                                                                                                 |
| :----------------------------------: | :-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------: |
|        [COMMON](./common.md)         |                                                                                                       Common properties can be set on any vertices within the graph.                                                                                                        |
|     [CONTAINER](./container.md)      |                                                                         A container image running on a Kubernetes pod. Containers in a Pod are co-located and co-scheduled to run on the same node.                                                                         |
|      [ENDPOINT](./endpoint.md)       |                                                                         A network endpoint exposed by a container accessible via a Kubernetes service, external node port or cluster IP/port tuple.                                                                         |
|      [IDENTITY](./identity.md)       |                                                                                                          Identity represents a Kubernetes user or service account.                                                                                                          |
|          [NODE](./node.md)           |                                                    A Kubernetes node. Kubernetes runs workloads by placing containers into Pods to run on Nodes. A node may be a virtual or physical machine, depending on the cluster.                                                     |
| [PERMISSION_SET](./permissionset.md) | A permission set represents a Kubernetes RBAC `Role` or `ClusterRole`, which contain rules that represent a set of permissions that has been bound to an identity via a `RoleBinding` or `ClusterRoleBinding`. Permissions are purely additive (there are no "deny" rules). |
|           [POD](./pod.md)            |                                                                                 A Kubernetes pod - the smallest deployable units of computing that you can create and manage in Kubernetes.                                                                                 |
|        [Volume](./volume.md)         |                                                                                                  Volume represents a volume mounted in a container and exposed by a node.                                                                                                   |
