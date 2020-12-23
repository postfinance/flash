package flash_test

import (
	"bufio"
	"bytes"
	"fmt"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/postfinance/flash"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/require"
	"github.com/tj/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// nolint: gochecknoglobals
var sink *memorySink

func TestMain(m *testing.M) {
	sink = &memorySink{new(bytes.Buffer)}
	err := zap.RegisterSink("memory", func(*url.URL) (zap.Sink, error) {
		return sink, nil
	})

	if err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}

func TestDefault(t *testing.T) {
	defer sink.Reset()

	l := flash.New(flash.WithSink("memory://"))
	l.Debug("debug")

	e := sink.parse()
	assert.Len(t, e, 0, "debug message logged")
	l.Info("info")

	e = sink.parse()
	assert.Equal(t, "INFO", e[0].level)
	assert.Equal(t, "info", e[0].msg)
	assert.NotEmpty(t, e[0].caller)
}

func TestWithColor(t *testing.T) {
	defer sink.Reset()

	l := flash.New(flash.WithSink("memory://"), flash.WithColor())

	l.Info("info")

	const blue = `1b[34m`

	output := sink.String()
	assert.True(t, strings.Contains(fmt.Sprintf("%q", output), blue))
}

func TestWithoutCaller(t *testing.T) {
	defer sink.Reset()

	l := flash.New(flash.WithSink("memory://"), flash.WithoutCaller())
	l.Info("info")
	require.NotEmpty(t, sink.String())
	e := sink.parse()
	assert.Empty(t, e[0].caller)
}

func TestWithStacktraceWithDebug(t *testing.T) {
	defer sink.Reset()

	l := flash.New(flash.WithSink("memory://"), flash.WithDebug(true), flash.WithStacktrace())
	l.Info("info")
	// only stacktraces for error in debug mode
	assert.False(t, sink.containsStackTrace(), "stack trace detected")
	l.Error("error")
	assert.True(t, sink.containsStackTrace(), "stack trace detected")
}

func TestSetDebugWithStacktrace(t *testing.T) {
	defer sink.Reset()

	l := flash.New(flash.WithSink("memory://"), flash.WithStacktrace())
	l.Debug("debug")
	assert.Len(t, sink.String(), 0, "debug message logged")
	l.Error("error")
	sink.Reset()
	assert.False(t, sink.containsStackTrace(), "stack trace detected")
	l.SetDebug(true)
	l.Debug("debug")
	assert.NotEmpty(t, sink.String(), "no debug message logged")
	l.Error("error")
	assert.True(t, sink.containsStackTrace(), "stack trace detected")
	sink.Reset()
	l.SetDebug(false)
	l.Debug("debug")
	assert.Empty(t, sink.String())
	l.Error("error")
	assert.False(t, sink.containsStackTrace())
}

func TestDisable(t *testing.T) {
	defer sink.Reset()

	l := flash.New(flash.WithSink("memory://"))
	l.Info("info")
	assert.NotEmpty(t, sink.String(), 0)
	sink.Reset()
	l.Disable()
	l.Debug("debug")
	l.Info("info")
	l.Error("error")
	assert.Empty(t, sink.String(), 0)
}

func TestSetLeve(t *testing.T) {
	defer sink.Reset()

	l := flash.New(flash.WithSink("memory://"))
	l.Debug("debug")
	assert.Empty(t, sink.String(), 0)
	l.SetLevel(zapcore.DebugLevel)
	l.Debug("debug")
	assert.NotEmpty(t, sink.String(), 0)
}

func TestWithPrometheus(t *testing.T) {
	defer sink.Reset()

	r := prometheus.NewRegistry()

	l := flash.New(flash.WithSink("memory://"), flash.WithPrometheus("appname", r))

	l.Info("info")
	l.Info("info")
	l.Error("error")
	l.Debug("debug")

	const metadata = `
		# HELP appname_log_messages_total How many log messages created, partitioned by log level.
        # TYPE appname_log_messages_total counter
	`

	expected := `
		appname_log_messages_total{level="error"} 1
		appname_log_messages_total{level="info"} 2
	`

	err := testutil.GatherAndCompare(r, strings.NewReader(metadata+expected), "appname_log_messages_total")
	require.NoError(t, err, "unexpected collecting result")
	l.SetDebug(true)
	l.Debug("debug")

	expected = `
		appname_log_messages_total{level="debug"} 1
		appname_log_messages_total{level="error"} 1
		appname_log_messages_total{level="info"} 2
	`
	err = testutil.GatherAndCompare(r, strings.NewReader(metadata+expected), "appname_log_messages_total")
	require.NoError(t, err, "unexpected collecting result")
}

type memorySink struct {
	*bytes.Buffer
}

// implement the zap.Sink interface.
func (m *memorySink) Close() error { return nil }
func (m *memorySink) Sync() error  { return nil }

func (m *memorySink) parse() []entry {
	s := bufio.NewScanner(m)
	s.Split(bufio.ScanLines)

	entries := []entry{}
	callerRegexp := regexp.MustCompile(`\.go:\d+`)

	for s.Scan() {
		e := entry{}
		fields := strings.Fields(s.Text())
		e.level = fields[1]
		msgIndex := 2

		if callerRegexp.MatchString(fields[2]) {
			e.caller = fields[2]
			msgIndex++
		}

		e.msg = strings.Join(fields[msgIndex:], `\t`)
		entries = append(entries, e)
	}

	return entries
}

func (m *memorySink) containsStackTrace() bool {
	currentPackageName := reflect.TypeOf(memorySink{}).PkgPath()

	return strings.Contains(m.String(), currentPackageName)
}

type entry struct {
	level  string
	caller string
	msg    string
}
