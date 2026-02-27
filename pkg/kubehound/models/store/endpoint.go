package store

import (
	"github.com/DataDog/KubeHound/pkg/kubehound/models/shared"
)

const (
	DefaultEndpointProtocol = "TCP"
	DefaultPortName         = ""
)

type Endpoint struct {
	Id           int64
	ContainerId  int64
	PodName      string
	PodNamespace string
	NodeName     string
	IsNamespaced bool
	Namespace    string
	Name         string
	HasSlice     bool
	ServiceName  string
	ServiceDns   string
	AddressType  string
	Addresses    []string
	Port         int
	PortName     string
	Protocol     string
	Exposure     shared.EndpointExposureType
	Ownership    OwnershipInfo
	Runtime      RuntimeInfo
}

// SafePort is a safe accessor for the endpoint port.
func (e *Endpoint) SafePort() int {
	return e.Port
}

// SafeProtocol is a safe accessor for the endpoint protocol.
func (e *Endpoint) SafeProtocol() string {
	if e.Protocol == "" {
		return DefaultEndpointProtocol
	}
	return e.Protocol
}

// SafePortName is a safe accessor for the endpoint port name.
func (e *Endpoint) SafePortName() string {
	return e.PortName
}
