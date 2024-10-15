package log

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// envOverride applies the
//   - LOG_LEVEL environment variable to this config.
//
// For visibility, if the variable is present but invalid, a warning is
// printed directly to stderr.
func newDefaultZapConfig() zap.Config {
	var err error
	level := DefaultLevel
	if env := os.Getenv("LOG_LEVEL"); len(env) > 0 {
		level, err = LevelFromString(env)
		if err != nil {
			fmt.Fprintf(os.Stderr, "WARN: invalid LOG_LEVEL setting %s\n", env)
		}
	}

	// NOTE: Avoid using common names like `stack_trace` to avoid field remapping by the Logs app
	// that might mask error message
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.StacktraceKey = "zap_stack_trace"

	return zap.Config{
		Level:            zap.NewAtomicLevelAt(level.zapLevel()),
		Development:      false,
		Sampling:         nil,
		EncoderConfig:    encoderConfig,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		// we don't want zap stacktraces because they are incredibly noisy
		DisableStacktrace: true,
	}
}

func newTextFormatterConfig() zap.Config {
	zc := newDefaultZapConfig()

	zc.Encoding = "text"
	zc.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	zc.EncoderConfig.CallerKey = ""
	zc.EncoderConfig.FunctionKey = ""
	zc.EncoderConfig.ConsoleSeparator = " "
	zc.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("15:04:05")
	zc.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder

	return zc
}

func newJSONFormatterConfig() zap.Config {
	// var zc zap.Config
	// zc = zap.NewProductionConfig()
	zc := newDefaultZapConfig()

	zc.Encoding = "json"
	// We want log.Duration("duration", ...) to naturally map to Datadog's 'duration' standard attribute.
	// Datadog displays it nicely and uses it as a default measure for trace search.
	// See https://docs.datadoghq.com/logs/log_configuration/attributes_naming_convention/#performance
	// The spec requires that it be encoded in nanoseconds (default is seconds).
	zc.EncoderConfig.EncodeDuration = zapcore.NanosDurationEncoder

	return zc
}

func legacyTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02T15:04:05.000000-07:00"))
}

func newZapConfig() zap.Config {
	// By default, we use text formatter
	zc := newTextFormatterConfig()

	switch logFormat := os.Getenv("KH_LOG_FORMAT"); {
	// Datadog require the logged field to be "message" and not "msg"
	case logFormat == logFormatDD:
		zc = newJSONFormatterConfig()
		zc.EncoderConfig.MessageKey = "message"
		zc.EncoderConfig.EncodeTime = legacyTimeEncoder
	case logFormat == logFormatJSON:
		zc = newJSONFormatterConfig()
	}

	zc.InitialFields = map[string]interface{}{
		FieldServiceKey: "kubehound",
	}

	return zc
}
