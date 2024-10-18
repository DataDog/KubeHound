package log

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"go.uber.org/zap"
	ddtrace "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// Trace returns a wrapped default logger that automatically adds trace
// related ids to log output.
func Trace(ctx context.Context) LoggerI {
	return TraceLogger(ctx, DefaultLogger())
}

// ddTraceTraceID creates a field with a string traceID with the common key "dd.trace_id".
// To keep precision in JSON we have to turn the very large ids into strings
func ddTraceTraceID(span ddtrace.Span) Field {
	traceID := span.Context().TraceID()
	traceIDStr := strconv.FormatUint(traceID, 10)
	return zap.String(traceIDKey, traceIDStr)
}

// ddTraceSpanID creates a field with a string spanID with the common key "dd.span_id".
// To keep precision in JSON we have to turn the very large ids into strings
func ddTraceSpanID(span ddtrace.Span) Field {
	spanID := span.Context().SpanID()
	spanIDStr := strconv.FormatUint(spanID, 10)
	return zap.String(spanIDKey, spanIDStr)
}

// traceLogger is a logger type that automatically adds trace and span ids
// to its output.  It implements `Level` and `LevelF` logging variants, using
// fields for `Level` funcs and pre-pending `k=v` output for trace & span
// ID in `LevelF` funcs.
type traceLogger struct {
	logger LoggerI

	// We assume fields to be immutable after trace logger is created,
	// hence they are directly referenced (and not copied) when deriving
	// a child logger with With.
	fields []Field
}

// TraceLogger returns a wrapped version of logger that automatically adds
// trace related ids to log output. If Logger was not created by this package, the caller
// line number information will be incorrect.
func TraceLogger(ctx context.Context, logger LoggerI) LoggerI {
	var fields []zap.Field
	span, ok := ddtrace.SpanFromContext(ctx)
	if !ok {
		return logger
	}

	fields = []zap.Field{ddTraceSpanID(span), ddTraceTraceID(span)}

	// Adding by default the runID and cluster to the logs
	runID := convertField(ctx.Value(ContextFieldRunID))
	if runID != "" {
		fields = append(fields, String(FieldRunIDKey, runID))
	}
	cluster := convertField(ctx.Value(ContextFieldCluster))
	if cluster != "" {
		fields = append(fields, String(FieldClusterKey, cluster))
	}
	component := convertField(ctx.Value(ContextFieldComponent))
	if component != "" {
		fields = append(fields, String(FieldComponentKey, component))
	}

	// no span: return the logger with no modifications
	if len(fields) == 0 {
		return logger
	}

	l := &traceLogger{
		logger: logger,
		fields: fields,
	}
	return l
}

func (t *traceLogger) appendTracingFields(msg string) string {
	var b strings.Builder
	b.Grow(len(msg))
	for i := range t.fields {
		b.WriteString(fmt.Sprintf("%s=%s ",
			t.fields[i].Key,
			t.fields[i].String))
	}
	b.WriteString(msg)
	return b.String()
}

func (t *traceLogger) With(fields ...Field) LoggerI {
	return &traceLogger{
		logger: t.logger.With(fields...),
		fields: t.fields,
	}
}

func (t *traceLogger) Debug(msg string, fields ...Field) {
	fields = append(fields, t.fields...)
	t.logger.Debug(msg, fields...)
}

func (t *traceLogger) Info(msg string, fields ...Field) {
	fields = append(fields, t.fields...)
	t.logger.Info(msg, fields...)
}

func (t *traceLogger) Warn(msg string, fields ...Field) {
	fields = append(fields, t.fields...)
	t.logger.Warn(msg, fields...)
}

func (t *traceLogger) Error(msg string, fields ...Field) {
	fields = append(fields, t.fields...)
	t.logger.Error(msg, fields...)
}

func (t *traceLogger) Panic(msg string, fields ...Field) {
	fields = append(fields, t.fields...)
	t.logger.Panic(msg, fields...)
}

func (t *traceLogger) Fatal(msg string, fields ...Field) {
	fields = append(fields, t.fields...)
	t.logger.Fatal(msg, fields...)
}

func (t *traceLogger) Debugf(msg string, params ...interface{}) {
	t.logger.With(t.fields...).Debugf(msg, params...)
}

func (t *traceLogger) Infof(msg string, params ...interface{}) {
	t.logger.With(t.fields...).Infof(msg, params...)
}

func (t *traceLogger) Warnf(msg string, params ...interface{}) {
	t.logger.With(t.fields...).Warnf(msg, params...)
}

func (t *traceLogger) Errorf(msg string, params ...interface{}) {
	t.logger.With(t.fields...).Errorf(msg, params...)
}

func (t *traceLogger) Panicf(msg string, params ...interface{}) {
	t.logger.With(t.fields...).Panicf(msg, params...)
}

func (t *traceLogger) Fatalf(msg string, params ...interface{}) {
	t.logger.With(t.fields...).Fatalf(msg, params...)
}
