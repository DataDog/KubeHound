package log

import (
	"context"
	"fmt"
	"sync/atomic"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	ddtrace "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// globalDefault contains the current global default logger and its configuration.
// It is set in ConfigureDefaultLogger, which is called in a package init() function.
// It will never be nil.
//
// When using this from package-level functions, we must use the zap loggers directly.
// This avoids adding an extra level of functions to the stack. This means we can use the same
// loggers, without changing the AddCallerSkip() configuration.
var globalDefault atomic.Pointer[traceLogger]

type LoggerI interface {
	// With returns a child logger structured with the provided fields.
	// Fields added to the child don't affect the parent, and vice versa.
	With(fields ...Field) LoggerI

	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Panic(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)

	Debugf(msg string, params ...interface{})
	Infof(msg string, params ...interface{})
	Warnf(msg string, params ...interface{})
	Errorf(msg string, params ...interface{})
	Panicf(msg string, params ...interface{})
	Fatalf(msg string, params ...interface{})
}

// Config options for logging.
type Config struct {
	logLevel  Level
	formatter string
	useColour bool

	// serializes a caller in /full/path/to/package/file:line format
	// instead of just the package/file:line format
	fullCallerPath bool

	disableCaller bool

	// stdout is a special case for the logger to output to stdout
	stdout bool
}

type KubehoundLogger struct {
	LoggerI
}

func Logger(ctx context.Context) LoggerI {
	logger := Trace(ctx)
	return &KubehoundLogger{
		LoggerI: logger,
	}
}

const (
	FieldK8sType          = "k8s_type"
	FieldCount            = "count"
	FieldNodeType         = "node_type"
	FieldVertexType       = "vertex_type"
	FieldCluster          = "cluster"
	FieldComponent        = "component"
	FieldRunID            = "run_id"
	FieldTeam             = "team"
	FieldService          = "service"
	FieldIngestorPipeline = "ingestor_pipeline"
	FieldDumpPipeline     = "dump_pipeline"
)

type contextKey int

const (
	ContextFieldRunID contextKey = iota
	ContextFieldCluster
)

// func TraceLogger(ctx context.Context) *CSALogger {
// 	logger := log.Trace(ctx)

// 	// Context based fields
// 	runID := convertTag(ctx.Value(ContextLogFieldRepo))
// 	if runID != "" {
// 		logger = logger.With(log.String(LogFieldRepo, runID))
// 	}

// 	return &CSALogger{Logger: logger}
// }

func convertTag(value any) string {
	val, err := value.(string)
	if !err {
		return ""
	}
	return val
}

func SpanSetDefaultTag(ctx context.Context, span ddtrace.Span) {
	runID := convertTag(ctx.Value(ContextFieldRunID))
	if runID != "" {
		span.SetTag(FieldRunID, convertTag(runID))
	}

	cluster := convertTag(ctx.Value(ContextFieldCluster))
	if cluster != "" {
		span.SetTag(FieldRunID, convertTag(cluster))
	}
}

func TagK8sType(k8sType string) string {
	return fmt.Sprintf("%s:%s", FieldK8sType, k8sType)
}

func TagCount(count int) string {
	return fmt.Sprintf("%s:%d", FieldCount, count)
}

func TagNodeType(nodeType string) string {
	return fmt.Sprintf("%s:%s", FieldNodeType, nodeType)
}

func TagVertexType(vertexType string) string {
	return fmt.Sprintf("%s:%s", FieldVertexType, vertexType)
}

func TagCluster(cluster string) string {
	return fmt.Sprintf("%s:%s", FieldCluster, cluster)
}

func TagComponent(component string) string {
	return fmt.Sprintf("%s:%s", FieldComponent, component)
}

func TagRunID(runID string) string {
	return fmt.Sprintf("%s:%s", FieldRunID, runID)
}

func TagTeam(team string) string {
	return fmt.Sprintf("%s:%s", FieldTeam, team)
}

func TagService(service string) string {
	return fmt.Sprintf("%s:%s", FieldService, service)
}

func TagIngestorPipeline(ingestorPipeline string) string {
	return fmt.Sprintf("%s:%s", FieldIngestorPipeline, ingestorPipeline)
}

func TagDumpPipeline(dumpPipeline string) string {
	return fmt.Sprintf("%s:%s", FieldDumpPipeline, dumpPipeline)
}

const (
	spanIDKey  = "dd.span_id"
	traceIDKey = "dd.trace_id"
)

// DefaultLogger returns the global logger
func DefaultLogger() LoggerI {
	return globalDefault.Load()
}

func NewTextEncoder(cfg zapcore.EncoderConfig) (zapcore.Encoder, error) {
	return zapcore.NewConsoleEncoder(cfg), nil
}

func init() {
	err := zap.RegisterEncoder("text", NewTextEncoder)
	if err != nil {
		panic(err)
	}

	cfg := &Config{
		logLevel:  LevelInfo,
		formatter: "text",
	}
	l := &traceLogger{
		logger: newLoggerWithSkip(cfg, 1),
		fields: []Field{},
	}
	globalDefault.Store(l)
}

func newLoggerWithSkip(cfg *Config, skip int) *zapLogger {
	// add 1 to skip: We wrap zap's functions with *zapLogger methods
	// skip += zapLoggerExtraCallerSkip

	zc := newZapConfig(cfg)
	zOptions := []zap.Option{
		zap.AddCallerSkip(skip),
		zap.AddStacktrace(zap.DPanicLevel),
	}
	// NOTE: Avoid using common names like `stack_trace` to avoid field remapping by the Logs app
	// that might mask error message
	zc.EncoderConfig.StacktraceKey = "zap_stack_trace"
	logger, err := zc.Build(zOptions...)

	// XXX: fall back to a basic printf-based logger?
	if err != nil {
		panic(err)
	}

	return &zapLogger{
		l: logger,
		s: logger.Sugar(),
	}
}

// type LoggerOption func(*logrus.Entry) *logrus.Entry

// type LoggerConfig struct {
// 	Tags logrus.Fields // Tags applied to all logs.
// 	Mu   *sync.Mutex   // Lock to enable safe runtime changes.
// 	DD   bool          // Whether Datadog integration is enabled.
// }

// var globalConfig = LoggerConfig{
// 	Tags: logrus.Fields{
// 		globals.TagService:   globals.DDServiceName,
// 		globals.TagComponent: globals.DefaultComponent,
// 	},
// 	Mu: &sync.Mutex{},
// 	DD: true,
// }

// I Global logger instance for use through the app
// var I = Base()

// Require our logger to append job or API related fields for easier filtering and parsing
// of logs within custom dashboards. Sticking to the "structured" log types also enables
// out of the box correlation of APM traces and log messages without the need for a custom
// index pipeline. See: https://docs.datadoghq.com/logs/log_collection/go/#configure-your-logger
// type KubehoundLogger struct {
// 	*logrus.Entry
// }

// // traceID retrieves the trace ID from the provided span.
// func traceID(span tracer.Span) string {
// 	traceID := span.Context().TraceID()

// 	return strconv.FormatUint(traceID, 10)
// }

// // traceID retrieves the span ID from the provided span.
// func spanID(span tracer.Span) string {
// 	spanID := span.Context().SpanID()

// 	return strconv.FormatUint(spanID, 10)
// }

// // Base returns the base logger for the application.
// func Base() *KubehoundLogger {
// 	logger := logrus.WithFields(globalConfig.Tags)
// 	logger.Logger.SetFormatter(GetLogrusFormatter())

// 	return &KubehoundLogger{logger}
// }

// // SetDD enables/disabled Datadog integration in the logger.
// func SetDD(enabled bool) {
// 	globalConfig.Mu.Lock()
// 	defer globalConfig.Mu.Unlock()

// 	globalConfig.DD = enabled

// 	// Replace the current logger instance to reflect changes
// 	I = Base()
// }

// // AddGlobalTags adds global tags to all application loggers.
// func AddGlobalTags(tags map[string]string) {
// 	globalConfig.Mu.Lock()
// 	defer globalConfig.Mu.Unlock()

// 	for tk, tv := range tags {
// 		globalConfig.Tags[tk] = tv
// 	}

// 	// Replace the current logger instance to reflect changes
// 	I = Base()
// }

// // WithComponent adds a component name tag to the logger.
// func WithComponent(name string) LoggerOption {
// 	return func(l *logrus.Entry) *logrus.Entry {
// 		return l.WithField(globals.TagComponent, name)
// 	}
// }

// // WithCollectedCluster adds a component name tag to the logger.
// func WithCollectedCluster(name string) LoggerOption {
// 	return func(l *logrus.Entry) *logrus.Entry {
// 		return l.WithField(globals.CollectedClusterComponent, name)
// 	}
// }

// // WithRunID adds a component name tag to the logger.
// func WithRunID(runid string) LoggerOption {
// 	return func(l *logrus.Entry) *logrus.Entry {
// 		return l.WithField(globals.RunID, runid)
// 	}
// }

// // Trace creates a logger from the current context, attaching trace and span IDs for use with APM.
// func Trace(ctx context.Context, opts ...LoggerOption) *KubehoundLogger {
// 	baseLogger := Base()

// 	span, ok := tracer.SpanFromContext(ctx)
// 	if !ok {
// 		return baseLogger
// 	}

// 	if !globalConfig.DD {
// 		return baseLogger
// 	}

// 	logger := baseLogger.WithFields(logrus.Fields{
// 		"dd.span_id":  spanID(span),
// 		"dd.trace_id": traceID(span),
// 	})

// 	for _, o := range opts {
// 		logger = o(logger)
// 	}

// 	return &KubehoundLogger{logger}
// }

// func GetLogrusFormatter() logrus.Formatter {
// 	customTextFormatter := NewFilteredTextFormatter(DefaultRemovedFields)

// 	switch logFormat := os.Getenv("KH_LOG_FORMAT"); {
// 	// Datadog require the logged field to be "message" and not "msg"
// 	case logFormat == "dd":
// 		formatter := &logrus.JSONFormatter{
// 			FieldMap: logrus.FieldMap{
// 				logrus.FieldKeyMsg: "message",
// 			},
// 		}

// 		return formatter
// 	case logFormat == "json":
// 		return &logrus.JSONFormatter{}
// 	case logFormat == "text":
// 		return customTextFormatter
// 	default:
// 		return customTextFormatter
// 	}
// }
