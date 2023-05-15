package config

import (
	"github.com/DataDog/KubeHound/pkg/globals"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
)

type KubehoundConfig struct {
}

func MustLoadDefaultConfig() *KubehoundConfig {
	log.I.Fatal(globals.ErrNotImplemented.Error())
	return nil
}
