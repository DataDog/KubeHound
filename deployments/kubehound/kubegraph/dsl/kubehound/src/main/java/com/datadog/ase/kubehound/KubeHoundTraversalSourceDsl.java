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
import org.apache.tinkerpop.gremlin.process.traversal.Path;

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

    /**
     * Starts a traversal that finds all vertices with a "Container" label and optionally allows filtering of those
     * vertices on the "name" property.
     *
     * @param names list of container names to filter on
     */
    public GraphTraversal<Vertex, Vertex> containers(String... names) {
        GraphTraversal traversal = this.clone().V();

        traversal = traversal.hasLabel("Container");
        if (names.length > 0) {
            traversal = traversal.has("name", P.within(names));
        } 

        return traversal;
    }

    /**
     * Starts a traversal that finds all vertices with a "Pod" label and optionally allows filtering of those
     * vertices on the "name" property.
     *
     * @param names list of pod names to filter on
     */
    public GraphTraversal<Vertex, Vertex> pods(String... names) {
        GraphTraversal traversal = this.clone().V();

        traversal = traversal.hasLabel("Pod");
        if (names.length > 0) {
            traversal = traversal.has("name", P.within(names));
        } 

        return traversal;
    }

    /**
     * Starts a traversal that finds all vertices with a "Node" label and optionally allows filtering of those
     * vertices on the "name" property.
     *
     * @param names list of node names to filter on
     */
    public GraphTraversal<Vertex, Vertex> nodes(String... names) {
        GraphTraversal traversal = this.clone().V();

        traversal = traversal.hasLabel("Node");
        if (names.length > 0) {
            traversal = traversal.has("name", P.within(names));
        } 

        return traversal;
    }

    /**
     * Starts a traversal that finds all container escape edges from a Container vertex to a Node vertex 
     * and optionally allows filtering of those vertices on the "nodeNames" property.
     *
     * @param nodeNames list of node names to filter on

     */
    public GraphTraversal<Vertex, Path> escapes(String... nodeNames) {
        GraphTraversal traversal = this.clone().V();

        traversal = traversal
            .hasLabel("Container")
            .outE()
            .inV()
            .hasLabel("Node");

        if (nodeNames.length > 0) {
            traversal = traversal.has("name", P.within(nodeNames));
        } 

        return traversal.path();
    }

    /**
     * Starts a traversal that finds all vertices with a "Endpoint" label.
     */
    public GraphTraversal<Vertex, Vertex> endpoints() {
        GraphTraversal traversal = this.clone().V();
    
        traversal = traversal.hasLabel("Endpoint");

        return traversal;
    }

    /**
     * Starts a traversal that finds all vertices with a "Endpoint" label and optionally allows filtering of those
     * vertices on the "exposure" property.
     *
     * @param exposure EndpointExposure enum value to filter on
     */
    public GraphTraversal<Vertex, Vertex> endpoints(EndpointExposure exposure) {
        if (exposure.ordinal() > EndpointExposure.Max.ordinal()) {
            throw new IllegalArgumentException(String.format("invalid exposure value (must be <= %d)", EndpointExposure.Max.ordinal()));
        }

        if (exposure.ordinal() < EndpointExposure.None.ordinal()) {
            throw new IllegalArgumentException(String.format("invalid exposure value (must be >= %d)", EndpointExposure.None.ordinal()));
        }

        GraphTraversal traversal = this.clone().V();
        
        traversal = traversal
            .hasLabel("Endpoint")
            .has("exposure", P.gte(exposure.ordinal()));

        return traversal;
    }

    /**
     * Starts a traversal that finds all vertices with a "Endpoint" label exposed OUTSIDE the cluster as a service 
     * and optionally allows filtering of those vertices on the "portName" property.
     *
     * @param portNames list of port names to filter on
     */
    public GraphTraversal<Vertex, Vertex> services(String... portNames) {
        GraphTraversal traversal = this.clone().V();
        
        traversal = traversal
            .hasLabel("Endpoint")
            .has("exposure", P.gte(EndpointExposure.External.ordinal()));

        if (portNames.length > 0) {
            traversal = traversal.has("portName", P.within(portNames));
        } 

        return traversal;
    }

    /**
     * Starts a traversal that finds all vertices with a "Volume" label and optionally allows filtering of those
     * vertices on the "name" property.
     *
     * @param names list of volume names to filter on
     */
    public GraphTraversal<Vertex, Vertex> volumes(String... names) {
        GraphTraversal traversal = this.clone().V();
        
        traversal = traversal.hasLabel("Volume");
        if (names.length > 0) {
            traversal = traversal.has("name", P.within(names));
        } 

        return traversal;
    }

    /**
     * Starts a traversal that finds all vertices representing volume host mounts and optionally allows filtering of those
     * vertices on the "sourcePath" property.
     *
     * @param sourcePaths list of host source paths to filter on
     */
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

    /**
     * Starts a traversal that finds all vertices with a "Identity" label and optionally allows filtering of those
     * vertices on the "name" property.
     *
     * @param names list of identity names to filter on
     */
    public GraphTraversal<Vertex, Vertex> identities(String... names) {
        GraphTraversal traversal = this.clone().V();
        
        traversal = traversal.hasLabel("Identity");
        if (names.length > 0) {
            traversal = traversal.has("name", P.within(names));
        } 

        return traversal;
    }

    /**
     * Starts a traversal that finds all vertices representing service accounts and optionally allows filtering of those
     * vertices on the "name" property.
     *
     * @param names list of service account names to filter on
     */
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

    /**
     * Starts a traversal that finds all vertices representing users and optionally allows filtering of those
     * vertices on the "name" property.
     *
     * @param names list of user names to filter on
     */
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

    /**
     * Starts a traversal that finds all vertices representing groups and optionally allows filtering of those
     * vertices on the "name" property.
     *
     * @param names list of groups names to filter on
     */
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

    /**
     * Starts a traversal that finds all vertices with a "PermissionSet" label and optionally allows filtering of those
     * vertices on the "role" property.
     *
     * @param roles list of underlying role names to filter on
     */
    public GraphTraversal<Vertex, Vertex> permissions(String... roles) {
        GraphTraversal traversal = this.clone().V();
        
        traversal = traversal.hasLabel("PermissionSet");
        if (roles.length > 0) {
            traversal = traversal.has("role", P.within(roles));
        } 

        return traversal;
    }
}
