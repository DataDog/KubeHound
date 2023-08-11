# Home 

Welcome to the KubeHound documentation!

KubeHound creates a graph of attack paths in a Kubernetes cluster, allowing you to identify direct and multi-hop routes an attacker is able to take, visually or through complex graph queries.

[![](./images/example-graph.png)](./images/example-graph.png)
<center>*A KubeHound graph showing attack paths between pods, nodes, and identities (click to enlarge)*</center>

After the data is ingested in a graph database, you can ask advanced questions such as:

- What are all possible container escapes in the cluster?
- What is the shortest exploitable path between an exposed service and a critical asset?
- Is there an attack path from a specific container to a node in the cluster?

KubeHound was built with efficiency in mind and can consequently handle very large clusters. Ingestion and computation of attack paths typically takes 1 minute for a cluster with 1'000 running pods, 15 minutes for 10'000 pods, and 25 minutes for 25'000 pods.

Next steps:

- Learn more about KubeHound [architecture](./architecture.md) and [terminology](./terminology.md)
- [Get started](./user-guide/getting-started.md) using KubeHound