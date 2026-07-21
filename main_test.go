package main

import "testing"

func TestFormatTarget(t *testing.T) {
	target := Target{Name: "Test", URL: "https://test.com", Timeout: 91}

	got := formatTarget(target)
	want := "Name: Test\nURL: https://test.com\nTimeout: 91 seconds"

	if got != want {
		t.Fatalf("formatTarget() = %q, want %q", got, want)
	}
}
