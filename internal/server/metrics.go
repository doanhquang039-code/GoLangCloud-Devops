package server

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Metrics struct {
	mu       sync.RWMutex
	requests map[metricKey]requestMetric
	started  time.Time
}

type metricKey struct {
	Method string
	Route  string
	Status int
}

type requestMetric struct {
	Count       int
	DurationSum float64
}

type metricsResponseWriter struct {
	http.ResponseWriter
	status int
}

func NewMetrics() *Metrics {
	return &Metrics{
		requests: make(map[metricKey]requestMetric),
		started:  time.Now().UTC(),
	}
}

func (m *Metrics) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &metricsResponseWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(recorder, r)

		m.record(r.Method, routeLabel(r.URL.Path), recorder.status, time.Since(start))
	})
}

func (m *Metrics) Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	m.mu.RLock()
	keys := make([]metricKey, 0, len(m.requests))
	for key := range m.requests {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].Route == keys[j].Route {
			if keys[i].Method == keys[j].Method {
				return keys[i].Status < keys[j].Status
			}
			return keys[i].Method < keys[j].Method
		}
		return keys[i].Route < keys[j].Route
	})

	fmt.Fprintln(w, "# HELP hr_cloud_service_up Whether the API process is serving metrics.")
	fmt.Fprintln(w, "# TYPE hr_cloud_service_up gauge")
	fmt.Fprintln(w, "hr_cloud_service_up 1")
	fmt.Fprintln(w, "# HELP hr_cloud_service_started_timestamp_seconds Unix timestamp when the API process started.")
	fmt.Fprintln(w, "# TYPE hr_cloud_service_started_timestamp_seconds gauge")
	fmt.Fprintf(w, "hr_cloud_service_started_timestamp_seconds %d\n", m.started.Unix())
	fmt.Fprintln(w, "# HELP http_requests_total Total HTTP requests by method, route, and status.")
	fmt.Fprintln(w, "# TYPE http_requests_total counter")
	for _, key := range keys {
		metric := m.requests[key]
		fmt.Fprintf(w, "http_requests_total{method=%q,route=%q,status=%q} %d\n", key.Method, key.Route, strconv.Itoa(key.Status), metric.Count)
	}
	fmt.Fprintln(w, "# HELP http_request_duration_seconds_sum Total HTTP request duration by method, route, and status.")
	fmt.Fprintln(w, "# TYPE http_request_duration_seconds_sum counter")
	for _, key := range keys {
		metric := m.requests[key]
		fmt.Fprintf(w, "http_request_duration_seconds_sum{method=%q,route=%q,status=%q} %.6f\n", key.Method, key.Route, strconv.Itoa(key.Status), metric.DurationSum)
	}
	fmt.Fprintln(w, "# HELP http_request_duration_seconds_count Total HTTP request duration observations by method, route, and status.")
	fmt.Fprintln(w, "# TYPE http_request_duration_seconds_count counter")
	for _, key := range keys {
		metric := m.requests[key]
		fmt.Fprintf(w, "http_request_duration_seconds_count{method=%q,route=%q,status=%q} %d\n", key.Method, key.Route, strconv.Itoa(key.Status), metric.Count)
	}
	m.mu.RUnlock()
}

func (m *Metrics) record(method string, route string, status int, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := metricKey{Method: method, Route: route, Status: status}
	metric := m.requests[key]
	metric.Count++
	metric.DurationSum += duration.Seconds()
	m.requests[key] = metric
}

func (w *metricsResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func routeLabel(path string) string {
	switch {
	case path == "/healthz", path == "/readyz", path == "/metrics":
		return path
	case path == "/api/v1/platform/summary", path == "/api/v1/platform/scorecards", path == "/api/v1/platform/environment-drift":
		return path
	case path == "/api/v1/employees":
		return path
	case strings.HasPrefix(path, "/api/v1/employees/"):
		return "/api/v1/employees/{id}"
	case path == "/api/v1/applications":
		return path
	case strings.HasPrefix(path, "/api/v1/applications/"):
		return "/api/v1/applications/{id}"
	case path == "/api/v1/clusters":
		return path
	case strings.HasPrefix(path, "/api/v1/clusters/"):
		return "/api/v1/clusters/{id}"
	case path == "/api/v1/environments":
		return path
	case strings.HasPrefix(path, "/api/v1/environments/"):
		return "/api/v1/environments/{id}"
	case path == "/api/v1/deployments":
		return path
	case strings.HasPrefix(path, "/api/v1/deployments/"):
		return "/api/v1/deployments/{id}"
	case path == "/api/v1/pipelines":
		return path
	case strings.HasPrefix(path, "/api/v1/pipelines/"):
		return "/api/v1/pipelines/{id}"
	case path == "/api/v1/incidents":
		return path
	case strings.HasPrefix(path, "/api/v1/incidents/"):
		return "/api/v1/incidents/{id}"
	default:
		return "unknown"
	}
}
