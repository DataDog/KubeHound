---
KEP: 0006
title: Large-scale Dataset Investigation Process
status: draft
author: Thibault Normand <thibault.normand@datadoghq.com>
created: 2025-03-12
updated: 2025-03-12
version: 1.0.0
---

# KEP-0006 - Large-scale Dataset Investigation Process

## Abstract

This proposal outlines an enhanced process for large-scale dataset investigations
within KubeHound. The current approach struggles with scalability due to dataset
size, cluster count, and query time explosion in JanusGraph. The proposed method
introduces a structured workflow to improve threat detection and mitigation by
enumerating attack profiles, assessing security risks, validating findings, and
monitoring mitigation steps effectively.

## Motivation

Investigating security risks at scale in KubeHound presents several challenges:

- Top-down reasoning limitations: The vast size and volume of ingested clusters 
  make it challenging to apply broad investigative approaches effectively.
- Query inefficiency: Most cluster-wide JanusGraph queries fail due to path 
  combination explosion, preventing meaningful results.
- Scalability issues: The current approach does not efficiently handle large-scale
  dataset investigations, making threat analysis ineffective.

## Proposal

The proposed approach introduces a structured and scalable investigation workflow 
consisting of the following steps:

- **Plan**: Define the objectives of the investigation, specifying the threat model 
  and detection goals.
- **Observe**: Investigate within the KubeHound dataset, focusing on specific 
  queries that align with the objectives focusing Kubernetes namespace as resource 
  attack path source.
- **Attest**: Validate findings to reduce false positives, ensuring that identified 
  issues are accurate and actionable.
- **Raise**: Inform resource owners about validated findings, ensuring they have 
  the necessary context to address security risks.
- **Mitigate**: Monitor and verify that mitigation actions are taken, reducing 
  the risk of exploitation.

This method provides a systematic approach to large-scale investigations while 
ensuring results remain actionable and scalable.

## Design

The investigation process is based on the following key enhancements:

- **Attack Profile Enumeration**: Identify attack scenarios at the namespace level 
  to focus investigations on relevant security concerns. Focusing on a namespace 
  allows the investigation to reduce the audited resources and align with a more 
  iterative approach.

- **Security Assessment of Concerned Resources**: Validate the initial step of 
  attack path exploitation by evaluating the security posture of identified 
  resources. The attack path exploitation is essentially based on the capability 
  to exploit the first step in the path. This step concerns the case that an attack 
  path exploitability is reduced due to environmental security measures that are
  not detectable by KubeHound.

- **Attack Path Enumeration**: Generate a list of possible attack paths based on
  the namespace, the exploitable resources and the resources they affect, 
  reducing query inefficiency by focusing on identified investigation starting 
  points.

- **Path Exploitation Assessment**: Analyse the feasibility and impact of identified
  attack paths to prioritise responses.

- **Resource Ownership Identification**: Map attack path sources to their owners 
  to streamline remediation efforts. We want to focus on cutting the exploitation 
  path from its root.

- **Detection Signal Raising**: Generate alerts for validated security issues to 
  ensure timely mitigation.

## Implementation

> [!NOTE]
> The following section will describe the approach based on the container escape 
> path profiles related investigation.

The approach uses the formal investigation process to identify hot points in the 
dataset, and focus on micro-investigation on them. This approach allows automation 
to be done, and overcome the blind approach by providing a variable focused 
investigation.

### Plan

Retrieve a list of exploitable containers used as entrypoint to access a Kubernetes 
node directly or indirectly.

### Observe

#### Determine the most populated namespaces

The purpose of this step is to identify which namespaces are available to audit, 
and reduce the scope if we decide to exclude one of them for size reasons.

```groovy
g.V().
  // Scope only "Container" vertices.
  has("class", "Container").
  // For a given ingestion ID.
  has("runID", $runID).
  // Group vertices by namespace and vertex cardinality
  groupCount().by("namespace").
  // Order by vertex cardinality count (DESC)
  orderBy(values, desc).
  // Unfold result to ensure one result per line
  unfold()
```

#### Enumerate attack path profiles for each namespace

> [!NOTE]
> This step must be done for each namespace.

The purpose is to be able to enumerate attack path profiles from a given namespace 
to identify which one are interesting from an investigation point of view.

```groovy
g.V().
  // Scope only "Container" vertices.
  has("class", "Container").
  // For a given ingestion ID.
  has("runID", $runID).
  // For a given namespace.
  has("namespace", $ns).
  // Jump using a path traversal (get out from the current vertex, enter the connex one)
  // Time limit the path traversal to prevent hanging response.
  repeat(outE().inV().simplePath().timeLimit(2000)).
  // Until the path reached a vertex with a class property "Node",
  // Prevent infinite loop by limiting the hops count to 10.
  until(has("class", "Node").or().loops().is(10)).
  // Ensure that the last step in a "Node" for truncated paths
  has("class", "Node").
  // Group by path profiles by using labels as path elements
  groupCount().by(path.by(label)).
  // Unfold result to ensure one result per line
  unfold()
```

#### Enumerate vulnerable container images

This purpose helps you to determine the container's instances which could be used 
as an initial move attempt to gain access to the running container. This step is 
hypothetical and can be exploited from multiple point of view.

> [!NOTE]
> Currently our RBAC is permissive and offer you a easy-to-exploit initial step 
> as insider.

```groovy
g.V().
  // Scope only "Container" vertices.
  has("class", "Container").
  // For a given ingestion ID.
  has("runID", $runID).
  // For a given namespace.
  has("namespace", $ns).
  // Filter on container that have identified attack path profiles attached to.
  where(
    repeat(outE().inV().simplePath().timeLimit(2000)).
    // Until the path reached a vertex with a class property "Node",
    // Prevent infinite loop by limiting the hops count to 10.
    until(has("class", "Node").or().loops().is(10)).
    // Ensure that the last step in a "Node" for truncated paths
    has("class", "Node").
    // Limit to one path reauired for the condition
    limit(1)
  ).
  // Deduplicate containers by image value
  dedup().by("image").
  // Transform the container to an object of 4 properties
  ValueMap("namespace", "app",  "team", "image") 
```

#### Enumerate concrete attack paths for vulnerable containers

> [!NOTE]
> This step must be done for each vulnerable image.

This purpose targets container-running exploitable images in a given namespace 
to compute all concrete attack paths. These paths are associated with the 
namespace-level attack profiles, which are observed in the 
[Enumerate attack path profiles for each namespace](#enumerate-attack-path-profiles-for-each-namespace) 
step.

```groovy
 g.V().
  // Scope only "Container" vertices.
  has("class", "Container").
  // For a given ingestion ID.
  has("runID", $runID).
  // For a given namespace.
  has("namespace", $ns).
  // For a given image.
  has("image", $image).
  // Jump using a path traversal (get out from the current vertex, enter the connex one)
  // Time limit the path traversal to prevent hanging response.
  repeat(outE().inV().simplePath().timeLimit(2000)).
  // Until the path reached a vertex with a class property "Node",
  // Prevent infinite loop by limiting the hops count to 10.
  until(has("class", "Node").or().loops().is(10)).
  // Ensure that the last step in a "Node" for truncated paths
  has("class", "Node").
  // Dump the path traversal and display elements by their properties
  path().by(elementMap)
```

### Attest

This phase is used to confirm the observed exploitation paths, and requires a lot 
of a manual execution to allow automated exploitation.

### Raise

> [!NOTE]
> Not affected by the data size.

### Mitigate

> [!NOTE]
> Not affected by the data size.

# History

- 2025-03-12: Initial draft.
