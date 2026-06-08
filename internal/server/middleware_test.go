package server

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWithRequestLoggingIncludesStatusAndResponseSize(t *testing.T) {
	var logs bytes.Buffer
	originalOutput := log.Writer()
	log.SetOutput(&logs)
	defer log.SetOutput(originalOutput)

	handler := WithRequestLogging(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		if _, err := w.Write([]byte("queued")); err != nil {
			t.Fatal(err)
		}
	}))

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/v1/deployments", nil))

	logLine := logs.String()
	for _, expected := range []string{"POST", "/api/v1/deployments", "202", "6B"} {
		if !strings.Contains(logLine, expected) {
			t.Fatalf("expected log line to contain %q, got %q", expected, logLine)
		}
	}
}

func TestWithRequestLoggingDefaultsStatusToOK(t *testing.T) {
	var logs bytes.Buffer
	originalOutput := log.Writer()
	log.SetOutput(&logs)
	defer log.SetOutput(originalOutput)

	handler := WithRequestLogging(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if !strings.Contains(logs.String(), "200") {
		t.Fatalf("expected default status 200 in log line, got %q", logs.String())
	}
}
