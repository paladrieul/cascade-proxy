package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestBreachLogIntegrationWithLogger verifies that breach log middleware
// composes correctly with the request logger middleware.
func TestBreachLogIntegrationWithLogger(t *testing.T) {
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	breachCfg := DefaultBreachLogConfig(logger)
	breachCfg.WindowSize = 4
	breachCfg.ErrorThreshold = 0.5

	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	})

	handler := Logger(logger)(NewBreachLogMiddleware(breachCfg)(backend))

	for i := 0; i < 4; i++ {
		req := httptest.NewRequest(http.MethodGet, "/downstream", nil)
		handler.ServeHTTP(httptest.NewRecorder(), req)
	}

	logs := logBuf.String()
	if !strings.Contains(logs, "breach") {
		t.Errorf("expected breach warning in logs, got: %s", logs)
	}
	if !strings.Contains(logs, "status") {
		t.Errorf("expected logger to record status, got: %s", logs)
	}
}

// TestBreachLogIntegrationWithRequestID verifies that breach log middleware
// works correctly when composed after request ID middleware.
func TestBreachLogIntegrationWithRequestID(t *testing.T) {
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logBuf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	breachCfg := DefaultBreachLogConfig(logger)
	breachCfg.WindowSize = 4
	breachCfg.ErrorThreshold = 0.5
	breachCfg.KeyFunc = func(r *http.Request) string {
		if id := RequestIDFromContext(r.Context()); id != "" {
			return "/traced"
		}
		return r.URL.Path
	}

	backend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	handler := NewRequestIDMiddleware()(NewBreachLogMiddleware(breachCfg)(backend))

	for i := 0; i < 4; i++ {
		req := httptest.NewRequest(http.MethodGet, "/traced", nil)
		handler.ServeHTTP(httptest.NewRecorder(), req)
	}

	if !strings.Contains(logBuf.String(), "breach") {
		t.Errorf("expected breach log entry, got: %s", logBuf.String())
	}
}
