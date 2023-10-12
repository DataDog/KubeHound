//nolint:unparam,gocritic
package risk

import (
	"sync"

	"github.com/DataDog/KubeHound/pkg/kubehound/models/store"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
)

var engineInstance *RiskEngine
var riOnce sync.Once

// Engine returns the risk engine singleton instance.
func Engine() *RiskEngine {
	var err error
	riOnce.Do(func() {
		engineInstance, err = newEngine()
		if err != nil {
			log.I.Fatalf("Risk engine initialization: %v", err)
		}
	})

	return engineInstance
}

// RiskEngine computes which assets are deemed critical based on a set of pre-configured rules.
type RiskEngine struct {
	roleMap map[string]bool // Map of critical roles
}

// newEngine creates a new risk engine instance. Should not be called directly.
func newEngine() (*RiskEngine, error) {
	return &RiskEngine{
		roleMap: CriticalRoleMap,
	}, nil
}

// IsCritical reports whether the provided asset should be marked as critical.
// The function expects a single store model input and currently only supports Roles.
func (ra *RiskEngine) IsCritical(model any) bool {
	switch o := model.(type) {
	case *store.PermissionSet:
		if ra.roleMap[o.RoleName] && !o.IsNamespaced {
			return true
		}
	}

	return false
}
