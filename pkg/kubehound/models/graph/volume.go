package graph

type Volume struct {
	StoreID      string `json:"storeID" mapstructure:"storeID"`
	App          string `json:"app" mapstructure:"app"`
	Team         string `json:"team" mapstructure:"team"`
	Service      string `json:"service" mapstructure:"service"`
	RunID        string `json:"runID" mapstructure:"runID"`
	Cluster      string `json:"cluster" mapstructure:"cluster"`
	IsNamespaced bool   `json:"isNamespaced" mapstructure:"isNamespaced"`
	Namespace    string `json:"namespace" mapstructure:"namespace"`
	Name         string `json:"name" mapstructure:"name"`
	Type         string `json:"type" mapstructure:"type"`
	SourcePath   string `json:"sourcePath" mapstructure:"sourcePath"`
	MountPath    string `json:"mountPath" mapstructure:"mountPath"`
	Readonly     bool   `json:"readonly" mapstructure:"readonly"`
}
