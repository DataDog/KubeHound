:remote connect tinkerpop.server conf/remote.yaml session
:remote console

try {
    mgmt = graph.openManagement();

    // Get all existing indices
    allIndices = mgmt.getGraphIndexes(Vertex.class);

    // Ensure there is at least one or we are too early!
    if (allIndices.size <= 0) {
        throw new Exception("Indexes not yet created")
    }

    // Query the state of each index and check it is enabled!
    allIndices.forEach { index ->
        mgmt.awaitGraphIndexStatus(graph, index.toString()).status(SchemaStatus.ENABLED).timeout(1, java.time.temporal.ChronoUnit.SECONDS).call();
    }

    mgmt.close();
    System.out.println("[KUBEHOUND] health check success");
} finally {
    :remote close
}
