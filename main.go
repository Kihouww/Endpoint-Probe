package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type Target struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	Timeout int    `json:"timeout"`
}

type ProbeResult struct {
	Name       string
	StatusCode int
	Duration   time.Duration
	Err        error
}

func probeAndSend(target Target, results chan<- ProbeResult) {
	result := probeTarget(target)
	results <- result
}

func probeTarget(target Target) ProbeResult {
	client := http.Client{
		Timeout: time.Duration(target.Timeout) * time.Second,
	}

	start := time.Now()
	resp, err := client.Get(target.URL)
	duration := time.Since(start)

	if err != nil {
		return ProbeResult{
			Name:     target.Name,
			Duration: duration,
			Err:      err,
		}
	}

	defer resp.Body.Close()
	return ProbeResult{
		Name:       target.Name,
		StatusCode: resp.StatusCode,
		Duration:   duration,
	}
}

func loadTargets(path string) ([]Target, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var targets []Target
	err = json.Unmarshal(data, &targets)
	if err != nil {
		return nil, err
	}
	return targets, nil
}

func formatTarget(target Target) string {
	return fmt.Sprintf("Name: %s\nURL: %s\nTimeout: %d seconds", target.Name, target.URL, target.Timeout)
}

func main() {
	os.Exit(run("configs/targets.json"))
}

func run(configPath string) int {
	targets, err := loadTargets(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load targets failed: %v\n", err)
		return 1
	}

	results := make(chan ProbeResult)
	totalStart := time.Now()

	for _, target := range targets {
		go probeAndSend(target, results)
	}

	failed := false

	for completed := 0; completed < len(targets); completed++ {
		result := <-results

		if result.Err != nil {
			failed = true
			fmt.Fprintf(
				os.Stderr,
				"%s: error after %v: %v\n",
				result.Name,
				result.Duration.Round(time.Millisecond),
				result.Err,
			)
			continue
		}

		fmt.Printf(
			"%s: status = %d duration = %v\n",
			result.Name,
			result.StatusCode,
			result.Duration.Round(time.Millisecond),
		)

		if result.StatusCode < http.StatusOK ||
			result.StatusCode >= http.StatusMultipleChoices {
			failed = true
		}
	}

	fmt.Printf(
		"total duration: %v\n",
		time.Since(totalStart).Round(time.Millisecond),
	)

	if failed {
		return 1
	}

	return 0
}
