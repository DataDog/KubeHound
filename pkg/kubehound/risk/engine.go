package risk

import (
	"sync"

	"github.com/DataDog/KubeHound/pkg/kubehound/models/store"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
)

var engineInstance *RiskEngine
var riOnce sync.Once

func Engine() *RiskEngine {
	var err error
	riOnce.Do(func() {
		engineInstance, err = NewEngine()
		if err != nil {
			log.I.Fatalf("Risk engine initialization: %v", err)
		}
	})

	return engineInstance
}

type RiskEngine struct {
	roleMap map[string]bool // Map of critical roles
}

func NewEngine() (*RiskEngine, error) {
	return &RiskEngine{
		roleMap: CriticalRoleMap,
	}, nil
}

func (ra *RiskEngine) IsCritical(model any) bool {
	switch o := model.(type) {
	case *store.Role:
		if ra.roleMap[o.Name] {
			return true
		}
	}

	return false
}
