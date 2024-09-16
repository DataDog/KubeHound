# Advanced configuration

This section covering all the flags available to use KubeHound. If you don't want to specify the bucket every time, you can set it up in your local config file (`./kubehound.yaml` or `$HOME/.config/kubehound.yaml`). See [getting started with KHaaS](./getting-started.md)

## Manual collection

In order to use `kubehound` with KHaaS, you need to specify the api endpoint you want to use:

- `--khaas-server` from the inline flags (by default `127.0.0.1:9000`)

Since this is not likely to change in your environment, we advise you to use the local config file. By default KubeHound will look for `./kubehound.yaml` or `$HOME/.config/kubehound.yaml`. As example here we set the default endpoint with disabled SSL.

## Dump and ingest

In order to use the collector with KHaaS you need to specify the dump location for the k8s resources:

- `--bucket_url` from the inline flags (i.e. `s3://<your_bucket>`). There is no default value for security reason.
- `--region` from the inline flags (i.e. `us-east-1`) to set the region to retrieve the configuration (only for s3).

!!! warning

    The `kubehound` binary needs to have push access to your cloud storage provider.

To dump and ingest the current k8s cluster, you just need to run the following command (i.e. using an AWS config):

```bash
kubehound dump remote --khaas-server 127.0.0.1:9000 --insecure --bucket_url s3://<your_bucket> --region  us-east-1
```

The last command will:

- **dump the k8s resources** to the cloud storage provider.
- send a grpc call to **run the ingestion on the KHaaS** grpc endpoint.

!!! note

    The ingestion will dump the current cluster being setup, if you want to skip the interactive mode, just specify `-y` or `--non-interactive`

### Manual ingestion

If you want to rehydrate (reingesting all the latest clusters dumps), you can use the `ingest` command to run it.

```bash
kubehound ingest remote --khaas-server 127.0.0.1:9000 --insecure
```

You can also specify a specific dump by using the `--cluster` and `run_id` flags.

```bash
kubehound ingest remote --khaas-server 127.0.0.1:9000 --insecure --cluster my-cluster-1 --run_id 01htdgjj34mcmrrksw4bjy2e94
```
