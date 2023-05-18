package pipeline

import (
	"github.com/DataDog/KubeHound/pkg/collector"
	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/graphdb"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
)

type Dependencies struct {
	Config    *config.KubehoundConfig
	Collector collector.CollectorClient
	Cache     cache.CacheProvider
	StoreDB   storedb.Provider
	GraphDB   graphdb.Provider
}
