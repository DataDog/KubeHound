# Datadog setup

The Datadog agent can be setup locally to provide some metrics and logs when developing on KubeHound.

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