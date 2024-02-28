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
    "cluster"       = "@cluster",
  }
}

import {
 # ID of the cloud resource
 # Check provider documentation for importable resources and format
 id = "kubehound.ingest.duration"
 # Resource address
 to = datadog_spans_metric.kubehound_ingest_duration
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

import {
 # ID of the cloud resource
 # Check provider documentation for importable resources and format
 id = "kubehound.graph.duration"
 # Resource address
 to = datadog_spans_metric.kubehound_graph_duration
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
    local.universal_tags,
    {
      # tag_name => path
      "entity" = "@entity"
  })
}

import {
 # ID of the cloud resource
 # Check provider documentation for importable resources and format
 id = "kubehound.collector.stream.duration"
 # Resource address
 to = datadog_spans_metric.kubehound_collector_stream_duration
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
    local.universal_tags,
    {
      # tag_name => path
      "label" = "@label"
  })
}

import {
 # ID of the cloud resource
 # Check provider documentation for importable resources and format
 id = "kubehound.graph.builder.edge.duration"
 # Resource address
 to = datadog_spans_metric.kubehound_graph_builder_edge_duration
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