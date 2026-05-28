package server

import "testing"

func TestRouteLabelKeepsMetricCardinalityLow(t *testing.T) {
	cases := map[string]string{
		"/healthz":                           "/healthz",
		"/api/v1/employees":                  "/api/v1/employees",
		"/api/v1/employees/emp-001":          "/api/v1/employees/{id}",
		"/api/v1/platform/scorecards":        "/api/v1/platform/scorecards",
		"/api/v1/platform/environment-drift": "/api/v1/platform/environment-drift",
		"/api/v1/incidents/inc-payroll":      "/api/v1/incidents/{id}",
		"/not-found":                         "unknown",
	}

	for path, expected := range cases {
		if got := routeLabel(path); got != expected {
			t.Fatalf("routeLabel(%q) = %q, want %q", path, got, expected)
		}
	}
}
