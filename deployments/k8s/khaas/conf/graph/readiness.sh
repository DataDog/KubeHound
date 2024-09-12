#!/bin/bash
set -u

API_ENDPOINT="127.0.0.1:{{ $.Values.services.graph.port }}?gremlin="

check_query() {
    [[ $# -gt 0 ]] || return 1
    status_code=$(curl -s -o /dev/null -w "%{http_code}" -XGET "$API_ENDPOINT$1")
    [[ $status_code -ge 200 && $status_code -lt 400 ]] || return 1
}

# Request from https://github.com/JanusGraph/janusgraph/issues/2807
check_query "graph.open" || exit 1
