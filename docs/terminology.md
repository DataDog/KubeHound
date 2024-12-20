# Terminology

## Graph theory

| **Term** | **Meaning** |
|:---:|:---:|
| Graph | A data type to represent complex, non-linear relationships between objects |
| Vertex | The fundamental unit of which graphs are formed (also known as "node") |
| Edge | A connection between vertices (also known as "relationship") |
| Path | A sequence of edges which joins a sequence of vertices |
| Traversal | The process of visiting (checking and/or updating) each vertex in a graph |

## KubeHound

**Entity**

An abstract representation of a Kubernetes component that form the vertices (nodes) of the attack graph. These do not necessarily have a one-to-mapping to Kubernetes objects, but represent a related construct in an attacker's mental model of the system. Each entity can be tied back to one (or more) Kubernetes object(s) from which it derived via vertex properties.

For instance, the [`PermissionSet`](./reference/entities/permissionset.md) entity abstracts the Kubernetes `Role` and `ClusterRole` objects that have a `RoleBinding` or `ClusterRoleBinding` attached.

**Attack**

All edges in the KubeHound graph represent a net "improvement" in an attacker's position or a lateral movement opportunity. Thus, if any two vertices in the graph are connected we know immediately that an attacker can move between them. As such ***attack*** and ***edge*** are used interchangeably throughout the project.

**Critical Asset**

An entity in KubeHound whose compromise would result in cluster admin (or equivalent) level access.
