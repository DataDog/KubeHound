# References

## 2024 - HackLu Workshop

### [KubeHound: Identifying attack paths in Kubernetes clusters at scale with no hustle](https://pretalx.com/hack-lu-2024/talk/HWDZGZ/)

[Slides :fontawesome-solid-file-pdf:{ .pdf } ](files/hacklu24/Kubehound-Workshop-HackLu24.pdf){ .md-button } [Jupyter notebook :fontawesome-brands-python:{ .python } ](https://github.com/DataDog/KubeHound/tree/main/deployments/kubehound/notebook/KubehoundDSL_101.ipynb){ .md-button }

Updated version of the Pass The Salt Workshop. Prerequisites are listed on [kubehound.io/workshop](https://kubehound.io/workshop/). The workshop is a hands-on session where participants will learn how to use KubeHound to identify attack paths in Kubernetes clusters. The workshop will cover the following topics:

- Hack a vulnerable Kubernetes cluster (exploiting 4 differents attacks in a local environement).
- Use KubeHound to identify specific resources in the vulnerable cluster.
- Use KubeHound to identify attack paths in the vulnerable cluster.

## 2024 - HackLu presentation

### [KubeHound: Identifying attack paths in Kubernetes clusters at scale with no hustle](https://pretalx.com/hack-lu-2024/talk/HWDZGZ/)

[Recording :fontawesome-brands-youtube:{ .youtube } ](https://www.youtube.com/watch?v=h-dD7PQC4NA){ .md-button .md-button--youtube } [Slides :fontawesome-solid-file-pdf:{ .pdf } ](files/hacklu24/Kubehound-HackLu24-slides.pdf){ .md-button }

This presentation explains the genesis behind the tool and a brief introduction to what Kubernetes security is. We showcase the three main usage for KubeHound:

- As a standalone tool to identify attack paths in a Kubernetes cluster from a laptop (the automatic mode and easy way to dump and ingest the data).
- As a blue teamer with **KubeHound as a Service** or **KHaaS** which allow using KubeHound to be used with a distributed model across multiple Kuberentes Clusters to generate continuously a security posture on a daily basis.
- As a consultant using the asynchronously mechanism to dump and rehydrate the KubeHound state from 2 different locations.

## 2024 - Pass The Salt (PTS) Workshop

### [KubeHound: Identifying attack paths in Kubernetes clusters at scale with no hustle](https://cfp.pass-the-salt.org/pts2024/talk/WA99YZ/)

[Slides :fontawesome-solid-file-pdf:{ .pdf } ](files/PassTheSalt24/Kubehound-Workshop-PassTheSalt_2024.pdf){ .md-button } [Jupyter notebook :fontawesome-brands-python:{ .python } ](https://github.com/DataDog/KubeHound/tree/main/deployments/kubehound/notebook/KubehoundDSL_101.ipynb){ .md-button }

The goal of the workshop was to showcase **how to use KubeHound to pinpoint security issues in a Kubernetes cluster and get a concrete security posture**.

But first, as attackers (or defenders), there's nothing better to understand an attack than to exploit it oneself. So the workshop started with some of the most common attacks (container escape and lateral movement) and **let attendees exploit them in our vulnerable cluster**.

After doing some introduction around Kubernetes basic and Graph theory, the attendees played with KubeHound to ingest data synchronously and asynchronously (dump and rehydrate the data). Then we **covered all the KubeHound DSL and basic gremlin usage**. The goal was to go over the possibilities of the KubeHound DSL like:

- List all the port and IP addresses being exposed outside of the k8s cluster
- Enumerate how attacks are present in the cluster
- List all attacks path from endpoints to node
- List all endpoint properties by port with serviceEndpoint and IP addresses that lead to a critical path
- ...

The workshop finished with some "real cases" scenario either from a red teamer or blue teamer point of view. The goal was to show how the tool can be used in different scenarios (initial recon, attack path analysis, assumed breach on compromised resources such as containers or credentials, ...)

All was done using the following notebook which is a step-by-step KubeHound DSL:

- A [specific notebook](https://github.com/DataDog/KubeHound/tree/main/deployments/kubehound/notebook/KubeHoundDSL_101.ipynb) to describe all KubeHound DSL queries and how you can leverage them. Also this notebook describes the basic Gremlin needed to handle the KubeHound DSL for specific cases.

## 2024 - Troopers presentation

### [Attacking and Defending Kubernetes Cluster with KubeHound, an Attack Graph Model](https://troopers.de/troopers24/talks/t8tc7m/)

[Recording :fontawesome-brands-youtube:{ .youtube } ](#){ .md-button .md-button--youtube } [Slides :fontawesome-solid-file-pdf:{ .pdf } ](files/Troopers24/Kubehound-Troopers_2024-slides.pdf){ .md-button } [Dashboard PoC :fontawesome-brands-python:{ .python } ](https://github.com/DataDog/KubeHound/tree/main/scripts/dashboard-demo){ .md-button }

This presentation explains the genesis behind the tool. A specific focus was made on the new version **KubeHound as a Service** or **KHaaS** which allow using KubeHound with a distributed model across multiple Kuberentes Clusters. We also introduce a new command that allows consultants to use KubeHound asynchronously (dumping and rehydration later, in office for instance).

2 demos were also shown:

- A [ PoC :fontawesome-brands-python:{ .python } of a dashboard](#) was created to show how interesting KPI can be extracted easily from KubeHound.
- A [specific notebook](https://github.com/DataDog/KubeHound/tree/main/deployments/kubehound/notebook/KubeHound_demo.ipynb) to show how to shift from a can of worms to the most critical vulnerability in a Kubernetes Cluster with a few KubeHound requests.

Also we showed how the tool has been built and lessons we have learned from the process.

## 2024 - InsomniHack 2024 presentation

### [Standing on the Shoulders of Giant(Dog)s: A Kubernetes Attack Graph Model](https://www.insomnihack.ch/talks-2024/#BZ3UA9)

[Recording :fontawesome-brands-youtube:{ .youtube } ](https://www.youtube.com/watch?v=sy_ijtW6wmQ){ .md-button .md-button--youtube } [Slides :fontawesome-solid-file-pdf:{ .pdf } ](files/insomnihack24/Kubehound - InsomniHack 2024 - slides.pdf){ .md-button } [Dashboard PoC :fontawesome-brands-python:{ .python } ](https://github.com/DataDog/KubeHound/tree/main/scripts/dashboard-demo){ .md-button }

This presentation explains why the tool was created and what problem it tries to solve. 2 demos were shown:

- A [ PoC :fontawesome-brands-python:{ .python } of a dashboard](#) was created to show how interesting KPI can be extracted easily from KubeHound.
- A [specific notebook](https://github.com/DataDog/KubeHound/tree/main/deployments/kubehound/notebook/InsomniHackDemo.ipynb) to show how to shift from a can of worms to the most critical vulnerability in a Kubernetes Cluster with a few KubeHound requests.

It also showed how the tool has been built and lessons we have learned from the process.

## 2023 - Release v1.0 annoucement

### [KubeHound: Identifying attack paths in Kubernetes clusters](https://securitylabs.datadoghq.com/articles/kubehound-identify-kubernetes-attack-paths/)

[Blog Article :fontawesome-brands-microblog: ](https://securitylabs.datadoghq.com/articles/kubehound-identify-kubernetes-attack-paths/){ .md-button }

Blog article published on [securitylabs](https://securitylabs.datadoghq.com) as a tutorial 101 on how to use the tools in different use cases:

- [Red team: Looking for low-hanging fruit](https://securitylabs.datadoghq.com/articles/kubehound-identify-kubernetes-attack-paths/#red-team-looking-for-low-hanging-fruit)
- [Blue team: Assessing the impact of a compromised container](https://securitylabs.datadoghq.com/articles/kubehound-identify-kubernetes-attack-paths/#blue-team-assessing-the-impact-of-a-compromised-container)
- [Blue team: Remediation](https://securitylabs.datadoghq.com/articles/kubehound-identify-kubernetes-attack-paths/#blue-team-remediation)
- [Blue team: Metrics and KPIs](https://securitylabs.datadoghq.com/articles/kubehound-identify-kubernetes-attack-paths/#blue-team-metrics-and-kpis)

It also explain briefly how the tools works (what is under the hood).
