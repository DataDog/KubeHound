# Security posture for kind-kubehound.test.local

Small dashboard to show how KubeHound can be used to build a custom KPI dashboard.

> [!NOTE] 
> **You need to install poetry in order to run the PoC**: [[https://python-poetry.org/](https://python-poetry.org/)]


## How to run it

The script will connect to the Janusgraph socket (by default `127.0.0.1:8182` defined in `GREMLIN_SOCKET`) and build the dashboard:

```bash
poetry install
poetry run panel serve main.py
```

It will serve a webpage on [http://localhost:5006/main](http://localhost:5006/main).

> [!WARNING]  
> **It is only serving localhost and not 127.0.0.1**: You can add a specific flag to allow CORS on 127.0.0.1 `BOKEH_ALLOW_WS_ORIGIN=127.0.0.1:5006`

## Prerequisites

In order to run the local kind Kubernetes Cluster you need to setup a "dev environment":

* Kind: https://kind.sigs.k8s.io/docs/user/quick-start/#installing-with-a-package-manager
* Kubectl: https://kubernetes.io/docs/tasks/tools/

### Step by step

You need to have data in KubeHound in order to generate metrics. The best way is to leverage the kind cluster. There is a script that bundle the local deployment of the kind cluster:

1. Checkout the latest branch from the Github repository: `git clone https://github.com/DataDog/KubeHound/`
2. Create a vulnerable local Kubernetes (kind) cluster: `make local-cluster-deploy`
3. Set connection to this local cluster: `KUBECONFIG=./kubehound/test/setup/.kube-config`

Then use the latest release to spawn KubeHound backend and run the binary:

1. Spawn KubeHound backend: `kubehound.sh backend-up`
2. Set connection to this local cluster: `export KUBECONFIG=./kubehound/test/setup/.kube-config`. This file is generated when spawning the local kind cluster.
2. Run KubeHound to ingest local cluster: `./kubehound`

### All in one

Script using main branch to deploy and ingest sample data from kind cluster:

```bash
git clone https://github.com/DataDog/KubeHound/
cd KubeHound
make local-cluster-deploy
export KUBECONFIG=./test/setup/.kube-config
make build
./bin/kubehound
```