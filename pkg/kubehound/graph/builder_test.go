package graph

// Create mock edge
// Test error case
// Test success case
// test one error cancels other works
import (
	"testing"

	graphdb "github.com/DataDog/KubeHound/pkg/kubehound/storage/graphdb/mocks"
)

func NewTestBuilder(t *testing.T) *Builder {
	gdb, err := graphdb.NewProvider(t)
}
