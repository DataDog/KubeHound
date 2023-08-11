package converter

import (
	"strconv"
	"strings"

	rbacv1 "k8s.io/api/rbac/v1"

	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/store"
	"github.com/DataDog/KubeHound/pkg/kubehound/risk"
)

// GraphConverter enables converting between an input store model to its equivalent graph model.
type GraphConverter struct {
}

// NewGraph returns a new graph converter instance.
func NewGraph() *GraphConverter {
	return &GraphConverter{}
}

// Container returns the graph representation of a container vertex from a store container model input.
func (c *GraphConverter) Container(input *store.Container, parent *store.Pod) (*graph.Container, error) {
	output := &graph.Container{
		StoreID:     input.Id.Hex(),
		App:         input.Ownership.Application,
		Team:        input.Ownership.Team,
		Service:     input.Ownership.Service,
		Namespace:   input.Inherited.Namespace,
		Name:        input.K8.Name,
		Image:       input.K8.Image,
		Command:     input.K8.Command,
		Args:        input.K8.Args,
		HostPID:     input.Inherited.HostPID,
		HostIPC:     input.Inherited.HostIPC,
		HostNetwork: input.Inherited.HostNetwork,
		Pod:         input.Inherited.PodName,
		Node:        input.Inherited.NodeName,
		RunAsUser:   input.Inherited.RunAsUser,
	}

	// Determine if a user is set in the security context
	if input.K8.SecurityContext != nil && input.K8.SecurityContext.RunAsUser != nil {
		output.RunAsUser = *input.K8.SecurityContext.RunAsUser
	}

	// Privileged pod/container
	if input.K8.SecurityContext != nil && input.K8.SecurityContext.Privileged != nil {
		output.Privileged = *input.K8.SecurityContext.Privileged
	}

	// Privilege escalation permitted
	if input.K8.SecurityContext != nil && input.K8.SecurityContext.AllowPrivilegeEscalation != nil {
		output.PrivEsc = *input.K8.SecurityContext.AllowPrivilegeEscalation
	}

	// Capabilities
	output.Capabilities = make([]string, 0)
	if input.K8.SecurityContext != nil && input.K8.SecurityContext.Capabilities != nil {
		for _, cap := range input.K8.SecurityContext.Capabilities.Add {
			output.Capabilities = append(output.Capabilities, string(cap))
		}
	}

	// Exposed ports
	output.Ports = make([]string, 0)
	if input.K8.Ports != nil {
		for _, p := range input.K8.Ports {
			// We map the integer port to a string to make our lives easier in the bullk gaph insert (#gremlin)
			output.Ports = append(output.Ports, strconv.Itoa((int(p.ContainerPort))))
		}
	}

	if output.Namespace != "" {
		output.IsNamespaced = true
	}

	return output, nil
}

// Node returns the graph representation of a node vertex from a store node model input.
func (c *GraphConverter) Node(input *store.Node) (*graph.Node, error) {
	output := &graph.Node{
		StoreID:  input.Id.Hex(),
		App:      input.Ownership.Application,
		Team:     input.Ownership.Team,
		Service:  input.Ownership.Service,
		Name:     input.K8.Name,
		Critical: risk.Engine().IsCritical(input),
	}

	if input.IsNamespaced {
		output.IsNamespaced = true
		output.Namespace = input.K8.Namespace
	}

	return output, nil
}

// Pod returns the graph representation of a pod vertex from a store pod model input.
func (c *GraphConverter) Pod(input *store.Pod) (*graph.Pod, error) {
	output := &graph.Pod{
		StoreID:        input.Id.Hex(),
		App:            input.Ownership.Application,
		Team:           input.Ownership.Team,
		Service:        input.Ownership.Service,
		Name:           input.K8.Name,
		Namespace:      input.K8.GetNamespace(),
		ServiceAccount: input.K8.Spec.ServiceAccountName,
		Node:           input.K8.Spec.NodeName,
		Critical:       risk.Engine().IsCritical(input),
	}
	if input.K8.Spec.ShareProcessNamespace != nil {
		output.SharedProcessNamespace = *input.K8.Spec.ShareProcessNamespace
	}
	if output.Namespace != "" {
		output.IsNamespaced = true
	}

	return output, nil
}

// Volume returns the graph representation of a volume vertex from a store volume model input.
func (c *GraphConverter) Volume(input *store.Volume, parent *store.Pod) (*graph.Volume, error) {
	output := &graph.Volume{
		StoreID:    input.Id.Hex(),
		App:        input.Ownership.Application,
		Team:       input.Ownership.Team,
		Service:    input.Ownership.Service,
		Name:       input.Name,
		Namespace:  parent.K8.Namespace,
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
// This is necessary as graph databases cannot typically handle complex data type attributes on nodes.
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
		StoreID:     input.Id.Hex(),
		App:         input.Ownership.Application,
		Team:        input.Ownership.Team,
		Service:     input.Ownership.Service,
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
		StoreID:   input.Id.Hex(),
		App:       input.Ownership.Application,
		Team:      input.Ownership.Team,
		Service:   input.Ownership.Service,
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
		StoreID:             input.Id.Hex(),
		App:                 input.Ownership.Application,
		Team:                input.Ownership.Team,
		Service:             input.Ownership.Service,
		Namespace:           input.Namespace,
		IsNamespaced:        input.IsNamespaced,
		Name:                input.Name,
		ServiceEndpointName: input.ServiceName,
		ServiceDnsName:      input.ServiceDns,
		AddressType:         string(input.AddressType),
		Addresses:           input.Backend.Addresses,
		Port:                input.SafePort(),
		PortName:            input.SafePortName(),
		Protocol:            input.SafeProtocol(),
		Exposure:            input.Exposure,
	}

	return output, nil
}
