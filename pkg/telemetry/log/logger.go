package log

import (
	"context"
	"os"
	"strconv"
	"sync"

	"github.com/DataDog/KubeHound/pkg/globals"
	logrus "github.com/sirupsen/logrus"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type LoggerOption func(*logrus.Entry) *logrus.Entry

type LoggerConfig struct {
	Tags logrus.Fields // Tags applied to all logs.
	Mu   *sync.Mutex   // Lock to enable safe runtime changes.
	DD   bool          // Whether Datadog integration is enabled.
}

var globalConfig = LoggerConfig{
	Tags: logrus.Fields{
		globals.TagService:   globals.DDServiceName,
		globals.TagComponent: globals.DefaultComponent,
	},
	Mu: &sync.Mutex{},
	DD: true,
}

// I Global logger instance for use through the app
var I = Base()

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

// Base returns the base logger for the application.
func Base() *KubehoundLogger {
	logger := logrus.WithFields(globalConfig.Tags)
	logger.Logger.SetFormatter(GetLogrusFormatter())

	return &KubehoundLogger{logger}
}

// SetDD enables/disabled Datadog integration in the logger.
func SetDD(enabled bool) {
	globalConfig.Mu.Lock()
	defer globalConfig.Mu.Unlock()

	globalConfig.DD = enabled

	// Replace the current logger instance to reflect changes
	I = Base()
}

// AddGlobalTags adds global tags to all application loggers.
func AddGlobalTags(tags map[string]string) {
	globalConfig.Mu.Lock()
	defer globalConfig.Mu.Unlock()

	for tk, tv := range tags {
		globalConfig.Tags[tk] = tv
	}

	// Replace the current logger instance to reflect changes
	I = Base()
}

// WithComponent adds a component name tag to the logger.
func WithComponent(name string) LoggerOption {
	return func(l *logrus.Entry) *logrus.Entry {
		return l.WithField(globals.TagComponent, name)
	}
}

// WithCollectedCluster adds a component name tag to the logger.
func WithCollectedCluster(name string) LoggerOption {
	return func(l *logrus.Entry) *logrus.Entry {
		return l.WithField(globals.CollectedClusterComponent, name)
	}
}

// Trace creates a logger from the current context, attaching trace and span IDs for use with APM.
func Trace(ctx context.Context, opts ...LoggerOption) *KubehoundLogger {
	baseLogger := Base()

	span, ok := tracer.SpanFromContext(ctx)
	if !ok {
		return baseLogger
	}

	if !globalConfig.DD {
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

func GetLogrusFormatter() logrus.Formatter {
	customTextFormatter := NewFilteredTextFormatter(DefaultRemovedFields)

	switch logFormat := os.Getenv("KH_LOG_FORMAT"); {
	// Datadog require the logged field to be "message" and not "msg"
	case logFormat == "dd":
		formatter := &logrus.JSONFormatter{
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyMsg: "message",
			},
		}

		return formatter
	case logFormat == "json":
		return &logrus.JSONFormatter{}
	case logFormat == "text":
		return customTextFormatter
	default:
		return customTextFormatter
	}
}
