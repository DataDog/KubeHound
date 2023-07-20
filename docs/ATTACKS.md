# Adding an Attack

To add a new attack to KubeHound, please do the following

+ Document the attack in the [edges documentation](./edges/) directory
+ Define the attack constraints in the graph database [schema builder](../deployments/kubehound/janusgraph/kubehound-db-init.groovy)
+ Create an implementation of the [edge.Builder](../pkg/kubehound/graph/edge/builder.go) interface that determines whether attacks are possible by quering the store database and writes any found as edges into the graph database
+ Create the [resources](../test/setup/test-cluster/attacks/) file in the test cluster that will introduce an instance of the attack into the test cluster 
+ Add an [edge system test](../test/system/graph_edge_test.go) that verifies the attack is correctly created by KubeHound
