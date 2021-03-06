// Package flash configures an opinionated zap logger.
package flash

import (
	"fmt"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is the flash logger which embeds a `zap.SugaredLogger`.
type Logger struct {
	*zap.SugaredLogger
	atom              zap.AtomicLevel
	m                 sync.Mutex
	currentLevel      zapcore.Level
	disableStackTrace bool
}

// Option configures zap.Config.
type Option func(c *config)

// WithColor enables color output.
func WithColor() Option {
	return func(c *config) {
		c.enableColor = true
	}
}

// WithoutCaller stops annotating logs with the calling function's file
// name and line number.
func WithoutCaller() Option {
	return func(c *config) {
		c.disableCaller = true
	}
}

// WithSinks changes the default zap `stderr` sink.
func WithSinks(sinks ...string) Option {
	return func(c *config) {
		c.sinks = sinks
	}
}

// WithDebug enables or disables `DebugLevel`.
func WithDebug(debug bool) Option {
	return func(c *config) {
		c.isDebug = debug
	}
}

// WithStacktrace completely enables automatic stacktrace capturing. Stacktraces
// are captured on `ErrorLevel` and above when in debug mode. When not in debug mode,
// only `FatalLevel` messages contain stacktraces.
func WithStacktrace() Option {
	return func(c *config) {
		c.disableStacktrace = false
	}
}

// WithPrometheus registers a prometheus log message counter.
//
// The created metrics are of the form:
//     <appName>_log_messages_total{level="info"} 4
func WithPrometheus(appName string, registry prometheus.Registerer) Option {
	return func(c *config) {
		counter := prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: fmt.Sprintf("%s_log_messages_total", appName),
				Help: "How many log messages created, partitioned by log level.",
			},
			[]string{"level"},
		)
		registry.MustRegister(counter)

		c.hook = func(e zapcore.Entry) error {
			counter.WithLabelValues(e.Level.String()).Inc()
			return nil
		}
	}
}

// New creates a new Logger. If no options are specified, stacktraces and color output are disabled and
// the confgured level is `InfoLevel`.
func New(opts ...Option) *Logger {
	l := zap.New(nil) // noop logger
	atom := zap.NewAtomicLevelAt(zap.InfoLevel)

	cfg := config{
		disableStacktrace: true,
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	zapConfig := zap.NewProductionConfig()
	zapConfig.DisableStacktrace = cfg.disableStacktrace
	zapConfig.Sampling = nil
	zapConfig.Encoding = "console"
	zapConfig.DisableCaller = cfg.disableCaller
	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zapConfig.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	zapConfig.Level = atom

	if cfg.enableColor {
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	if len(cfg.sinks) > 0 {
		zapConfig.OutputPaths = cfg.sinks
	}

	var err error

	l, err = zapConfig.Build()
	if err != nil {
		panic(fmt.Sprintf("could not create zap logger: %s", err))
	}

	stackTraceLevel := zap.FatalLevel

	if cfg.isDebug {
		atom.SetLevel(zap.DebugLevel)
		stackTraceLevel = zap.ErrorLevel
	}

	// fix level for stack traces
	if !cfg.disableStacktrace {
		l = l.WithOptions(zap.AddStacktrace(stackTraceLevel))
	}

	if cfg.hook != nil {
		l = l.WithOptions(zap.Hooks(cfg.hook))
	}

	defer func() {
		_ = l.Sync()
	}()

	return &Logger{
		SugaredLogger:     l.Sugar(),
		atom:              atom,
		disableStackTrace: cfg.disableStacktrace,
	}
}

// SetDebug enables or disables `DebugLevel`.
func (l *Logger) SetDebug(d bool) {
	level := zap.DebugLevel
	stackTraceLevel := zap.ErrorLevel

	if !d {
		l.m.Lock()
		level = l.currentLevel
		l.m.Unlock()

		stackTraceLevel = zap.FatalLevel
	}

	l.atom.SetLevel(level)
	l.stackTrace(stackTraceLevel)
}

// Disable disables (nearly) all output. Only `FatalLevel` errors are logged.
func (l *Logger) Disable() {
	l.m.Lock()
	l.currentLevel = zapcore.FatalLevel
	l.m.Unlock()
	l.atom.SetLevel(zap.FatalLevel)
}

// SetLevel sets the chosen level. If stacktraces are enabled, it adjusts stacktrace levels accordingly.
func (l *Logger) SetLevel(level zapcore.Level) {
	l.m.Lock()
	oldLevel := l.currentLevel
	l.currentLevel = level
	l.m.Unlock()
	l.atom.SetLevel(level)

	if level == zap.DebugLevel {
		l.stackTrace(zap.ErrorLevel)
		return
	}

	if oldLevel == zap.DebugLevel && level != zap.DebugLevel {
		l.stackTrace(zap.FatalLevel)
	}
}

// Get returns the embedded zap.Logger
func (l *Logger) Get() *zap.SugaredLogger {
	return l.SugaredLogger
}

func (l *Logger) stackTrace(lvl zapcore.Level) {
	if l.disableStackTrace {
		return
	}

	l.m.Lock()
	l.SugaredLogger = l.Get().Desugar().WithOptions(zap.AddStacktrace(lvl)).Sugar()
	l.m.Unlock()
}

type config struct {
	enableColor       bool
	disableCaller     bool
	disableStacktrace bool
	isDebug           bool
	hook              func(zapcore.Entry) error
	sinks             []string
}
