package graph

type Role struct {
	StoreID      string   `json:"storeID" mapstructure:"storeID"`
	App          string   `json:"app" mapstructure:"app"`
	Team         string   `json:"team" mapstructure:"team"`
	Service      string   `json:"service" mapstructure:"service"`
	Name         string   `json:"name" mapstructure:"name"`
	IsNamespaced bool     `json:"isNamespaced" mapstructure:"isNamespaced"`
	Namespace    string   `json:"namespace" mapstructure:"namespace"`
	Rules        []string `json:"rules" mapstructure:"rules"`
	Critical     bool     `json:"critical" mapstructure:"critical"`
}