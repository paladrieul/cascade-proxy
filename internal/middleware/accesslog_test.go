package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestAccessLogMiddleware(t *testing.T) (*bytes.Buffer, func(http.Handler) http.Handler) {
	t.Helper()
	buf := &bytes.Buffer{}
	logger := slog.New(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	cfg := DefaultAccessLogConfig(logger)
	return buf, NewAccessLogMiddleware(cfg)
}

func TestAccessLogWritesEntryForRequest(t *testing.T) {
	buf, mw := newTestAccessLogMiddleware(t)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("hello"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	if !strings.Contains(buf.String(), "/api/test") {
		t.Errorf("expected log to contain path, got: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "200") {
		t.Errorf("expected log to contain status 200, got: %s", buf.String())
	}
}

func TestAccessLogSkipsHealthPath(t *testing.T) {
	buf, mw := newTestAccessLogMiddleware(t)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	if buf.Len() != 0 {
		t.Errorf("expected no log output for skipped path, got: %s", buf.String())
	}
}

func TestAccessLogUsesWarnLevelFor4xx(t *testing.T) {
	buf, mw := newTestAccessLogMiddleware(t)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse log JSON: %v", err)
	}
	if entry["level"] != "WARN" {
		t.Errorf("expected WARN level for 4xx, got: %v", entry["level"])
	}
}

func TestAccessLogUsesErrorLevelFor5xx(t *testing.T) {
	buf, mw := newTestAccessLogMiddleware(t)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	req := httptest.NewRequest(http.MethodGet, "/boom", nil)
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("failed to parse log JSON: %v", err)
	}
	if entry["level"] != "ERROR" {
		t.Errorf("expected ERROR level for 5xx, got: %v", entry["level"])
	}
}
