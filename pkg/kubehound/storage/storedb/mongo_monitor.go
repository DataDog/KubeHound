package storedb

import (
	"context"

	"go.mongodb.org/mongo-driver/event"
	mongotrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/go.mongodb.org/mongo-driver/mongo"
)

type SizedMonitor struct {
	mon *event.CommandMonitor
}

func NewSizedMonitor(opts ...mongotrace.Option) *event.CommandMonitor {
	m := &SizedMonitor{
		mon: mongotrace.NewMonitor(opts...),
	}

	return &event.CommandMonitor{
		Started:   m.Started,
		Succeeded: m.Succeeded,
		Failed:    m.Failed,
	}
}

func (m *SizedMonitor) Started(ctx context.Context, evt *event.CommandStartedEvent) {
	if len(evt.Command) > 1*1024 {
		// b, _ := bson.MarshalExtJSON(evt.Command, false, false)
		// d, err := json.Unmarshal(data, v)
		// // TODO marshal raw BSON and drop the documents piece
		// log.I.Debugf("%s", b)
		// evt.Command = bson.Raw{}
		m.mon.Started(ctx, evt)
	}

	m.mon.Started(ctx, evt)
}

func (m *SizedMonitor) Succeeded(ctx context.Context, evt *event.CommandSucceededEvent) {
	m.mon.Succeeded(ctx, evt)
}

func (m *SizedMonitor) Failed(ctx context.Context, evt *event.CommandFailedEvent) {
	m.mon.Failed(ctx, evt)
}
