package events

import (
	"context"
	"fmt"

	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	kstatsd "github.com/DataDog/KubeHound/pkg/telemetry/statsd"
	"github.com/DataDog/KubeHound/pkg/telemetry/tag"
	"github.com/DataDog/datadog-go/v5/statsd"
)

const (
	IngestSkip = iota
	IngestStarted
	IngestFinished
	IngestorInit
	IngestorFailed
	DumpStarted
	DumpFinished
	DumpFailed
)

const (
	EventActionFail   = "fail"
	EventActionInit   = "init"
	EventActionStart  = "start"
	EventActionSkip   = "skip"
	EventActionFinish = "finish"
)

type EventAction int

type EventActionDetails struct {
	Title  string
	Text   string
	Level  statsd.EventAlertType
	Action string
}

// Could also be a format stirng template in this case, if needed?
var map2msg = map[EventAction]EventActionDetails{
	IngestorFailed: {Title: "Ingestor/grpc endpoint init failed", Level: statsd.Error, Action: EventActionFail},
	IngestorInit:   {Title: "Ingestor/grpc endpoint initiated", Level: statsd.Info, Action: EventActionInit},
	IngestStarted:  {Title: "Ingestion started", Level: statsd.Info, Action: EventActionStart},
	IngestSkip:     {Title: "Ingestion skipped", Level: statsd.Info, Action: EventActionSkip},
	IngestFinished: {Title: "Ingestion finished", Level: statsd.Info, Action: EventActionFinish},

	DumpStarted:  {Title: "Dump started", Level: statsd.Info, Action: EventActionStart},
	DumpFinished: {Title: "Dump finished", Level: statsd.Info, Action: EventActionFinish},
	DumpFailed:   {Title: "Dump failed", Level: statsd.Error, Action: EventActionFail},
}

func (ea EventAction) Tags(ctx context.Context) []string {
	tags := tag.GetDefaultTags(ctx)
	tags = append(tags, fmt.Sprintf("%s:%s", tag.ActionTypeTag, map2msg[ea].Action))

	return tags
}

func (ea EventAction) Level() statsd.EventAlertType {
	return map2msg[ea].Level
}

func (ea EventAction) Title(ctx context.Context) string {
	title, _ := getTitleTextMsg(ctx, map2msg[ea].Title)

	return title
}

func (ea EventAction) DefaultMessage(ctx context.Context) string {
	_, msg := getTitleTextMsg(ctx, map2msg[ea].Title)

	return msg
}

func getTitleTextMsg(ctx context.Context, actionMsg string) (string, string) {
	cluster := log.GetClusterFromContext(ctx)
	runId := log.GetRunIDFromContext(ctx)
	title := fmt.Sprintf("%s for %s", actionMsg, cluster)
	text := fmt.Sprintf("%s for %s with run_id %s", actionMsg, cluster, runId)

	return title, text
}

func PushEvent(ctx context.Context, action EventAction, text string) error {
	if text == "" {
		text = action.DefaultMessage(ctx)
	}

	return kstatsd.Event(&statsd.Event{
		Title:     action.Title(ctx),
		Text:      text,
		Tags:      action.Tags(ctx),
		AlertType: action.Level(),
	})
}
