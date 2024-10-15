package log

import (
	"go.uber.org/zap"
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
