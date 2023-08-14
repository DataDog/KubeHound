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
package ${package};

import org.apache.tinkerpop.gremlin.process.traversal.dsl.GremlinDsl;
import org.apache.tinkerpop.gremlin.process.traversal.dsl.GremlinDsl.AnonymousMethod;
import org.apache.tinkerpop.gremlin.process.traversal.Traversal;
import org.apache.tinkerpop.gremlin.process.traversal.dsl.graph.GraphTraversal;
import org.apache.tinkerpop.gremlin.structure.Vertex;
import org.apache.tinkerpop.gremlin.process.traversal.P;

/**
 * This Social DSL is meant to be used with the TinkerPop "modern" toy graph.
 * <p/>
 * All DSLs should extend {@code GraphTraversal.Admin} and be suffixed with "TraversalDsl". Simply add DSL traversal
 * methods to this interface. Use Gremlin's steps to build the underlying traversal in these methods to ensure
 * compatibility with the rest of the TinkerPop stack and provider implementations.
 * <p/>
 * Arguments provided to the {@code GremlinDsl} annotation are all optional. In this case, a {@code traversalSource} is
 * specified which points to a specific implementation to use. Had that argument not been specified then a default
 * {@code TraversalSource} would have been generated.
 */
@GremlinDsl(traversalSource = "${package}.SocialTraversalSourceDsl")
public interface SocialTraversalDsl<S, E> extends GraphTraversal.Admin<S, E> {

    /**
     * From a {@code Vertex} traverse "knows" edges to adjacent "person" vertices and filter those vertices on the
     * "name" property.
     *
     * @param personName the name of the person to filter on
     */
    public default GraphTraversal<S, Vertex> knows(String personName) {
        return ((SocialTraversalDsl) out("knows")).person().has("name", personName);
    }

    /**
     * From a {@code Vertex} traverse "knows" edges to adjacent "person" vertices and determine the youngest age of
     * those persons.
     */
    public default <E2 extends Number> GraphTraversal<S, E2> youngestFriendsAge() {
        return ((SocialTraversalDsl) out("knows")).person().values("age").min();
    }

    /**
     * Designed to be used as a filter for "person" vertices based on the number of "created" edges encountered.
     *
     * @param number the minimum number of projects a person created
     */
    public default GraphTraversal<S, Long> createdAtLeast(int number) {
        return outE("created").count().is(P.gte(number));
    }

    /**
     * Filters objects by the "person" label. This step is designed to work with incoming vertices.
     */
    @GremlinDsl.AnonymousMethod(returnTypeParameters = {"A", "A"}, methodTypeParameters = {"A"})
    public default GraphTraversal<S, E> person() {
        return hasLabel("person");
    }
}
