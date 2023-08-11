# Architecture

KubeHound works in 3 steps:

1. Connect to your Kubernetes cluster and read API resources (pods, nodes, permissions...)
2. Compute attack paths
3. Write the results to a local graph database (JanusGraph)

After the initial ingestion is done, you use a compatible client (such as [gdotv](https://gdotv.com/)) to visualize and query attack paths in your cluster.

<!-- TODO: use proper captioning instead of center-->
<center>
[![KubeHound architecture](./images/kubehound-high-level.png)](./images/kubehound-high-level.png)
<p><em>KubeHound architecture (click to enlarge)</em></p>
</center>

Under the hood, KubeHound leverages a caching and persistence layer (Redis and MongoDB) while computing attack paths. As an end user, this is mostly transparent to you.

<!-- TODO: use proper captioning instead of center-->
<center>
[![KubeHound architecture](./images/kubehound-detailed.png)](./images/kubehound-detailed.png)
<p><em>KubeHound detailed architecture (click to enlarge)</em></p>
</center>
