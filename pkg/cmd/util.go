package cmd

import (
	"fmt"

	"strings"

	"github.com/DataDog/KubeHound/pkg/telemetry/log"
)

func AskForConfirmation() (bool, error) {
	var response string

	_, err := fmt.Scanln(&response)
	if err != nil && err.Error() != "unexpected newline" {
		return false, fmt.Errorf("scanln: %w", err)
	}

	switch strings.ToLower(response) {
	case "y", "yes":
		return true, nil
	case "n", "no":
		return false, nil
	default:
		log.I.Error("Please type (y)es or (n)o and then press enter:")

		return AskForConfirmation()
	}
}
