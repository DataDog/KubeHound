All paths between a volume and identity

```
g.V().hasLabel("Volume").repeat(out().simplePath()).until(hasLabel("Identity")).path()
```

All container escapes

```
g.V().hasLabel("Container").repeat(out().simplePath()).until(hasLabel("Node")).path()
```

Paths from container to any critical asset

```
g.V().hasLabel("Container").repeat(out().simplePath()).until(has("critical", true)).path()
```