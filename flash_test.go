package flash_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
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

	l := flash.New(flash.WithSinks("memory://"))
	l.Debug("debug")
	assert.Empty(t, sink.String(), "default should not log debug messages")

	l.Info("info")

	// log entry should not be in json format since in tests we do not have a terminal
	e, err := sink.parse()
	require.NoError(t, err)

	assert.Equal(t, "INFO", e[0].Level)
	assert.Equal(t, "info", e[0].Msg)
	assert.NotEmpty(t, e[0].Caller)
	assert.NotEmpty(t, e[0].TS)

	t.Run("default console should not contain color without terminal", func(t *testing.T) {
		sink.Reset()
		l := flash.New(flash.WithSinks("memory://"), flash.WithEncoder(flash.Console))
		l.Info("a log message")
		want := fmt.Sprintf("%s\t%s\t%s", "INFO", "flash/flash_test.go:60", "a log message")
		entry := sink.String()
		assert.True(t, strings.Contains(entry, want), "got: %s, want:%s", entry, want)
	})

	t.Run("console output with color option should produce color output", func(t *testing.T) {
		sink.Reset()
		l := flash.New(flash.WithSinks("memory://"), flash.WithEncoder(flash.Console), flash.WithColor())
		l.Info("a log message")

		const blue = `1b[34m`
		assert.True(t, strings.Contains(fmt.Sprintf("%q", sink.String()), blue))
	})

	t.Run("default console without timestmaps", func(t *testing.T) {
		sink.Reset()
		l := flash.New(flash.WithSinks("memory://"), flash.WithEncoder(flash.Console), flash.WithoutTimestamps())
		l.Info("a log message")
		assert.Equal(t, "INFO\tflash/flash_test.go:78\ta log message\n", sink.String())
	})
}

func TestWithoutCaller(t *testing.T) {
	defer sink.Reset()

	l := flash.New(flash.WithSinks("memory://"), flash.WithoutCaller())
	l.Info("info")
	require.NotEmpty(t, sink.String())
	e, err := sink.parse()
	require.NoError(t, err)
	assert.Empty(t, e[0].Caller)
}

func TestLogFmt(t *testing.T) {
	defer sink.Reset()

	l := flash.New(flash.WithSinks("memory://"), flash.WithEncoder(flash.LogFmt))
	l.Info("info")
	require.NotEmpty(t, sink.String())
	fmt.Println(sink.String())
}

func TestWithStacktraceWithDebug(t *testing.T) {
	defer sink.Reset()

	l := flash.New(flash.WithSinks("memory://"), flash.WithDebug(true), flash.WithStacktrace())

	l.Info("info")

	e, err := sink.parse()
	require.NoError(t, err)
	// only stacktraces for error in debug mode
	assert.Empty(t, e[0].Stacktrace, "stack trace logged")

	sink.Reset()

	l.Error("error")

	e, err = sink.parse()
	require.NoError(t, err)
	assert.NotEmpty(t, e[0].Stacktrace, "no stack trace logged")
}

func TestSetDebugWithStacktrace(t *testing.T) {
	defer sink.Reset()

	l := flash.New(flash.WithSinks("memory://"), flash.WithStacktrace())

	t.Run("it should not log stack traces on errors when not in debug mode", func(t *testing.T) {
		l.Debug("debug")

		assert.Len(t, sink.String(), 0, "debug message logged")

		l.Error("error")

		e, err := sink.parse()
		require.NoError(t, err)

		assert.Empty(t, e[0].Stacktrace)
	})

	sink.Reset()

	t.Run("it should log stack traces on errors when in debug mode", func(t *testing.T) {
		l.SetDebug(true)
		l.Debug("debug")
		assert.NotEmpty(t, sink.String(), "no debug message logged")

		l.Error("error")

		e, err := sink.parse()
		require.NoError(t, err)

		assert.NotEmpty(t, e[1].Stacktrace)
	})

	sink.Reset()

	t.Run("it should not log stack traces on errors when not in info mode", func(t *testing.T) {
		l.SetDebug(false)
		l.Error("error")

		e, err := sink.parse()
		require.NoError(t, err)

		assert.Empty(t, e[0].Stacktrace)
	})
}

func TestDisable(t *testing.T) {
	defer sink.Reset()

	l := flash.New(flash.WithSinks("memory://"))
	l.Info("info")
	assert.NotEmpty(t, sink.String(), 0)
	sink.Reset()
	l.Disable()
	l.Debug("debug")
	l.Info("info")
	l.Error("error")
	assert.Empty(t, sink.String(), 0)
}

func TestSetLevelWithStacktrace(t *testing.T) {
	defer sink.Reset()

	l := flash.New(flash.WithSinks("memory://"), flash.WithStacktrace())

	t.Run("it should not log stack traces on errors when not in debug mode", func(t *testing.T) {
		l.Debug("debug")
		assert.Empty(t, sink.String(), 0)

		l.Error("error")

		e, err := sink.parse()
		require.NoError(t, err)

		assert.Empty(t, e[0].Stacktrace)
	})

	sink.Reset()

	t.Run("it should log stack traces on errors when in debug mode", func(t *testing.T) {
		l.SetLevel(zapcore.DebugLevel)
		l.Debug("debug")
		assert.NotEmpty(t, sink.String(), 0)
		l.Error("error")

		e, err := sink.parse()
		require.NoError(t, err)

		assert.NotEmpty(t, e[1].Stacktrace)
	})

	sink.Reset()

	t.Run("it should not log stack traces on errors when not in info mode", func(t *testing.T) {
		l.SetLevel(zapcore.InfoLevel)
		l.Error("error")

		e, err := sink.parse()
		require.NoError(t, err)

		assert.Empty(t, e[0].Stacktrace)
	})
}

func TestWithPrometheus(t *testing.T) {
	defer sink.Reset()

	r := prometheus.NewRegistry()

	l := flash.New(flash.WithSinks("memory://"), flash.WithPrometheus("appname", r))

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

func TestWithFileConfig(t *testing.T) {
	file, err := ioutil.TempFile("", "*test.log")
	require.NoError(t, err)

	defer func() {
		_ = file.Close()
		_ = os.Remove(file.Name())
	}()

	l := flash.New(flash.WithFile(flash.FileConfig{
		Path: file.Name(),
	}))

	l.Info("hello world")

	d, err := os.ReadFile(file.Name())
	require.NoError(t, err)
	assert.Contains(t, string(d), "INFO")
}

type memorySink struct {
	*bytes.Buffer
}

// implement the zap.Sink interface.
func (m *memorySink) Close() error { return nil }
func (m *memorySink) Sync() error  { return nil }

func (m *memorySink) parse() ([]logEntry, error) {
	s := bufio.NewScanner(m)
	s.Split(bufio.ScanLines)

	entries := []logEntry{}

	for s.Scan() {
		e := logEntry{}
		if err := json.Unmarshal(s.Bytes(), &e); err != nil {
			return nil, err
		}

		entries = append(entries, e)
	}

	return entries, nil
}

type logEntry struct {
	Level      string `json:"level"`
	TS         string `json:"ts"`
	Caller     string `json:"caller"`
	Msg        string `json:"msg"`
	Stacktrace string `json:"stacktrace"`
}
