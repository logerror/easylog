package izap

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type StdLogger interface {
	Log(lvl zapcore.Level, msg string, fields ...zap.Field)
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Panic(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)

	DPanic(msg string, fields ...zap.Field)
}

type Logger interface {
	StdLogger
	WithContext(ctx context.Context) StdLogger
	With(fields ...zap.Field) Logger
	WithOptions(opts ...zap.Option) Logger
	Sugar() SugaredLogger
}

type StdSugaredLogger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	DPanic(args ...interface{})
	Panic(args ...interface{})
	Fatal(args ...interface{})

	Debugf(template string, args ...interface{})
	Infof(template string, args ...interface{})
	Warnf(template string, args ...interface{})
	Errorf(template string, args ...interface{})
	DPanicf(template string, args ...interface{})
	Panicf(template string, args ...interface{})
	Fatalf(template string, args ...interface{})

	Debugw(msg string, keysAndValues ...interface{})
	Infow(msg string, keysAndValues ...interface{})
	Warnw(msg string, keysAndValues ...interface{})
	Errorw(msg string, keysAndValues ...interface{})
	DPanicw(msg string, keysAndValues ...interface{})
	Panicw(msg string, keysAndValues ...interface{})
	Fatalw(msg string, keysAndValues ...interface{})

	Debugln(args ...interface{})
	Infoln(args ...interface{})
	Warnln(args ...interface{})
	Errorln(args ...interface{})
	DPanicln(args ...interface{})
	Panicln(args ...interface{})
	Fatalln(args ...interface{})
}

type SugaredLogger interface {
	StdSugaredLogger
	WithContext(ctx context.Context) StdSugaredLogger
	With(args ...interface{}) SugaredLogger
	WithOptions(opts ...zap.Option) SugaredLogger
	Desugar() Logger
}
