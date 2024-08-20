package otel

import (
	"context"
	"fmt"
	"runtime"
	"strconv"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/logerror/easylog/pkg/izap"
)

const (
	defaultTraceIdKey = "trace_id"
	defaultSpanIdKey  = "span_id"
	defaultSampledKey = "sampled"
)

var (
	logSeverityKey = attribute.Key("log.severity")
	logMessageKey  = attribute.Key("log.message")
)

var _ izap.StdLogger = (*stdLogger)(nil)

type stdLogger struct {
	*zap.Logger
	ctx context.Context

	LogLevel         zapcore.Level
	ErrorStatusLevel zapcore.Level
	CallerDepth      int8
	CallerSkip       uint8
}

func (l *stdLogger) Log(lvl zapcore.Level, msg string, fields ...zap.Field) {
	l.traceInfo(lvl, msg)
	l.Logger.Log(lvl, msg, fields...)
}

func (l *stdLogger) Debug(msg string, fields ...zap.Field) {
	l.traceInfo(zapcore.DebugLevel, msg)
	l.Logger.Debug(msg, fields...)
}

func (l *stdLogger) Info(msg string, fields ...zap.Field) {
	l.traceInfo(zapcore.InfoLevel, msg)
	l.Logger.Info(msg, fields...)
}

func (l *stdLogger) Warn(msg string, fields ...zap.Field) {
	l.traceInfo(zapcore.WarnLevel, msg)
	l.Logger.Warn(msg, fields...)
}

func (l *stdLogger) Error(msg string, fields ...zap.Field) {
	l.traceInfo(zapcore.ErrorLevel, msg)
	l.Logger.Error(msg, fields...)
}

func (l *stdLogger) Panic(msg string, fields ...zap.Field) {
	l.traceInfo(zapcore.PanicLevel, msg)
	l.Logger.Panic(msg, fields...)
}

func (l *stdLogger) Fatal(msg string, fields ...zap.Field) {
	l.traceInfo(zapcore.FatalLevel, msg)
	l.Logger.Fatal(msg, fields...)
}

func (l *stdLogger) DPanic(msg string, fields ...zap.Field) {
	l.traceInfo(zapcore.DPanicLevel, msg)
	l.Logger.DPanic(msg, fields...)
}

func (l *stdLogger) traceInfo(lvl zapcore.Level, msg string) {
	span := trace.SpanFromContext(l.ctx)
	if !span.IsRecording() {
		return
	}

	if lvl >= l.LogLevel {
		var attrs []attribute.KeyValue
		attrs = append(attrs, logSeverityKey.String(lvl.String()))
		attrs = append(attrs, logMessageKey.String(msg))
		attrs = recordCaller(attrs, l.CallerDepth, int(l.CallerSkip+3))
		span.AddEvent("log", trace.WithAttributes(attrs...))
	}

	if lvl >= l.ErrorStatusLevel {
		span.SetStatus(codes.Error, msg)
	}
}

func recordCaller(attrs []attribute.KeyValue, callerDepth int8, skip int) []attribute.KeyValue {
	if callerDepth >= 0 {
		var stack bool
		var pc []uintptr
		if callerDepth == 0 {
			pc = make([]uintptr, 1)
			stack = false
		} else {
			pc = make([]uintptr, callerDepth)
			stack = true
		}
		cc := runtime.Callers(skip+1, pc)
		frames := runtime.CallersFrames(pc)

		var stackStr strings.Builder
		for i := 0; i < cc; i++ {
			next, more := frames.Next()
			if !more {
				break
			}
			if i == 0 { //first frame
				attrs = append(attrs, semconv.CodeFunctionKey.String(next.Function))
				attrs = append(attrs, semconv.CodeFilepathKey.String(next.File))
				attrs = append(attrs, semconv.CodeLineNumberKey.Int(next.Line))
			}
			if stack {
				stackStr.WriteString(next.Function)
				stackStr.WriteString(" ")
				stackStr.WriteString(next.File)
				stackStr.WriteString(":")
				stackStr.WriteString(strconv.Itoa(next.Line))
				stackStr.WriteString("\n")
			}
		}
		if stack {
			attrs = append(attrs, semconv.ExceptionStacktraceKey.String(stackStr.String()))
		}
	}
	return attrs
}

func WithContext(ctx context.Context, zLogger *zap.Logger, opts ...Option) izap.StdLogger {
	if ctx == nil {
		return zLogger
	}

	spanContext := trace.SpanContextFromContext(ctx)
	if !spanContext.IsValid() { // must be !isRecording()
		return zLogger
	}

	cfg := applyConfig(opts...)

	var fields []zap.Field
	if cfg.LogTraceId {
		traceIdField := zap.String(defaultTraceIdKey, spanContext.TraceID().String())
		fields = append(fields, traceIdField)
	}
	if cfg.LogSpanId {
		spanIdField := zap.String(defaultSpanIdKey, spanContext.SpanID().String())
		fields = append(fields, spanIdField)
	}
	if cfg.LogSampled {
		sampledField := zap.String(defaultSampledKey, spanContext.TraceFlags().String())
		fields = append(fields, sampledField)
	}

	return &stdLogger{
		Logger:           zLogger.WithOptions(zap.Fields(fields...), zap.AddCallerSkip(1)),
		ctx:              ctx,
		LogLevel:         cfg.LogLevel,
		ErrorStatusLevel: cfg.ErrorStatusLevel,
		CallerDepth:      cfg.CallerDepth,
		CallerSkip:       cfg.CallerSkip,
	}
}

var _ izap.StdSugaredLogger = (*stdSugaredLogger)(nil)

type stdSugaredLogger struct {
	*zap.SugaredLogger
	ctx              context.Context
	LogLevel         zapcore.Level
	ErrorStatusLevel zapcore.Level
	CallerDepth      int8
	CallerSkip       uint8
}

func (s *stdSugaredLogger) sugaredTraceInfo(lvl zapcore.Level, msg string, ln bool, args []interface{}) {
	span := trace.SpanFromContext(s.ctx)
	if !span.IsRecording() {
		return
	}

	//first return for reduce call format
	if lvl < s.LogLevel && lvl < s.ErrorStatusLevel {
		return
	}

	if ln {
		msg = getMessageln(args)
	} else {
		msg = getMessage(msg, args)
	}

	if lvl >= s.LogLevel {
		var attrs []attribute.KeyValue
		attrs = append(attrs, logSeverityKey.String(lvl.String()))
		attrs = append(attrs, logMessageKey.String(msg))
		attrs = recordCaller(attrs, s.CallerDepth, int(3+s.CallerSkip))

		//TODO record caller
		span.AddEvent("log", trace.WithAttributes(attrs...))
	}

	if lvl >= s.ErrorStatusLevel {
		span.SetStatus(codes.Error, msg)
	}
}

// getMessage copy from zap.
func getMessage(template string, fmtArgs []interface{}) string {
	if len(fmtArgs) == 0 {
		return template
	}

	if template != "" {
		return fmt.Sprintf(template, fmtArgs...)
	}

	if len(fmtArgs) == 1 {
		if str, ok := fmtArgs[0].(string); ok {
			return str
		}
	}
	return fmt.Sprint(fmtArgs...)
}

// getMessageln copy from zap.
func getMessageln(fmtArgs []interface{}) string {
	msg := fmt.Sprintln(fmtArgs...)
	return msg[:len(msg)-1]
}

func (s *stdSugaredLogger) Debug(args ...interface{}) {
	s.sugaredTraceInfo(zapcore.DebugLevel, "", false, args)
	s.SugaredLogger.Debug(args...)
}

func (s *stdSugaredLogger) Info(args ...interface{}) {
	s.sugaredTraceInfo(zapcore.InfoLevel, "", false, args)
	s.SugaredLogger.Info(args...)
}

func (s *stdSugaredLogger) Warn(args ...interface{}) {
	s.sugaredTraceInfo(zapcore.WarnLevel, "", false, args)
	s.SugaredLogger.Warn(args...)
}

func (s *stdSugaredLogger) Error(args ...interface{}) {
	s.sugaredTraceInfo(zapcore.ErrorLevel, "", false, args)
	s.SugaredLogger.Error(args...)
}

func (s *stdSugaredLogger) DPanic(args ...interface{}) {
	s.sugaredTraceInfo(zapcore.DPanicLevel, "", false, args)
	s.SugaredLogger.DPanic(args...)
}

func (s *stdSugaredLogger) Panic(args ...interface{}) {
	s.sugaredTraceInfo(zapcore.PanicLevel, "", false, args)
	s.SugaredLogger.Panic(args...)
}

func (s *stdSugaredLogger) Fatal(args ...interface{}) {
	s.sugaredTraceInfo(zapcore.FatalLevel, "", false, args)
	s.SugaredLogger.Fatal(args...)
}

func (s *stdSugaredLogger) Debugf(template string, args ...interface{}) {
	s.sugaredTraceInfo(zapcore.DebugLevel, template, false, args)
	s.SugaredLogger.Debugf(template, args...)
}

func (s *stdSugaredLogger) Infof(template string, args ...interface{}) {
	s.sugaredTraceInfo(zapcore.InfoLevel, template, false, args)
	s.SugaredLogger.Infof(template, args...)
}

func (s *stdSugaredLogger) Warnf(template string, args ...interface{}) {
	s.sugaredTraceInfo(zapcore.WarnLevel, template, false, args)
	s.SugaredLogger.Warnf(template, args...)
}

func (s *stdSugaredLogger) Errorf(template string, args ...interface{}) {
	s.sugaredTraceInfo(zapcore.ErrorLevel, template, false, args)
	s.SugaredLogger.Errorf(template, args...)
}

func (s *stdSugaredLogger) DPanicf(template string, args ...interface{}) {
	s.sugaredTraceInfo(zapcore.DPanicLevel, template, false, args)
	s.SugaredLogger.DPanicf(template, args...)
}

func (s *stdSugaredLogger) Panicf(template string, args ...interface{}) {
	s.sugaredTraceInfo(zapcore.PanicLevel, template, false, args)
	s.SugaredLogger.Panicf(template, args...)
}

func (s *stdSugaredLogger) Fatalf(template string, args ...interface{}) {
	s.sugaredTraceInfo(zapcore.FatalLevel, template, false, args)
	s.SugaredLogger.Fatalf(template, args...)
}

func (s *stdSugaredLogger) Debugw(msg string, keysAndValues ...interface{}) {
	s.sugaredTraceInfo(zapcore.DebugLevel, msg, false, nil)
	s.SugaredLogger.Debugw(msg, keysAndValues...)
}

func (s *stdSugaredLogger) Infow(msg string, keysAndValues ...interface{}) {
	s.sugaredTraceInfo(zapcore.InfoLevel, msg, false, nil)
	s.SugaredLogger.Infow(msg, keysAndValues...)
}

func (s *stdSugaredLogger) Warnw(msg string, keysAndValues ...interface{}) {
	s.sugaredTraceInfo(zapcore.WarnLevel, msg, false, nil)
	s.SugaredLogger.Warnw(msg, keysAndValues...)
}

func (s *stdSugaredLogger) Errorw(msg string, keysAndValues ...interface{}) {
	s.sugaredTraceInfo(zapcore.ErrorLevel, msg, false, nil)
	s.SugaredLogger.Errorw(msg, keysAndValues...)
}

func (s *stdSugaredLogger) DPanicw(msg string, keysAndValues ...interface{}) {
	s.sugaredTraceInfo(zapcore.DPanicLevel, msg, false, nil)
	s.SugaredLogger.DPanicw(msg, keysAndValues...)
}

func (s *stdSugaredLogger) Panicw(msg string, keysAndValues ...interface{}) {
	s.sugaredTraceInfo(zapcore.PanicLevel, msg, false, nil)
	s.SugaredLogger.Panicw(msg, keysAndValues...)
}

func (s *stdSugaredLogger) Fatalw(msg string, keysAndValues ...interface{}) {
	s.sugaredTraceInfo(zapcore.FatalLevel, msg, false, nil)
	s.SugaredLogger.Fatalw(msg, keysAndValues...)
}

func (s *stdSugaredLogger) Debugln(args ...interface{}) {
	s.sugaredTraceInfo(zapcore.DebugLevel, "", true, args)
	s.SugaredLogger.Debugln(args...)
}

func (s *stdSugaredLogger) Infoln(args ...interface{}) {
	s.sugaredTraceInfo(zapcore.InfoLevel, "", true, args)
	s.SugaredLogger.Infoln(args...)
}

func (s *stdSugaredLogger) Warnln(args ...interface{}) {
	s.sugaredTraceInfo(zapcore.WarnLevel, "", true, args)
	s.SugaredLogger.Warnln(args...)
}

func (s *stdSugaredLogger) Errorln(args ...interface{}) {
	s.sugaredTraceInfo(zapcore.ErrorLevel, "", true, args)
	s.SugaredLogger.Errorln(args...)
}

func (s *stdSugaredLogger) DPanicln(args ...interface{}) {
	s.sugaredTraceInfo(zapcore.DPanicLevel, "", true, args)
	s.SugaredLogger.DPanicln(args...)
}

func (s *stdSugaredLogger) Panicln(args ...interface{}) {
	s.sugaredTraceInfo(zapcore.PanicLevel, "", true, args)
	s.SugaredLogger.Panicln(args...)
}

func (s *stdSugaredLogger) Fatalln(args ...interface{}) {
	s.sugaredTraceInfo(zapcore.FatalLevel, "", true, args)
	s.SugaredLogger.Fatalln(args...)
}

func SugarWithContext(ctx context.Context, zsLogger *zap.SugaredLogger, opts ...Option) izap.StdSugaredLogger {
	if ctx == nil {
		return zsLogger
	}

	spanContext := trace.SpanContextFromContext(ctx)
	if !spanContext.IsValid() { // must be !isRecording()
		return zsLogger
	}

	cfg := applyConfig(opts...)

	var fields []zap.Field
	if cfg.LogTraceId {
		traceIdField := zap.String(defaultTraceIdKey, spanContext.TraceID().String())
		fields = append(fields, traceIdField)
	}
	if cfg.LogSpanId {
		spanIdField := zap.String(defaultSpanIdKey, spanContext.SpanID().String())
		fields = append(fields, spanIdField)
	}
	if cfg.LogSampled {
		sampledField := zap.String(defaultSampledKey, spanContext.TraceFlags().String())
		fields = append(fields, sampledField)
	}

	return &stdSugaredLogger{
		SugaredLogger:    zsLogger.WithOptions(zap.Fields(fields...), zap.AddCallerSkip(1)),
		ctx:              ctx,
		LogLevel:         cfg.LogLevel,
		ErrorStatusLevel: cfg.ErrorStatusLevel,
		CallerDepth:      cfg.CallerDepth,
		CallerSkip:       cfg.CallerSkip,
	}
}

func applyConfig(opts ...Option) config {
	cfg := config{
		LogTraceId:       true,
		LogLevel:         zapcore.ErrorLevel,
		ErrorStatusLevel: zapcore.ErrorLevel,
		CallerDepth:      8,
	}
	for _, opt := range opts {
		opt.apply(&cfg)
	}
	return cfg
}

var _ izap.Logger = (*logger)(nil)

type logger struct {
	*zap.Logger
	cfg config
}

func NewLogger(log *zap.Logger, opts ...Option) izap.Logger {
	cfg := applyConfig(opts...)
	return &logger{
		Logger: log,
		cfg:    cfg,
	}
}

func (l *logger) WithContext(ctx context.Context) izap.StdLogger {
	spanContext := trace.SpanContextFromContext(ctx)
	if !spanContext.IsValid() { // must be !isRecording()
		return l
	}
	var fields []zap.Field

	if l.cfg.LogTraceId {
		traceIdField := zap.String(defaultTraceIdKey, spanContext.TraceID().String())
		fields = append(fields, traceIdField)
	}
	if l.cfg.LogSpanId {
		spanIdField := zap.String(defaultSpanIdKey, spanContext.SpanID().String())
		fields = append(fields, spanIdField)
	}
	if l.cfg.LogSampled {
		sampledField := zap.String(defaultSampledKey, spanContext.TraceFlags().String())
		fields = append(fields, sampledField)
	}
	return &stdLogger{
		Logger:           l.Logger.WithOptions(zap.Fields(fields...), zap.AddCallerSkip(1)),
		ctx:              ctx,
		LogLevel:         l.cfg.LogLevel,
		ErrorStatusLevel: l.cfg.ErrorStatusLevel,
		CallerDepth:      l.cfg.CallerDepth,
		CallerSkip:       l.cfg.CallerSkip,
	}
}

func (l *logger) With(fields ...zap.Field) izap.Logger {
	newL := l.Logger.With(fields...)
	return &logger{
		Logger: newL,
		cfg:    l.cfg,
	}
}

func (l *logger) WithOptions(opts ...zap.Option) izap.Logger {
	newL := l.Logger.WithOptions(opts...)
	return &logger{
		Logger: newL,
		cfg:    l.cfg,
	}
}

func (l *logger) Sugar() izap.SugaredLogger {
	sl := l.Logger.Sugar()
	return &sugaredLogger{
		SugaredLogger: sl,
		cfg:           l.cfg,
	}
}

func NewSugaredLogger(log *zap.SugaredLogger, opts ...Option) izap.SugaredLogger {
	cfg := applyConfig(opts...)
	return &sugaredLogger{
		SugaredLogger: log,
		cfg:           cfg,
	}
}

var _ izap.SugaredLogger = (*sugaredLogger)(nil)

type sugaredLogger struct {
	*zap.SugaredLogger
	cfg config
}

func (o *sugaredLogger) WithContext(ctx context.Context) izap.StdSugaredLogger {
	spanContext := trace.SpanContextFromContext(ctx)
	if !spanContext.IsValid() { // must be !isRecording()
		return o
	}
	var fields []zap.Field
	if o.cfg.LogTraceId {
		traceIdField := zap.String(defaultTraceIdKey, spanContext.TraceID().String())
		fields = append(fields, traceIdField)
	}
	if o.cfg.LogSpanId {
		spanIdField := zap.String(defaultSpanIdKey, spanContext.SpanID().String())
		fields = append(fields, spanIdField)
	}
	if o.cfg.LogSampled {
		sampledField := zap.String(defaultSampledKey, spanContext.TraceFlags().String())
		fields = append(fields, sampledField)
	}
	return &stdSugaredLogger{
		SugaredLogger:    o.SugaredLogger.WithOptions(zap.Fields(fields...), zap.AddCallerSkip(1)),
		ctx:              ctx,
		LogLevel:         o.cfg.LogLevel,
		ErrorStatusLevel: o.cfg.ErrorStatusLevel,
		CallerDepth:      o.cfg.CallerDepth,
		CallerSkip:       o.cfg.CallerSkip,
	}
}

func (o *sugaredLogger) With(args ...interface{}) izap.SugaredLogger {
	sl := o.SugaredLogger.With(args)
	return &sugaredLogger{
		SugaredLogger: sl,
		cfg:           o.cfg,
	}
}

func (o *sugaredLogger) WithOptions(opts ...zap.Option) izap.SugaredLogger {
	sl := o.SugaredLogger.WithOptions(opts...)
	return &sugaredLogger{
		SugaredLogger: sl,
		cfg:           o.cfg,
	}
}

func (o *sugaredLogger) Desugar() izap.Logger {
	l := o.SugaredLogger.Desugar()
	return &logger{
		Logger: l,
		cfg:    o.cfg,
	}
}
