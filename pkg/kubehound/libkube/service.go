package libkube

import (
	"strings"

	"github.com/DataDog/KubeHound/pkg/globals/types"
)

func ServiceName(ep types.EndpointType) string {
	return ep.Labels["kubernetes.io/service-name"]
}

// TODO service DNS name
func ServiceDns(ep types.EndpointType) string {
	var sb strings.Builder

	sb.WriteString(ep.Labels["kubernetes.io/service-name"])
	sb.WriteString(".")
	sb.WriteString("stripe.us1.staging.dog")

	return sb.String()
}

// // Exposed port from containerPort
// func ExposedPort() int {

// }
