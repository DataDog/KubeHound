# KubeHound as a Service - ingestor - k8s deployment

To deploy KHaaS ingestor services in a Kubernetes environment please refer to [k8s-deployment](https://kubehound.io/user-guide/khaas-101/#k8s-deployment)

All the Helm charts and templates are provided as example. You should tweak them to your own environment (resources limitation, endpoint configuration, ...). This will depend of the number/size of the clusters you want to ingest.

* [Jupyter resources estimation](https://tljh.jupyter.org/en/latest/howto/admin/resource-estimation.html)
* [MongoDB hardware considerations](https://www.mongodb.com/docs/manual/administration/production-notes/#hardware-considerations)
* [Janusgraph InMemory Storage Backend](https://docs.janusgraph.org/storage-backend/inmemorybackend/)