---
KEP: 0007
title: MCP Agent for KubeHound
status: draft
author: Thibault Normand <thibault.normand@datadoghq.com>
created: 2025-03-13
updated: 2025-03-13
version: 1.0.0
---

# KEP-0007 - MCP Agent for KubeHound

## Abstract

This KEP proposes the introduction of a [Model Context Protocol](https://modelcontextprotocol.io/introduction) (MCP) agent for 
KubeHound. With its standardised mechanism for querying and managing contextual 
information within the KubeHound graph database, the MCP agent revolutionises how 
we approach security monitoring, AI-driven investigations for blue teams, and 
exploitations for red teams.

## Motivation

KubeHound lacks a dedicated mechanism for querying its graph database using 
natural language or AI-driven workflows. This results in:

- Limited accessibility for non-technical users who need security insights.
- Challenges in integrating with AI-driven security analysis tools.
- Inefficient investigation workflows requiring manual query composition.
- Custom and repetitive attack path validation. 

The introduction of an MCP agent will address these issues by enabling:

- AI assistance during KubeHound investigations through various MCP servers.
- The ability to query the KubeHound graph using natural language.
- Allowing the customers to compose an AI-powered investigation lab.
- Automated attack path exploitation for validation.

## Model Context Protocol (MCP)

MCP is an open protocol that standardises how applications provide context to 
large language models (LLMs). Think of MCP as a USB-C port for AI applications. 
Just as USB-C offers a standardised way to connect your devices to various 
peripherals and accessories, MCP provides a standardised way to connect AI 
models to different data sources and tools.

### Why MCP?

MCP helps build agents and complex workflows on top of LLMs. Since LLMs 
frequently need to integrate with data and tools, MCP provides:

- A growing list of pre-built integrations that an LLM can directly plug into.
- The flexibility to switch between LLM providers and vendors.
- Best practices for securing data within existing infrastructure.

By integrating MCP with KubeHound, security investigations can leverage AI-driven 
insights with greater flexibility and efficiency.

## Proposal

The MCP agent will act as an intermediary layer that enables dynamic, AI-driven 
queries against the KubeHound graph database. 

The MCP must follow the large-scale dataset methodology from the 
[KEP-0006](KEP0006-Large-scale-Dataset-Investigation-Process.md).

## Implementation

The MCP server will be built with Go and the MCP SDK - 
[mark3labs/mcp-go](https://github.com/mark3labs/mcp-go).

The server will implemement STDIO and HTTP/SSE protocols provided by the SDK to 
offer the most compatible surface for integration.

### Exposed services

MCP exposes various entities to LLM:

- **tool**: a function which is callable by the LLM model to do action.
- **prompt**: a specific prompt used to enhance the relevance of of queries.
- **resource**: to declare specific additional context used during LLM prompt 
  generation and reduce hallucination (similar to RAG).

This is the list of exposed services by the MCP:

| MCP Type | ID                                                    | Description                                                                                                                                            |
| -------- | ----------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Tool     | kh_list_runids                                        | Function used to enumerate known imported dataset in the JanusGraph database.                                                                          |
| Tool     | kh_enumerate_namespaces(cluster, runID)               | Function used to retrieve the list of namespace from an imported dataset.                                                                              |
| Tool     | kh_container_escape_profiles(cluster,runID,namespace) | Function to enumerate attack path profiles for a given namespace.                                                                                      |
| Tool     | kh_vulnerable_images(cluster,runID,namespace)         | Function to enumerate container images concerned by an attack path. This function allows you to focus on the initial exploitability of an attack path. |

### Local setup

To enforce privacy and security of the exposed data to the LLM, we encourage you 
to use a local setup.

We are going to use:

- Ollama as a model runner
- MCPHost as a MCP client / Prompt to Ollama model
- `qwen3:14b` model as glue model supporting function calls, structured responses 
  and reasoning.

## History

- 2025-03-13: Initial draft.
