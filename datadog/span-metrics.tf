locals {
  universal_tags = {
    # tag_name => path
    "run_id"        = "@run_id",
    "env"           = "env",
    "resource_name" = "resource_name",
    "service"       = "service",
    "team"          = "team",
    "version"       = "version",
    "platform"      = "platform",
  }
}

// Duration of the data ingest
resource "datadog_spans_metric" "kubehound_ingest_duration" {
  name = "kubehound.ingest.duration"

  compute {
    aggregation_type    = "distribution"
    include_percentiles = false
    path                = "@duration"
  }

  filter {
    query = "service:kubehound operation_name:kubehound.ingestData"
  }

  dynamic "group_by" {
    for_each = local.universal_tags
    content {
      tag_name = group_by.key
      path     = group_by.value
    }
  }
}

// Duration of the graph construction
resource "datadog_spans_metric" "kubehound_graph_duration" {
  name = "kubehound.graph.duration"

  compute {
    aggregation_type    = "distribution"
    include_percentiles = false
    path                = "@duration"
  }

  filter {
    query = "service:kubehound operation_name:kubehound.buildGraph"
  }

  dynamic "group_by" {
    for_each = local.universal_tags
    content {
      tag_name = group_by.key
      path     = group_by.value
    }
  }
}

locals {
  stream_tags = merge(
    universal_tags,
    {
      # tag_name => path
      "entity" = "@entity"
  })
}

// Collector stream duration grouped by entity
resource "datadog_spans_metric" "kubehound_collector_stream_duration" {
  name = "kubehound.collector.stream.duration"

  compute {
    aggregation_type    = "distribution"
    include_percentiles = false
    path                = "@duration"
  }

  filter {
    query = "service:kubehound operation_name:kubehound.collector.stream"
  }

  dynamic "group_by" {
    for_each = local.stream_tags
    content {
      tag_name = group_by.key
      path     = group_by.value
    }
  }
}


locals {
  edge_tags = merge(
    universal_tags,
    {
      # tag_name => path
      "label" = "@label"
  })
}

// Edge builder duration grouped by label
resource "datadog_spans_metric" "kubehound_graph_builder_edge_duration" {
  name = "kubehound.graph.builder.edge.duration"

  compute {
    aggregation_type    = "distribution"
    include_percentiles = false
    path                = "@duration"
  }

  filter {
    query = "service:kubehound operation_name:kubehound.graph.builder.edge"
  }

  dynamic "group_by" {
    for_each = local.edge_tags
    content {
      tag_name = group_by.key
      path     = group_by.value
    }
  }
}