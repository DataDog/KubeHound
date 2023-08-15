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

import org.apache.tinkerpop.gremlin.process.remote.RemoteConnection;
import org.apache.tinkerpop.gremlin.process.traversal.P;
import org.apache.tinkerpop.gremlin.process.traversal.TraversalStrategies;
import org.apache.tinkerpop.gremlin.process.traversal.dsl.GremlinDsl;
import org.apache.tinkerpop.gremlin.process.traversal.dsl.graph.DefaultGraphTraversal;
import org.apache.tinkerpop.gremlin.process.traversal.dsl.graph.GraphTraversal;
import org.apache.tinkerpop.gremlin.process.traversal.dsl.graph.GraphTraversalSource;
import org.apache.tinkerpop.gremlin.process.traversal.step.map.GraphStep;
import org.apache.tinkerpop.gremlin.process.traversal.step.util.HasContainer;
import org.apache.tinkerpop.gremlin.process.traversal.util.TraversalHelper;
import org.apache.tinkerpop.gremlin.structure.Graph;
import org.apache.tinkerpop.gremlin.structure.T;
import org.apache.tinkerpop.gremlin.structure.Vertex;

/**
 * See {@code KubeHoundTraversalDsl} for more information about this DSL.
 */
public class KubeHoundTraversalSourceDsl extends GraphTraversalSource {

    public KubeHoundTraversalSourceDsl(final Graph graph, final TraversalStrategies traversalStrategies) {
        super(graph, traversalStrategies);
    }

    public KubeHoundTraversalSourceDsl(final Graph graph) {
        super(graph);
    }

    public KubeHoundTraversalSourceDsl(final RemoteConnection connection) {
        super(connection);
    }

    // /**
    //  * Starts a traversal that finds all vertices with a "person" label and optionally allows filtering of those
    //  * vertices on the "name" property.
    //  *
    //  * @param names list of person names to filter on
    //  */
    // public GraphTraversal<Vertex, Vertex> persons(String... names) {
    //     GraphTraversalSource clone = this.clone();

    //     // Manually add a "start" step for the traversal in this case the equivalent of V(). GraphStep is marked
    //     // as a "start" step by passing "true" in the constructor.
    //     clone.getBytecode().addStep(GraphTraversal.Symbols.V);
    //     GraphTraversal<Vertex, Vertex> traversal = new DefaultGraphTraversal<>(clone);
    //     traversal.asAdmin().addStep(new GraphStep<>(traversal.asAdmin(), Vertex.class, true));

    //     traversal = traversal.hasLabel("person");
    //     if (names.length > 0) traversal = traversal.has("name", P.within(names));

    //     return traversal;
    // }

    // private GraphTraversal<Vertex, Vertex> startTraversal() {
    //     GraphTraversalSource clone = this.clone();

    //     // Manually add a "start" step for the traversal in this case the equivalent of V(). GraphStep is marked
    //     // as a "start" step by passing "true" in the constructor.
    //     clone.getBytecode().addStep(GraphTraversal.Symbols.V);
    //     GraphTraversal<Vertex, Vertex> traversal = new DefaultGraphTraversal<>(clone);
    //     traversal.asAdmin().addStep(new GraphStep<>(traversal.asAdmin(), Vertex.class, true));

    //     return traversal;
    // }

    public GraphTraversal<Vertex, Vertex> containers(String... names) {
        GraphTraversal traversal = this.clone().V();

        traversal = traversal.hasLabel("Container");
        if (names.length > 0) {
            traversal = traversal.has("name", P.within(names));
        } 

        return traversal;
    }

     public GraphTraversal<Vertex, Vertex> pods(String... names) {
        GraphTraversal traversal = this.clone().V();

        traversal = traversal.hasLabel("Pod");
        if (names.length > 0) {
            traversal = traversal.has("name", P.within(names));
        } 

        return traversal;
    }

     public GraphTraversal<Vertex, Vertex> nodes(String... names) {
        GraphTraversal traversal = this.clone().V();

        traversal = traversal.hasLabel("Node");
        if (names.length > 0) {
            traversal = traversal.has("name", P.within(names));
        } 

        return traversal;
    }

    public GraphTraversal<Vertex, Vertex> endpoints(EndpointExposure exposure) {
        GraphTraversal traversal = this.clone().V();
        
        traversal = traversal
            .hasLabel("Endpoint")
            .has("exposure", P.gte(exposure.ordinal()));

        return traversal;
    }

    public GraphTraversal<Vertex, Vertex> services(String... names) {
        GraphTraversal traversal = this.clone().V();
        
        traversal = traversal
            .hasLabel("Endpoint")
            .has("exposure", P.gte(EndpointExposure.External.ordinal()));

        if (names.length > 0) {
            traversal = traversal.has("name", P.within(names));
        } 

        return traversal;
    }


    public GraphTraversal<Vertex, Vertex> volumes() {
        GraphTraversal traversal = this.clone().V();
        
        traversal = traversal.hasLabel("Volume");

        return traversal;
    }

    public GraphTraversal<Vertex, Vertex> hostMounts(String... sourcePaths) {
        GraphTraversal traversal = this.clone().V();
        
        traversal = traversal
            .hasLabel("Volume")
            .has("type", "HostPath");

        if (sourcePaths.length > 0) {
            traversal = traversal.has("sourcePath", P.within(sourcePaths));
        } 

        return traversal;
    }

    public GraphTraversal<Vertex, Vertex> identities(String... names) {
        GraphTraversal traversal = this.clone().V();
        
        traversal = traversal.hasLabel("Identity");
        if (names.length > 0) {
            traversal = traversal.has("name", P.within(names));
        } 

        return traversal;
    }

    public GraphTraversal<Vertex, Vertex> sas(String... names) {
        GraphTraversal traversal = this.clone().V();
        
        traversal = traversal
            .hasLabel("Identity")
            .has("type", "ServiceAccount");

        if (names.length > 0) {
            traversal = traversal.has("name", P.within(names));
        } 

        return traversal;
    }

    public GraphTraversal<Vertex, Vertex> users(String... names) {
        GraphTraversal traversal = this.clone().V();
        
        traversal = traversal
            .hasLabel("Identity")
            .has("type", "User");

        if (names.length > 0) {
            traversal = traversal.has("name", P.within(names));
        } 

        return traversal;
    }

    public GraphTraversal<Vertex, Vertex> groups(String... names) {
        GraphTraversal traversal = this.clone().V();
         
        traversal = traversal
            .hasLabel("Identity")
            .has("type", "Group");

        if (names.length > 0) {
            traversal = traversal.has("name", P.within(names));
        } 

        return traversal;
    }

    public GraphTraversal<Vertex, Vertex> permissions(String... roles) {
        GraphTraversal traversal = this.clone().V();
        
        traversal = traversal.hasLabel("PermissionSet");
        if (roles.length > 0) {
            traversal = traversal.has("name", P.within(roles));
        } 

        return traversal;
    }
}
