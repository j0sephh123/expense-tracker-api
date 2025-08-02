package main

import (
	"fmt"
	"net/http"
	"os"
)

var logger *Logger

func main() {
	logger = NewLogger()

	logger.Info("Starting expense tracker API server")
	logger.Info("Server running at http://localhost:8082")

	http.HandleFunc("/api/v1/", func(w http.ResponseWriter, r *http.Request) {
		logger.Info(fmt.Sprintf("Request received: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr))
		fmt.Fprintln(w, "Hello, world123!")
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logger.Info(fmt.Sprintf("Request received: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr))
		fmt.Fprintln(w, "API is running. Use /api/v1/ for versioned endpoints.")
	})

	logger.Info("Server started successfully")
	if err := http.ListenAndServe(":8082", nil); err != nil {
		logger.Error(fmt.Sprintf("Server failed to start: %v", err))
		os.Exit(1)
	}
}
