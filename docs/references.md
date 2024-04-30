# References

## 2024 - InsomniHack 2024 presentation
### [Standing on the Shoulders of Giant(Dog)s: A Kubernetes Attack Graph Model](https://www.insomnihack.ch/talks-2024/#BZ3UA9) 

[Recording :fontawesome-brands-youtube:{ .youtube } ](#){ .md-button  .md-button--youtube } [Slides :fontawesome-solid-file-pdf:{ .pdf } ](files/insomnihack24/Kubehound - Insomni'Hack 2024 - slides.pdf){ .md-button } [Dashboard PoC :fontawesome-brands-python:{ .python } ](https://github.com/DataDog/KubeHound/tree/main/deployments/kubehound/scripts/dashboard-insomnihack24k){ .md-button }

This presentation explains why the tool was created and what problem it tries to solve. 2 demos were showed:

* A [ PoC :fontawesome-brands-python:{ .python } of a dashboard](#) was created to show how interesting KPI can be extracted easily from KubeHound.
* A [specific notebook](https://github.com/DataDog/KubeHound/tree/main/deployments/kubehound/notebook/InsomniHackDemo.ipynb) to show how to shift from a can of worms to the most critical vulnerability in a Kubernetes Cluster with a few KubeHound requests.

It also showed how the tool has been built and lessons we have learned from the process.

##  2023 - Release v1.0 annoucement 
### [KubeHound: Identifying attack paths in Kubernetes clusters](https://securitylabs.datadoghq.com/articles/kubehound-identify-kubernetes-attack-paths/)

[Blog Article :fontawesome-brands-microblog: ](https://securitylabs.datadoghq.com/articles/kubehound-identify-kubernetes-attack-paths/){ .md-button }

Blog article published on [securitylabs](https://securitylabs.datadoghq.com) as a tutorial 101 on how to use the tools in different use cases:

* [Red team: Looking for low-hanging fruit](https://securitylabs.datadoghq.com/articles/kubehound-identify-kubernetes-attack-paths/#red-team-looking-for-low-hanging-fruit)
* [Blue team: Assessing the impact of a compromised container](https://securitylabs.datadoghq.com/articles/kubehound-identify-kubernetes-attack-paths/#blue-team-assessing-the-impact-of-a-compromised-container)
* [Blue team: Remediation](https://securitylabs.datadoghq.com/articles/kubehound-identify-kubernetes-attack-paths/#blue-team-remediation)
* [Blue team: Metrics and KPIs](https://securitylabs.datadoghq.com/articles/kubehound-identify-kubernetes-attack-paths/#blue-team-metrics-and-kpis)

It also explain briefly how the tools works (what is under the hood).