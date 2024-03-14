package main

import (
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.I.Fatal(err.Error())
	}
}
