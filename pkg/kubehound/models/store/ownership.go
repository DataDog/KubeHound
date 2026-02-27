package store

// OwnershipInfo encapsulates internal ownership information of Kubernetes assets.
type OwnershipInfo struct {
	Application string
	Team        string
	Service     string
}

// ExtractOwnership extracts ownership information from a provided Kubernets labels map.
func ExtractOwnership(labels map[string]string) OwnershipInfo {
	return OwnershipInfo{
		Application: labels["app"],
		Team:        labels["team"],
		Service:     labels["service"],
	}
}
