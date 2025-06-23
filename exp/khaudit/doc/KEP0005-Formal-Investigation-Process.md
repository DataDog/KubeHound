---
KEP: 0005
title: Formal Investigation Process
status: draft
author: Thibault Normand <thibault.normand@datadoghq.com>
created: 2025-03-12
updated: 2025-03-12
version: 1.0.0
---

# KEP-0005 - Formal Investigation Process

## Abstract

This proposal introduces a formal investigation process within KubeHound, a 
significant step towards ensuring structured analysis, validation, and mitigation 
of security findings. The process, consisting of five key phases: Plan, Observe, 
Attest, Raise, and Mitigate, promises to significantly improve the reliability 
and response to security incidents. By reducing false positives and ensuring 
timely remediation, this standardised methodology will bring about a more secure
KubeHound.

## Motivation

Currently, security investigations in KubeHound lack a standardised framework, 
leading to inconsistent analysis, validation, and resolution of findings. 

By formalising the investigation process, we can:

- Improve the efficiency and accuracy of security investigations.
- Reduce false positives through structured validation steps.
- Provide clear accountability for issue resolution.
- Enhance communication with resource owners regarding security concerns.

## Proposal

This proposal defines a five-phase investigation framework:

- Plan: Define investigation objectives and scope to ensure focus and efficiency.
- Observe: Analyse the KubeHound dataset to identify security findings.
- Attest: Validate findings to minimise false positives.
- Raise: Notify resource owners of confirmed security risks.
- Mitigate: Monitor and ensure mitigation efforts are completed.

This structured approach enables security teams to handle a wide range of threats
within the Kubernetes ecosystem systematically, from misconfigurations to 
vulnerabilities and compliance issues.

## Design

### Plan Phase

- Define investigation objectives, such as identifying misconfigurations, 
  detecting vulnerabilities, or assessing compliance.
- Establish scope by specifying which namespaces, clusters, or resources are 
  under review.
- Determine the expected data sources within KubeHound.

### Observe Phase

- Query the KubeHound dataset using predefined heuristics or detection rules.
- Collect relevant security findings and categorise them based on severity.
- Cross-reference findings with historical data to assess patterns.

### Attest Phase

- Validate findings using multiple data points to confirm legitimacy.
- Apply heuristics and automated validation techniques to filter out false positives.
- If necessary, perform manual reviews to ensure accuracy.

### Raise Phase

- Notify the respective resource owners of validated findings.
- Provide detailed reports, including risk assessment and recommended actions.
- Implement a tracking mechanism to ensure findings are acknowledged and addressed.

### Mitigate Phase

- Monitor the resolution process to ensure timely remediation.
- Verify mitigation actions through follow-up assessments.
- Escalate unresolved critical findings as needed.

## History

- 2025-03-12: Initial draft created.
