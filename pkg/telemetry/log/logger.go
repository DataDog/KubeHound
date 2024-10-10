package log

import (
	"context"
	"sync/atomic"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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
	// TOOD: use the env var to setup the formatter (json / text / dd ...)

	cfg := &Config{
		logLevel:  LevelInfo,
		formatter: "text",
		useColour: true,
	}
	l := &traceLogger{
		logger: newLoggerWithSkip(cfg, 1),
		fields: []Field{},
	}
	globalDefault.Store(l)
}

func newLoggerWithSkip(cfg *Config, skip int) *zapLogger {
	// add 1 to skip: We wrap zap's functions with *zapLogger methods
	skip += 1

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
