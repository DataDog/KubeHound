package globals

import (
	"os"
)

const (
	DDServiceName = "kubehound"
	DefaultDDEnv  = "dev"
)

func GetDDEnv() string {
	env := os.Getenv("DD_ENV")
	if env == "" {
		return DefaultDDEnv
	}

	return env
}

func GetDDServiceName() string {
	serviceName := os.Getenv("DD_SERVICE_NAME")
	if serviceName == "" {
		return DDServiceName
	}

	return serviceName
}
