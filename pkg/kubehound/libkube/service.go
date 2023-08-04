package libkube

import (
	"fmt"

	"github.com/DataDog/KubeHound/pkg/globals/types"
)

// ServiceName returns the name of the service associated with the provided EndpointSlice.
func ServiceName(ep types.EndpointType) string {
	return ep.Labels["kubernetes.io/service-name"]
}

// ServiceDns provides the DNS name of the service associated with the provided EndpointSlice.
func ServiceDns(ep types.EndpointType) string {
	clusterName := "stripe.us1.staging.dog" // TODO dynamic
	namespace := ep.Namespace

	return fmt.Sprintf("%s.%s.%s", ep.Labels["kubernetes.io/service-name"], namespace, clusterName)
}
