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
node = mgmt.makeVertexLabel('Node').make();
pod = mgmt.makeVertexLabel('Pod').make();
role = mgmt.makeVertexLabel('Role').make();
volume = mgmt.makeVertexLabel('Volume').make();

// Create our edge labels and connections
roleGrant = mgmt.makeEdgeLabel('ROLE_GRANT').multiplicity(MULTI).make();
mgmt.addConnection(roleGrant, identity, role);

volmeMount = mgmt.makeEdgeLabel('VOLUME_MOUNT').multiplicity(MULTI).make();
mgmt.addConnection(volmeMount, container, volume);
mgmt.addConnection(volmeMount, node, volume);

sharedPs = mgmt.makeEdgeLabel('SHARED_PS_NAMESPACE').multiplicity(MULTI).make();
mgmt.addConnection(sharedPs, container, container);

containerAttach = mgmt.makeEdgeLabel('CONTAINER_ATTACH').multiplicity(ONE2MANY).make();
mgmt.addConnection(containerAttach, pod, container);

idAssume = mgmt.makeEdgeLabel('IDENTITY_ASSUME').multiplicity(MANY2ONE).make();
mgmt.addConnection(idAssume, container, identity);

idImpersonate = mgmt.makeEdgeLabel('IDENTITY_IMPERSONATE').multiplicity(MANY2ONE).make();
mgmt.addConnection(idImpersonate, role, identity);

roleBind = mgmt.makeEdgeLabel('ROLE_BIND').multiplicity(MANY2ONE).make();
mgmt.addConnection(roleBind, role, role);

podAttach = mgmt.makeEdgeLabel('POD_ATTACH').multiplicity(ONE2MANY).make();
mgmt.addConnection(podAttach, node, pod);

podCreate = mgmt.makeEdgeLabel('POD_CREATE').multiplicity(MULTI).make();
mgmt.addConnection(podCreate, role, node);
mgmt.addConnection(podCreate, role, role); // self-referencing for large cluster optimizations

podPatch = mgmt.makeEdgeLabel('POD_PATCH').multiplicity(MULTI).make();
mgmt.addConnection(podPatch, role, node);
mgmt.addConnection(podPatch, role, role); // self-referencing for large cluster optimizations

podExec = mgmt.makeEdgeLabel('POD_EXEC').multiplicity(MULTI).make();
mgmt.addConnection(podExec, role, pod);
mgmt.addConnection(podExec, role, role); // self-referencing for large cluster optimizations

tokenSteal = mgmt.makeEdgeLabel('TOKEN_STEAL').multiplicity(MULTI).make();
mgmt.addConnection(tokenSteal, volume, identity);

tokenBruteforce = mgmt.makeEdgeLabel('TOKEN_BRUTEFORCE').multiplicity(MULTI).make();
mgmt.addConnection(tokenBruteforce, role, identity);

tokenList = mgmt.makeEdgeLabel('TOKEN_LIST').multiplicity(MULTI).make();
mgmt.addConnection(tokenList, role, identity);

tokenVarLog = mgmt.makeEdgeLabel('TOKEN_VAR_LOG_SYMLINK').multiplicity(ONE2MANY).make();
mgmt.addConnection(tokenVarLog, container, volume);

nsenter = mgmt.makeEdgeLabel('CE_NSENTER').multiplicity(MANY2ONE).make();
mgmt.addConnection(nsenter, container, node);

moduleLoad = mgmt.makeEdgeLabel('CE_MODULE_LOAD').multiplicity(MANY2ONE).make();
mgmt.addConnection(moduleLoad, container, node);

umhCorePattern = mgmt.makeEdgeLabel('CE_UMH_CORE_PATTERN').multiplicity(MANY2ONE).make();
mgmt.addConnection(umhCorePattern, container, node);

privMount = mgmt.makeEdgeLabel('CE_PRIV_MOUNT').multiplicity(MANY2ONE).make();
mgmt.addConnection(privMount, container, node);

sysPtrace = mgmt.makeEdgeLabel('CE_SYS_PTRACE').multiplicity(MANY2ONE).make();
mgmt.addConnection(sysPtrace, container, node);


// All properties we will index on
cls = mgmt.makePropertyKey('class').dataType(String.class).cardinality(Cardinality.SINGLE).make();
storeID = mgmt.makePropertyKey('storeID').dataType(String.class).cardinality(Cardinality.SINGLE).make();
app = mgmt.makePropertyKey('app').dataType(String.class).cardinality(Cardinality.SINGLE).make();
team = mgmt.makePropertyKey('team').dataType(String.class).cardinality(Cardinality.SINGLE).make();
service = mgmt.makePropertyKey('service').dataType(String.class).cardinality(Cardinality.SINGLE).make();
name = mgmt.makePropertyKey('name').dataType(String.class).cardinality(Cardinality.SINGLE).make();
namespace = mgmt.makePropertyKey('namespace').dataType(String.class).cardinality(Cardinality.SINGLE).make();
type = mgmt.makePropertyKey('type').dataType(String.class).cardinality(Cardinality.SINGLE).make();
critical = mgmt.makePropertyKey('critical').dataType(Boolean.class).cardinality(Cardinality.SINGLE).make();

// All properties that we want to be able to search on
isNamespaced = mgmt.makePropertyKey('isNamespaced').dataType(Boolean.class).cardinality(Cardinality.SINGLE).make();
compromised = mgmt.makePropertyKey('compromised').dataType(Integer.class).cardinality(Cardinality.SINGLE).make();
path = mgmt.makePropertyKey('path').dataType(String.class).cardinality(Cardinality.SINGLE).make();
nodeName = mgmt.makePropertyKey('node').dataType(String.class).cardinality(Cardinality.SINGLE).make();
sharedPs = mgmt.makePropertyKey('sharedProcessNamespace').dataType(Boolean.class).cardinality(Cardinality.SINGLE).make();
serviceAccount = mgmt.makePropertyKey('serviceAccount').dataType(String.class).cardinality(Cardinality.SINGLE).make();
image = mgmt.makePropertyKey('image').dataType(String.class).cardinality(Cardinality.SINGLE).make();
podName = mgmt.makePropertyKey('pod').dataType(String.class).cardinality(Cardinality.SINGLE).make();
hostNetwork = mgmt.makePropertyKey('hostNetwork').dataType(Boolean.class).cardinality(Cardinality.SINGLE).make();
hostPath = mgmt.makePropertyKey('hostPath').dataType(Boolean.class).cardinality(Cardinality.SINGLE).make();
hostPid = mgmt.makePropertyKey('hostPid').dataType(Boolean.class).cardinality(Cardinality.SINGLE).make();
hostIpc = mgmt.makePropertyKey('hostIpc').dataType(Boolean.class).cardinality(Cardinality.SINGLE).make();
privesc = mgmt.makePropertyKey('privesc').dataType(Boolean.class).cardinality(Cardinality.SINGLE).make();
privileged = mgmt.makePropertyKey('privileged').dataType(Boolean.class).cardinality(Cardinality.SINGLE).make();
runAsUser = mgmt.makePropertyKey('runAsUser').dataType(Long.class).cardinality(Cardinality.SINGLE).make();
rules = mgmt.makePropertyKey('rules').dataType(String.class).cardinality(Cardinality.LIST).make();
command = mgmt.makePropertyKey('command').dataType(String.class).cardinality(Cardinality.LIST).make();
args = mgmt.makePropertyKey('args').dataType(String.class).cardinality(Cardinality.LIST).make();
capabilities = mgmt.makePropertyKey('capabilities').dataType(String.class).cardinality(Cardinality.LIST).make();
ports = mgmt.makePropertyKey('ports').dataType(String.class).cardinality(Cardinality.LIST).make();
identityName = mgmt.makePropertyKey('identity').dataType(String.class).cardinality(Cardinality.SINGLE).make();

// Define properties for each vertex 
mgmt.addProperties(container, cls, storeID, app, team, service, isNamespaced, namespace, name, image, privileged, privesc, hostPid, hostPath, 
    hostIpc, hostNetwork, runAsUser, podName, nodeName, compromised, command, args, capabilities, ports);
mgmt.addProperties(identity, cls, storeID, app, team, service, name, isNamespaced, namespace, type, critical);
mgmt.addProperties(node, cls, storeID, app, team, service, name, isNamespaced, namespace, compromised, critical);
mgmt.addProperties(pod, cls, storeID, app, team, service, name, isNamespaced, namespace, sharedPs, serviceAccount, nodeName, compromised, critical);
mgmt.addProperties(role, cls, storeID, app, team, service, name, isNamespaced, namespace, rules, critical);
mgmt.addProperties(volume, cls, storeID, app, team, service, name, isNamespaced, namespace, type, path);


// Create the indexes on vertex properties
// NOTE: labels cannot be indexed so we create the class property to mirror the vertex label and allow indexing
mgmt.buildIndex('byClass', Vertex.class).addKey(cls).buildCompositeIndex();
mgmt.buildIndex('byStoreIDUnique', Vertex.class).addKey(storeID).unique().buildCompositeIndex();
mgmt.buildIndex('byApp', Vertex.class).addKey(app).buildCompositeIndex();
mgmt.buildIndex('byTeam', Vertex.class).addKey(team).buildCompositeIndex();
mgmt.buildIndex('byService', Vertex.class).addKey(service).buildCompositeIndex();
mgmt.buildIndex('byName', Vertex.class).addKey(name).buildCompositeIndex();
mgmt.buildIndex('byNamespace', Vertex.class).addKey(namespace).buildCompositeIndex();
mgmt.buildIndex('byType', Vertex.class).addKey(type).buildCompositeIndex();
mgmt.buildIndex('byCritical', Vertex.class).addKey(critical).buildCompositeIndex();

mgmt.commit();

// Wait for indexes to become available
ManagementSystem.awaitGraphIndexStatus(graph, 'byClass').status(SchemaStatus.ENABLED).call();
ManagementSystem.awaitGraphIndexStatus(graph, 'byStoreIDUnique').status(SchemaStatus.ENABLED).call();
ManagementSystem.awaitGraphIndexStatus(graph, 'byApp').status(SchemaStatus.ENABLED).call();
ManagementSystem.awaitGraphIndexStatus(graph, 'byTeam').status(SchemaStatus.ENABLED).call();
ManagementSystem.awaitGraphIndexStatus(graph, 'byService').status(SchemaStatus.ENABLED).call();
ManagementSystem.awaitGraphIndexStatus(graph, 'byName').status(SchemaStatus.ENABLED).call();
ManagementSystem.awaitGraphIndexStatus(graph, 'byNamespace').status(SchemaStatus.ENABLED).call();
ManagementSystem.awaitGraphIndexStatus(graph, 'byType').status(SchemaStatus.ENABLED).call();
ManagementSystem.awaitGraphIndexStatus(graph, 'byCritical').status(SchemaStatus.ENABLED).call();

System.out.println("[KUBEHOUND] graph schema and indexes ready");
mgmt.close();

// Close the open connection
:remote close