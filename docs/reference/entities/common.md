# Common Properties

Common properties can be set on any vertices within the graph.

## Ownership Information

| Property | Type     | Description                                        |
| -------- | -------- | -------------------------------------------------- |
| app      | `string` | Internal app name extracted from object labels     |
| team     | `string` | Internal team name extracted from object labels    |
| service  | `string` | Internal service name extracted from object labels |

## Risk Information

| Property    | Type   | Description                                                                                                                                                                                  |
| ----------- | ------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| critical    | `bool` | Whether the vertex is a critical asset within the cluster. Critical assets form the termination condition of an attack path and represent an asset that leads to complete cluster compromise |
| compromised | `int`  | Enum defining asset compromise for scenario-based simulations                                                                                                                                |

## Store Information

| Property | Type     | Description                                                                  |
| -------- | -------- | ---------------------------------------------------------------------------- |
| storeID  | `string` | Unique store database identifier of the store objected generating the vertex |

## Namespace Information

| Property     | Type     | Description                                                      |
| ------------ | -------- | ---------------------------------------------------------------- |
| namespace    | `string` | Kubernetes namespace to which the object (or its parent) belongs |
| isNamespaced | `bool`   | Whether or not the object has an associated namespace            |

## Run Information

| Property | Type     | Description                                    |
| -------- | -------- | ---------------------------------------------- |
| runID    | `string` | Unique ULID identifying a KubeHound run        |
| cluster  | `string` | Kubernetes cluster to which the entity belongs |
