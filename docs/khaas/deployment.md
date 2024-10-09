# Deploying KHaaS - Ingestor stack

!!! warning "deprecated"

    The `kubehound-ingestor` has been deprecated since **v1.5.0** and renamed to `kubehound-binary`.

## Docker deployment

To run the KubeHound as a Service with `docker` just use the following [compose files](https://github.com/DataDog/KubeHound/tree/main/deployments/kubehound). First you need to set the environment variables in the `kubehound.env` file. There is a template file `kubehound.env.template` that you can use as a reference.

```bash
cd ./deployments/kubehound
docker compose -f docker-compose.yaml -f docker-compose.release.yaml -f docker-compose.release.ingestor.yaml --profile jupyter up -d
```

By default the endpoints are only exposed locally:

- `127.0.0.1:9000` for ingestor endpoint.
- `127.0.0.1:8888` for the UI.

For the UI 2 profiles (`--profile`) are available, you need to pick one:

- `jupyter` to spawn a Jupyter backend compatible with Janusgraph endpoint ([aws graph-notebook](https://github.com/aws/graph-notebook)).
- `invana` to spawn the [Invana Studio](https://github.com/invana/invana-studio), a dedicated UI for Janusgraph (this is also deploying the invana backend). **We do not encourage to use as it is not maintained anymore**.

!!! warning

    You should change the default password by editing `NOTEBOOK_PASSWORD=<your_password>` in the `docker-compose.yaml`

## k8s deployment

To run the KubeHound as a Service on Kubernetes just use the following [helm files](https://github.com/DataDog/KubeHound/tree/main/deployments/k8s):

```bash
cd ./deployments/k8s
helm install khaas khaas --namespace khaas --create-namespace
```

If it succeeded you should see the deployment listed:

```bash
$ helm ls -A
NAME    NAMESPACE       REVISION        UPDATED                                 STATUS          CHART              APP VERSION
khaas   khaas           1               2024-07-30 19:04:37.0575 +0200 CEST     deployed        kubehound-0.0.1
```

!!! Note

    This is an example to deploy KubeHound as a Service in k8s cluster, but you will need to adapt it to your own environment.

## k8s collector

When deploying the collector inside a k8s cluster, we need to configure one of the following variable:

- `KH_K8S_CLUSTER_NAME`: variable indicating the name of the targetted k8s cluster

### RBAC requirements

In order for the collector to work it needs access to the k8s API and the following k8s ClusterRole:

| apiGroups                 | resources                                                    | verb        |
| ------------------------- | ------------------------------------------------------------ | ----------- |
| rbac.authorization.k8s.io | roles<br>rolebindings<br>clusterroles<br>clusterrolebindings | get<br>list |
|                           | pods<br>nodes<br>                                            | get<br>list |
| discovery.k8s.io          | endpointslices                                               | get<br>list |

The definition of the k8s RBAC can find here:

- [clusterRole](https://github.com/DataDog/KubeHound/tree/main/deployments/k8s/khaas/templates/cluster_role.yaml)
- [clusterRoleBinding](https://github.com/DataDog/KubeHound/tree/main/deployments/k8s/khaas/templates/cluster_role_binding.yaml)
- [serviceAccount](https://github.com/DataDog/KubeHound/tree/main/deployments/k8s/khaas/templates/service_account.yaml)
