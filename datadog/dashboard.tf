resource "datadog_dashboard" "ordered_dashboard" {
  title        = "Kubehound V2"
  layout_type  = "ordered"
  is_read_only = false
  template_variable {
    name    = "run"
    prefix  = "run_id"
    default = "*"
  }
  widget {
    group_definition {
      title            = "Summary"
      background_color = "vivid_blue"
      layout_type      = "ordered"
      widget {
        query_value_definition {
          title       = "Run Duration"
          title_size  = "16"
          title_align = "left"
          request {
            query {
              metric_query {
                name        = "query1"
                data_source = "metrics"
                query       = "sum:kubehound.graph.duration{$run }.as_count()"
                aggregator  = "sum"
              }
            }
            query {
              metric_query {
                name        = "query2"
                data_source = "metrics"
                query       = "sum:kubehound.ingest.duration{$run }.as_count()"
                aggregator  = "sum"
              }
            }
            formula {
              formula_expression = "query1 + query2"
            }
          }
          autoscale = true
          precision = 1
        }
        widget_layout {
          x      = 0
          y      = 0
          width  = 2
          height = 2
        }
      }
      widget {
        query_value_definition {
          title       = "Ingest Duration"
          title_size  = "16"
          title_align = "left"
          request {
            query {
              metric_query {
                name        = "query1"
                data_source = "metrics"
                query       = "sum:kubehound.ingest.duration{$run }.as_count()"
                aggregator  = "sum"
              }
            }
            formula {
              formula_expression = "query1"
            }
          }
          autoscale = true
          precision = 1
        }
        widget_layout {
          x      = 2
          y      = 0
          width  = 2
          height = 2
        }
      }
      widget {
        query_value_definition {
          title       = "Graph Duration"
          title_size  = "16"
          title_align = "left"
          request {
            query {
              metric_query {
                name        = "query1"
                data_source = "metrics"
                query       = "sum:kubehound.graph.duration{$run }.as_count()"
                aggregator  = "sum"
              }
            }
            formula {
              formula_expression = "query1"
            }
          }
          autoscale = true
          precision = 1
        }
        widget_layout {
          x      = 4
          y      = 0
          width  = 2
          height = 2
        }
      }
      widget {
        query_value_definition {
          title       = "Objects Ingested"
          title_size  = "16"
          title_align = "left"
          request {
            query {
              metric_query {
                name        = "query1"
                data_source = "metrics"
                query       = "sum:kubehound.collector.count{$run }.as_count()"
                aggregator  = "sum"
              }
            }
            formula {
              formula_expression = "query1"
            }
          }
          autoscale = true
          precision = 1
        }
        widget_layout {
          x      = 6
          y      = 0
          width  = 2
          height = 2
        }
      }
      widget {
        query_value_definition {
          title       = "Vertices"
          title_size  = "16"
          title_align = "left"
          request {
            query {
              metric_query {
                name        = "query1"
                data_source = "metrics"
                query       = "sum:kubehound.storage.batchwrite.vertex{$run }.as_count()"
                aggregator  = "sum"
              }
            }
            formula {
              formula_expression = "query1"
            }
          }
          autoscale = true
          precision = 1
        }
        widget_layout {
          x      = 8
          y      = 0
          width  = 2
          height = 2
        }
      }
      widget {
        query_value_definition {
          title       = "Edges"
          title_size  = "16"
          title_align = "left"
          request {
            query {
              metric_query {
                name        = "query1"
                data_source = "metrics"
                query       = "sum:kubehound.storage.batchwrite.edge{$run }.as_count()"
                aggregator  = "sum"
              }
            }
            formula {
              formula_expression = "query1"
            }
          }
          autoscale = true
          precision = 1
        }
        widget_layout {
          x      = 10
          y      = 0
          width  = 2
          height = 2
        }
      }
    }
    widget_layout {
      x      = 0
      y      = 0
      width  = 12
      height = 3
    }
  }
  widget {
    group_definition {
      title            = "Collector"
      background_color = "vivid_green"
      layout_type      = "ordered"
      widget {
        treemap_definition {
          title = "K8 Objects"
          request {
            query {
              metric_query {
                data_source = "metrics"
                name        = "query1"
                query       = "sum:kubehound.collector.count{$run} by {entity}.as_count()"
                aggregator  = "sum"
              }
            }
            style {
              palette = "datadog16"
            }
            formula {
              formula_expression = "query1"
            }
          }
        }
        widget_layout {
          x      = 0
          y      = 0
          width  = 6
          height = 4
        }
      }
      widget {
        treemap_definition {
          title = "Ingest Pipeline Duration"
          request {
            query {
              metric_query {
                data_source = "metrics"
                name        = "query1"
                query       = "sum:kubehound.collector.stream.duration{$run } by {entity}.as_count()"
                aggregator  = "sum"
              }
            }
            style {
              palette = "datadog16"
            }
            formula {
              formula_expression = "query1"
            }
          }
        }
        widget_layout {
          x      = 6
          y      = 0
          width  = 6
          height = 4
        }
      }
    }
    widget_layout {
      x               = 0
      y               = 0
      width           = 12
      height          = 5
      is_column_break = true
    }
  }
  widget {
    group_definition {
      title            = "Graph"
      background_color = "vivid_blue"
      layout_type      = "ordered"
      widget {
        treemap_definition {
          title = "Vertex Writes"
          request {
            query {
              metric_query {
                data_source = "metrics"
                name        = "query1"
                query       = "sum:kubehound.storage.batchwrite.vertex{$run} by {label}.as_count()"
                aggregator  = "sum"
              }
            }
            style {
              palette = "datadog16"
            }
            formula {
              formula_expression = "query1"
            }
          }
        }
        widget_layout {
          x      = 0
          y      = 0
          width  = 6
          height = 4
        }
      }
      widget {
        treemap_definition {
          title = "Edge Writes"
          request {
            query {
              metric_query {
                data_source = "metrics"
                name        = "query1"
                query       = "sum:kubehound.storage.batchwrite.edge{$run} by {label}.as_count()"
                aggregator  = "sum"
              }
            }
            style {
              palette = "datadog16"
            }
            formula {
              formula_expression = "query1"
            }
          }
        }
        widget_layout {
          x      = 6
          y      = 0
          width  = 6
          height = 4
        }
      }
    }
    widget_layout {
      x      = 0
      y      = 5
      width  = 12
      height = 5
    }
  }
  widget {
    treemap_definition {
      title = "Edge Duration"
      request {
        query {
          metric_query {
            data_source = "metrics"
            name        = "query1"
            query       = "sum:kubehound.graph.builder.edge.duration{$run } by {label}.as_count()"
            aggregator  = "sum"
          }
        }
        style {
          palette = "datadog16"
        }
        formula {
          formula_expression = "query1"
        }
      }
    }
    widget_layout {
      x      = 0
      y      = 0
      width  = 9
      height = 4
    }
  }
  widget {
    toplist_definition {
      title       = "Slowest Edge Insert Rate"
      title_size  = "16"
      title_align = "left"
      request {
        formula {
          formula_expression = "query2 / query1"
        }
        query {
          metric_query {
            data_source = "metrics"
            name        = "query2"
            query       = "sum:kubehound.storage.batchwrite.edge{$run } by {label}.as_count()"
            aggregator  = "sum"
          }
        }
        query {
          metric_query {
            data_source = "metrics"
            name        = "query1"
            query       = "sum:kubehound.graph.builder.edge.duration{$run } by {label}.as_count()"
            aggregator  = "avg"
          }
        }
      }
      style {
      }
    }
    widget_layout {
      x      = 9
      y      = 0
      width  = 3
      height = 4
    }
  }
  widget {
    group_definition {
      title            = "Cache"
      background_color = "vivid_purple"
      layout_type      = "ordered"
      widget {
        toplist_definition {
          title       = "Cache Write"
          title_size  = "16"
          title_align = "left"
          request {
            formula {
              formula_expression = "query1"
            }
            query {
              metric_query {
                data_source = "metrics"
                name        = "query1"
                query       = "avg:kubehound.cache.write{$run} by {cache_key}.as_count()"
                aggregator  = "avg"
              }
            }
          }
          style {
          }
        }
        widget_layout {
          x      = 0
          y      = 0
          width  = 3
          height = 2
        }
      }
      widget {
        toplist_definition {
          title       = "Cache Hit"
          title_size  = "16"
          title_align = "left"
          request {
            formula {
              formula_expression = "query1"
            }
            query {
              metric_query {
                data_source = "metrics"
                name        = "query1"
                query       = "avg:kubehound.cache.hit{$run} by {cache_key}.as_count()"
                aggregator  = "avg"
              }
            }
          }
          style {
          }
        }
        widget_layout {
          x      = 3
          y      = 0
          width  = 3
          height = 2
        }
      }
      widget {
        toplist_definition {
          title       = "Cache Miss"
          title_size  = "16"
          title_align = "left"
          request {
            formula {
              formula_expression = "query1"
            }
            query {
              metric_query {
                data_source = "metrics"
                name        = "query1"
                query       = "avg:kubehound.cache.miss{$run} by {cache_key}.as_count()"
                aggregator  = "avg"
              }
            }
          }
          style {
          }
        }
        widget_layout {
          x      = 6
          y      = 0
          width  = 3
          height = 2
        }
      }
      widget {
        toplist_definition {
          title       = "Cache Overwrite"
          title_size  = "16"
          title_align = "left"
          request {
            formula {
              formula_expression = "query1"
            }
            query {
              metric_query {
                data_source = "metrics"
                name        = "query1"
                query       = "sum:kubehound.cache.duplicate{$run} by {cache_key}"
                aggregator  = "avg"
              }
            }
            conditional_formats {
              comparator = ">"
              value      = 0
              palette    = "white_on_red"
            }
            conditional_formats {
              comparator = "<="
              value      = 0
              palette    = "white_on_green"
            }
          }
          style {
          }
        }
        widget_layout {
          x      = 9
          y      = 0
          width  = 3
          height = 2
        }
      }
    }
    widget_layout {
      x      = 0
      y      = 14
      width  = 12
      height = 3
    }
  }
  notify_list = []
  reflow_type = "fixed"
}