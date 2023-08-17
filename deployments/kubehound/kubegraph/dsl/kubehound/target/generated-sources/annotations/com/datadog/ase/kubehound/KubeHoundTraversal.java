package com.datadog.ase.kubehound;

import java.lang.Comparable;
import java.lang.Double;
import java.lang.Integer;
import java.lang.Long;
import java.lang.Number;
import java.lang.Object;
import java.lang.Override;
import java.lang.String;
import java.util.Collection;
import java.util.Comparator;
import java.util.Iterator;
import java.util.List;
import java.util.Map;
import java.util.function.BiFunction;
import java.util.function.Consumer;
import java.util.function.Function;
import java.util.function.Predicate;
import org.apache.tinkerpop.gremlin.process.computer.VertexProgram;
import org.apache.tinkerpop.gremlin.process.traversal.Order;
import org.apache.tinkerpop.gremlin.process.traversal.P;
import org.apache.tinkerpop.gremlin.process.traversal.Path;
import org.apache.tinkerpop.gremlin.process.traversal.Pop;
import org.apache.tinkerpop.gremlin.process.traversal.Scope;
import org.apache.tinkerpop.gremlin.process.traversal.Traversal;
import org.apache.tinkerpop.gremlin.process.traversal.Traverser;
import org.apache.tinkerpop.gremlin.process.traversal.step.util.Tree;
import org.apache.tinkerpop.gremlin.process.traversal.traverser.util.TraverserSet;
import org.apache.tinkerpop.gremlin.process.traversal.util.TraversalMetrics;
import org.apache.tinkerpop.gremlin.structure.Column;
import org.apache.tinkerpop.gremlin.structure.Direction;
import org.apache.tinkerpop.gremlin.structure.Edge;
import org.apache.tinkerpop.gremlin.structure.Element;
import org.apache.tinkerpop.gremlin.structure.Property;
import org.apache.tinkerpop.gremlin.structure.T;
import org.apache.tinkerpop.gremlin.structure.Vertex;
import org.apache.tinkerpop.gremlin.structure.VertexProperty;

public interface KubeHoundTraversal<S, E> extends KubeHoundTraversalDsl<S, E> {
  @Override
  default KubeHoundTraversal<S, Path> attacks() {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.attacks();
  }

  @Override
  default KubeHoundTraversal<S, E> critical() {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.critical();
  }

  @Override
  default KubeHoundTraversal<S, Path> criticalPaths(int maxHops) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.criticalPaths(maxHops);
  }

  @Override
  default KubeHoundTraversal<S, Path> criticalPaths() {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.criticalPaths();
  }

  @Override
  default KubeHoundTraversal<S, Path> criticalPathsFilter(String[] exclusions, int maxHops) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.criticalPathsFilter(exclusions,maxHops);
  }

  @Override
  default KubeHoundTraversal<S, E> hasCriticalPath() {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.hasCriticalPath();
  }

  @Override
  default <E2> KubeHoundTraversal<S, E2> map(Function<Traverser<E>, E2> function) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.map(function);
  }

  @Override
  default <E2> KubeHoundTraversal<S, E2> map(Traversal<?, E2> mapTraversal) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.map(mapTraversal);
  }

  @Override
  default <E2> KubeHoundTraversal<S, E2> flatMap(Function<Traverser<E>, Iterator<E2>> function) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.flatMap(function);
  }

  @Override
  default <E2> KubeHoundTraversal<S, E2> flatMap(Traversal<?, E2> flatMapTraversal) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.flatMap(flatMapTraversal);
  }

  @Override
  default KubeHoundTraversal<S, Object> id() {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.id();
  }

  @Override
  default KubeHoundTraversal<S, String> label() {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.label();
  }

  @Override
  default KubeHoundTraversal<S, E> identity() {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.identity();
  }

  @Override
  default <E2> KubeHoundTraversal<S, E2> constant(E2 e) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.constant(e);
  }

  @Override
  default KubeHoundTraversal<S, Vertex> V(Object... vertexIdsOrElements) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.V(vertexIdsOrElements);
  }

  @Override
  default KubeHoundTraversal<S, Vertex> to(Direction direction, String... edgeLabels) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.to(direction,edgeLabels);
  }

  @Override
  default KubeHoundTraversal<S, Vertex> out(String... edgeLabels) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.out(edgeLabels);
  }

  @Override
  default KubeHoundTraversal<S, Vertex> in(String... edgeLabels) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.in(edgeLabels);
  }

  @Override
  default KubeHoundTraversal<S, Vertex> both(String... edgeLabels) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.both(edgeLabels);
  }

  @Override
  default KubeHoundTraversal<S, Edge> toE(Direction direction, String... edgeLabels) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.toE(direction,edgeLabels);
  }

  @Override
  default KubeHoundTraversal<S, Edge> outE(String... edgeLabels) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.outE(edgeLabels);
  }

  @Override
  default KubeHoundTraversal<S, Edge> inE(String... edgeLabels) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.inE(edgeLabels);
  }

  @Override
  default KubeHoundTraversal<S, Edge> bothE(String... edgeLabels) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.bothE(edgeLabels);
  }

  @Override
  default KubeHoundTraversal<S, Vertex> toV(Direction direction) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.toV(direction);
  }

  @Override
  default KubeHoundTraversal<S, Vertex> inV() {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.inV();
  }

  @Override
  default KubeHoundTraversal<S, Vertex> outV() {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.outV();
  }

  @Override
  default KubeHoundTraversal<S, Vertex> bothV() {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.bothV();
  }

  @Override
  default KubeHoundTraversal<S, Vertex> otherV() {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.otherV();
  }

  @Override
  default KubeHoundTraversal<S, E> order() {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.order();
  }

  @Override
  default KubeHoundTraversal<S, E> order(Scope scope) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.order(scope);
  }

  @Override
  default <E2> KubeHoundTraversal<S, ? extends Property<E2>> properties(String... propertyKeys) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.properties(propertyKeys);
  }

  @Override
  default <E2> KubeHoundTraversal<S, E2> values(String... propertyKeys) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.values(propertyKeys);
  }

  @Override
  default <E2> KubeHoundTraversal<S, Map<String, E2>> propertyMap(String... propertyKeys) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.propertyMap(propertyKeys);
  }

  @Override
  default <E2> KubeHoundTraversal<S, Map<Object, E2>> elementMap(String... propertyKeys) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.elementMap(propertyKeys);
  }

  @Override
  default <E2> KubeHoundTraversal<S, Map<Object, E2>> valueMap(String... propertyKeys) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.valueMap(propertyKeys);
  }

  @Override
  default <E2> KubeHoundTraversal<S, Map<Object, E2>> valueMap(boolean includeTokens,
      String... propertyKeys) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.valueMap(includeTokens,propertyKeys);
  }

  @Override
  default KubeHoundTraversal<S, String> key() {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.key();
  }

  @Override
  default <E2> KubeHoundTraversal<S, E2> value() {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.value();
  }

  @Override
  default KubeHoundTraversal<S, Path> path() {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.path();
  }

  @Override
  default <E2> KubeHoundTraversal<S, Map<String, E2>> match(Traversal<?, ?>... matchTraversals) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.match(matchTraversals);
  }

  @Override
  default <E2> KubeHoundTraversal<S, E2> sack() {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.sack();
  }

  @Override
  default KubeHoundTraversal<S, Integer> loops() {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.loops();
  }

  @Override
  default KubeHoundTraversal<S, Integer> loops(String loopName) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.loops(loopName);
  }

  @Override
  default <E2> KubeHoundTraversal<S, Map<String, E2>> project(String projectKey,
      String... otherProjectKeys) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.project(projectKey,otherProjectKeys);
  }

  @Override
  default <E2> KubeHoundTraversal<S, Map<String, E2>> select(Pop pop, String selectKey1,
      String selectKey2, String... otherSelectKeys) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.select(pop,selectKey1,selectKey2,otherSelectKeys);
  }

  @Override
  default <E2> KubeHoundTraversal<S, Map<String, E2>> select(String selectKey1, String selectKey2,
      String... otherSelectKeys) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.select(selectKey1,selectKey2,otherSelectKeys);
  }

  @Override
  default <E2> KubeHoundTraversal<S, E2> select(Pop pop, String selectKey) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.select(pop,selectKey);
  }

  @Override
  default <E2> KubeHoundTraversal<S, E2> select(String selectKey) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.select(selectKey);
  }

  @Override
  default <E2> KubeHoundTraversal<S, E2> select(Pop pop, Traversal<S, E2> keyTraversal) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.select(pop,keyTraversal);
  }

  @Override
  default <E2> KubeHoundTraversal<S, E2> select(Traversal<S, E2> keyTraversal) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.select(keyTraversal);
  }

  @Override
  default <E2> KubeHoundTraversal<S, Collection<E2>> select(Column column) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.select(column);
  }

  @Override
  default <E2> KubeHoundTraversal<S, E2> unfold() {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.unfold();
  }

  @Override
  default KubeHoundTraversal<S, List<E>> fold() {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.fold();
  }

  @Override
  default <E2> KubeHoundTraversal<S, E2> fold(E2 seed, BiFunction<E2, E, E2> foldFunction) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.fold(seed,foldFunction);
  }

  @Override
  default KubeHoundTraversal<S, Long> count() {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.count();
  }

  @Override
  default KubeHoundTraversal<S, Long> count(Scope scope) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.count(scope);
  }

  @Override
  default <E2 extends Number> KubeHoundTraversal<S, E2> sum() {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.sum();
  }

  @Override
  default <E2 extends Number> KubeHoundTraversal<S, E2> sum(Scope scope) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.sum(scope);
  }

  @Override
  default <E2 extends Comparable> KubeHoundTraversal<S, E2> max() {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.max();
  }

  @Override
  default <E2 extends Comparable> KubeHoundTraversal<S, E2> max(Scope scope) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.max(scope);
  }

  @Override
  default <E2 extends Comparable> KubeHoundTraversal<S, E2> min() {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.min();
  }

  @Override
  default <E2 extends Comparable> KubeHoundTraversal<S, E2> min(Scope scope) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.min(scope);
  }

  @Override
  default <E2 extends Number> KubeHoundTraversal<S, E2> mean() {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.mean();
  }

  @Override
  default <E2 extends Number> KubeHoundTraversal<S, E2> mean(Scope scope) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.mean(scope);
  }

  @Override
  default <K, V> KubeHoundTraversal<S, Map<K, V>> group() {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.group();
  }

  @Override
  default <K> KubeHoundTraversal<S, Map<K, Long>> groupCount() {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.groupCount();
  }

  @Override
  default KubeHoundTraversal<S, Tree> tree() {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.tree();
  }

  @Override
  default KubeHoundTraversal<S, Vertex> addV(String vertexLabel) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.addV(vertexLabel);
  }

  @Override
  default KubeHoundTraversal<S, Vertex> addV(Traversal<?, String> vertexLabelTraversal) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.addV(vertexLabelTraversal);
  }

  @Override
  default KubeHoundTraversal<S, Vertex> addV() {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.addV();
  }

  @Override
  default KubeHoundTraversal<S, Vertex> mergeV() {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.mergeV();
  }

  @Override
  default KubeHoundTraversal<S, Vertex> mergeV(Map<Object, Object> searchCreate) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.mergeV(searchCreate);
  }

  @Override
  default KubeHoundTraversal<S, Vertex> mergeV(Traversal<?, Map<Object, Object>> searchCreate) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.mergeV(searchCreate);
  }

  @Override
  default KubeHoundTraversal<S, Edge> mergeE() {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.mergeE();
  }

  @Override
  default KubeHoundTraversal<S, Edge> mergeE(Map<Object, Object> searchCreate) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.mergeE(searchCreate);
  }

  @Override
  default KubeHoundTraversal<S, Edge> mergeE(Traversal<?, Map<Object, Object>> searchCreate) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.mergeE(searchCreate);
  }

  @Override
  default KubeHoundTraversal<S, Edge> addE(String edgeLabel) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.addE(edgeLabel);
  }

  @Override
  default KubeHoundTraversal<S, Edge> addE(Traversal<?, String> edgeLabelTraversal) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.addE(edgeLabelTraversal);
  }

  @Override
  default KubeHoundTraversal<S, E> to(String toStepLabel) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.to(toStepLabel);
  }

  @Override
  default KubeHoundTraversal<S, E> from(String fromStepLabel) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.from(fromStepLabel);
  }

  @Override
  default KubeHoundTraversal<S, E> to(Traversal<?, Vertex> toVertex) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.to(toVertex);
  }

  @Override
  default KubeHoundTraversal<S, E> from(Traversal<?, Vertex> fromVertex) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.from(fromVertex);
  }

  @Override
  default KubeHoundTraversal<S, E> to(Vertex toVertex) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.to(toVertex);
  }

  @Override
  default KubeHoundTraversal<S, E> from(Vertex fromVertex) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.from(fromVertex);
  }

  @Override
  default KubeHoundTraversal<S, Double> math(String expression) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.math(expression);
  }

  @Override
  default KubeHoundTraversal<S, Element> element() {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.element();
  }

  @Override
  default <E> KubeHoundTraversal<S, E> call(String service) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.call(service);
  }

  @Override
  default <E> KubeHoundTraversal<S, E> call(String service, Map params) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.call(service,params);
  }

  @Override
  default <E> KubeHoundTraversal<S, E> call(String service,
      Traversal<?, Map<?, ?>> childTraversal) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.call(service,childTraversal);
  }

  @Override
  default <E> KubeHoundTraversal<S, E> call(String service, Map params,
      Traversal<?, Map<?, ?>> childTraversal) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.call(service,params,childTraversal);
  }

  @Override
  default KubeHoundTraversal<S, E> filter(Predicate<Traverser<E>> predicate) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.filter(predicate);
  }

  @Override
  default KubeHoundTraversal<S, E> filter(Traversal<?, ?> filterTraversal) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.filter(filterTraversal);
  }

  @Override
  default KubeHoundTraversal<S, E> none() {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.none();
  }

  @Override
  default KubeHoundTraversal<S, E> or(Traversal<?, ?>... orTraversals) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.or(orTraversals);
  }

  @Override
  default KubeHoundTraversal<S, E> and(Traversal<?, ?>... andTraversals) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.and(andTraversals);
  }

  @Override
  default KubeHoundTraversal<S, E> inject(E... injections) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.inject(injections);
  }

  @Override
  default KubeHoundTraversal<S, E> dedup(Scope scope, String... dedupLabels) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.dedup(scope,dedupLabels);
  }

  @Override
  default KubeHoundTraversal<S, E> dedup(String... dedupLabels) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.dedup(dedupLabels);
  }

  @Override
  default KubeHoundTraversal<S, E> where(String startKey, P<String> predicate) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.where(startKey,predicate);
  }

  @Override
  default KubeHoundTraversal<S, E> where(P<String> predicate) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.where(predicate);
  }

  @Override
  default KubeHoundTraversal<S, E> where(Traversal<?, ?> whereTraversal) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.where(whereTraversal);
  }

  @Override
  default KubeHoundTraversal<S, E> has(String propertyKey, P<?> predicate) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.has(propertyKey,predicate);
  }

  @Override
  default KubeHoundTraversal<S, E> has(T accessor, P<?> predicate) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.has(accessor,predicate);
  }

  @Override
  default KubeHoundTraversal<S, E> has(String propertyKey, Object value) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.has(propertyKey,value);
  }

  @Override
  default KubeHoundTraversal<S, E> has(T accessor, Object value) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.has(accessor,value);
  }

  @Override
  default KubeHoundTraversal<S, E> has(String label, String propertyKey, P<?> predicate) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.has(label,propertyKey,predicate);
  }

  @Override
  default KubeHoundTraversal<S, E> has(String label, String propertyKey, Object value) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.has(label,propertyKey,value);
  }

  @Override
  default KubeHoundTraversal<S, E> has(T accessor, Traversal<?, ?> propertyTraversal) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.has(accessor,propertyTraversal);
  }

  @Override
  default KubeHoundTraversal<S, E> has(String propertyKey, Traversal<?, ?> propertyTraversal) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.has(propertyKey,propertyTraversal);
  }

  @Override
  default KubeHoundTraversal<S, E> has(String propertyKey) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.has(propertyKey);
  }

  @Override
  default KubeHoundTraversal<S, E> hasNot(String propertyKey) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.hasNot(propertyKey);
  }

  @Override
  default KubeHoundTraversal<S, E> hasLabel(String label, String... otherLabels) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.hasLabel(label,otherLabels);
  }

  @Override
  default KubeHoundTraversal<S, E> hasLabel(P<String> predicate) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.hasLabel(predicate);
  }

  @Override
  default KubeHoundTraversal<S, E> hasId(Object id, Object... otherIds) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.hasId(id,otherIds);
  }

  @Override
  default KubeHoundTraversal<S, E> hasId(P<Object> predicate) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.hasId(predicate);
  }

  @Override
  default KubeHoundTraversal<S, E> hasKey(String label, String... otherLabels) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.hasKey(label,otherLabels);
  }

  @Override
  default KubeHoundTraversal<S, E> hasKey(P<String> predicate) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.hasKey(predicate);
  }

  @Override
  default KubeHoundTraversal<S, E> hasValue(Object value, Object... otherValues) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.hasValue(value,otherValues);
  }

  @Override
  default KubeHoundTraversal<S, E> hasValue(P<Object> predicate) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.hasValue(predicate);
  }

  @Override
  default KubeHoundTraversal<S, E> is(P<E> predicate) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.is(predicate);
  }

  @Override
  default KubeHoundTraversal<S, E> is(Object value) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.is(value);
  }

  @Override
  default KubeHoundTraversal<S, E> not(Traversal<?, ?> notTraversal) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.not(notTraversal);
  }

  @Override
  default KubeHoundTraversal<S, E> coin(double probability) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.coin(probability);
  }

  @Override
  default KubeHoundTraversal<S, E> range(long low, long arg1) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.range(low,arg1);
  }

  @Override
  default <E2> KubeHoundTraversal<S, E2> range(Scope scope, long low, long arg2) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.range(scope,low,arg2);
  }

  @Override
  default KubeHoundTraversal<S, E> limit(long limit) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.limit(limit);
  }

  @Override
  default <E2> KubeHoundTraversal<S, E2> limit(Scope scope, long limit) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.limit(scope,limit);
  }

  @Override
  default KubeHoundTraversal<S, E> tail() {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.tail();
  }

  @Override
  default KubeHoundTraversal<S, E> tail(long limit) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.tail(limit);
  }

  @Override
  default <E2> KubeHoundTraversal<S, E2> tail(Scope scope) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.tail(scope);
  }

  @Override
  default <E2> KubeHoundTraversal<S, E2> tail(Scope scope, long limit) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.tail(scope,limit);
  }

  @Override
  default KubeHoundTraversal<S, E> skip(long skip) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.skip(skip);
  }

  @Override
  default <E2> KubeHoundTraversal<S, E2> skip(Scope scope, long skip) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.skip(scope,skip);
  }

  @Override
  default KubeHoundTraversal<S, E> timeLimit(long timeLimit) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.timeLimit(timeLimit);
  }

  @Override
  default KubeHoundTraversal<S, E> simplePath() {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.simplePath();
  }

  @Override
  default KubeHoundTraversal<S, E> cyclicPath() {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.cyclicPath();
  }

  @Override
  default KubeHoundTraversal<S, E> sample(int amountToSample) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.sample(amountToSample);
  }

  @Override
  default KubeHoundTraversal<S, E> sample(Scope scope, int amountToSample) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.sample(scope,amountToSample);
  }

  @Override
  default KubeHoundTraversal<S, E> drop() {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.drop();
  }

  @Override
  default KubeHoundTraversal<S, E> sideEffect(Consumer<Traverser<E>> consumer) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.sideEffect(consumer);
  }

  @Override
  default KubeHoundTraversal<S, E> sideEffect(Traversal<?, ?> sideEffectTraversal) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.sideEffect(sideEffectTraversal);
  }

  @Override
  default <E2> KubeHoundTraversal<S, E2> cap(String sideEffectKey, String... sideEffectKeys) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.cap(sideEffectKey,sideEffectKeys);
  }

  @Override
  default KubeHoundTraversal<S, Edge> subgraph(String sideEffectKey) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.subgraph(sideEffectKey);
  }

  @Override
  default KubeHoundTraversal<S, E> aggregate(String sideEffectKey) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.aggregate(sideEffectKey);
  }

  @Override
  default KubeHoundTraversal<S, E> aggregate(Scope scope, String sideEffectKey) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.aggregate(scope,sideEffectKey);
  }

  @Override
  default KubeHoundTraversal<S, E> group(String sideEffectKey) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.group(sideEffectKey);
  }

  @Override
  default KubeHoundTraversal<S, E> groupCount(String sideEffectKey) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.groupCount(sideEffectKey);
  }

  @Override
  default KubeHoundTraversal<S, E> fail() {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.fail();
  }

  @Override
  default KubeHoundTraversal<S, E> fail(String message) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.fail(message);
  }

  @Override
  default KubeHoundTraversal<S, E> tree(String sideEffectKey) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.tree(sideEffectKey);
  }

  @Override
  default <V, U> KubeHoundTraversal<S, E> sack(BiFunction<V, U, V> sackOperator) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.sack(sackOperator);
  }

  @Override
  default KubeHoundTraversal<S, E> store(String sideEffectKey) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.store(sideEffectKey);
  }

  @Override
  default KubeHoundTraversal<S, E> profile(String sideEffectKey) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.profile(sideEffectKey);
  }

  @Override
  default KubeHoundTraversal<S, TraversalMetrics> profile() {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.profile();
  }

  @Override
  default KubeHoundTraversal<S, E> property(VertexProperty.Cardinality cardinality, Object key,
      Object value, Object... keyValues) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.property(cardinality,key,value,keyValues);
  }

  @Override
  default KubeHoundTraversal<S, E> property(Object key, Object value, Object... keyValues) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.property(key,value,keyValues);
  }

  @Override
  default KubeHoundTraversal<S, E> property(Map<Object, Object> value) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.property(value);
  }

  @Override
  default <M, E2> KubeHoundTraversal<S, E2> branch(Traversal<?, M> branchTraversal) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.branch(branchTraversal);
  }

  @Override
  default <M, E2> KubeHoundTraversal<S, E2> branch(Function<Traverser<E>, M> function) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.branch(function);
  }

  @Override
  default <M, E2> KubeHoundTraversal<S, E2> choose(Traversal<?, M> choiceTraversal) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.choose(choiceTraversal);
  }

  @Override
  default <E2> KubeHoundTraversal<S, E2> choose(Traversal<?, ?> traversalPredicate,
      Traversal<?, E2> trueChoice, Traversal<?, E2> falseChoice) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.choose(traversalPredicate,trueChoice,falseChoice);
  }

  @Override
  default <E2> KubeHoundTraversal<S, E2> choose(Traversal<?, ?> traversalPredicate,
      Traversal<?, E2> trueChoice) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.choose(traversalPredicate,trueChoice);
  }

  @Override
  default <M, E2> KubeHoundTraversal<S, E2> choose(Function<E, M> choiceFunction) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.choose(choiceFunction);
  }

  @Override
  default <E2> KubeHoundTraversal<S, E2> choose(Predicate<E> choosePredicate,
      Traversal<?, E2> trueChoice, Traversal<?, E2> falseChoice) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.choose(choosePredicate,trueChoice,falseChoice);
  }

  @Override
  default <E2> KubeHoundTraversal<S, E2> choose(Predicate<E> choosePredicate,
      Traversal<?, E2> trueChoice) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.choose(choosePredicate,trueChoice);
  }

  @Override
  default <E2> KubeHoundTraversal<S, E2> optional(Traversal<?, E2> optionalTraversal) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.optional(optionalTraversal);
  }

  @Override
  default <E2> KubeHoundTraversal<S, E2> union(Traversal<?, E2>... unionTraversals) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.union(unionTraversals);
  }

  @Override
  default <E2> KubeHoundTraversal<S, E2> coalesce(Traversal<?, E2>... coalesceTraversals) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.coalesce(coalesceTraversals);
  }

  @Override
  default KubeHoundTraversal<S, E> repeat(Traversal<?, E> repeatTraversal) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.repeat(repeatTraversal);
  }

  @Override
  default KubeHoundTraversal<S, E> repeat(String loopName, Traversal<?, E> repeatTraversal) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.repeat(loopName,repeatTraversal);
  }

  @Override
  default KubeHoundTraversal<S, E> emit(Traversal<?, ?> emitTraversal) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.emit(emitTraversal);
  }

  @Override
  default KubeHoundTraversal<S, E> emit(Predicate<Traverser<E>> emitPredicate) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.emit(emitPredicate);
  }

  @Override
  default KubeHoundTraversal<S, E> emit() {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.emit();
  }

  @Override
  default KubeHoundTraversal<S, E> until(Traversal<?, ?> untilTraversal) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.until(untilTraversal);
  }

  @Override
  default KubeHoundTraversal<S, E> until(Predicate<Traverser<E>> untilPredicate) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.until(untilPredicate);
  }

  @Override
  default KubeHoundTraversal<S, E> times(int maxLoops) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.times(maxLoops);
  }

  @Override
  default <E2> KubeHoundTraversal<S, E2> local(Traversal<?, E2> localTraversal) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.local(localTraversal);
  }

  @Override
  default KubeHoundTraversal<S, E> pageRank() {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.pageRank();
  }

  @Override
  default KubeHoundTraversal<S, E> pageRank(double alpha) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.pageRank(alpha);
  }

  @Override
  default KubeHoundTraversal<S, E> peerPressure() {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.peerPressure();
  }

  @Override
  default KubeHoundTraversal<S, E> connectedComponent() {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.connectedComponent();
  }

  @Override
  default KubeHoundTraversal<S, Path> shortestPath() {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.shortestPath();
  }

  @Override
  default KubeHoundTraversal<S, E> program(VertexProgram<?> vertexProgram) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.program(vertexProgram);
  }

  @Override
  default KubeHoundTraversal<S, E> as(String stepLabel, String... stepLabels) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.as(stepLabel,stepLabels);
  }

  @Override
  default KubeHoundTraversal<S, E> barrier() {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.barrier();
  }

  @Override
  default KubeHoundTraversal<S, E> barrier(int maxBarrierSize) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.barrier(maxBarrierSize);
  }

  @Override
  default <E2> KubeHoundTraversal<S, E2> index() {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.index();
  }

  @Override
  default KubeHoundTraversal<S, E> barrier(Consumer<TraverserSet<Object>> barrierConsumer) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.barrier(barrierConsumer);
  }

  @Override
  default KubeHoundTraversal<S, E> with(String key) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.with(key);
  }

  @Override
  default KubeHoundTraversal<S, E> with(String key, Object value) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.with(key,value);
  }

  @Override
  default KubeHoundTraversal<S, E> by() {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.by();
  }

  @Override
  default KubeHoundTraversal<S, E> by(Traversal<?, ?> traversal) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.by(traversal);
  }

  @Override
  default KubeHoundTraversal<S, E> by(T token) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.by(token);
  }

  @Override
  default KubeHoundTraversal<S, E> by(String key) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.by(key);
  }

  @Override
  default <V> KubeHoundTraversal<S, E> by(Function<V, Object> function) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.by(function);
  }

  @Override
  default <V> KubeHoundTraversal<S, E> by(Traversal<?, ?> traversal, Comparator<V> comparator) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.by(traversal,comparator);
  }

  @Override
  default KubeHoundTraversal<S, E> by(Comparator<E> comparator) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.by(comparator);
  }

  @Override
  default KubeHoundTraversal<S, E> by(Order order) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.by(order);
  }

  @Override
  default <V> KubeHoundTraversal<S, E> by(String key, Comparator<V> comparator) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.by(key,comparator);
  }

  @Override
  default <U> KubeHoundTraversal<S, E> by(Function<U, Object> function, Comparator comparator) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.by(function,comparator);
  }

  @Override
  default <M, E2> KubeHoundTraversal<S, E> option(M token, Traversal<?, E2> traversalOption) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.option(token,traversalOption);
  }

  @Override
  default <M, E2> KubeHoundTraversal<S, E> option(M token, Map<Object, Object> m) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.option(token,m);
  }

  @Override
  default <E2> KubeHoundTraversal<S, E> option(Traversal<?, E2> traversalOption) {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.option(traversalOption);
  }

  @Override
  default KubeHoundTraversal<S, E> read() {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.read();
  }

  @Override
  default KubeHoundTraversal<S, E> write() {
    return (KubeHoundTraversal) KubeHoundTraversalDsl.super.write();
  }

  @Override
  default KubeHoundTraversal<S, E> iterate() {
    KubeHoundTraversalDsl.super.iterate();
    return this;
  }
}
