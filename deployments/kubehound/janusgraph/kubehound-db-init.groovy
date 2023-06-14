:remote connect tinkerpop.server conf/remote.yaml session
:remote console
:remote config timeout none

mgmt = graph.openManagement();

System.out.println("[KUBEHOUND] Creating graph schema and indexes");

// Create our edge labels
mgmt.makeEdgeLabel('CE_NSENTER').multiplicity(MULTI).make();
mgmt.makeEdgeLabel('CONTAINER_ATTACH').multiplicity(MULTI).make();
mgmt.makeEdgeLabel('IDENTITY_ASSUME').multiplicity(MULTI).make();
mgmt.makeEdgeLabel('TOKEN_STEAL').multiplicity(MULTI).make();
mgmt.makeEdgeLabel('TOKEN_BRUTEFORCE').multiplicity(MULTI).make();

// Create our vertex labels
mgmt.makeVertexLabel('Container').make();
mgmt.makeVertexLabel('Identity').make();
mgmt.makeVertexLabel('Node').make();
mgmt.makeVertexLabel('Pod').make();
mgmt.makeVertexLabel('Role').make();
mgmt.makeVertexLabel('Token').make();
mgmt.makeVertexLabel('Volume').make();

// Create the indexes on vertex properties
// NOTE: labels cannot be indexed so we create the class property to mirror the vertex label and allow indexing
cls = mgmt.makePropertyKey('class').dataType(String.class).cardinality(Cardinality.SINGLE).make();
mgmt.buildIndex('byClass', Vertex.class).addKey(cls).buildCompositeIndex();

storeID = mgmt.makePropertyKey('storeID').dataType(String.class).cardinality(Cardinality.SINGLE).make();
mgmt.buildIndex('byStoreIDUnique', Vertex.class).addKey(storeID).unique().buildCompositeIndex();

name = mgmt.makePropertyKey('name').dataType(String.class).cardinality(Cardinality.SINGLE).make();
mgmt.buildIndex('byName', Vertex.class).addKey(name).buildCompositeIndex();

namespace = mgmt.makePropertyKey('namespace').dataType(String.class).cardinality(Cardinality.SINGLE).make();
mgmt.buildIndex('byNamespace', Vertex.class).addKey(namespace).buildCompositeIndex();

type = mgmt.makePropertyKey('type').dataType(String.class).cardinality(Cardinality.SINGLE).make();
mgmt.buildIndex('byType', Vertex.class).addKey(type).buildCompositeIndex();

critical = mgmt.makePropertyKey('critical').dataType(Boolean.class).cardinality(Cardinality.SINGLE).make();
mgmt.buildIndex('byCritical', Vertex.class).addKey(critical).buildCompositeIndex();

mgmt.commit();

ManagementSystem.awaitGraphIndexStatus(graph, 'byClass').status(SchemaStatus.ENABLED).call();
ManagementSystem.awaitGraphIndexStatus(graph, 'byStoreIDUnique').status(SchemaStatus.ENABLED).call();
ManagementSystem.awaitGraphIndexStatus(graph, 'byName').status(SchemaStatus.ENABLED).call();
ManagementSystem.awaitGraphIndexStatus(graph, 'byNamespace').status(SchemaStatus.ENABLED).call();
ManagementSystem.awaitGraphIndexStatus(graph, 'byType').status(SchemaStatus.ENABLED).call();
ManagementSystem.awaitGraphIndexStatus(graph, 'byCritical').status(SchemaStatus.ENABLED).call();

System.out.println("[KUBEHOUND] graph schema and indexes ready");

:remote close