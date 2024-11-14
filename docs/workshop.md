# Workshop

## Requirements

In order to run the workshop you need to install the following tools:

- kubectl - [https://kubernetes.io/docs/tasks/tools/](https://kubernetes.io/docs/tasks/tools/)
- kind - [https://kind.sigs.k8s.io/docs/user/quick-start](https://kind.sigs.k8s.io/docs/user/quick-start)
- docker - [https://docs.docker.com/engine/install](https://docs.docker.com/engine/install )
- make - package (sourceforge for Windows)

Those following packages are needed to spin the lab used during the workshop. We are reusing our developing environment.

For Mac user we have a oneliner to install everything (if you are using brew):

```shell
brew update && brew install kubectl, kind, docker`
```

The last requirements is of course kubehound. You need to download the latest release from our repository:

```shell
wget https://github.com/DataDog/KubeHound/releases/latest/download/kubehound-$(uname -o | sed 's/GNU\///g')-$(uname -m) -O kubehound
chmod +x kubehound
```

or

```shell
brew update && brew install kubehound
```

## Cheatsheet

### Starting the lab

First you need to run spin our dev environement with a vulnerable cluster:

```shell
cd $HOME
git clone https://github.com/DataDog/KubeHound.git
cd kubehound
make local-cluster-deploy
```

### Initiating Kubehound

As the images used by KubeHound are quite heavy (due to Jupyter and Janusgraph), we want to make sure that we have them downloaded before starting the workshop. To do so, we can run the following command:

```shell
./kubehound
```

This will pull all the images that will be needed during the workshop.


### Running the workshop

In order to use our vulnerable cluster, we need to use the `kubeconfig` file generated when we created (with kind) our cluster. This variable needs to be exported in all the shell you will be using during the workshop.

```shell
export KUBECONFIG=./test/setup/.kube-config
# Checking the clustername
kubectl config current-context
# Checking the pods deployed
kubectl get pods
```

During the workshop we will be playing with Kubernetes resources. We advise you to install [k9s](https://github.com/derailed/k9s) which is a great tool made by the community - provides a terminal UI to interact with k8s cluster.

In order to test the attacks, we will assume breach of the containers. To execute a command you can jump into a container/pod using the following command:

```shell
kubectl exec -it <pod_name> -- bash
```
!!! note

    You can also use k9s (typing on `s` key when highlighting a pod).
