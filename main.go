package main

import (
	"fmt"
	"os"
)

type Target struct {
	Name    string
	URL     string
	Timeout int
}

func readConfig(path string) (string, error) {
	s, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(s), nil
}

func formatTarget(target Target) string {
	return fmt.Sprintf("Name: %s\nURL: %s\nTimeout: %d seconds", target.Name, target.URL, target.Timeout)
}

func main() {
	jsonStr, err := readConfig("configs/targets.json")
	if err != nil {
		fmt.Printf("read config failed: %v\n", err)
		return
	}
	fmt.Printf("Config JSON:\n%s\n\n", jsonStr)

	targets := []Target{
		{Name: "GitHub", URL: "https://github.com", Timeout: 3},
		{Name: "Go", URL: "https://go.dev", Timeout: 2},
		{Name: "Local", URL: "http://localhost:8080", Timeout: 1},
	}

	fmt.Printf("before append: %d\n", len(targets))

	targets = append(targets, Target{Name: "Example", URL: "https://example.com", Timeout: 4})

	fmt.Printf("after append: %d\n\n", len(targets))

	for _, target := range targets {
		fmt.Printf("%s\n\n", formatTarget(target))
	}
}
