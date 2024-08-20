package otel

import "go.uber.org/zap/zapcore"

// config is used to configure the iris middleware.
type config struct {
	LogTraceId bool
	LogSpanId  bool
	LogSampled bool

	LogLevel         zapcore.Level
	ErrorStatusLevel zapcore.Level
	CallerDepth      int8
	CallerSkip       uint8
}

// Option specifies instrumentation configuration options.
type Option interface {
	apply(*config)
}

type optionFunc func(*config)

func (o optionFunc) apply(c *config) {
	o(c)
}

func WithLogTraceId(enabled bool) Option {
	return optionFunc(func(cfg *config) {
		cfg.LogTraceId = enabled
	})
}

func WithLogSpanId(enabled bool) Option {
	return optionFunc(func(cfg *config) {
		cfg.LogSpanId = enabled
	})
}

func WithLogSampled(enabled bool) Option {
	return optionFunc(func(cfg *config) {
		cfg.LogSampled = enabled
	})
}

func WithLogLevel(logLevel zapcore.Level) Option {
	return optionFunc(func(cfg *config) {
		cfg.LogLevel = logLevel
	})
}

func WithErrorStatusLevel(errorStatusLevel zapcore.Level) Option {
	return optionFunc(func(cfg *config) {
		cfg.ErrorStatusLevel = errorStatusLevel
	})
}

// WithCallerDepth for add caller info.
// 0. only recode Caller
// -1. not recode
// > 0.
func WithCallerDepth(depth int) Option {
	return optionFunc(func(cfg *config) {
		cfg.CallerDepth = int8(depth)
	})
}

func WithCallerSkip(skip int) Option {
	if skip > 0 {
		return optionFunc(func(cfg *config) {
			cfg.CallerSkip = uint8(skip)
		})
	}
	return optionFunc(func(c *config) {})
}
