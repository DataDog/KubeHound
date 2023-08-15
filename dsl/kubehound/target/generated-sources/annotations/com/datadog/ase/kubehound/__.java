package com.datadog.ase.kubehound;

import java.lang.Comparable;
import java.lang.Double;
import java.lang.Integer;
import java.lang.Long;
import java.lang.Object;
import java.lang.String;
import java.util.Collection;
import java.util.Iterator;
import java.util.List;
import java.util.Map;
import java.util.function.BiFunction;
import java.util.function.Consumer;
import java.util.function.Function;
import java.util.function.Predicate;
import org.apache.tinkerpop.gremlin.process.traversal.P;
import org.apache.tinkerpop.gremlin.process.traversal.Path;
import org.apache.tinkerpop.gremlin.process.traversal.Pop;
import org.apache.tinkerpop.gremlin.process.traversal.Scope;
import org.apache.tinkerpop.gremlin.process.traversal.Traversal;
import org.apache.tinkerpop.gremlin.process.traversal.Traverser;
import org.apache.tinkerpop.gremlin.process.traversal.step.util.Tree;
import org.apache.tinkerpop.gremlin.process.traversal.traverser.util.TraverserSet;
import org.apache.tinkerpop.gremlin.structure.Column;
import org.apache.tinkerpop.gremlin.structure.Direction;
import org.apache.tinkerpop.gremlin.structure.Edge;
import org.apache.tinkerpop.gremlin.structure.Element;
import org.apache.tinkerpop.gremlin.structure.Property;
import org.apache.tinkerpop.gremlin.structure.T;
import org.apache.tinkerpop.gremlin.structure.Vertex;
import org.apache.tinkerpop.gremlin.structure.VertexProperty;

public final class __ {
  private __() {
  }

  public static <A> KubeHoundTraversal<A, A> start() {
    return new DefaultKubeHoundTraversal<>();
  }

  public static <S> KubeHoundTraversal<S, Edge> attacks() {
    return __.<S>start().attacks();
  }

  public static <A> KubeHoundTraversal<A, A> critical() {
    return __.<A>start().critical();
  }

  public static <A> KubeHoundTraversal<A, A> __(A... starts) {
    return inject(starts);
  }

  public static <A, B> KubeHoundTraversal<A, B> map(Function<Traverser<A>, B> function) {
    return __.<A>start().map(function);
  }

  public static <A, B> KubeHoundTraversal<A, B> map(Traversal<?, B> mapTraversal) {
    return __.<A>start().map(mapTraversal);
  }

  public static <A, B> KubeHoundTraversal<A, B> flatMap(
      Function<Traverser<A>, Iterator<B>> function) {
    return __.<A>start().flatMap(function);
  }

  public static <A, B> KubeHoundTraversal<A, B> flatMap(Traversal<?, B> flatMapTraversal) {
    return __.<A>start().flatMap(flatMapTraversal);
  }

  public static <A> KubeHoundTraversal<A, A> identity() {
    return __.<A>start().identity();
  }

  public static <A> KubeHoundTraversal<A, A> constant(A a) {
    return __.<A>start().constant(a);
  }

  public static <A extends Element> KubeHoundTraversal<A, String> label() {
    return __.<A>start().label();
  }

  public static <A extends Element> KubeHoundTraversal<A, Object> id() {
    return __.<A>start().id();
  }

  public static <A> KubeHoundTraversal<A, Vertex> V(Object... vertexIdsOrElements) {
    return __.<A>start().V(vertexIdsOrElements);
  }

  public static KubeHoundTraversal<Vertex, Vertex> to(Direction direction, String... edgeLabels) {
    return __.<Vertex>start().to(direction,edgeLabels);
  }

  public static KubeHoundTraversal<Vertex, Vertex> out(String... edgeLabels) {
    return __.<Vertex>start().out(edgeLabels);
  }

  public static KubeHoundTraversal<Vertex, Vertex> in(String... edgeLabels) {
    return __.<Vertex>start().in(edgeLabels);
  }

  public static KubeHoundTraversal<Vertex, Vertex> both(String... edgeLabels) {
    return __.<Vertex>start().both(edgeLabels);
  }

  public static KubeHoundTraversal<Vertex, Edge> toE(Direction direction, String... edgeLabels) {
    return __.<Vertex>start().toE(direction,edgeLabels);
  }

  public static KubeHoundTraversal<Vertex, Edge> outE(String... edgeLabels) {
    return __.<Vertex>start().outE(edgeLabels);
  }

  public static KubeHoundTraversal<Vertex, Edge> inE(String... edgeLabels) {
    return __.<Vertex>start().inE(edgeLabels);
  }

  public static KubeHoundTraversal<Vertex, Edge> bothE(String... edgeLabels) {
    return __.<Vertex>start().bothE(edgeLabels);
  }

  public static KubeHoundTraversal<Edge, Vertex> toV(Direction direction) {
    return __.<Edge>start().toV(direction);
  }

  public static KubeHoundTraversal<Edge, Vertex> inV() {
    return __.<Edge>start().inV();
  }

  public static KubeHoundTraversal<Edge, Vertex> outV() {
    return __.<Edge>start().outV();
  }

  public static KubeHoundTraversal<Edge, Vertex> bothV() {
    return __.<Edge>start().bothV();
  }

  public static KubeHoundTraversal<Edge, Vertex> otherV() {
    return __.<Edge>start().otherV();
  }

  public static <A> KubeHoundTraversal<A, A> order() {
    return __.<A>start().order();
  }

  public static <A> KubeHoundTraversal<A, A> order(Scope scope) {
    return __.<A>start().order(scope);
  }

  public static <A extends Element, B> KubeHoundTraversal<A, ? extends Property<B>> properties(
      String... propertyKeys) {
    return __.<A>start().properties(propertyKeys);
  }

  public static <A extends Element, B> KubeHoundTraversal<A, B> values(String... propertyKeys) {
    return __.<A>start().values(propertyKeys);
  }

  public static <A extends Element, B> KubeHoundTraversal<A, Map<String, B>> propertyMap(
      String... propertyKeys) {
    return __.<A>start().propertyMap(propertyKeys);
  }

  public static <A extends Element, B> KubeHoundTraversal<A, Map<Object, B>> elementMap(
      String... propertyKeys) {
    return __.<A>start().elementMap(propertyKeys);
  }

  public static <A extends Element, B> KubeHoundTraversal<A, Map<Object, B>> valueMap(
      String... propertyKeys) {
    return __.<A>start().valueMap(propertyKeys);
  }

  public static <A extends Element, B> KubeHoundTraversal<A, Map<Object, B>> valueMap(
      boolean includeTokens, String... propertyKeys) {
    return __.<A>start().valueMap(includeTokens,propertyKeys);
  }

  public static <A, B> KubeHoundTraversal<A, Map<String, B>> project(String projectKey,
      String... projectKeys) {
    return __.<A>start().project(projectKey,projectKeys);
  }

  public static <A, B> KubeHoundTraversal<A, Collection<B>> select(Column column) {
    return __.<A>start().select(column);
  }

  public static <A extends Property> KubeHoundTraversal<A, String> key() {
    return __.<A>start().key();
  }

  public static <A extends Property, B> KubeHoundTraversal<A, B> value() {
    return __.<A>start().value();
  }

  public static <A> KubeHoundTraversal<A, Path> path() {
    return __.<A>start().path();
  }

  public static <A, B> KubeHoundTraversal<A, Map<String, B>> match(
      Traversal<?, ?>... matchTraversals) {
    return __.<A>start().match(matchTraversals);
  }

  public static <A, B> KubeHoundTraversal<A, B> sack() {
    return __.<A>start().sack();
  }

  public static <A> KubeHoundTraversal<A, Integer> loops() {
    return __.<A>start().loops();
  }

  public static <A> KubeHoundTraversal<A, Integer> loops(String loopName) {
    return __.<A>start().loops(loopName);
  }

  public static <A, B> KubeHoundTraversal<A, B> select(Pop pop, String selectKey) {
    return __.<A>start().select(pop,selectKey);
  }

  public static <A, B> KubeHoundTraversal<A, B> select(String selectKey) {
    return __.<A>start().select(selectKey);
  }

  public static <A, B> KubeHoundTraversal<A, Map<String, B>> select(Pop pop, String selectKey1,
      String selectKey2, String... otherSelectKeys) {
    return __.<A>start().select(pop,selectKey1,selectKey2,otherSelectKeys);
  }

  public static <A, B> KubeHoundTraversal<A, Map<String, B>> select(String selectKey1,
      String selectKey2, String... otherSelectKeys) {
    return __.<A>start().select(selectKey1,selectKey2,otherSelectKeys);
  }

  public static <A, B> KubeHoundTraversal<A, B> select(Pop pop, Traversal<A, B> keyTraversal) {
    return __.<A>start().select(pop,keyTraversal);
  }

  public static <A, B> KubeHoundTraversal<A, B> select(Traversal<A, B> keyTraversal) {
    return __.<A>start().select(keyTraversal);
  }

  public static <A> KubeHoundTraversal<A, A> unfold() {
    return __.<A>start().unfold();
  }

  public static <A> KubeHoundTraversal<A, List<A>> fold() {
    return __.<A>start().fold();
  }

  public static <A, B> KubeHoundTraversal<A, B> fold(B seed, BiFunction<B, A, B> foldFunction) {
    return __.<A>start().fold(seed,foldFunction);
  }

  public static <A> KubeHoundTraversal<A, Long> count() {
    return __.<A>start().count();
  }

  public static <A> KubeHoundTraversal<A, Long> count(Scope scope) {
    return __.<A>start().count(scope);
  }

  public static <A> KubeHoundTraversal<A, Double> sum() {
    return __.<A>start().sum();
  }

  public static <A> KubeHoundTraversal<A, Double> sum(Scope scope) {
    return __.<A>start().sum(scope);
  }

  public static <A, B extends Comparable> KubeHoundTraversal<A, B> min() {
    return __.<A>start().min();
  }

  public static <A, B extends Comparable> KubeHoundTraversal<A, B> min(Scope scope) {
    return __.<A>start().min(scope);
  }

  public static <A, B extends Comparable> KubeHoundTraversal<A, B> max() {
    return __.<A>start().max();
  }

  public static <A, B extends Comparable> KubeHoundTraversal<A, B> max(Scope scope) {
    return __.<A>start().max(scope);
  }

  public static <A> KubeHoundTraversal<A, Double> mean() {
    return __.<A>start().mean();
  }

  public static <A> KubeHoundTraversal<A, Double> mean(Scope scope) {
    return __.<A>start().mean(scope);
  }

  public static <A, K, V> KubeHoundTraversal<A, Map<K, V>> group() {
    return __.<A>start().group();
  }

  public static <A, K> KubeHoundTraversal<A, Map<K, Long>> groupCount() {
    return __.<A>start().groupCount();
  }

  public static <A> KubeHoundTraversal<A, Tree> tree() {
    return __.<A>start().tree();
  }

  public static <A> KubeHoundTraversal<A, Vertex> addV(String vertexLabel) {
    return __.<A>start().addV(vertexLabel);
  }

  public static <A> KubeHoundTraversal<A, Vertex> addV(Traversal<?, String> vertexLabelTraversal) {
    return __.<A>start().addV(vertexLabelTraversal);
  }

  public static <A> KubeHoundTraversal<A, Vertex> addV() {
    return __.<A>start().addV();
  }

  public static <A> KubeHoundTraversal<A, Vertex> mergeV() {
    return __.<A>start().mergeV();
  }

  public static <A> KubeHoundTraversal<A, Vertex> mergeV(Map<Object, Object> searchCreate) {
    return __.<A>start().mergeV(searchCreate);
  }

  public static <A> KubeHoundTraversal<A, Vertex> mergeV(
      Traversal<?, Map<Object, Object>> searchCreate) {
    return __.<A>start().mergeV(searchCreate);
  }

  public static <A> KubeHoundTraversal<A, Edge> addE(String edgeLabel) {
    return __.<A>start().addE(edgeLabel);
  }

  public static <A> KubeHoundTraversal<A, Edge> addE(Traversal<?, String> edgeLabelTraversal) {
    return __.<A>start().addE(edgeLabelTraversal);
  }

  public static <A> KubeHoundTraversal<A, Edge> mergeE() {
    return __.<A>start().mergeE();
  }

  public static <A> KubeHoundTraversal<A, Edge> mergeE(Map<Object, Object> searchCreate) {
    return __.<A>start().mergeE(searchCreate);
  }

  public static <A> KubeHoundTraversal<A, Edge> mergeE(
      Traversal<?, Map<Object, Object>> searchCreate) {
    return __.<A>start().mergeE(searchCreate);
  }

  public static <A> KubeHoundTraversal<A, Double> math(String expression) {
    return __.<A>start().math(expression);
  }

  public static <A> KubeHoundTraversal<A, A> filter(Predicate<Traverser<A>> predicate) {
    return __.<A>start().filter(predicate);
  }

  public static <A> KubeHoundTraversal<A, A> filter(Traversal<?, ?> filterTraversal) {
    return __.<A>start().filter(filterTraversal);
  }

  public static <A> KubeHoundTraversal<A, A> and(Traversal<?, ?>... andTraversals) {
    return __.<A>start().and(andTraversals);
  }

  public static <A> KubeHoundTraversal<A, A> or(Traversal<?, ?>... orTraversals) {
    return __.<A>start().or(orTraversals);
  }

  public static <A> KubeHoundTraversal<A, A> inject(A... injections) {
    return __.<A>start().inject(injections);
  }

  public static <A> KubeHoundTraversal<A, A> dedup(String... dedupLabels) {
    return __.<A>start().dedup(dedupLabels);
  }

  public static <A> KubeHoundTraversal<A, A> dedup(Scope scope, String... dedupLabels) {
    return __.<A>start().dedup(scope,dedupLabels);
  }

  public static <A> KubeHoundTraversal<A, A> has(String propertyKey, P<?> predicate) {
    return __.<A>start().has(propertyKey,predicate);
  }

  public static <A> KubeHoundTraversal<A, A> has(T accessor, P<?> predicate) {
    return __.<A>start().has(accessor,predicate);
  }

  public static <A> KubeHoundTraversal<A, A> has(String propertyKey, Object value) {
    return __.<A>start().has(propertyKey,value);
  }

  public static <A> KubeHoundTraversal<A, A> has(T accessor, Object value) {
    return __.<A>start().has(accessor,value);
  }

  public static <A> KubeHoundTraversal<A, A> has(String label, String propertyKey, Object value) {
    return __.<A>start().has(label,propertyKey,value);
  }

  public static <A> KubeHoundTraversal<A, A> has(String label, String propertyKey, P<?> predicate) {
    return __.<A>start().has(label,propertyKey,predicate);
  }

  public static <A> KubeHoundTraversal<A, A> has(T accessor, Traversal<?, ?> propertyTraversal) {
    return __.<A>start().has(accessor,propertyTraversal);
  }

  public static <A> KubeHoundTraversal<A, A> has(String propertyKey,
      Traversal<?, ?> propertyTraversal) {
    return __.<A>start().has(propertyKey,propertyTraversal);
  }

  public static <A> KubeHoundTraversal<A, A> has(String propertyKey) {
    return __.<A>start().has(propertyKey);
  }

  public static <A> KubeHoundTraversal<A, A> hasNot(String propertyKey) {
    return __.<A>start().hasNot(propertyKey);
  }

  public static <A> KubeHoundTraversal<A, A> hasLabel(String label, String... otherLabels) {
    return __.<A>start().hasLabel(label,otherLabels);
  }

  public static <A> KubeHoundTraversal<A, A> hasLabel(P<String> predicate) {
    return __.<A>start().hasLabel(predicate);
  }

  public static <A> KubeHoundTraversal<A, A> hasId(Object id, Object... otherIds) {
    return __.<A>start().hasId(id,otherIds);
  }

  public static <A> KubeHoundTraversal<A, A> hasId(P<Object> predicate) {
    return __.<A>start().hasId(predicate);
  }

  public static <A> KubeHoundTraversal<A, A> hasKey(String label, String... otherLabels) {
    return __.<A>start().hasKey(label,otherLabels);
  }

  public static <A> KubeHoundTraversal<A, A> hasKey(P<String> predicate) {
    return __.<A>start().hasKey(predicate);
  }

  public static <A> KubeHoundTraversal<A, A> hasValue(Object value, Object... values) {
    return __.<A>start().hasValue(value,values);
  }

  public static <A> KubeHoundTraversal<A, A> hasValue(P<Object> predicate) {
    return __.<A>start().hasValue(predicate);
  }

  public static <A> KubeHoundTraversal<A, A> where(String startKey, P<String> predicate) {
    return __.<A>start().where(startKey,predicate);
  }

  public static <A> KubeHoundTraversal<A, A> where(P<String> predicate) {
    return __.<A>start().where(predicate);
  }

  public static <A> KubeHoundTraversal<A, A> where(Traversal<?, ?> whereTraversal) {
    return __.<A>start().where(whereTraversal);
  }

  public static <A> KubeHoundTraversal<A, A> is(P<A> predicate) {
    return __.<A>start().is(predicate);
  }

  public static <A> KubeHoundTraversal<A, A> is(Object value) {
    return __.<A>start().is(value);
  }

  public static <A> KubeHoundTraversal<A, A> not(Traversal<?, ?> notTraversal) {
    return __.<A>start().not(notTraversal);
  }

  public static <A> KubeHoundTraversal<A, A> coin(double probability) {
    return __.<A>start().coin(probability);
  }

  public static <A> KubeHoundTraversal<A, A> range(long low, long arg1) {
    return __.<A>start().range(low,arg1);
  }

  public static <A> KubeHoundTraversal<A, A> range(Scope scope, long low, long arg2) {
    return __.<A>start().range(scope,low,arg2);
  }

  public static <A> KubeHoundTraversal<A, A> limit(long limit) {
    return __.<A>start().limit(limit);
  }

  public static <A> KubeHoundTraversal<A, A> limit(Scope scope, long limit) {
    return __.<A>start().limit(scope,limit);
  }

  public static <A> KubeHoundTraversal<A, A> skip(long skip) {
    return __.<A>start().skip(skip);
  }

  public static <A> KubeHoundTraversal<A, A> skip(Scope scope, long skip) {
    return __.<A>start().skip(scope,skip);
  }

  public static <A> KubeHoundTraversal<A, A> tail() {
    return __.<A>start().tail();
  }

  public static <A> KubeHoundTraversal<A, A> tail(long limit) {
    return __.<A>start().tail(limit);
  }

  public static <A> KubeHoundTraversal<A, A> tail(Scope scope) {
    return __.<A>start().tail(scope);
  }

  public static <A> KubeHoundTraversal<A, A> tail(Scope scope, long limit) {
    return __.<A>start().tail(scope,limit);
  }

  public static <A> KubeHoundTraversal<A, A> simplePath() {
    return __.<A>start().simplePath();
  }

  public static <A> KubeHoundTraversal<A, A> cyclicPath() {
    return __.<A>start().cyclicPath();
  }

  public static <A> KubeHoundTraversal<A, A> sample(int amountToSample) {
    return __.<A>start().sample(amountToSample);
  }

  public static <A> KubeHoundTraversal<A, A> sample(Scope scope, int amountToSample) {
    return __.<A>start().sample(scope,amountToSample);
  }

  public static <A> KubeHoundTraversal<A, A> drop() {
    return __.<A>start().drop();
  }

  public static <A> KubeHoundTraversal<A, A> sideEffect(Consumer<Traverser<A>> consumer) {
    return __.<A>start().sideEffect(consumer);
  }

  public static <A> KubeHoundTraversal<A, A> sideEffect(Traversal<?, ?> sideEffectTraversal) {
    return __.<A>start().sideEffect(sideEffectTraversal);
  }

  public static <A, B> KubeHoundTraversal<A, B> cap(String sideEffectKey,
      String... sideEffectKeys) {
    return __.<A>start().cap(sideEffectKey,sideEffectKeys);
  }

  public static <A> KubeHoundTraversal<A, Edge> subgraph(String sideEffectKey) {
    return __.<A>start().subgraph(sideEffectKey);
  }

  public static <A> KubeHoundTraversal<A, A> aggregate(String sideEffectKey) {
    return __.<A>start().aggregate(sideEffectKey);
  }

  public static <A> KubeHoundTraversal<A, A> aggregate(Scope scope, String sideEffectKey) {
    return __.<A>start().aggregate(scope,sideEffectKey);
  }

  public static <A> KubeHoundTraversal<A, A> fail() {
    return __.<A>start().fail();
  }

  public static <A> KubeHoundTraversal<A, A> fail(String message) {
    return __.<A>start().fail(message);
  }

  public static <A> KubeHoundTraversal<A, A> group(String sideEffectKey) {
    return __.<A>start().group(sideEffectKey);
  }

  public static <A> KubeHoundTraversal<A, A> groupCount(String sideEffectKey) {
    return __.<A>start().groupCount(sideEffectKey);
  }

  public static <A> KubeHoundTraversal<A, A> timeLimit(long timeLimit) {
    return __.<A>start().timeLimit(timeLimit);
  }

  public static <A> KubeHoundTraversal<A, A> tree(String sideEffectKey) {
    return __.<A>start().tree(sideEffectKey);
  }

  public static <A, V, U> KubeHoundTraversal<A, A> sack(BiFunction<V, U, V> sackOperator) {
    return __.<A>start().sack(sackOperator);
  }

  public static <A> KubeHoundTraversal<A, A> store(String sideEffectKey) {
    return __.<A>start().store(sideEffectKey);
  }

  public static <A> KubeHoundTraversal<A, A> property(Object key, Object value,
      Object... keyValues) {
    return __.<A>start().property(key,value,keyValues);
  }

  public static <A> KubeHoundTraversal<A, A> property(VertexProperty.Cardinality cardinality,
      Object key, Object value, Object... keyValues) {
    return __.<A>start().property(cardinality,key,value,keyValues);
  }

  public static <A> KubeHoundTraversal<A, A> property(Map<Object, Object> value) {
    return __.<A>start().property(value);
  }

  public static <A, M, B> KubeHoundTraversal<A, B> branch(Function<Traverser<A>, M> function) {
    return __.<A>start().branch(function);
  }

  public static <A, M, B> KubeHoundTraversal<A, B> branch(Traversal<?, M> traversalFunction) {
    return __.<A>start().branch(traversalFunction);
  }

  public static <A, B> KubeHoundTraversal<A, B> choose(Predicate<A> choosePredicate,
      Traversal<?, B> trueChoice, Traversal<?, B> falseChoice) {
    return __.<A>start().choose(choosePredicate,trueChoice,falseChoice);
  }

  public static <A, B> KubeHoundTraversal<A, B> choose(Predicate<A> choosePredicate,
      Traversal<?, B> trueChoice) {
    return __.<A>start().choose(choosePredicate,trueChoice);
  }

  public static <A, M, B> KubeHoundTraversal<A, B> choose(Function<A, M> choiceFunction) {
    return __.<A>start().choose(choiceFunction);
  }

  public static <A, M, B> KubeHoundTraversal<A, B> choose(Traversal<?, M> traversalFunction) {
    return __.<A>start().choose(traversalFunction);
  }

  public static <A, M, B> KubeHoundTraversal<A, B> choose(Traversal<?, M> traversalPredicate,
      Traversal<?, B> trueChoice, Traversal<?, B> falseChoice) {
    return __.<A>start().choose(traversalPredicate,trueChoice,falseChoice);
  }

  public static <A, M, B> KubeHoundTraversal<A, B> choose(Traversal<?, M> traversalPredicate,
      Traversal<?, B> trueChoice) {
    return __.<A>start().choose(traversalPredicate,trueChoice);
  }

  public static <A> KubeHoundTraversal<A, A> optional(Traversal<?, A> optionalTraversal) {
    return __.<A>start().optional(optionalTraversal);
  }

  public static <A, B> KubeHoundTraversal<A, B> union(Traversal<?, B>... traversals) {
    return __.<A>start().union(traversals);
  }

  public static <A, B> KubeHoundTraversal<A, B> coalesce(Traversal<?, B>... traversals) {
    return __.<A>start().coalesce(traversals);
  }

  public static <A> KubeHoundTraversal<A, A> repeat(Traversal<?, A> traversal) {
    return __.<A>start().repeat(traversal);
  }

  public static <A> KubeHoundTraversal<A, A> repeat(String loopName, Traversal<?, A> traversal) {
    return __.<A>start().repeat(loopName,traversal);
  }

  public static <A> KubeHoundTraversal<A, A> emit(Traversal<?, ?> emitTraversal) {
    return __.<A>start().emit(emitTraversal);
  }

  public static <A> KubeHoundTraversal<A, A> emit(Predicate<Traverser<A>> emitPredicate) {
    return __.<A>start().emit(emitPredicate);
  }

  public static <A> KubeHoundTraversal<A, A> until(Traversal<?, ?> untilTraversal) {
    return __.<A>start().until(untilTraversal);
  }

  public static <A> KubeHoundTraversal<A, A> until(Predicate<Traverser<A>> untilPredicate) {
    return __.<A>start().until(untilPredicate);
  }

  public static <A> KubeHoundTraversal<A, A> times(int maxLoops) {
    return __.<A>start().times(maxLoops);
  }

  public static <A> KubeHoundTraversal<A, A> emit() {
    return __.<A>start().emit();
  }

  public static <A, B> KubeHoundTraversal<A, B> local(Traversal<?, B> localTraversal) {
    return __.<A>start().local(localTraversal);
  }

  public static <A> KubeHoundTraversal<A, A> as(String label, String... labels) {
    return __.<A>start().as(label,labels);
  }

  public static <A> KubeHoundTraversal<A, A> barrier() {
    return __.<A>start().barrier();
  }

  public static <A> KubeHoundTraversal<A, A> barrier(int maxBarrierSize) {
    return __.<A>start().barrier(maxBarrierSize);
  }

  public static <A> KubeHoundTraversal<A, A> barrier(
      Consumer<TraverserSet<Object>> barrierConsumer) {
    return __.<A>start().barrier(barrierConsumer);
  }

  public static <A, B> KubeHoundTraversal<A, B> index() {
    return __.<A>start().index();
  }

  public static <A, B> KubeHoundTraversal<A, Element> element() {
    return __.<A>start().element();
  }

  public static <A, B> KubeHoundTraversal<A, B> call(String service) {
    return __.<A>start().call(service);
  }

  public static <A, B> KubeHoundTraversal<A, B> call(String service, Map params) {
    return __.<A>start().call(service,params);
  }

  public static <A, B> KubeHoundTraversal<A, B> call(String service,
      Traversal<?, Map<?, ?>> childTraversal) {
    return __.<A>start().call(service,childTraversal);
  }

  public static <A, B> KubeHoundTraversal<A, B> call(String service, Map params,
      Traversal<?, Map<?, ?>> childTraversal) {
    return __.<A>start().call(service,params,childTraversal);
  }
}
