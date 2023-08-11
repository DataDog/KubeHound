# Comparison with other tools

## Lyft's [Cartography](https://github.com/lyft/cartography)

Cartography has a [Kubernetes module](https://lyft.github.io/cartography/modules/kubernetes/index.html). While useful to vizualize a cluster, it only has a few types of resources and relationships. Consequently, it cannot be used for mapping attack paths in a Kubernetes cluster.

## [BloodHound](https://github.com/SpecterOps/BloodHound)

BloodHound is one of the first projects (and certainly the most popular) that introduced attack graphs mapping. It is currently focused on Active Directory and [Azure](https://bloodhound.readthedocs.io/en/latest/data-collection/azurehound.html) environments, and does not support Kubernetes.

## [BOtB](https://github.com/brompwnie/botb)

BOtB is a pentesting tool that attempts to exploit common weaknesses. It runs from inside a compromised container. While very useful when performing a blackbox assessment, it doesn't have a full view of the cluster and does not attempt to find cluster-wide attack paths.


## [BOtB](https://github.com/brompwnie/botb)

BOtB is a pentesting tool that attempts to exploit common weaknesses. It runs from inside a compromised container. While very useful when performing a blackbox assessment, it doesn't have a full view of the cluster and does not attempt to find cluster-wide attack paths.

## [peirates](https://github.com/inguardians/peirates)

Similarly to BOtB, peirates is an offensive tool running from inside a pod. It doesn't have a full view of the cluster and does not attempt to find cluster-wide attack paths.

## [rbac-police](https://github.com/PaloAltoNetworks/rbac-police)

rbac-police allows you to retrieve the permissions associated to a specific identity in the cluster, which makes it easier to understand who has access to what. However, it does not look for attack paths in the cluster - it's focused on showing effective permissions of an identity.

## [KubiScan](https://github.com/cyberark/KubiScan)

KubeScan scans a Kubernetes cluster for risky permissions that allow an identity to escalate its privileges inside the cluster. It does not look for other types of attacks in the cluster, nor does it attempt to build an attack graph.

## [kdigger](https://github.com/quarkslab/kdigger)

Similarly to BOtB and peirates, kdigger runs from a compromised pod and attempts to retrieve information about the cluster and potential weaknesses. It does not have a full cluster view, nor does it attempt to build attack paths on a graph.