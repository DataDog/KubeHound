package main

import (
	"github.com/DataDog/KubeHound/pkg/cmd"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/telemetry/tag"
)

func main() {
	tag.SetupBaseTags()
	err := rootCmd.Execute()
	cmd.CloseKubehoundConfig(rootCmd.Context())
	if err != nil {
		log.Logger(rootCmd.Context()).Fatal(err.Error())
	}
}
