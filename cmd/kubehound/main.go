package main

import (
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/telemetry/tag"
)

func main() {
	tag.SetupBaseTags()
	if err := rootCmd.Execute(); err != nil {
		log.Logger(rootCmd.Context()).Fatal(err.Error())
	}
}
