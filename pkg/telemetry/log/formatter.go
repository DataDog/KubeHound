package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func newTextFormatterConfig(cfg *Config) zap.Config {
	var zc zap.Config

	zc = zap.NewProductionConfig()

	if cfg.useColour {
		zc.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	zc.EncoderConfig.CallerKey = ""
	zc.EncoderConfig.FunctionKey = ""
	zc.EncoderConfig.ConsoleSeparator = " "
	zc.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05")
	zc.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	zc.Level.SetLevel(cfg.logLevel.zapLevel())
	zc.Encoding = cfg.formatter

	// we don't want zap stacktraces because they are incredibly noisy
	zc.DisableStacktrace = true
	return zc
}

func newJSONFormatterConfig(cfg *Config) zap.Config {
	var zc zap.Config
	zc = zap.NewProductionConfig()
	// We want log.Duration("duration", ...) to naturally map to Datadog's 'duration' standard attribute.
	// Datadog displays it nicely and uses it as a default measure for trace search.
	// See https://docs.datadoghq.com/logs/log_configuration/attributes_naming_convention/#performance
	// The spec requires that it be encoded in nanoseconds (default is seconds).
	zc.EncoderConfig.EncodeDuration = zapcore.NanosDurationEncoder
	zc.Level.SetLevel(cfg.logLevel.zapLevel())
	zc.Encoding = cfg.formatter
	// we don't want zap stacktraces because they are incredibly noisy
	zc.DisableStacktrace = true
	return zc
}
