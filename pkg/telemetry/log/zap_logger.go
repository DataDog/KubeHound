package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type zapLogger struct {
	l *zap.Logger
	s *zap.SugaredLogger
}

func (z *zapLogger) With(fields ...Field) LoggerI {
	childLogger := z.l.With(fields...)

	return &zapLogger{
		l: childLogger,
		s: childLogger.Sugar(),
	}
}

// newZapConfig creates a zap.Config
func newZapConfig(cfg *Config) zap.Config {
	var zc zap.Config

	zc = zap.NewProductionConfig()

	// We want log.Duration("duration", ...) to naturally map to Datadog's 'duration' standard attribute.
	// Datadog displays it nicely and uses it as a default measure for trace search.
	// See https://docs.datadoghq.com/logs/log_configuration/attributes_naming_convention/#performance
	// The spec requires that it be encoded in nanoseconds (default is seconds).
	zc.EncoderConfig.EncodeDuration = zapcore.NanosDurationEncoder

	// never use color with JSON output: the JSON encoder escapes it
	if cfg.useColour && cfg.formatter != "json" {
		zc.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		zc.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	}
	zc.Level.SetLevel(cfg.logLevel.zapLevel())
	zc.Encoding = cfg.formatter

	// we don't want zap stacktraces because they are incredibly noisy
	zc.DisableStacktrace = true

	return zc
}

func (z *zapLogger) Debug(msg string, fields ...Field) {
	z.l.Debug(msg, fields...)
}

func (z *zapLogger) Info(msg string, fields ...Field) {
	z.l.Info(msg, fields...)
}

func (z *zapLogger) Warn(msg string, fields ...Field) {
	z.l.Warn(msg, fields...)
}

func (z *zapLogger) Error(msg string, fields ...Field) {
	z.l.Error(msg, fields...)
}

func (z *zapLogger) Panic(msg string, fields ...Field) {
	z.l.Panic(msg, fields...)
}

func (z *zapLogger) Fatal(msg string, fields ...Field) {
	z.l.Fatal(msg, fields...)
}

func (z *zapLogger) Debugf(msg string, params ...interface{}) {
	z.s.Debugf(msg, params...)
}

func (z *zapLogger) Infof(msg string, params ...interface{}) {
	z.s.Infof(msg, params...)
}

func (z *zapLogger) Warnf(msg string, params ...interface{}) {
	z.s.Warnf(msg, params...)
}

func (z *zapLogger) Errorf(msg string, params ...interface{}) {
	z.s.Errorf(msg, params...)
}

func (z *zapLogger) Panicf(msg string, params ...interface{}) {
	z.s.Panicf(msg, params...)
}

func (z *zapLogger) Fatalf(msg string, params ...interface{}) {
	z.s.Fatalf(msg, params...)
}
