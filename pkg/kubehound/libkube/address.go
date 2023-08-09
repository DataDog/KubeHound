package libkube

import (
	"fmt"
	"net"
	"regexp"

	discoveryv1 "k8s.io/api/discovery/v1"
)

const (
	AddressFQDNMatcher = `^(([a-z0-9][a-z0-9\-]*[a-z0-9])|[a-z0-9]+\.)*([a-z]+|xn\-\-[a-z0-9]+)\.?$`
)

// AddressType returns the discoveryv1.AddressType identifier corresponding to the provided address.
func AddressType(address string) (discoveryv1.AddressType, error) {
	ip := net.ParseIP(address)
	switch {
	case ip.To4() != nil:
		return discoveryv1.AddressTypeIPv4, nil
	case ip.To16() != nil:
		return discoveryv1.AddressTypeIPv6, nil
	}

	if match, _ := regexp.MatchString(AddressFQDNMatcher, address); match {
		return discoveryv1.AddressTypeFQDN, nil
	}

	return "", fmt.Errorf("invalid addrress input: %s", address)
}
