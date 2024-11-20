package cmd

import (
	"context"
	"fmt"

	"strings"

	"github.com/DataDog/KubeHound/pkg/telemetry/events"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
)

func AskForConfirmation(ctx context.Context) (bool, error) {
	l := log.Logger(ctx)

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
		l.Info("Please type (y)es or (n)o and then press enter:")

		return AskForConfirmation(ctx)
	}
}

func ReportError(ctx context.Context, action events.EventAction, err error) {
	l := log.Logger(ctx)
	if err != nil {
		errMsg := fmt.Errorf("fatal error: %w", err)
		l.Error("Error occurred", log.ErrorField(errMsg))
		_ = events.PushEvent(ctx, action, fmt.Sprintf("%s", errMsg))
	}
}
