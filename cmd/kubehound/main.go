package main

import (
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
)

func main() {
	if err := Execute(); err != nil {
		log.DefaultLogger().Fatal("unable to execute command", log.ErrorField(err))
	}
}
