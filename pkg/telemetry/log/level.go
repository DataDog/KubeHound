package log

import (
	"errors"
	"strings"

	"go.uber.org/zap/zapcore"
)

// Level of log emission.  A logger at any level will ignore all levels
// below it in value.
type Level byte

// Logging levels
const (
	LevelTrace Level = iota
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
	LevelPanic
	LevelFatal
)

// DefaultLevel is the logging level if nothing is configured.
const DefaultLevel = LevelInfo

func (lvl Level) String() string {
	switch lvl {
	case LevelTrace:
		return "trace"
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelWarn:
		return "warn"
	case LevelError:
		return "error"
	case LevelPanic:
		return "panic"
	case LevelFatal:
		return "fatal"
	default:
		return ""
	}
}

// MarshalText implements encoding.TextMarshaler for Level.
func (lvl Level) MarshalText() ([]byte, error) {
	return []byte(lvl.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler for Level.
func (lvl *Level) UnmarshalText(text []byte) error {
	l, err := LevelFromString(string(text))
	if err != nil {
		return err
	}
	*lvl = l

	return nil
}

// LevelFromString returns the level for the given string.  If the string
// is not valid, an error is returned.
func LevelFromString(str string) (Level, error) {
	s := strings.ToLower(str)
	switch s {
	case "trace":
		return LevelTrace, nil
	case "debug":
		return LevelDebug, nil
	case "info":
		return LevelInfo, nil
	case "warn":
		return LevelWarn, nil
	case "error":
		return LevelError, nil
	case "panic":
		return LevelPanic, nil
	case "fatal":
		return LevelFatal, nil
	}

	return 0, errors.New("invalid log level")
}

func (lvl Level) zapLevel() zapcore.Level {
	switch lvl {
	case LevelTrace:
		return zapcore.DebugLevel
	case LevelDebug:
		return zapcore.DebugLevel
	case LevelInfo:
		return zapcore.InfoLevel
	case LevelWarn:
		return zapcore.WarnLevel
	case LevelError:
		return zapcore.ErrorLevel
	case LevelPanic:
		return zapcore.PanicLevel
	case LevelFatal:
		return zapcore.FatalLevel
	}

	// default to InfoLevel if we have something weird
	return zapcore.InfoLevel
}
