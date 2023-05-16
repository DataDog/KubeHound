package log

import (
	"context"
	"strconv"

	"github.com/DataDog/KubeHound/pkg/globals"
	logrus "github.com/sirupsen/logrus"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

const (
	DefaultComponent = "kubehound"
)

// Default logger instance for use through the app
var I *KubehoundLogger = Default()

// Require our logger to append job or API related fields for easier filtering and parsing
// of logs within custom dashboards. Sticking to the "structured" log types also enables
// out of the box correlation of APM traces and log messages without the need for a custom
// index pipeline. See: https://docs.datadoghq.com/logs/log_collection/go/#configure-your-logger
type KubehoundLogger struct {
	*logrus.Entry
}

// traceID retrieves the trace ID from the provided span.
func traceID(span tracer.Span) string {
	traceID := span.Context().TraceID()
	return strconv.FormatUint(traceID, 10)
}

// traceID retrieves the span ID from the provided span.
func spanID(span tracer.Span) string {
	spanID := span.Context().SpanID()
	return strconv.FormatUint(spanID, 10)
}

// Default returns the default logger for the application.
func Default() *KubehoundLogger {
	fields := logrus.Fields{
		globals.TagService:   globals.DDServiceName,
		globals.TagComponent: DefaultComponent,
		globals.TagTeam:      globals.DDTeamName,
	}

	logger := logrus.WithFields(fields)
	if globals.ProdEnv() {
		logger.Logger.SetFormatter(&logrus.JSONFormatter{})
	}

	return &KubehoundLogger{logger}
}

type LoggerOption func(*logrus.Entry) *logrus.Entry

func WithComponent(name string) LoggerOption {
	return func(l *logrus.Entry) *logrus.Entry {
		return l.WithField(globals.TagComponent, name)
	}
}

// Trace creates a logger from the current context, attaching trace and span IDs for use with APM.
func Trace(ctx context.Context, opts ...LoggerOption) *KubehoundLogger {
	baseLogger := Default()
	span, ok := tracer.SpanFromContext(ctx)
	if !ok {
		return baseLogger
	}

	logger := baseLogger.WithFields(logrus.Fields{
		"dd.span_id":  spanID(span),
		"dd.trace_id": traceID(span),
	})

	for _, o := range opts {
		logger = o(logger)
	}

	return &KubehoundLogger{logger}
}
