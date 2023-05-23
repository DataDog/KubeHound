package collector

import (
	"context"
	"testing"

	mocks "github.com/DataDog/KubeHound/pkg/collector/mockingest"
	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFileCollector_Constructor(t *testing.T) {
	cfg, err := config.NewConfig("testdata/kubehound-test.yaml")
	assert.NoError(t, err)

	c, err := NewFileCollector(context.Background(), cfg)
	assert.NoError(t, err)
	assert.IsType(t, &FileCollector{}, c)
	assert.Equal(t, cfg.Collector.File.Directory, c.(*FileCollector).cfg.Directory)
}

func TestFileCollector_HealthCheck(t *testing.T) {
	c := &FileCollector{
		cfg: &config.FileCollectorConfig{
			Directory: "does-not-exist/",
		},
	}

	ok, err := c.HealthCheck(context.Background())
	assert.False(t, ok)
	assert.ErrorContains(t, err, "no such file or directory")

	c = &FileCollector{
		cfg: &config.FileCollectorConfig{
			Directory: "testdata/kubehound-test.yaml",
		},
	}

	ok, err = c.HealthCheck(context.Background())
	assert.False(t, ok)
	assert.ErrorContains(t, err, "is not a directory")

	c = &FileCollector{
		cfg: &config.FileCollectorConfig{
			Directory: "testdata/test-cluster/",
		},
	}

	ok, err = c.HealthCheck(context.Background())
	assert.True(t, ok)
	assert.NoError(t, err)
}

func NewTestFileCollector(t *testing.T) *FileCollector {
	cfg, err := config.NewConfig("testdata/kubehound-test.yaml")
	assert.NoError(t, err)

	c, err := NewFileCollector(context.Background(), cfg)
	assert.NoError(t, err)

	return c.(*FileCollector)
}

func TestFileCollector_StreamNodes(t *testing.T) {
	c := NewTestFileCollector(t)
	ctx := context.Background()
	i := mocks.NewNodeIngestor(t)

	i.EXPECT().IngestNode(ctx, mock.AnythingOfType("types.NodeType")).Return(nil)
	i.EXPECT().Complete(ctx).Return(nil).Once()

	err := c.StreamNodes(ctx, i)
	assert.NoError(t, err)
}

func TestFileCollector_StreamPods(t *testing.T) {
	c := NewTestFileCollector(t)
	ctx := context.Background()
	i := mocks.NewPodIngestor(t)

	i.EXPECT().IngestPod(ctx, mock.AnythingOfType("types.PodType")).Return(nil).Twice()
	i.EXPECT().Complete(ctx).Return(nil).Once()

	err := c.StreamPods(ctx, i)
	assert.NoError(t, err)
}

func TestFileCollector_StreamRoles(t *testing.T) {
	c := NewTestFileCollector(t)
	ctx := context.Background()
	i := mocks.NewRoleIngestor(t)

	i.EXPECT().IngestRole(ctx, mock.AnythingOfType("types.RoleType")).Return(nil).Twice()
	i.EXPECT().Complete(ctx).Return(nil).Once()

	err := c.StreamRoles(ctx, i)
	assert.NoError(t, err)
}

func TestFileCollector_StreamRoleBindings(t *testing.T) {
	c := NewTestFileCollector(t)
	ctx := context.Background()
	i := mocks.NewRoleBindingIngestor(t)

	i.EXPECT().IngestRoleBinding(ctx, mock.AnythingOfType("types.RoleBindingType")).Return(nil).Twice()
	i.EXPECT().Complete(ctx).Return(nil).Once()

	err := c.StreamRoleBindings(ctx, i)
	assert.NoError(t, err)
}

func TestFileCollector_StreamClusterRoles(t *testing.T) {
	c := NewTestFileCollector(t)
	ctx := context.Background()
	i := mocks.NewClusterRoleIngestor(t)

	i.EXPECT().IngestClusterRole(ctx, mock.AnythingOfType("types.ClusterRoleType")).Return(nil)
	i.EXPECT().Complete(ctx).Return(nil).Once()

	err := c.StreamClusterRoles(ctx, i)
	assert.NoError(t, err)
}

func TestFileCollector_StreamClusterRoleBindings(t *testing.T) {
	c := NewTestFileCollector(t)
	ctx := context.Background()
	i := mocks.NewClusterRoleBindingIngestor(t)

	i.EXPECT().IngestClusterRoleBinding(ctx, mock.AnythingOfType("types.ClusterRoleBindingType")).Return(nil)
	i.EXPECT().Complete(ctx).Return(nil).Once()

	err := c.StreamClusterRoleBindings(ctx, i)
	assert.NoError(t, err)
}
