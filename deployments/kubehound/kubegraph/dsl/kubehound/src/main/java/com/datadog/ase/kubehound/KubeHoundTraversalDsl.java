/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */
package com.datadog.ase.kubehound;

import org.apache.tinkerpop.gremlin.process.traversal.dsl.GremlinDsl;
import org.apache.tinkerpop.gremlin.process.traversal.dsl.GremlinDsl.AnonymousMethod;
import org.apache.tinkerpop.gremlin.process.traversal.Traversal;
import org.apache.tinkerpop.gremlin.process.traversal.dsl.graph.GraphTraversal;
import org.apache.tinkerpop.gremlin.structure.Vertex;
import org.apache.tinkerpop.gremlin.structure.Edge;
import org.apache.tinkerpop.gremlin.process.traversal.P;
import org.apache.tinkerpop.gremlin.process.traversal.Path;

/**
 * This KubeHound DSL is meant to be used with the Kubernetes attack graph created by the KubeHound application.
 * <p/>
 * All DSLs should extend {@code GraphTraversal.Admin} and be suffixed with "TraversalDsl". Simply add DSL traversal
 * methods to this interface. Use Gremlin's steps to build the underlying traversal in these methods to ensure
 * compatibility with the rest of the TinkerPop stack and provider implementations.
 * <p/>
 * Arguments provided to the {@code GremlinDsl} annotation are all optional. In this case, a {@code traversalSource} is
 * specified which points to a specific implementation to use. Had that argument not been specified then a default
 * {@code TraversalSource} would have been generated.
 */
@GremlinDsl(traversalSource = "com.datadog.ase.kubehound.KubeHoundTraversalSourceDsl")
public interface KubeHoundTraversalDsl<S, E> extends GraphTraversal.Admin<S, E> {
    public static final int PATH_HOPS_MIN = 1;
    public static final int PATH_HOPS_MAX = 15;
    public static final int PATH_HOPS_DEFAULT = 10;

    /**
     * From a {@code Vertex} traverse "knows" edges to adjacent "person" vertices and filter those vertices on the
     * "name" property.
     *
     * @param personName the name of the person to filter on
     */
    public default GraphTraversal<S, Path> attacks() {
        return outE().inV().path();
    }

    @GremlinDsl.AnonymousMethod(returnTypeParameters = {"A", "A"}, methodTypeParameters = {"A"})
    public default GraphTraversal<S, E> critical() {
        return has("critical", true);
    }

    public default GraphTraversal<S, Path> criticalPaths(int maxHops) {
        if (maxHops < PATH_HOPS_MIN) throw new IllegalArgumentException(String.format("maxHops must be >= %d", PATH_HOPS_MIN));
        if (maxHops > PATH_HOPS_MAX) throw new IllegalArgumentException(String.format("maxHops must be <= %d", PATH_HOPS_MAX));

        return repeat((
                (KubeHoundTraversalDsl) __.out())
                .simplePath()
            ).until(
                __.has("critical", true)
                .or()
                .loops()
                .is(maxHops)
            ).has("critical", true)
            .path();
    }

    public default GraphTraversal<S, Path> criticalPaths() {
        return repeat((
                (KubeHoundTraversalDsl) __.out())
                .simplePath()
            ).until(
                __.has("critical", true)
                .or()
                .loops()
                .is(PATH_HOPS_DEFAULT)
            ).has("critical", true)
            .path();
    }

    public default GraphTraversal<S, Path> criticalPathsFilter(String[] exclusions, int maxHops) {
        if (exclusions.length <= 0) throw new IllegalArgumentException("exclusions must be provided (otherwise use criticalPaths())");
        if (maxHops < PATH_HOPS_MIN) throw new IllegalArgumentException(String.format("maxHops must be >= %d", PATH_HOPS_MIN));
        if (maxHops > PATH_HOPS_MAX) throw new IllegalArgumentException(String.format("maxHops must be <= %d", PATH_HOPS_MAX));

         return repeat((
                (KubeHoundTraversalDsl) __.outE())
                .hasLabel(P.not(P.within(exclusions)))
                .inV()
                .simplePath()
            ).until(
                __.has("critical", true)
                .or()
                .loops()
                .is(maxHops)
            ).has("critical", true)
            .path();
    }

    @GremlinDsl.AnonymousMethod(returnTypeParameters = {"A", "A"}, methodTypeParameters = {"A"})
    public default GraphTraversal<S, E> hasCriticalPath() {
        return where(criticalPaths().limit(1)); 
    }

    // public default <E2 extends Number> GraphTraversal<S, E2> shortestCriticalPath() {
    //      return repeat((
    //             (KubeHoundTraversalDsl) __.out())
    //             .simplePath()
    //         ).until(
    //             __.has("critical", true)
    //             .or()
    //             .loops()
    //             .is(PATH_HOPS_DEFAULT)
    //         ).has("critical", true)
    //         .path()
    //         .count(local)
    //         .min();
    // }

    // public default GraphTraversal<S, E> pathTo() {
    //     TODO g.V().hasLabel("Container", "Identity", "Node").repeat(out().simplePath()).until(has("name", "system:auth-delegator").or().loops().is(5)).has("name", "system:auth-delegator").hasLabel("Role").path()
    // }

    // public default GraphTraversal<Vertex, Map<String, Long>> attackPathMap() {
    //      return repeat((
    //             (KubeHoundTraversalDsl) __.outE())
    //             .inV()
    //             .simplePath()
    //         ).emit()
    //         .until(
    //             __.has("critical", true)
    //             .or()
    //             .loops()
    //             .is(PATH_HOPS_DEFAULT)
    //         ).has("critical", true)
    //         .path()
    //         .by(label())
    //         .groupCount();
    // }
}
