# Troubleshooting Guide

## Backend Issues

The most common issues can usually be resolved by restarting the backend. See [Common Operations](./common-operations.md#restarting-the-backend)

## Janusgraph (kubegraph container) won't start

Make sure you have enough disk space. About 5GB is necessary for JanusGraph to start and ingest a (large) Kubernetes cluster data.

On Linux, and if you have a "strongly" partitioned system, you should make sure your Docker setup has enough space available, one quick check can be done with:
```bash
df $(docker info| grep "Root Dir" | cut -d":" -f2)
```