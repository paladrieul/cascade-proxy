package middleware

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestLogger(buf *bytes.Buffer) *log.Logger {
	return log.New(buf, "", 0)
}

func TestLoggerLogsRequest(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf)

	handler := Logger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	output := buf.String()
	if !strings.Contains(output, "GET") {
		t.Errorf("expected log to contain method GET, got: %s", output)
	}
	if !strings.Contains(output, "/health") {
		t.Errorf("expected log to contain path /health, got: %s", output)
	}
	if !strings.Contains(output, "200") {
		t.Errorf("expected log to contain status 200, got: %s", output)
	}
}

func TestLoggerCapturesNon200Status(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf)

	handler := Logger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))

	req := httptest.NewRequest(http.MethodPost, "/proxy", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	output := buf.String()
	if !strings.Contains(output, "502") {
		t.Errorf("expected log to contain status 502, got: %s", output)
	}
}

func TestResponseRecorderDefaultStatus(t *testing.T) {
	rec := httptest.NewRecorder()
	rr := NewResponseRecorder(rec)

	if rr.StatusCode != http.StatusOK {
		t.Errorf("expected default status 200, got %d", rr.StatusCode)
	}
}

func TestResponseRecorderTracksWrittenBytes(t *testing.T) {
	rec := httptest.NewRecorder()
	rr := NewResponseRecorder(rec)

	body := []byte("hello world")
	_, _ = rr.Write(body)

	if rr.Written != int64(len(body)) {
		t.Errorf("expected %d bytes written, got %d", len(body), rr.Written)
	}
}
