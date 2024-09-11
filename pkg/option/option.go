package option

import (
	"strings"

	"go.uber.org/zap/zapcore"
)

var (
	LogFilePath string

	// LogFileSizeMB is the maximum size in megabytes of the log file before it gets
	// rotated. It defaults to 100 megabytes.
	LogFileSizeMB = 6

	// Compress determines if the rotated log files should be compressed
	// using gzip. The default is not to perform compression.
	Compress bool

	// MaxBackups is the maximum number of old log files to retain.  The default
	// is to retain all old log files (though MaxAge may still cause them to get
	// deleted.)
	MaxBackups int

	// MaxAge is the maximum number of days to retain old log files based on the
	// timestamp encoded in their filename.  Note that a day is defined as 24
	// hours and may not exactly correspond to calendar days due to daylight
	// savings, leap seconds, etc. The default is not to remove old log files
	// based on age.
	MaxAge int

	LogLevel = "info"

	ConsoleRequired = true

	CallerSkip = 2
)

type (
	Level = zapcore.Level
)

var (
	DebugLevel = zapcore.DebugLevel
	InfoLevel  = zapcore.InfoLevel
	WarnLevel  = zapcore.WarnLevel
	ErrorLevel = zapcore.ErrorLevel
	PanicLevel = zapcore.PanicLevel
	FatalLevel = zapcore.FatalLevel
)

var LevelMapping = map[string]Level{
	DebugLevel.String(): DebugLevel,
	InfoLevel.String():  InfoLevel,
	WarnLevel.String():  WarnLevel,
	ErrorLevel.String(): ErrorLevel,
	PanicLevel.String(): PanicLevel,
	FatalLevel.String(): FatalLevel,
}

// Option is a functional option for configuring the logger.
type Option interface {
	Apply()
}

type logFileOption struct {
	LogFilePath   string
	LogFileSizeMB int
	Compress      bool
	MaxBackups    int
	MaxAge        int
}

func WithLogFilePath(logFilePath string) Option {
	return &logFileOption{
		LogFilePath: logFilePath,
	}
}

// WithLogFile configures the logger to write logs to a file using Lumberjack.
func WithLogFile(logFilePath string, logFileSizeMB, maxBackups, maxAge int, compress bool) Option {
	return &logFileOption{
		LogFilePath:   logFilePath,
		LogFileSizeMB: logFileSizeMB,
		Compress:      compress,
		MaxBackups:    maxBackups,
		MaxAge:        maxAge,
	}
}

func (o *logFileOption) Apply() {
	if o.LogFilePath != "" {
		LogFilePath = o.LogFilePath
		if o.LogFileSizeMB == 0 {
			LogFileSizeMB = 100
		} else {
			LogFileSizeMB = o.LogFileSizeMB
		}

		Compress = o.Compress
		MaxBackups = o.MaxBackups
		MaxAge = o.MaxAge
	}
}

type logLevelOption struct {
	LogLevel string
}

func WithLogLevel(level string) Option {
	return &logLevelOption{
		LogLevel: strings.ToLower(level),
	}
}

func (o *logLevelOption) Apply() {
	if o.LogLevel != "" {
		LogLevel = o.LogLevel
	}
}

type logConsoleOption struct {
	Required bool
}

func WithConsole(required bool) Option {
	return &logConsoleOption{
		Required: required,
	}
}

func (o *logConsoleOption) Apply() {
	ConsoleRequired = o.Required
}

// AddCallerSkip increases the number of callers skipped by caller annotation
// (as enabled by the AddCaller option). When building wrappers around the
// Logger and SugaredLogger, supplying this Option prevents zap from always
// reporting the wrapper code as the caller.
type logCallerSkipOption struct {
	CallerSkip int
}

func WithCallerSkip(callerSkip int) Option {
	return &logCallerSkipOption{
		CallerSkip: callerSkip,
	}
}

func (o *logCallerSkipOption) Apply() {
	CallerSkip = o.CallerSkip
}
