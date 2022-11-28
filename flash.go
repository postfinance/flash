// Package flash configures an opinionated zap logger.
package flash

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/mattn/go-isatty"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	lumberjackSinkURIPrefix = "lumberjack"
)

// EncoderType is a zap encoder.
type EncoderType int

// All supported encoder types.
const (
	Console EncoderType = iota
	JSON
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

// WithEncoder configures the zap encoder.
func WithEncoder(e EncoderType) Option {
	return func(c *config) {
		c.encoder = e
	}
}

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
//
//	<appName>_log_messages_total{level="info"} 4
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

// WithFile configures the logger to log output into a file.
func WithFile(cfg FileConfig) Option {
	return func(c *config) {
		c.fileConfig = &cfg
	}
}

// WithoutTimestamps configures the logger to log without timestamps.
func WithoutTimestamps() Option {
	return func(c *config) {
		c.disableTimestamps = true
	}
}

// FileConfig holds the configuration for logging into a file. The size is in Megabytes and
// MaxAge is in days. If compress is true the rotated files are compressed.
type FileConfig struct {
	Path       string
	MaxSize    int
	MaxBackups int
	MaxAge     int
	Compress   bool
}

// New creates a new Logger. If no options are specified, stacktraces and color output are disabled and
// the confgured level is `InfoLevel`.
func New(opts ...Option) *Logger {
	l := zap.New(nil) // noop logger
	atom := zap.NewAtomicLevelAt(zap.InfoLevel)

	cfg := config{
		disableStacktrace: true,
		encoder:           Console,
	}

	// set encoder to json and disable color output when no terminal is detected
	if !isatty.IsTerminal(os.Stdout.Fd()) {
		cfg.encoder = JSON
		cfg.enableColor = false
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	if cfg.encoder != Console {
		cfg.enableColor = false
	}

	zapConfig := zap.NewProductionConfig()
	zapConfig.DisableStacktrace = cfg.disableStacktrace
	zapConfig.Sampling = nil
	zapConfig.DisableCaller = cfg.disableCaller
	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zapConfig.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	zapConfig.Level = atom

	switch cfg.encoder {
	case Console:
		zapConfig.Encoding = "console"
	case JSON:
		zapConfig.Encoding = "json"
	}

	// no colors when logging to file
	if cfg.enableColor && cfg.fileConfig == nil {
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	if len(cfg.sinks) > 0 {
		zapConfig.OutputPaths = cfg.sinks
	}

	if cfg.disableTimestamps {
		zapConfig.EncoderConfig.TimeKey = ""
	}

	if cfg.fileConfig != nil {
		if err := cfg.registerFileSink(); err != nil {
			panic(err)
		}

		zapConfig.OutputPaths = []string{cfg.fileConfig.sinkURI()}
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
	disableTimestamps bool
	isDebug           bool
	hook              func(zapcore.Entry) error
	sinks             []string
	fileConfig        *FileConfig
	encoder           EncoderType
}

func (cfg FileConfig) sinkURI() string {
	return fmt.Sprintf("%s://localhost/%s", lumberjackSinkURIPrefix, cfg.Path)
}

func pathFromURI(u *url.URL) string {
	return strings.Replace(u.Path, "/", "", 1)
}

type lumberjackSink struct {
	*lumberjack.Logger
}

// Sync implements zap.Sink. The remaining methods are implemented
// by the embedded *lumberjack.Logger.
func (lumberjackSink) Sync() error { return nil }

func (c config) registerFileSink() error {
	return zap.RegisterSink(lumberjackSinkURIPrefix, func(u *url.URL) (zap.Sink, error) {
		return lumberjackSink{
			Logger: &lumberjack.Logger{
				Filename:   pathFromURI(u),
				MaxSize:    c.fileConfig.MaxSize,
				MaxAge:     c.fileConfig.MaxAge,
				MaxBackups: c.fileConfig.MaxBackups,
				Compress:   c.fileConfig.Compress,
			},
		}, nil
	})
}
