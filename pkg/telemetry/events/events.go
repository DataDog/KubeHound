package events

import (
	kstatsd "github.com/DataDog/KubeHound/pkg/telemetry/statsd"
	"github.com/DataDog/datadog-go/v5/statsd"
)

const (
	DumperRun  = "kubehound.dumper.run"
	DumperStop = "kubehound.dumper.stop"
)

func PushEvent(title string, text string, tags []string) {
	_ = kstatsd.Event(&statsd.Event{
		Title: title,
		Text:  text,
		Tags:  tags,
	})
}
