# KubeHound docker stack

Under the hood, KubeHound is running dockers for the storedb, graphdb and UI.

## KubeHound backend

The docker backend can be handled directly with the `kubehound` binary. Check the [Backend](https://kubehound.io/user-guide/common-operations/#backend) documentation.

If you want you can also use directly the compose files without `kubehound` binary (running `docker compose ...` commands).

## KubeHound as a Service - ingestor - Docker deployment

To deploy KHaaS ingestor services please refer to [docker-deployment](https://kubehound.io/user-guide/khaas-101/#docker-deployment)