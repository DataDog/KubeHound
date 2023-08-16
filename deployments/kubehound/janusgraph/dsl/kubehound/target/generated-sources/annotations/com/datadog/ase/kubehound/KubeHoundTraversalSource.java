package com.datadog.ase.kubehound;

import java.lang.Class;
import java.lang.Object;
import java.lang.Override;
import java.lang.String;
import java.util.Optional;
import java.util.function.BinaryOperator;
import java.util.function.Supplier;
import java.util.function.UnaryOperator;
import org.apache.tinkerpop.gremlin.process.computer.Computer;
import org.apache.tinkerpop.gremlin.process.computer.GraphComputer;
import org.apache.tinkerpop.gremlin.process.remote.RemoteConnection;
import org.apache.tinkerpop.gremlin.process.traversal.Path;
import org.apache.tinkerpop.gremlin.process.traversal.Traversal;
import org.apache.tinkerpop.gremlin.process.traversal.TraversalStrategies;
import org.apache.tinkerpop.gremlin.process.traversal.TraversalStrategy;
import org.apache.tinkerpop.gremlin.process.traversal.dsl.graph.GraphTraversal;
import org.apache.tinkerpop.gremlin.process.traversal.step.map.AddEdgeStartStep;
import org.apache.tinkerpop.gremlin.process.traversal.step.map.AddVertexStartStep;
import org.apache.tinkerpop.gremlin.process.traversal.step.map.GraphStep;
import org.apache.tinkerpop.gremlin.process.traversal.step.sideEffect.InjectStep;
import org.apache.tinkerpop.gremlin.structure.Edge;
import org.apache.tinkerpop.gremlin.structure.Graph;
import org.apache.tinkerpop.gremlin.structure.Vertex;

public class KubeHoundTraversalSource extends KubeHoundTraversalSourceDsl {
  public KubeHoundTraversalSource(Graph graph) {
    super(graph);
  }

  public KubeHoundTraversalSource(Graph graph, TraversalStrategies strategies) {
    super(graph, strategies);
  }

  public KubeHoundTraversalSource(RemoteConnection connection) {
    super(connection);
  }

  @Override
  public KubeHoundTraversalSource clone() {
    return (KubeHoundTraversalSource) super.clone();
  }

  @Override
  public KubeHoundTraversalSource with(String key) {
    return (KubeHoundTraversalSource) super.with(key);
  }

  @Override
  public KubeHoundTraversalSource with(String key, Object value) {
    return (KubeHoundTraversalSource) super.with(key,value);
  }

  @Override
  public KubeHoundTraversalSource withStrategies(TraversalStrategy... traversalStrategies) {
    return (KubeHoundTraversalSource) super.withStrategies(traversalStrategies);
  }

  @Override
  public KubeHoundTraversalSource withoutStrategies(
      Class<? extends TraversalStrategy>... traversalStrategyClasses) {
    return (KubeHoundTraversalSource) super.withoutStrategies(traversalStrategyClasses);
  }

  @Override
  public KubeHoundTraversalSource withComputer(Computer computer) {
    return (KubeHoundTraversalSource) super.withComputer(computer);
  }

  @Override
  public KubeHoundTraversalSource withComputer(Class<? extends GraphComputer> graphComputerClass) {
    return (KubeHoundTraversalSource) super.withComputer(graphComputerClass);
  }

  @Override
  public KubeHoundTraversalSource withComputer() {
    return (KubeHoundTraversalSource) super.withComputer();
  }

  @Override
  public <A> KubeHoundTraversalSource withSideEffect(String key, Supplier<A> initialValue,
      BinaryOperator<A> reducer) {
    return (KubeHoundTraversalSource) super.withSideEffect(key,initialValue,reducer);
  }

  @Override
  public <A> KubeHoundTraversalSource withSideEffect(String key, A initialValue,
      BinaryOperator<A> reducer) {
    return (KubeHoundTraversalSource) super.withSideEffect(key,initialValue,reducer);
  }

  @Override
  public <A> KubeHoundTraversalSource withSideEffect(String key, A initialValue) {
    return (KubeHoundTraversalSource) super.withSideEffect(key,initialValue);
  }

  @Override
  public <A> KubeHoundTraversalSource withSideEffect(String key, Supplier<A> initialValue) {
    return (KubeHoundTraversalSource) super.withSideEffect(key,initialValue);
  }

  @Override
  public <A> KubeHoundTraversalSource withSack(Supplier<A> initialValue,
      UnaryOperator<A> splitOperator, BinaryOperator<A> mergeOperator) {
    return (KubeHoundTraversalSource) super.withSack(initialValue,splitOperator,mergeOperator);
  }

  @Override
  public <A> KubeHoundTraversalSource withSack(A initialValue, UnaryOperator<A> splitOperator,
      BinaryOperator<A> mergeOperator) {
    return (KubeHoundTraversalSource) super.withSack(initialValue,splitOperator,mergeOperator);
  }

  @Override
  public <A> KubeHoundTraversalSource withSack(A initialValue) {
    return (KubeHoundTraversalSource) super.withSack(initialValue);
  }

  @Override
  public <A> KubeHoundTraversalSource withSack(Supplier<A> initialValue) {
    return (KubeHoundTraversalSource) super.withSack(initialValue);
  }

  @Override
  public <A> KubeHoundTraversalSource withSack(Supplier<A> initialValue,
      UnaryOperator<A> splitOperator) {
    return (KubeHoundTraversalSource) super.withSack(initialValue,splitOperator);
  }

  @Override
  public <A> KubeHoundTraversalSource withSack(A initialValue, UnaryOperator<A> splitOperator) {
    return (KubeHoundTraversalSource) super.withSack(initialValue,splitOperator);
  }

  @Override
  public <A> KubeHoundTraversalSource withSack(Supplier<A> initialValue,
      BinaryOperator<A> mergeOperator) {
    return (KubeHoundTraversalSource) super.withSack(initialValue,mergeOperator);
  }

  @Override
  public <A> KubeHoundTraversalSource withSack(A initialValue, BinaryOperator<A> mergeOperator) {
    return (KubeHoundTraversalSource) super.withSack(initialValue,mergeOperator);
  }

  @Override
  public KubeHoundTraversalSource withBulk(boolean useBulk) {
    return (KubeHoundTraversalSource) super.withBulk(useBulk);
  }

  @Override
  public KubeHoundTraversalSource withPath() {
    return (KubeHoundTraversalSource) super.withPath();
  }

  @Override
  public KubeHoundTraversal<Vertex, Vertex> containers(String... names) {
    KubeHoundTraversalSource clone = this.clone();
    return new DefaultKubeHoundTraversal (clone, super.containers(names).asAdmin());
  }

  @Override
  public KubeHoundTraversal<Vertex, Vertex> pods(String... names) {
    KubeHoundTraversalSource clone = this.clone();
    return new DefaultKubeHoundTraversal (clone, super.pods(names).asAdmin());
  }

  @Override
  public KubeHoundTraversal<Vertex, Vertex> nodes(String... names) {
    KubeHoundTraversalSource clone = this.clone();
    return new DefaultKubeHoundTraversal (clone, super.nodes(names).asAdmin());
  }

  @Override
  public KubeHoundTraversal<Vertex, Path> escapes(String... nodeNames) {
    KubeHoundTraversalSource clone = this.clone();
    return new DefaultKubeHoundTraversal (clone, super.escapes(nodeNames).asAdmin());
  }

  @Override
  public KubeHoundTraversal<Vertex, Vertex> endpoints() {
    KubeHoundTraversalSource clone = this.clone();
    return new DefaultKubeHoundTraversal (clone, super.endpoints().asAdmin());
  }

  @Override
  public KubeHoundTraversal<Vertex, Vertex> endpoints(EndpointExposure exposure) {
    KubeHoundTraversalSource clone = this.clone();
    return new DefaultKubeHoundTraversal (clone, super.endpoints(exposure).asAdmin());
  }

  @Override
  public KubeHoundTraversal<Vertex, Vertex> services(String... portNames) {
    KubeHoundTraversalSource clone = this.clone();
    return new DefaultKubeHoundTraversal (clone, super.services(portNames).asAdmin());
  }

  @Override
  public KubeHoundTraversal<Vertex, Vertex> volumes() {
    KubeHoundTraversalSource clone = this.clone();
    return new DefaultKubeHoundTraversal (clone, super.volumes().asAdmin());
  }

  @Override
  public KubeHoundTraversal<Vertex, Vertex> hostMounts(String... sourcePaths) {
    KubeHoundTraversalSource clone = this.clone();
    return new DefaultKubeHoundTraversal (clone, super.hostMounts(sourcePaths).asAdmin());
  }

  @Override
  public KubeHoundTraversal<Vertex, Vertex> identities(String... names) {
    KubeHoundTraversalSource clone = this.clone();
    return new DefaultKubeHoundTraversal (clone, super.identities(names).asAdmin());
  }

  @Override
  public KubeHoundTraversal<Vertex, Vertex> sas(String... names) {
    KubeHoundTraversalSource clone = this.clone();
    return new DefaultKubeHoundTraversal (clone, super.sas(names).asAdmin());
  }

  @Override
  public KubeHoundTraversal<Vertex, Vertex> users(String... names) {
    KubeHoundTraversalSource clone = this.clone();
    return new DefaultKubeHoundTraversal (clone, super.users(names).asAdmin());
  }

  @Override
  public KubeHoundTraversal<Vertex, Vertex> groups(String... names) {
    KubeHoundTraversalSource clone = this.clone();
    return new DefaultKubeHoundTraversal (clone, super.groups(names).asAdmin());
  }

  @Override
  public KubeHoundTraversal<Vertex, Vertex> permissions(String... roles) {
    KubeHoundTraversalSource clone = this.clone();
    return new DefaultKubeHoundTraversal (clone, super.permissions(roles).asAdmin());
  }

  @Override
  public KubeHoundTraversal<Vertex, Vertex> addV() {
    KubeHoundTraversalSource clone = this.clone();
    clone.getBytecode().addStep(GraphTraversal.Symbols.addV);
    DefaultKubeHoundTraversal traversal = new DefaultKubeHoundTraversal(clone);
    return (KubeHoundTraversal) traversal.asAdmin().addStep(new AddVertexStartStep(traversal, (String) null));
  }

  @Override
  public KubeHoundTraversal<Vertex, Vertex> addV(String label) {
    KubeHoundTraversalSource clone = this.clone();
    clone.getBytecode().addStep(GraphTraversal.Symbols.addV, label);
    DefaultKubeHoundTraversal traversal = new DefaultKubeHoundTraversal(clone);
    return (KubeHoundTraversal) traversal.asAdmin().addStep(new AddVertexStartStep(traversal, label));
  }

  @Override
  public KubeHoundTraversal<Vertex, Vertex> addV(Traversal vertexLabelTraversal) {
    KubeHoundTraversalSource clone = this.clone();
    clone.getBytecode().addStep(GraphTraversal.Symbols.addV, vertexLabelTraversal);
    DefaultKubeHoundTraversal traversal = new DefaultKubeHoundTraversal(clone);
    return (KubeHoundTraversal) traversal.asAdmin().addStep(new AddVertexStartStep(traversal, vertexLabelTraversal));
  }

  @Override
  public KubeHoundTraversal<Edge, Edge> addE(String label) {
    KubeHoundTraversalSource clone = this.clone();
    clone.getBytecode().addStep(GraphTraversal.Symbols.addE, label);
    DefaultKubeHoundTraversal traversal = new DefaultKubeHoundTraversal(clone);
    return (KubeHoundTraversal) traversal.asAdmin().addStep(new AddEdgeStartStep(traversal, label));
  }

  @Override
  public KubeHoundTraversal<Edge, Edge> addE(Traversal edgeLabelTraversal) {
    KubeHoundTraversalSource clone = this.clone();
    clone.getBytecode().addStep(GraphTraversal.Symbols.addE, edgeLabelTraversal);
    DefaultKubeHoundTraversal traversal = new DefaultKubeHoundTraversal(clone);
    return (KubeHoundTraversal) traversal.asAdmin().addStep(new AddEdgeStartStep(traversal, edgeLabelTraversal));
  }

  @Override
  public KubeHoundTraversal<Vertex, Vertex> V(Object... vertexIds) {
    KubeHoundTraversalSource clone = this.clone();
    clone.getBytecode().addStep(GraphTraversal.Symbols.V, vertexIds);
    DefaultKubeHoundTraversal traversal = new DefaultKubeHoundTraversal(clone);
    return (KubeHoundTraversal) traversal.asAdmin().addStep(new GraphStep(traversal, Vertex.class, true, vertexIds));
  }

  @Override
  public KubeHoundTraversal<Edge, Edge> E(Object... edgeIds) {
    KubeHoundTraversalSource clone = this.clone();
    clone.getBytecode().addStep(GraphTraversal.Symbols.E, edgeIds);
    DefaultKubeHoundTraversal traversal = new DefaultKubeHoundTraversal(clone);
    return (KubeHoundTraversal) traversal.asAdmin().addStep(new GraphStep(traversal, Edge.class, true, edgeIds));
  }

  @Override
  public <S> KubeHoundTraversal<S, S> inject(S... starts) {
    KubeHoundTraversalSource clone = this.clone();
    clone.getBytecode().addStep(GraphTraversal.Symbols.inject, starts);
    DefaultKubeHoundTraversal traversal = new DefaultKubeHoundTraversal(clone);
    return (KubeHoundTraversal) traversal.asAdmin().addStep(new InjectStep(traversal, starts));
  }

  @Override
  public Optional<Class<?>> getAnonymousTraversalClass() {
    return Optional.of(__.class);
  }
}
