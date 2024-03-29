package easylog

import (
	"context"

	"github.com/logerror/easylog/pkg/izap"
	"github.com/logerror/easylog/pkg/option"
	otelzap "github.com/logerror/easylog/pkg/otel"
	"go.uber.org/zap"
)

func DefaultLogger() Logger {
	return globalLogger
}

func DefaultSugaredLogger() SugaredLogger {
	return globalSugaredLogger
}

func DefaultOtelLogger() izap.Logger {
	return globalOtelLogger
}

func DefaultOtelSugaredLogger() izap.SugaredLogger {
	return globalOtelSugaredLogger
}

func SetLevel(lvl option.Level) {
	globalLoggerLevel.SetLevel(lvl)
}

func SetDebug() {
	SetLevel(option.DebugLevel)
}

func Named(s string) Logger {
	return globalLogger.Named(s)
}

func (l *logger) Named(s string) Logger {
	lg := l.logger.Named(s)
	return &logger{
		level:             l.level,
		logger:            lg,
		sugaredLogger:     lg.Sugar(),
		otelLogger:        otelzap.NewLogger(lg),
		otelSugaredLogger: otelzap.NewSugaredLogger(lg.Sugar()),
	}
}

func With(fields ...Field) Logger {
	return globalLogger.With(fields...)
}

func (l *logger) With(fields ...Field) Logger {
	lg := l.logger.With(fields...)
	return &logger{
		level:             l.level,
		logger:            lg,
		sugaredLogger:     lg.Sugar(),
		otelLogger:        otelzap.NewLogger(lg),
		otelSugaredLogger: otelzap.NewSugaredLogger(lg.Sugar()),
	}
}

func N(ctx context.Context, name string) izap.StdLogger {
	l := globalRawLogger.logger.Named(name)
	return otelzap.NewLogger(l).WithContext(ctx)
}

func G(ctx context.Context) izap.StdLogger {
	return WithContext(ctx)
}

func GS(ctx context.Context) izap.StdSugaredLogger {
	return globalOtelSugaredLogger.WithContext(ctx)
}
func WithContext(ctx context.Context) izap.StdLogger {
	return globalOtelLogger.WithContext(ctx)
}
func (l *logger) WithContext(ctx context.Context) izap.StdLogger {
	return l.otelLogger.WithContext(ctx)
}

func Debug(msg string, fields ...Field) {
	globalLogger.Debug(msg, fields...)
}
func (l *logger) Debug(msg string, fields ...Field) {
	l.logger.Debug(msg, fields...)
}

func Info(msg string, fields ...Field) {
	globalLogger.Info(msg, fields...)
}
func (l *logger) Info(msg string, fields ...Field) {
	l.logger.Info(msg, fields...)
}

func Warn(msg string, fields ...Field) {
	globalLogger.Warn(msg, fields...)
}
func (l *logger) Warn(msg string, fields ...Field) {
	l.logger.Warn(msg, fields...)
}

func Error(msg string, fields ...Field) {
	globalLogger.Error(msg, fields...)
}
func (l *logger) Error(msg string, fields ...Field) {
	l.logger.Error(msg, fields...)
}

func (l *logger) Clone() Logger {
	copyLogger := *l.logger
	copySugaredLogger := *l.sugaredLogger
	return &logger{
		level:         l.level,
		logger:        &copyLogger,
		sugaredLogger: &copySugaredLogger,
	}
}

func (l *logger) Level() string {
	return l.level
}

func IsDebug() bool {
	return globalLogger.IsDebug()
}
func (l *logger) IsDebug() bool {
	return l.level == option.DebugLevel.String()
}

func ReplaceLogger(l Logger) {
	globalLogger = l
	globalSugaredLogger = l.SugaredLogger()
	zap.ReplaceGlobals(globalLogger.CoreLogger())
}

func Sync() {
	globalLogger.Sync()
}

func (l *logger) Sync() {
	_ = l.logger.Sync()
	_ = l.sugaredLogger.Sync()
}

func GetSugaredLogger() SugaredLogger {
	return globalLogger.SugaredLogger()
}
func (l *logger) SugaredLogger() SugaredLogger {
	return &sugaredLogger{
		sugaredLogger: l.sugaredLogger,
	}
}

func CoreLogger() *zap.Logger {
	return globalLogger.CoreLogger()
}
func (l *logger) CoreLogger() *zap.Logger {
	return l.logger
}

// --- sugared logger ---

func (s *sugaredLogger) Named(name string) SugaredLogger {
	l := s.sugaredLogger.Named(name)
	return &sugaredLogger{sugaredLogger: l}
}

func (s *sugaredLogger) With(args ...interface{}) SugaredLogger {
	l := s.sugaredLogger.With(args...)
	return &sugaredLogger{sugaredLogger: l}
}

func (s *sugaredLogger) Debug(args ...interface{}) {
	s.sugaredLogger.Debug(args...)
}

func (s *sugaredLogger) Info(args ...interface{}) {
	s.sugaredLogger.Info(args...)
}

func (s *sugaredLogger) Warn(args ...interface{}) {
	s.sugaredLogger.Warn(args...)
}

func (s *sugaredLogger) Error(args ...interface{}) {
	s.sugaredLogger.Error(args...)
}

func Panic(args ...interface{}) {
	globalSugaredLogger.Panic(args...)
}
func (s *sugaredLogger) Panic(args ...interface{}) {
	s.sugaredLogger.Panic(args...)
}

func Fatal(args ...interface{}) {
	globalSugaredLogger.Fatal(args...)
}
func (s *sugaredLogger) Fatal(args ...interface{}) {
	s.sugaredLogger.Fatal(args...)
}

func Debugf(format string, args ...interface{}) {
	globalSugaredLogger.Debugf(format, args...)
}
func (s *sugaredLogger) Debugf(format string, args ...interface{}) {
	s.sugaredLogger.Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	globalSugaredLogger.Infof(format, args...)
}
func (s *sugaredLogger) Infof(format string, args ...interface{}) {
	s.sugaredLogger.Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
	globalSugaredLogger.Warnf(format, args...)
}
func (s *sugaredLogger) Warnf(format string, args ...interface{}) {
	s.sugaredLogger.Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	globalSugaredLogger.Errorf(format, args...)
}
func (s *sugaredLogger) Errorf(format string, args ...interface{}) {
	s.sugaredLogger.Errorf(format, args...)
}

func Panicf(format string, args ...interface{}) {
	globalSugaredLogger.Panicf(format, args...)
}
func (s *sugaredLogger) Panicf(format string, args ...interface{}) {
	s.sugaredLogger.Panicf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	globalSugaredLogger.Fatalf(format, args...)
}
func (s *sugaredLogger) Fatalf(format string, args ...interface{}) {
	s.sugaredLogger.Fatalf(format, args...)
}

func (s *sugaredLogger) Sync() {
	_ = s.sugaredLogger.Sync()
}
