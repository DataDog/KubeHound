# UI

## Jupyter Notebooks

This UI is using Amazon's [Graph Notebook](https://github.com/aws/graph-notebook) tooling to query JanusGraph.

As a first step, we recommend following the [KubeHound DSL 101 notebook](./KubehoundDSL_101.ipynb). You will find details on how to connect, default configuration and how to construct requests there.

## Gremlin tricks

### Cell magic

As stated in the 101 notebook, to launch gremlin queries against the KubeHound graph, you need to start the cell with the jupyter cell magic `%%gremlin`

### Display help

This magic can take a bunch of parameters, you can get the help message by running a cell containing just
```python
%%gremlin --help
g
```

### Common arguments

Most of the pre-written cells in the notebooks will already have magic arguments defined:
- `-d`: Displayed value on the node, defaults to the `label`, but often set to `name` or `class`
- `-g`: Group by - when set (often to `critical`), groups the nodes to reduce visual clobber
- `le 50`: Max labels length (default to 10, so overridden to 50 to fit the attacks names)
- `-p inv,oute`: [Visualization hints](https://docs.aws.amazon.com/neptune/latest/userguide/notebooks-visualization.html#notebooks-visualization-Gremlin)

### Save result

You can save the output of the gremlin query to manipulate it in python with the `--store-to <var>` argument
```python
%%gremlin --store-to nsVertexCount
g.V().groupCount().by('namespace')
```
And in a following cell
```python
print(nsVertexCount[0])
```
