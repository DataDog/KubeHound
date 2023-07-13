package store

type OwnershipInfo struct {
	Application string `bson:"application"`
	Team        string `bson:"team"`
	Service     string `bson:"service"`
}

func ExtractOwnership(labels map[string]string) OwnershipInfo {
	return OwnershipInfo{
		Application: labels["app"],
		Team:        labels["team"],
		Service:     labels["service"],
	}
}
