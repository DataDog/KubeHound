# Developer

## Requirements

To sucessufully build and run the test for kubehound, you need:

+ [Golang](https://go.dev/doc/install) `>= 1.22`
+ [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installing-with-a-package-manager)
+ [Kubectl](https://kubernetes.io/docs/tasks/tools/)

## Build

Build the application via:

```bash
make build
```

All binaries will be output to the [bin](./bin/) folder

## Release build

Build the release packages locally using [goreleaser](https://goreleaser.com/install):

```bash
make local-release
```

## Unit Testing

The full suite of unit tests can be run locally via:

```bash
make test
```

## System Testing

The repository includes a suite of system tests that will do the following:
+ create a local kubernetes cluster
+ collect kubernetes API data from the cluster
+ run KubeHound using the file collector to create a working graph database
+ query the graph database to ensure all expected vertices and edges have been created correctly

The cluster setup and running instances can be found under [test/setup](./test/setup/)

If you need to manually access the system test environment with kubectl and other commands, you'll need to set (assuming you are at the root dir):

```bash
cd test/setup/ && export KUBECONFIG=$(pwd)/.kube-config
```

### Environment variable:
- `DD_API_KEY` (optional): set to the datadog API key used to submit metrics and other observability data.

### Setup

Setup the test kind cluster (you only need to do this once!) via:

```bash
make local-cluster-deploy
```

Then run the system tests via:

```bash
make system-test
```

To cleanup the environment you can destroy the cluster via:

```bash
make local-cluster-destroy
```

To list all the available commands, run:

```bash
make help
```

In case of conflict/error, or just if you want to free some of your RAM. You can destroy the backend stack dedicated to the system-test. 
Simply run:
```bash
make system-test-clean
```

Note: if you are running on Linux but you don't want to run `sudo` for `kind` and `docker` command, you can overwrite this behavior by editing the following var in `test/setup/.config`:
* `DOCKER_CMD="docker"` for docker command
* `KIND_CMD="kind"` for kind command 

### CI Testing

System tests will be run in CI via the [system-test](./.github/workflows/system-test.yml) github action 


## Metrics and logs

To have some in-depth metrics and log correlation, all the components are now linked to datadog.  To configure it you just need to add your Datadog API key (`DD_API_KEY`) in the environment variable in the `deployments/kubehound/.env`. When the API key is configured, a docker will be created `kubehound-dev-datadog`. 

All the information being gathered are available at:

* Metrics: https://app.datadoghq.com/metric/summary?filter=kubehound.janusgraph
* Logs: https://app.datadoghq.com/logs?query=service%3Akubehound%20&cols=host%2Cservice&index=%2A&messageDisplay=inline&stream_sort=desc&viz=stream&from_ts=1688140043795&to_ts=1688140943795&live=true

To collect the metrics for Janusgraph an exporter from Prometheus is being used:
* https://github.com/prometheus/jmx_exporter

They are exposed here:
* Locally: http://127.0.0.1:8099/metrics
* Datadog: https://app.datadoghq.com/metric/summary?filter=kubehound.janusgraph
