package core

import (
	"context"
	"fmt"

	"github.com/DataDog/KubeHound/pkg/collector"
	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/graphdb"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
	"github.com/DataDog/KubeHound/pkg/telemetry"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/telemetry/tag"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type ProvidersFactoryConfig struct {
	CacheProvider cache.CacheProvider
	StoreProvider storedb.Provider
	GraphProvider graphdb.Provider
}

// Initiating all the providers need for KubeHound (cache, store, graph)
func NewProvidersFactoryConfig(ctx context.Context, lc *LaunchConfig) (*ProvidersFactoryConfig, error) {
	// Create the cache client
	log.I.Info("Loading cache provider")
	cp, err := cache.Factory(ctx, lc.Cfg)
	if err != nil {
		return nil, fmt.Errorf("cache client creation: %w", err)
	}
	log.I.Infof("Loaded %s cache provider", cp.Name())

	// Create the store client
	log.I.Info("Loading store database provider")
	sp, err := storedb.Factory(ctx, lc.Cfg)
	if err != nil {
		return nil, fmt.Errorf("store database client creation: %w", err)
	}
	log.I.Infof("Loaded %s store provider", sp.Name())

	err = sp.Prepare(ctx)
	if err != nil {
		return nil, fmt.Errorf("store database prepare: %w", err)
	}

	// Create the graph client
	log.I.Info("Loading graph database provider")
	gp, err := graphdb.Factory(ctx, lc.Cfg)
	if err != nil {
		return nil, fmt.Errorf("graph database client creation: %w", err)
	}
	log.I.Infof("Loaded %s graph provider", gp.Name())

	err = gp.Prepare(ctx)
	if err != nil {
		return nil, fmt.Errorf("graph database prepare: %w", err)
	}

	return &ProvidersFactoryConfig{
		CacheProvider: cp,
		StoreProvider: sp,
		GraphProvider: gp,
	}, nil
}

func (fc *ProvidersFactoryConfig) Close(ctx context.Context) {
	fc.CacheProvider.Close(ctx)
	fc.StoreProvider.Close(ctx)
	fc.GraphProvider.Close(ctx)
}

type LaunchConfig struct {
	Cfg      *config.KubehoundConfig
	ts       *telemetry.State
	mainSpan ddtrace.Span
	RunID    *config.RunID
	opName   string
}

// Initiating configuration of Kubehound (from file or inline)
// This includes all the telemetry aspects (tags, spans, tracer, runID, logs, ...)
func NewLaunchConfig(ctx context.Context, opName string, configPath string, inline bool) (context.Context, *LaunchConfig) {
	// Configuration initialization
	cfg := config.NewKubehoundConfig(configPath, inline)

	lc := &LaunchConfig{
		Cfg:    cfg,
		opName: opName,
	}

	return ctx, lc
}

func (l *LaunchConfig) Initialize(ctx context.Context, generateRunID bool) context.Context {
	// We define a unique run id this so we can measure run by run in addition of version per version.
	// Useful when rerunning the same binary (same version) on different dataset or with different databases...
	// In the case of KHaaS, the runID is taken from the GRPC request argument
	if generateRunID {
		l.RunID = config.NewRunID()
	}

	l.InitTags(ctx)
	l.InitTelemetry()

	l.mainSpan, ctx = tracer.StartSpanFromContext(ctx, l.opName, tracer.Measured())

	return ctx

}

func (l *LaunchConfig) InitTelemetry() {
	var err error
	log.I.Info("Initializing application telemetry")
	l.ts, err = telemetry.Initialize(l.Cfg)
	if err != nil {
		log.I.Warnf("failed telemetry initialization: %v", err)
	}
}

func (l *LaunchConfig) InitTags(ctx context.Context) {

	clusterName, err := collector.GetClusterName(ctx)
	if err == nil {
		tag.AppendBaseTags(tag.ClusterName(clusterName))
	} else {
		log.I.Errorf("collector cluster info: %v", err)
	}

	if l.RunID != nil {
		// We update the base tags to include that run id, so we have it available for metrics
		tag.AppendBaseTags(tag.RunID(l.RunID.String()))

		// Set the run ID as a global log tag
		log.AddGlobalTags(map[string]string{
			tag.RunIdTag: l.RunID.String(),
		})
	}

	// Update the logger behaviour from configuration
	log.SetDD(l.Cfg.Telemetry.Enabled)
	log.AddGlobalTags(l.Cfg.Telemetry.Tags)
}

func (l *LaunchConfig) Close() {
	l.mainSpan.Finish()
	telemetry.Shutdown(l.ts)
}
