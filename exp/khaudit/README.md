# KubeHound Auditor

KubeHound Auditor is a tool that audits KubeHound dataset at scale. Used to 
automate the audit process done via Jupyter Notebooks.

## KEPs

* [KEP-0005 - Formal Investigation Process](doc/KEP0005-Formal-Investigation-Process.md)
* [KEP-0006 - Large-scale Dataset Investigation Process](doc/KEP0006-Large-scale-Dataset-Investigation-Process.md)
* [KEP-0007 - MCP Agent for KubeHound](doc/KEP0007-MCP-Agent-for-KubeHound.md)

## Investigation Process

### Container Escape Investigation

The container escape investigation process is a process that investigates container
escapes in a Kubernetes cluster. It is a process that is used to investigate
container escapes in a Kubernetes cluster.

#### Enumerate interesting namespaces and attack path profiles

> [!NOTE]
> This command will enumerate interesting namespaces and attack path profiles
> for the given cluster. It could take a while to complete.

```sh
khaudit container-escape profiles
```

Sample output:

```csv
cluster,run_id,namespace,pod_count,attack,attack_path,path_count
test.cluster.local,01jpasma5zrtby5q87v8c70ak1,default,102,container-escape,Container-->CE_MODULE_LOAD-->Node,11
test.cluster.local,01jpasma5zrtby5q87v8c70ak1,default,102,container-escape,Container-->VOLUME_DISCOVER-->Volume-->TOKEN_STEAL-->Identity-->PERMISSION_DISCOVER-->PermissionSet-->POD_PATCH-->Pod-->CONTAINER_ATTACH-->Container-->CE_PRIV_MOUNT-->Node,11
test.cluster.local,01jpasma5zrtby5q87v8c70ak1,default,102,container-escape,Container-->IDENTITY_ASSUME-->Identity-->PERMISSION_DISCOVER-->PermissionSet-->POD_PATCH-->Pod-->CONTAINER_ATTACH-->Container-->CE_PRIV_MOUNT-->Node,11
test.cluster.local,01jpasma5zrtby5q87v8c70ak1,default,102,container-escape,Container-->VOLUME_DISCOVER-->Volume-->TOKEN_STEAL-->Identity-->PERMISSION_DISCOVER-->PermissionSet-->POD_PATCH-->Pod-->CONTAINER_ATTACH-->Container-->CE_MODULE_LOAD-->Node,11
test.cluster.local,01jpasma5zrtby5q87v8c70ak1,default,102,container-escape,Container-->CE_PRIV_MOUNT-->Node,11
test.cluster.local,01jpasma5zrtby5q87v8c70ak1,default,102,container-escape,Container-->IDENTITY_ASSUME-->Identity-->PERMISSION_DISCOVER-->PermissionSet-->POD_PATCH-->Pod-->CONTAINER_ATTACH-->Container-->CE_MODULE_LOAD-->Node,11
```

#### Enumerate exploitable container images

> [!NOTE]
> This command will enumerate exploitable container images for the given namespace.
> It will provide a list of images that are the initial vector for the container
> escape attack. 
> 
> You can assume that the container images are exploitable if they are part of
> the attack path profiles enumerated in the previous step.

```sh
khaudit container-escape containers <namespace>
```

Sample output:

```csv
cluster,run_id,namespace,app,team,image
test.cluster.local,01jpasma5zrtby5q87v8c70ak1,default,postgres-haproxy-general,postgres,registry.cluster.local/postgres-haproxy:v2.4.18
test.cluster.local,01jpasma5zrtby5q87v8c70ak1,default,toolbox,data-science,registry.cluster.local/toolbox/py-toolbox:0.1.142
```

#### Enumerate exploitable container escape paths

> [!NOTE]
> This command will enumerate exploitable container escape paths for the given
> container image. It will provide the attack path to exploit as a HexTuples
> array.

```sh
khaudit container-escape paths \
    --run-id $RUN_ID \
    --cluster $CLUSTER \
    --namespace $NAMESPACE \
    <image>
```

#### Generate a markdown report

> [!NOTE]
> This command will generate a markdown report for the given cluster and run ID.
> It will provide a list of exploitable container images for the given namespace.

```sh
khaudit report md <cluster> <run-id> --namespaces <namespace1>,<namespace2>,... > report.md
```

#### Convert the attack path to a Mitre Attack Flow

Extracted paths from JanusGraph are in the format of HexTuples array.

HexTuples is an NDJSON (Newline Delimited JSON) based RDF serialisation format 
designed to achieve the best possible performance by allowing JSON streaming. 
It uses plain JSON arrays, in which the position of the items denotes 
subject, predicate, object, datatype, language, and graph.

```json
[ "<subject>", "<predicate>", "<object>", "<dataType>", "<language>", "<graph>"]
```

A JSONLD Object representation:

```json
  {
    "@context": "https://kubehound.io/schemas/v1/vertices#Container",
    "@id": "urn:vertex:10475208976",
    "app": "postgres-haproxy-terry",
    "cluster": "test.cluster.local",
    "team": "postgres"
  }
```

Becomes in HexTuples format:

```json
[
  ["urn:vertex:10475208976", "label", "Container", "", "", ""],
  ["urn:vertex:10475208976", "app", "postgres-haproxy-terry", "", "", ""],
  ["urn:vertex:10475208976", "cluster", "test.cluster.local", "", "", ""],
  ["urn:vertex:10475208976", "team", "postgres", "", "", ""]
]
```

> [!NOTE]
> This format is compatible with jsonl (json lines) format used for data streaming.

Extract the attack path from the container escape path as HexTuples array.

```sh
khaudit ce paths --run-id <run-id> --cluster <cluster> --namespace <namespace> <image> > attack.path.json
```

Filter the attack path to only include the attack path that matches the profile and convert the output to a Mitre Attack Flow.

```sh
khaudit path filter attack.path.json --profile "Container-->CE_PRIV_MOUNT-->Node" | khaudit path attackflow -
```

## Test MCP locally

Register `khaudit` as a MCP tool, by adding the following to your `~/.mcp.json`:

```json
{
  "mcpServers": {
    "kubehound": {
      "command": "$PATH_TO_KHAUDIT",
      "args":["mcp"]
    }
  }
}
```

> [!NOTE]
> Don't forget to replace `$PATH_TO_KHAUDIT` with the actual path to the `khaudit` command.

Use the [mcphost](https://github.com/mark3labs/mcphost) command to get a MCP client with prompt capabilities:

```sh
# For local setup using ollama and a model that supports function calling.
mcphost -m ollama:qwen3:14b
```

> [!NOTE]
> You can replace `ollama:qwen3:14b` with the actual model you want to use as soon 
> as the model supports function calling.

### Using WhiteRabbitNeo fine-tuned from Qwen2.5

[WhiteRabbitNeo](https://www.whiterabbitneo.com/) is a fine-tuned version of the Qwen2.5 model that supports function 
calling and is optimised for cybersecurity use cases.

You need to create an Ollama model file for the WhiteRabbitNeo model.

```modelfile
FROM hf.co/bartowski/WhiteRabbitNeo-2.5-Qwen-2.5-Coder-7B-GGUF:Q5_K_L

TEMPLATE "{{- if .Messages }}
{{- if or .System .Tools }}<|im_start|>system
{{- if .System }}
{{ .System }}
{{- end }}
{{- if .Tools }}

# Tools

You may call one or more functions to assist with the user query.

You are provided with function signatures within <tools></tools> XML tags:
<tools>
{{- range .Tools }}
{"type": "function", "function": {{ .Function }}}
{{- end }}
</tools>

For each function call, return a json object with function name and arguments within <tool_call></tool_call> XML tags:
<tool_call>
{"name": <function-name>, "arguments": <args-json-object>}
</tool_call>
{{- end }}<|im_end|>
{{ end }}
{{- range $i, $_ := .Messages }}
{{- $last := eq (len (slice $.Messages $i)) 1 -}}
{{- if eq .Role "user" }}<|im_start|>user
{{ .Content }}<|im_end|>
{{ else if eq .Role "assistant" }}<|im_start|>assistant
{{ if .Content }}{{ .Content }}
{{- else if .ToolCalls }}<tool_call>
{{ range .ToolCalls }}{"name": "{{ .Function.Name }}", "arguments": {{ .Function.Arguments }}}
{{ end }}</tool_call>
{{- end }}{{ if not $last }}<|im_end|>
{{ end }}
{{- else if eq .Role "tool" }}<|im_start|>user
<tool_response>
{{ .Content }}
</tool_response><|im_end|>
{{ end }}
{{- if and (ne .Role "assistant") $last }}<|im_start|>assistant
{{ end }}
{{- end }}
{{- else }}
{{- if .System }}<|im_start|>system
{{ .System }}<|im_end|>
{{ end }}{{ if .Prompt }}<|im_start|>user
{{ .Prompt }}<|im_end|>
{{ end }}<|im_start|>assistant
{{ end }}{{ .Response }}{{ if .Response }}<|im_end|>{{ end }}"

PARAMETER num_keep 24
PARAMETER stop <|start_header_id|>
PARAMETER stop <|end_header_id|>
PARAMETER stop <|eot_id|>
# Token context number
PARAMETER num_ctx 4096
```

Then, you can use the following command to build the Ollama model:

```sh
ollama create whiterabbitneo:7b -f Modelfile
```

You can replace the `qwen2.5:3b` with the `whiterabbitneo:7b` model during the 
`mcphost` command.

```sh
mcphost -m ollama:whiterabbitneo:7b
```
