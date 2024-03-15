package easylog

import (
	"context"

	"os"
	"time"

	"github.com/logerror/easylog/pkg/izap"
	"github.com/logerror/easylog/pkg/option"
	otelzap "github.com/logerror/easylog/pkg/otel"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	globalLogger        Logger
	globalRawLogger     *logger
	globalSugaredLogger SugaredLogger

	globalLoggerLevel zap.AtomicLevel

	globalOtelLogger        izap.Logger
	globalOtelSugaredLogger izap.SugaredLogger
)

type (
	// Field is an alias of zap.Field. Aliasing this type dramatically
	// improves the navigability of this package's API documentation.
	Field = zap.Field
)

type SugaredLogger interface {
	Named(name string) SugaredLogger
	With(args ...interface{}) SugaredLogger

	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Panic(args ...interface{})
	Fatal(args ...interface{})

	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Panicf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})

	Sync()
}

// Logger defines methods of writing log
type Logger interface {
	Named(s string) Logger
	With(fields ...Field) Logger
	WithContext(ctx context.Context) izap.StdLogger

	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)

	Clone() Logger
	Level() string
	IsDebug() bool
	Sync()

	SugaredLogger() SugaredLogger
	CoreLogger() *zap.Logger
}

type logger struct {
	level       string
	atomicLevel zap.AtomicLevel

	logger            *zap.Logger
	sugaredLogger     *zap.SugaredLogger
	otelLogger        izap.Logger
	otelSugaredLogger izap.SugaredLogger
}

type sugaredLogger struct {
	sugaredLogger *zap.SugaredLogger
}

func InitLogger(options ...option.Option) Logger {
	return initLogger(options...)
}

func InitGlobalLogger(options ...option.Option) Logger {
	globalRawLogger = initLogger(options...)
	globalLogger = globalRawLogger
	globalSugaredLogger = globalLogger.SugaredLogger()
	globalLoggerLevel = globalRawLogger.atomicLevel
	globalOtelLogger = globalRawLogger.otelLogger
	globalOtelSugaredLogger = globalRawLogger.otelSugaredLogger
	zap.ReplaceGlobals(globalLogger.CoreLogger())
	return globalRawLogger
}

func initLogger(options ...option.Option) *logger {
	l := &logger{}

	encoder := zapcore.EncoderConfig{
		TimeKey:       "time",
		LevelKey:      "level",
		NameKey:       "name",
		CallerKey:     "caller",
		MessageKey:    "msg",
		StacktraceKey: "stacktrace",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.LowercaseLevelEncoder,
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			encodeTimeLayout(t, "2006-01-02 15:04:05.000", enc)
		},
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Apply additional options
	for _, o := range options {
		o.Apply()
	}

	consoleSyncer := zapcore.AddSync(os.Stdout)
	multiWriteSyncer := zapcore.NewMultiWriteSyncer(consoleSyncer)
	if option.LogFilePath != "" && option.LogFileSizeMB != 0 {
		lumberjackLogger := &lumberjack.Logger{
			Filename:   option.LogFilePath,
			MaxSize:    option.LogFileSizeMB, // MaxSize in megabytes
			MaxBackups: option.MaxBackups,    // Max number of old log files to retain
			MaxAge:     option.MaxAge,        // Max number of days to retain old log files
			Compress:   option.Compress,      // Whether to compress the old log files
		}

		fileSyncer := zapcore.AddSync(lumberjackLogger)
		if option.ConsoleRequired {
			multiWriteSyncer = zapcore.NewMultiWriteSyncer(consoleSyncer, fileSyncer)
		} else {
			multiWriteSyncer = zapcore.NewMultiWriteSyncer(fileSyncer)
		}
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoder),
		multiWriteSyncer,
		ParseLevel(option.LogLevel),
	)

	l.logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1), zap.AddStacktrace(zapcore.ErrorLevel))
	l.sugaredLogger = l.logger.Sugar()
	l.otelLogger = otelzap.NewLogger(l.logger)
	l.otelSugaredLogger = otelzap.NewSugaredLogger(l.sugaredLogger)

	return l
}

func ParseLevel(level string) option.Level {
	lvl, ok := option.LevelMapping[level]
	if ok {
		return lvl
	}
	// default level: info
	return option.InfoLevel
}

func encodeTimeLayout(t time.Time, layout string, enc zapcore.PrimitiveArrayEncoder) {
	type appendTimeEncoder interface {
		AppendTimeLayout(time.Time, string)
	}

	if enc, ok := enc.(appendTimeEncoder); ok {
		enc.AppendTimeLayout(t, layout)
		return
	}

	enc.AppendString(t.Format(layout))
}

func init() {
	globalRawLogger = initLogger()
	globalLogger = globalRawLogger
	globalSugaredLogger = globalLogger.SugaredLogger()
	globalLoggerLevel = globalRawLogger.atomicLevel
	globalOtelLogger = globalRawLogger.otelLogger
	globalOtelSugaredLogger = globalRawLogger.otelSugaredLogger
	//zap.ReplaceGlobals(globalLogger.CoreLogger())
}
