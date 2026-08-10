package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"time"
)

type PredictRequest struct {
	Features []float64 `json:"features"`
}

type PredictResponse struct {
	Class       int     `json:"class"`
	Probability float64 `json:"probability"`
	Version     string  `json:"version"`
}

var weights = []float64{0.8, -0.4, 0.6}

const bias = -0.2

func predictHandler(version string) http.HandlerFunc {
	return func(
		responseWriter http.ResponseWriter,
		request *http.Request,
	) {
		var input PredictRequest
		if err := json.NewDecoder(request.Body).Decode(&input); err != nil {
			http.Error(responseWriter, "invalid JSON", http.StatusBadRequest)
			return
		}

		if len(input.Features) != len(weights) {
			http.Error(responseWriter, "need exactly 3 features", http.StatusBadRequest)
			return
		}

		score := bias
		for index, feature := range input.Features {
			score += weights[index] * feature
		}

		probability := 1 / (1 + math.Exp(-score))
		class := 0
		if probability >= 0.5 {
			class = 1
		}

		responseWriter.Header().Set("Content-Type", "application/json")
		json.NewEncoder(responseWriter).Encode(PredictResponse{
			Class:       class,
			Probability: probability,
			Version:     version,
		})
	}
}

func main() {
	version := os.Getenv("APP_VERSION")
	if version == "" {
		version = "local"
	}

	mux := http.NewServeMux()

	mux.HandleFunc("POST /predict", predictHandler(version))

	mux.HandleFunc("GET /healthz", func(
		responseWriter http.ResponseWriter,
		_ *http.Request,
	) {
		responseWriter.WriteHeader(http.StatusOK)
		fmt.Fprintln(responseWriter, "ok")
	})

	mux.HandleFunc("GET /version", func(
		responseWriter http.ResponseWriter,
		_ *http.Request,
	) {
		fmt.Fprintln(responseWriter, version)
	})

	server := http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("predictor version=%s listening on :8080", version)
	log.Fatal(server.ListenAndServe())
}
