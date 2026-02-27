package converter

import (
	"strconv"
	"strings"

	rbacv1 "k8s.io/api/rbac/v1"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/store"
	"github.com/DataDog/KubeHound/pkg/kubehound/risk"
)

// GraphConverter enables converting between an input store model to its equivalent graph model.
type GraphConverter struct {
	runtime *config.DynamicConfig
}

// NewGraph returns a new graph converter instance.
func NewGraph(cfg *config.KubehoundConfig) *GraphConverter {
	return &GraphConverter{
		runtime: &cfg.Dynamic,
	}
}

// Container returns the graph representation of a container vertex from a store container model input.
func (c *GraphConverter) Container(input *store.Container, parent *store.Pod) (*graph.Container, error) {
	output := &graph.Container{
		StoreID:     store.Hex(input.Id),
		App:         input.Ownership.Application,
		Team:        input.Ownership.Team,
		Service:     input.Ownership.Service,
		RunID:       c.runtime.RunID.String(),
		Cluster:     c.runtime.Cluster.Name,
		Namespace:   input.Inherited.Namespace,
		Name:        input.Name,
		Image:       input.Image,
		Command:     input.Command,
		Args:        input.Args,
		HostPID:     input.Inherited.HostPID,
		HostIPC:     input.Inherited.HostIPC,
		HostNetwork: input.Inherited.HostNetwork,
		Pod:         input.Inherited.PodName,
		Node:        input.Inherited.NodeName,
		RunAsUser:   input.Inherited.RunAsUser,
		Privileged:  input.Privileged,
		PrivEsc:     input.PrivEsc,
	}

	// Capabilities
	output.Capabilities = make([]string, len(input.Capabilities))
	copy(output.Capabilities, input.Capabilities)

	// Ports
	output.Ports = make([]string, 0, len(input.Ports))
	for _, p := range input.Ports {
		output.Ports = append(output.Ports, strconv.Itoa(int(p)))
	}

	if output.Namespace != "" {
		output.IsNamespaced = true
	}

	return output, nil
}

// Node returns the graph representation of a node vertex from a store node model input.
func (c *GraphConverter) Node(input *store.Node) (*graph.Node, error) {
	output := &graph.Node{
		StoreID:  store.Hex(input.Id),
		App:      input.Ownership.Application,
		Team:     input.Ownership.Team,
		Service:  input.Ownership.Service,
		RunID:    c.runtime.RunID.String(),
		Cluster:  c.runtime.Cluster.Name,
		Name:     input.Name,
		Critical: risk.Engine().IsCritical(input),
	}

	if input.IsNamespaced {
		output.IsNamespaced = true
		output.Namespace = input.Namespace
	}

	return output, nil
}

// Pod returns the graph representation of a pod vertex from a store pod model input.
func (c *GraphConverter) Pod(input *store.Pod) (*graph.Pod, error) {
	output := &graph.Pod{
		StoreID:               store.Hex(input.Id),
		App:                   input.Ownership.Application,
		Team:                  input.Ownership.Team,
		Service:               input.Ownership.Service,
		RunID:                 c.runtime.RunID.String(),
		Cluster:               c.runtime.Cluster.Name,
		Name:                  input.Name,
		Namespace:             input.Namespace,
		ServiceAccount:        input.ServiceAccount,
		Node:                  input.NodeName,
		ShareProcessNamespace: input.ShareProcessNamespace,
		Critical:              risk.Engine().IsCritical(input),
	}
	if output.Namespace != "" {
		output.IsNamespaced = true
	}

	return output, nil
}

// Volume returns the graph representation of a volume vertex from a store volume model input.
func (c *GraphConverter) Volume(input *store.Volume, parent *store.Pod) (*graph.Volume, error) {
	output := &graph.Volume{
		StoreID:    store.Hex(input.Id),
		App:        input.Ownership.Application,
		Team:       input.Ownership.Team,
		Service:    input.Ownership.Service,
		RunID:      c.runtime.RunID.String(),
		Cluster:    c.runtime.Cluster.Name,
		Name:       input.Name,
		Namespace:  parent.Namespace,
		Type:       input.Type,
		SourcePath: input.SourcePath,
		MountPath:  input.MountPath,
		Readonly:   input.ReadOnly,
	}

	if output.Namespace != "" {
		output.IsNamespaced = true
	}

	return output, nil
}

// flattenPolicyRules flattens the policy rule array into a string array.
func (c *GraphConverter) flattenPolicyRules(input []rbacv1.PolicyRule) []string {
	rules := make([]string, 0, len(input))

	for _, i := range input {
		var sb strings.Builder

		sb.WriteString("API(")
		sb.WriteString(strings.Join(i.APIGroups, ","))
		sb.WriteString(")::")

		sb.WriteString("R(")
		sb.WriteString(strings.Join(i.Resources, ","))
		sb.WriteString(")::")

		sb.WriteString("N(")
		sb.WriteString(strings.Join(i.ResourceNames, ","))
		sb.WriteString(")::")

		sb.WriteString("V(")
		sb.WriteString(strings.Join(i.Verbs, ","))
		sb.WriteString(")")

		rules = append(rules, sb.String())
	}

	return rules
}

// PermissionSet returns the graph representation of a role vertex from a store role model input.
func (c *GraphConverter) PermissionSet(input *store.PermissionSet) (*graph.PermissionSet, error) {
	output := &graph.PermissionSet{
		StoreID:     store.Hex(input.Id),
		App:         input.Ownership.Application,
		Team:        input.Ownership.Team,
		Service:     input.Ownership.Service,
		RunID:       c.runtime.RunID.String(),
		Cluster:     c.runtime.Cluster.Name,
		Name:        input.Name,
		Namespace:   input.Namespace,
		Role:        input.RoleName,
		RoleBinding: input.RoleBindingName,
		Rules:       c.flattenPolicyRules(input.Rules),
		Critical:    risk.Engine().IsCritical(input),
	}

	if output.Namespace != "" {
		output.IsNamespaced = true
	}

	return output, nil
}

// Identity returns the graph representation of an identity vertex from a store identity model input.
func (c *GraphConverter) Identity(input *store.Identity) (*graph.Identity, error) {
	output := &graph.Identity{
		StoreID:   store.Hex(input.Id),
		App:       input.Ownership.Application,
		Team:      input.Ownership.Team,
		Service:   input.Ownership.Service,
		RunID:     c.runtime.RunID.String(),
		Cluster:   c.runtime.Cluster.Name,
		Name:      input.Name,
		Namespace: input.Namespace,
		Type:      input.Type,
		Critical:  risk.Engine().IsCritical(input),
	}

	if output.Namespace != "" {
		output.IsNamespaced = true
	}

	return output, nil
}

// Endpoint returns the graph representation of an endpoint vertex from a store endpoint model input.
func (c *GraphConverter) Endpoint(input *store.Endpoint) (*graph.Endpoint, error) {
	output := &graph.Endpoint{
		StoreID:             store.Hex(input.Id),
		App:                 input.Ownership.Application,
		Team:                input.Ownership.Team,
		Service:             input.Ownership.Service,
		RunID:               c.runtime.RunID.String(),
		Cluster:             c.runtime.Cluster.Name,
		Namespace:           input.Namespace,
		IsNamespaced:        input.IsNamespaced,
		Name:                input.Name,
		ServiceEndpointName: input.ServiceName,
		ServiceDnsName:      input.ServiceDns,
		AddressType:         input.AddressType,
		Addresses:           input.Addresses,
		Port:                input.Port,
		PortName:            input.PortName,
		Protocol:            input.Protocol,
		Exposure:            input.Exposure,
	}

	return output, nil
}
