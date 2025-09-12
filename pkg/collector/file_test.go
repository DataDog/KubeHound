//nolint:forcetypeassert
package collector

import (
	"testing"

	mocks "github.com/DataDog/KubeHound/pkg/collector/mockingest"
	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFileCollector_Constructor(t *testing.T) {
	t.Parallel()

	v := viper.New()
	cfg, err := config.NewConfig(t.Context(), v, "testdata/kubehound-test.yaml")
	assert.NoError(t, err)

	c, err := NewFileCollector(t.Context(), cfg)
	assert.NoError(t, err)
	assert.IsType(t, &FileCollector{}, c)
	assert.Equal(t, cfg.Collector.File.Directory, c.(*FileCollector).cfg.Directory)
}

func TestFileCollector_HealthCheck(t *testing.T) {
	t.Parallel()

	c := &FileCollector{
		cfg: &config.FileCollectorConfig{
			Directory: "does-not-exist/",
		},
	}

	ok, err := c.HealthCheck(t.Context())
	assert.False(t, ok)
	assert.ErrorContains(t, err, "no such file or directory")

	c = &FileCollector{
		cfg: &config.FileCollectorConfig{
			Directory: "testdata/kubehound-test.yaml",
		},
	}

	ok, err = c.HealthCheck(t.Context())
	assert.False(t, ok)
	assert.ErrorContains(t, err, "is not a directory")

	c = &FileCollector{
		cfg: &config.FileCollectorConfig{
			Directory: "testdata/test-cluster/",
		},
		clusterName: "test-cluster",
	}

	ok, err = c.HealthCheck(t.Context())
	assert.True(t, ok)
	assert.NoError(t, err)
}

func NewTestFileCollector(t *testing.T) *FileCollector {
	t.Helper()

	v := viper.New()
	cfg, err := config.NewConfig(t.Context(), v, "testdata/kubehound-test.yaml")
	assert.NoError(t, err)

	c, err := NewFileCollector(t.Context(), cfg)
	assert.NoError(t, err)

	return c.(*FileCollector)
}

func TestFileCollector_ClusterInfo(t *testing.T) {
	t.Parallel()

	c := NewTestFileCollector(t)
	ctx := t.Context()

	info, err := c.ClusterInfo(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "test-cluster", info.Name)
}

func TestFileCollector_StreamNodes(t *testing.T) {
	t.Parallel()

	c := NewTestFileCollector(t)
	ctx := t.Context()
	i := mocks.NewNodeIngestor(t)

	i.EXPECT().IngestNode(mock.Anything, mock.AnythingOfType("types.NodeType")).Return(nil)
	i.EXPECT().Complete(mock.Anything).Return(nil).Once()

	err := c.StreamNodes(ctx, i)
	assert.NoError(t, err)
}

func TestFileCollector_StreamPods(t *testing.T) {
	t.Parallel()

	c := NewTestFileCollector(t)
	ctx := t.Context()
	i := mocks.NewPodIngestor(t)

	i.EXPECT().IngestPod(mock.Anything, mock.AnythingOfType("types.PodType")).Return(nil).Twice()
	i.EXPECT().Complete(mock.Anything).Return(nil).Once()

	err := c.StreamPods(ctx, i)
	assert.NoError(t, err)
}

func TestFileCollector_StreamRoles(t *testing.T) {
	t.Parallel()

	c := NewTestFileCollector(t)
	ctx := t.Context()
	i := mocks.NewRoleIngestor(t)

	i.EXPECT().IngestRole(mock.Anything, mock.AnythingOfType("types.RoleType")).Return(nil).Twice()
	i.EXPECT().Complete(mock.Anything).Return(nil).Once()

	err := c.StreamRoles(ctx, i)
	assert.NoError(t, err)
}

func TestFileCollector_StreamRoleBindings(t *testing.T) {
	t.Parallel()

	c := NewTestFileCollector(t)
	ctx := t.Context()
	i := mocks.NewRoleBindingIngestor(t)

	i.EXPECT().IngestRoleBinding(mock.Anything, mock.AnythingOfType("types.RoleBindingType")).Return(nil).Twice()
	i.EXPECT().Complete(mock.Anything).Return(nil).Once()

	err := c.StreamRoleBindings(ctx, i)
	assert.NoError(t, err)
}

func TestFileCollector_StreamClusterRoles(t *testing.T) {
	t.Parallel()

	c := NewTestFileCollector(t)
	ctx := t.Context()
	i := mocks.NewClusterRoleIngestor(t)

	i.EXPECT().IngestClusterRole(mock.Anything, mock.AnythingOfType("types.ClusterRoleType")).Return(nil)
	i.EXPECT().Complete(mock.Anything).Return(nil).Once()

	err := c.StreamClusterRoles(ctx, i)
	assert.NoError(t, err)
}

func TestFileCollector_StreamClusterRoleBindings(t *testing.T) {
	t.Parallel()

	c := NewTestFileCollector(t)
	ctx := t.Context()
	i := mocks.NewClusterRoleBindingIngestor(t)

	i.EXPECT().IngestClusterRoleBinding(mock.Anything, mock.AnythingOfType("types.ClusterRoleBindingType")).Return(nil)
	i.EXPECT().Complete(mock.Anything).Return(nil).Once()

	err := c.StreamClusterRoleBindings(ctx, i)
	assert.NoError(t, err)
}

func TestFileCollector_StreamEndpoints(t *testing.T) {
	t.Parallel()

	c := NewTestFileCollector(t)
	ctx := t.Context()
	i := mocks.NewEndpointIngestor(t)

	i.EXPECT().IngestEndpoint(mock.Anything, mock.AnythingOfType("types.EndpointType")).Return(nil)
	i.EXPECT().Complete(mock.Anything).Return(nil).Once()

	err := c.StreamEndpoints(ctx, i)
	assert.NoError(t, err)
}
