package store

// OwnershipInfo encapsulates internal ownership information of Kubernetes assets.
type OwnershipInfo struct {
	Application string `bson:"application"`
	Team        string `bson:"team"`
	Service     string `bson:"service"`
}

// ExtractOwnership extracts ownership information from a provided Kubernets labels map.
func ExtractOwnership(labels map[string]string) OwnershipInfo {
	return OwnershipInfo{
		Application: labels["app"],
		Team:        labels["team"],
		Service:     labels["service"],
	}
}
