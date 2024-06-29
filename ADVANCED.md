# Advanced usage

### Infrastructure Setup

First create and populate a .env file with the required variables:

```bash
cp deployments/kubehound/.env.tpl deployments/kubehound/.env
```

Edit the variables (datadog env `DD_*` related and `KUBEHOUND_ENV`):

* `KUBEHOUND_ENV`: `dev` or `release` 
* `DD_API_KEY`: api key you created from https://app.datadoghq.com/ website

Note:
* `KUBEHOUND_ENV=dev` will build the images locally
* `KUBEHOUND_ENV=release` will use prebuilt images from ghcr.io 

### Running KubeHound

To replicate the automated command and run KubeHound step-by-step. First build the application:

```bash
make build
```

Next create a configuration file:

```yaml
collector:
  type: live-k8s-api-collector
telemetry:
  enabled: true
```

A tailored sample configuration file can be found [here](./configs/etc/kubehound.yaml), a full configuration reference containing all possible parameters [here](./configs/etc/kubehound-reference.yaml). 

Finally run the KubeHound binary, passing in the desired configuration:

```bash
bin/kubehound -c <config path>
```

Remember the targeted cluster must be set via `kubectx` or setting the `KUBECONFIG` environment variable. Additional functionality for managing the application can be found via:

```bash
make help
```
