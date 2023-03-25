package zerolog

import (
	"github.com/pkritiotis/outbox/logs"
	"github.com/rs/zerolog"
	"io"
	"path/filepath"
	"runtime"
)

type zerologLogger struct {
	logger zerolog.Logger
	prefix string
}

// NewZerologAdapter returns a new Log instance
func NewZerologAdapter(prefix string, out io.Writer) logs.LoggerAdapter {
	zl := zerolog.New(out).With().Str("prefix", prefix).Timestamp().Logger()

	return &zerologLogger{prefix: prefix, logger: zl}
}

// Error implements the LoggerAdapter interface
func (l zerologLogger) Error(msg string, fields map[string]interface{}) {
	lf := l.logWithFields(fields)
	lf.Warn().Msg(msg)
}

// Trace implements the LoggerAdapter interface
func (l zerologLogger) Trace(msg string, fields map[string]interface{}) {
	lf := l.logWithFields(fields)
	lf.Trace().Msg(msg)
}

// Warn implements the LoggerAdapter interface
func (l zerologLogger) Warn(msg string, fields map[string]interface{}) {
	lf := l.logWithFields(fields)
	lf.Warn().Msg(msg)
}

// Info implements the LoggerAdapter interface
func (l zerologLogger) Info(msg string, fields map[string]interface{}) {
	lf := l.logWithFields(fields)
	lf.Info().Msg(msg)
}

// Debug implements the LoggerAdapter interface
func (l zerologLogger) Debug(msg string, fields map[string]interface{}) {
	lf := l.logWithFields(fields)
	lf.Debug().Msg(msg)
}

// Fatal implements the LoggerAdapter interface
func (l zerologLogger) Fatal(msg string, fields map[string]interface{}) {
	lf := l.logWithFields(fields)
	lf.Fatal().Msg(msg)
}

func (l zerologLogger) logWithFields(fields map[string]interface{}) zerolog.Logger {
	ll := l.logger
	_, f, no, ok := runtime.Caller(2)
	if ok {
		f = filepath.Base(f)
		ll = l.logger.With().Str("file", f).Int("line", no).Fields(fields).Logger()
	}

	return ll
}
