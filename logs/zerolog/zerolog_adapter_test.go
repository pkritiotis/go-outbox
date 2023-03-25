package zerolog

import (
	"bytes"
	"github.com/pkritiotis/outbox/logs"
	"regexp"
	"strings"
	"sync"
	"testing"
)

func TestLogger(t *testing.T) {
	testCases := []struct {
		name        string
		fields      map[string]interface{}
		outputLog   func(l logs.LoggerAdapter, fields map[string]interface{})
		wantPattern string
	}{
		{
			name:   "Info logging",
			fields: map[string]interface{}{"user": "Alice", "request_id": "123"},
			outputLog: func(l logs.LoggerAdapter, fields map[string]interface{}) {
				l.Info("info test", fields)
			},
			wantPattern: `^\{"level":"[a-z]+","prefix":"[a-zA-Z]+","file":"[a-zA-Z0-9_.-]+","line":[0-9]+,"request_id":"[a-zA-Z0-9]+","user":"[a-zA-Z]+","time":"[0-9TZ:+-]+","message":"info test"\}$`,
		},
		{
			name:   "Debug logging",
			fields: map[string]interface{}{"user": "Alice", "request_id": "123"},
			outputLog: func(l logs.LoggerAdapter, fields map[string]interface{}) {
				l.Debug("debug test", fields)
			},
			wantPattern: `^\{"level":"[a-z]+","prefix":"[a-zA-Z]+","file":"[a-zA-Z0-9_.-]+","line":[0-9]+,"request_id":"[a-zA-Z0-9]+","user":"[a-zA-Z]+","time":"[0-9TZ:+-]+","message":"debug test"\}$`,
		},
		{
			name:   "Trace logging",
			fields: map[string]interface{}{"user": "Alice", "request_id": "123"},
			outputLog: func(l logs.LoggerAdapter, fields map[string]interface{}) {
				l.Trace("trace test", fields)
			},
			wantPattern: `^\{"level":"[a-z]+","prefix":"[a-zA-Z]+","file":"[a-zA-Z0-9_.-]+","line":[0-9]+,"request_id":"[a-zA-Z0-9]+","user":"[a-zA-Z]+","time":"[0-9TZ:+-]+","message":"trace test"\}$`,
		},
		{
			name:   "Error logging",
			fields: map[string]interface{}{"user": "Alice", "request_id": "123"},
			outputLog: func(l logs.LoggerAdapter, fields map[string]interface{}) {
				l.Error("error test", fields)
			},
			wantPattern: `^\{"level":"[a-z]+","prefix":"[a-zA-Z]+","file":"[a-zA-Z0-9_.-]+","line":[0-9]+,"request_id":"[a-zA-Z0-9]+","user":"[a-zA-Z]+","time":"[0-9TZ:+-]+","message":"error test"\}$`,
		},
		{
			name:   "Warn logging",
			fields: map[string]interface{}{"user": "Alice", "request_id": "123"},
			outputLog: func(l logs.LoggerAdapter, fields map[string]interface{}) {
				l.Warn("warn test", fields)
			},
			wantPattern: `^\{"level":"[a-z]+","prefix":"[a-zA-Z]+","file":"[a-zA-Z0-9_.-]+","line":[0-9]+,"request_id":"[a-zA-Z0-9]+","user":"[a-zA-Z]+","time":"[0-9TZ:+-]+","message":"warn test"\}$`,
		},
	}

	buf := bytes.NewBuffer([]byte{})
	mu := sync.Mutex{}
	wantCheckActualLog := true

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mu.Lock()
			defer mu.Unlock()
			buf.Reset()

			zerologger := NewZerologAdapter("testPrefix", buf)

			fields := map[string]interface{}{"user": "Alice", "request_id": "123"}
			tc.outputLog(zerologger, fields)

			logs := buf.String()

			if wantCheckActualLog { // for debugging purpose
				t.Log(logs)
			}

			if tc.wantPattern != "" {
				matched, err := regexp.MatchString(tc.wantPattern, strings.TrimSpace(logs))
				if err != nil {
					t.Fatalf("Error compiling regex pattern: %v", err)
				}
				if !matched {
					t.Errorf("Log does not match expected pattern. got %q, want pattern %q", logs, tc.wantPattern)
				}
			}
		})
	}
}
