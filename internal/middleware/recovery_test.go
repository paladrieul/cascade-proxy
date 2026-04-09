package middleware_test

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/user/cascade-proxy/internal/middleware"
)

func newTestRecoveryMiddleware(buf *bytes.Buffer) func(http.Handler) http.Handler {
	logger := slog.New(slog.NewTextHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
	return middleware.NewRecoveryMiddleware(logger)
}

func TestRecoveryReturns500OnPanic(t *testing.T) {
	var buf bytes.Buffer
	mw := newTestRecoveryMiddleware(&buf)

	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("something went wrong")
	})

	req := httptest.NewRequest(http.MethodGet, "/crash", nil)
	rec := httptest.NewRecorder()
	mw(panicHandler).ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Internal Server Error") {
		t.Errorf("expected body to contain 'Internal Server Error', got %q", rec.Body.String())
	}
}

func TestRecoveryLogsStackTrace(t *testing.T) {
	var buf bytes.Buffer
	mw := newTestRecoveryMiddleware(&buf)

	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})

	req := httptest.NewRequest(http.MethodPost, "/explode", nil)
	rec := httptest.NewRecorder()
	mw(panicHandler).ServeHTTP(rec, req)

	logOutput := buf.String()
	if !strings.Contains(logOutput, "panic recovered") {
		t.Errorf("expected log to contain 'panic recovered', got %q", logOutput)
	}
	if !strings.Contains(logOutput, "boom") {
		t.Errorf("expected log to contain panic value 'boom', got %q", logOutput)
	}
}

func TestRecoveryPassesThroughNormalRequest(t *testing.T) {
	var buf bytes.Buffer
	mw := newTestRecoveryMiddleware(&buf)

	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	req := httptest.NewRequest(http.MethodGet, "/healthy", nil)
	rec := httptest.NewRecorder()
	mw(okHandler).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if buf.Len() != 0 {
		t.Errorf("expected no log output for normal request, got %q", buf.String())
	}
}
