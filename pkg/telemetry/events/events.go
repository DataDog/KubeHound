package events

import (
	"context"
	"fmt"
	"time"

	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	kstatsd "github.com/DataDog/KubeHound/pkg/telemetry/statsd"
	"github.com/DataDog/KubeHound/pkg/telemetry/tag"
	"github.com/DataDog/datadog-go/v5/statsd"
)

const (
	EventActionFail   = "fail"
	EventActionInit   = "init"
	EventActionStart  = "start"
	EventActionSkip   = "skip"
	EventActionFinish = "finish"
)

func pushEventInfo(title string, text string, tags []string) {
	_ = kstatsd.Event(&statsd.Event{
		Title:     title,
		Text:      text,
		Tags:      tags,
		AlertType: statsd.Info,
	})
}

func pushEventError(title string, text string, tags []string) {
	_ = kstatsd.Event(&statsd.Event{
		Title:     title,
		Text:      text,
		Tags:      tags,
		AlertType: statsd.Error,
	})
}

func getTitleTextMsg(ctx context.Context, actionMsg string) (string, string) {
	cluster := log.GetClusterFromContext(ctx)
	runId := log.GetRunIDFromContext(ctx)
	title := fmt.Sprintf("%s for %s", actionMsg, cluster)
	text := fmt.Sprintf("%s for %s with run_id %s", actionMsg, cluster, runId)

	return title, text
}

func PushEventIngestSkip(ctx context.Context) {
	tags := tag.GetDefaultTags(ctx)
	tags = append(tags, tag.ActionType(EventActionSkip))

	title, text := geTitleTextMsg(ctx, "Ingestion skipped")
	pushEventInfo(title, text, tags)
}

func PushEventIngestStarted(ctx context.Context) {
	tags := tag.GetDefaultTags(ctx)
	tags = append(tags, tag.ActionType(EventActionStart))

	title, text := geTitleTextMsg(ctx, "Ingestion started")
	pushEventInfo(title, text, tags)
}

func PushEventIngestFinished(ctx context.Context, start time.Time) {
	tags := tag.GetDefaultTags(ctx)
	tags = append(tags, tag.ActionType(EventActionFinish))

	title, _ := geTitleTextMsg(ctx, "Ingest finished")
	text := fmt.Sprintf("KubeHound ingestion has been completed in %s", time.Since(start))
	pushEventInfo(title, text, tags)
}

func PushEventDumpStarted(ctx context.Context) {
	tags := tag.GetDefaultTags(ctx)
	tags = append(tags, tag.ActionType(EventActionStart))

	title, text := geTitleTextMsg(ctx, "Dump started")
	pushEventInfo(title, text, tags)
}

func PushEventDumpFinished(ctx context.Context, start time.Time) {
	tags := tag.GetDefaultTags(ctx)
	tags = append(tags, tag.ActionType(EventActionFinish))

	title, _ := geTitleTextMsg(ctx, "Dump finished")

	text := fmt.Sprintf("KubeHound dump run has been completed in %s", time.Since(start))
	pushEventInfo(title, text, tags)
}

func PushEventIngestorInit(ctx context.Context) {
	tags := tag.GetDefaultTags(ctx)
	tags = append(tags, tag.ActionType(EventActionInit))

	msg := "Ingestor/grpc endpoint initiated"
	pushEventInfo(msg, msg, tags)
}

func PushEventIngestorFailed(ctx context.Context) {
	tags := tag.GetDefaultTags(ctx)
	tags = append(tags, tag.ActionType(EventActionFail))

	msg := "Ingestor/grpc endpoint init failed"
	pushEventError(msg, msg, tags)
}
