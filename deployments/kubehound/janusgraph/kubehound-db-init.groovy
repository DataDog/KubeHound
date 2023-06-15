:remote connect tinkerpop.server conf/remote.yaml session
:remote console
:remote config timeout none

//
// Graph schema and index definition for the KubeHound graph mode
// See details of the janus graph APIs here https://docs.janusgraph.org/schema/
// See details of the underlying graph model here: @@DOCLINK: https://datadoghq.atlassian.net/wiki/spaces/ASE/pages/2871886994/Kube+Graph+Model
//

graph.tx().rollback()
mgmt = graph.openManagement();

System.out.println("[KUBEHOUND] Creating graph schema and indexes");

// Create our vertex labels
container = mgmt.makeVertexLabel('Container').make();
identity = mgmt.makeVertexLabel('Identity').make();
node =mgmt.makeVertexLabel('Node').make();
pod = mgmt.makeVertexLabel('Pod').make();
role = mgmt.makeVertexLabel('Role').make();
token = mgmt.makeVertexLabel('Token').make();
volume = mgmt.makeVertexLabel('Volume').make();

// Create our edge labels and connections
roleGrant = mgmt.makeEdgeLabel('ROLE_GRANT').multiplicity(ONE2MANY).make();
mgmt.addConnection(roleGrant, identity, role);

volmeMount = mgmt.makeEdgeLabel('VOLUME_MOUNT').multiplicity(MANY2ONE).make();
mgmt.addConnection(volmeMount, container, volume);
mgmt.addConnection(volmeMount, node, volume);

sharedPs = mgmt.makeEdgeLabel('SHARED_PS_NAMESPACE').multiplicity(MULTI).make();
mgmt.addConnection(sharedPs, container, container);

containerAttach = mgmt.makeEdgeLabel('CONTAINER_ATTACH').multiplicity(ONE2MANY).make();
mgmt.addConnection(containerAttach, pod, container);

idAssume = mgmt.makeEdgeLabel('IDENTITY_ASSUME').multiplicity(MANY2ONE).make();
mgmt.addConnection(idAssume, pod, identity);
mgmt.addConnection(idAssume, token, identity);

idImpersonate = mgmt.makeEdgeLabel('IDENTITY_IMPERSONATE').multiplicity(MANY2ONE).make();
mgmt.addConnection(idImpersonate, role, identity);

roleBind = mgmt.makeEdgeLabel('ROLE_BIND').multiplicity(MANY2ONE).make();
mgmt.addConnection(roleBind, role, role);

podAttach = mgmt.makeEdgeLabel('POD_ATTACH').multiplicity(ONE2MANY).make();
mgmt.addConnection(podAttach, node, container);

podCreate = mgmt.makeEdgeLabel('POD_CREATE').multiplicity(ONE2MANY).make();
mgmt.addConnection(podCreate, role, pod);

tokenSteal = mgmt.makeEdgeLabel('TOKEN_STEAL').multiplicity(ONE2MANY).make();
mgmt.addConnection(tokenSteal, volume, token);

tokenBruteforce = mgmt.makeEdgeLabel('TOKEN_BRUTEFORCE').multiplicity(ONE2MANY).make();
mgmt.addConnection(tokenBruteforce, role, identity);

tokenList = mgmt.makeEdgeLabel('TOKEN_LIST').multiplicity(ONE2MANY).make();
mgmt.addConnection(tokenBruteforce, role, identity);

tokenVarLog = mgmt.makeEdgeLabel('TOKEN_VAR_LOG_SYMLINK').multiplicity(ONE2MANY).make();
mgmt.addConnection(tokenVarLog, container, token);

nsenter = mgmt.makeEdgeLabel('CE_NSENTER').multiplicity(MANY2ONE).make();
mgmt.addConnection(nsenter, container, node);

moduleLoad = mgmt.makeEdgeLabel('CE_MODULE_LOAD').multiplicity(MANY2ONE).make();
mgmt.addConnection(moduleLoad, container, node);

umhCorePattern = mgmt.makeEdgeLabel('CE_UMH_CORE_PATTERN').multiplicity(MANY2ONE).make();
mgmt.addConnection(umhCorePattern, container, node);

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

// Specify remaining properties that will NOT be indexed
mgmt.makePropertyKey('isNamespaced').dataType(Boolean.class).cardinality(Cardinality.SINGLE).make();
mgmt.makePropertyKey('compromised').dataType(Long.class).cardinality(Cardinality.SINGLE).make();
mgmt.makePropertyKey('path').dataType(String.class).cardinality(Cardinality.SINGLE).make();
mgmt.makePropertyKey('node').dataType(String.class).cardinality(Cardinality.SINGLE).make();
mgmt.makePropertyKey('sharedProcessNamespace').dataType(Boolean.class).cardinality(Cardinality.SINGLE).make();
mgmt.makePropertyKey('serviceAccount').dataType(String.class).cardinality(Cardinality.SINGLE).make();
mgmt.makePropertyKey('image').dataType(String.class).cardinality(Cardinality.SINGLE).make();
mgmt.makePropertyKey('pod').dataType(String.class).cardinality(Cardinality.SINGLE).make();
mgmt.makePropertyKey('hostNetwork').dataType(Boolean.class).cardinality(Cardinality.SINGLE).make();
mgmt.makePropertyKey('hostPath').dataType(Boolean.class).cardinality(Cardinality.SINGLE).make();
mgmt.makePropertyKey('hostPid').dataType(Boolean.class).cardinality(Cardinality.SINGLE).make();
mgmt.makePropertyKey('privesc').dataType(Boolean.class).cardinality(Cardinality.SINGLE).make();
mgmt.makePropertyKey('privileged').dataType(Boolean.class).cardinality(Cardinality.SINGLE).make();
mgmt.makePropertyKey('runAsUser').dataType(Long.class).cardinality(Cardinality.SINGLE).make();
mgmt.makePropertyKey('rules').dataType(String.class).cardinality(Cardinality.SET).make();
mgmt.makePropertyKey('command').dataType(String.class).cardinality(Cardinality.SET).make();
mgmt.makePropertyKey('args').dataType(String.class).cardinality(Cardinality.SET).make();
mgmt.makePropertyKey('capabilities').dataType(String.class).cardinality(Cardinality.SET).make();
mgmt.makePropertyKey('ports').dataType(Long.class).cardinality(Cardinality.SET).make();

mgmt.commit();

// Wait for indexes to become available
ManagementSystem.awaitGraphIndexStatus(graph, 'byClass').status(SchemaStatus.ENABLED).call();
ManagementSystem.awaitGraphIndexStatus(graph, 'byStoreIDUnique').status(SchemaStatus.ENABLED).call();
ManagementSystem.awaitGraphIndexStatus(graph, 'byName').status(SchemaStatus.ENABLED).call();
ManagementSystem.awaitGraphIndexStatus(graph, 'byNamespace').status(SchemaStatus.ENABLED).call();
ManagementSystem.awaitGraphIndexStatus(graph, 'byType').status(SchemaStatus.ENABLED).call();
ManagementSystem.awaitGraphIndexStatus(graph, 'byCritical').status(SchemaStatus.ENABLED).call();

System.out.println("[KUBEHOUND] graph schema and indexes ready");

// Close the open connection
:remote close