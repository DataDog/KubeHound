package main

import (
	"errors"

	"github.com/DataDog/KubeHound/pkg/cmd"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/telemetry/tag"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

func main() {
	tag.SetupBaseTags()
	err := rootCmd.Execute()
	err = errors.Join(err, cmd.CloseKubehoundConfig(rootCmd.Context()))
	if err != nil {
		log.Logger(rootCmd.Context()).Fatal(err.Error())
	}
}
