package com.datadog.ase.kubehound;

import java.lang.Override;
import org.apache.tinkerpop.gremlin.process.traversal.dsl.graph.GraphTraversal;
import org.apache.tinkerpop.gremlin.process.traversal.util.DefaultTraversal;
import org.apache.tinkerpop.gremlin.structure.Graph;

public class DefaultKubeHoundTraversal<S, E> extends DefaultTraversal<S, E> implements KubeHoundTraversal<S, E> {
  public DefaultKubeHoundTraversal() {
    super();
  }

  public DefaultKubeHoundTraversal(Graph graph) {
    super(graph);
  }

  public DefaultKubeHoundTraversal(KubeHoundTraversalSource traversalSource) {
    super(traversalSource);
  }

  public DefaultKubeHoundTraversal(KubeHoundTraversalSource traversalSource,
      GraphTraversal.Admin traversal) {
    super(traversalSource, traversal.asAdmin());
  }

  @Override
  public KubeHoundTraversal<S, E> iterate() {
    return (KubeHoundTraversal) super.iterate();
  }

  @Override
  public GraphTraversal.Admin<S, E> asAdmin() {
    return (GraphTraversal.Admin) super.asAdmin();
  }

  @Override
  public DefaultKubeHoundTraversal<S, E> clone() {
    return (DefaultKubeHoundTraversal) super.clone();
  }
}
