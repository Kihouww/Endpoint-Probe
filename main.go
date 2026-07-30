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
	targets, err := loadTargets("configs/targets.json")
	if err != nil {
		fmt.Printf("load targets failed: %v\n", err)
		return
	}

	results := make(chan ProbeResult)
	totalStart := time.Now()
	for _, target := range targets {
		go probeAndSend(target, results)
	}
	for completed := 0; completed < len(targets); completed++ {
		result := <-results

		if result.Err != nil {
			fmt.Printf("%s: error after %v: %v\n",
				result.Name,
				result.Duration.Round(time.Millisecond),
				result.Err,
			)
			continue
		}

		fmt.Printf("%s: status = %d duration = %v\n",
			result.Name,
			result.StatusCode,
			result.Duration.Round(time.Millisecond),
		)
	}

	fmt.Printf(
		"total duration: %v\n",
		time.Since(totalStart).Round(time.Millisecond),
	)
}
