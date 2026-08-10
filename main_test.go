package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestRunExitCode(t *testing.T) {
	testCases := []struct {
		name       string
		statusCode int
		want       int
	}{
		{name: "success", statusCode: http.StatusOK, want: 0},
		{name: "failure", statusCode: http.StatusServiceUnavailable, want: 1},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(
				func(responseWriter http.ResponseWriter, _ *http.Request) {
					responseWriter.WriteHeader(testCase.statusCode)
				},
			))
			defer server.Close()

			configData, err := json.Marshal([]Target{
				{
					Name:    "Local",
					URL:     server.URL,
					Timeout: 1,
				},
			})
			if err != nil {
				t.Fatalf("json.Marshal() error = %v", err)
			}

			configPath := filepath.Join(t.TempDir(), "targets.json")
			if err := os.WriteFile(configPath, configData, 0o600); err != nil {
				t.Fatalf("os.WriteFile() error = %v", err)
			}

			got := run(configPath)
			if got != testCase.want {
				t.Fatalf("run() = %d, want %d", got, testCase.want)
			}
		})
	}
}
func TestProbeTargetInvalidURL(t *testing.T) {
	target := Target{
		Name:    "Invalid",
		URL:     "://invalid",
		Timeout: 1,
	}

	result := probeTarget(target)

	if result.Err == nil {
		t.Fatal("probeTarget() error = nil, want non-nil")
	}
}
func TestProbeTargetStatusCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(responseWriter http.ResponseWriter, _ *http.Request) {
			responseWriter.WriteHeader(http.StatusTeapot)
		},
	))
	defer server.Close()

	target := Target{
		Name:    "Local",
		URL:     server.URL,
		Timeout: 1,
	}

	result := probeTarget(target)

	if result.Err != nil {
		t.Fatalf("probeTarget() returned unexpected error: %v", result.Err)
	}

	got := result.StatusCode
	want := http.StatusTeapot

	if got != want {
		t.Fatalf("probeTarget() status = %d, want %d", got, want)
	}

}
func TestFormatTarget(t *testing.T) {
	target := Target{Name: "Test", URL: "https://test.com", Timeout: 91}

	got := formatTarget(target)
	want := "Name: Test\nURL: https://test.com\nTimeout: 91 seconds"

	if got != want {
		t.Fatalf("formatTarget() = %q, want %q", got, want)
	}
}
