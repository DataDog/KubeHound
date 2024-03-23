package cmd

import (
	"context"
	"fmt"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/telemetry"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/telemetry/tag"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func GetConfig() (*config.KubehoundConfig, error) {
	// Passing the Kubehound config from viper
	khCfg := config.KubehoundConfig{}
	err := viper.Unmarshal(&khCfg)
	if err != nil {
		return nil, fmt.Errorf("unmarshal viper: %w", err)
	}

	return &khCfg, nil
}

func InitializeKubehoundConfig(ctx context.Context, configPath string, generateRunID bool) error {
	// We define a unique run id this so we can measure run by run in addition of version per version.
	// Useful when rerunning the same binary (same version) on different dataset or with different databases...
	// In the case of KHaaS, the runID is taken from the GRPC request argument
	if generateRunID {
		viper.Set(config.DynamicRunID, config.NewRunID())
	}

	clusterName, err := config.GetClusterName(ctx)
	if err == nil {
		viper.Set(config.DynamicClusterName, clusterName)
	} else {
		log.I.Errorf("collector cluster info: %v", err)
	}

	khCfg := config.NewKubehoundConfig(configPath, true)

	// Activate debug mode if needed
	if khCfg.Debug {
		log.I.Logger.SetLevel(logrus.DebugLevel)
	}

	InitTags(ctx, khCfg)
	InitTelemetry(khCfg)

	return nil

}

func InitTelemetry(khCfg *config.KubehoundConfig) {
	log.I.Info("Initializing application telemetry")
	err := telemetry.Initialize(khCfg)
	if err != nil {
		log.I.Warnf("failed telemetry initialization: %v", err)
	}
}

func InitTags(ctx context.Context, khCfg *config.KubehoundConfig) {

	if khCfg.Dynamic.ClusterName != "" {
		tag.AppendBaseTags(tag.ClusterName(khCfg.Dynamic.ClusterName))
	}

	if khCfg.Dynamic.RunID != nil {
		// We update the base tags to include that run id, so we have it available for metrics
		tag.AppendBaseTags(tag.RunID(khCfg.Dynamic.RunID.String()))

		// Set the run ID as a global log tag
		log.AddGlobalTags(map[string]string{
			tag.RunIdTag: khCfg.Dynamic.RunID.String(),
		})
	}

	// Update the logger behaviour from configuration
	log.SetDD(khCfg.Telemetry.Enabled)
	log.AddGlobalTags(khCfg.Telemetry.Tags)
}

func CloseKubehoundConfig() error {
	khCfg, err := GetConfig()
	if err != nil {
		return err
	}

	telemetry.Shutdown(khCfg.Telemetry.Enabled)

	return nil
}
