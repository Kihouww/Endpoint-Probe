package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

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
