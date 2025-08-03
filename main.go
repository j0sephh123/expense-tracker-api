package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

var logger *Logger

func main() {
	logger = NewLogger()

	if err := godotenv.Load(); err != nil {
		logger.Info("No .env file found, using default environment variables")
	}

	host := os.Getenv("HOST")
	port := os.Getenv("PORT")

	logger.Info("Starting expense tracker API server")

	if err := initDB(); err != nil {
		logger.Error(fmt.Sprintf("Failed to initialize database: %v", err))
		os.Exit(1)
	}

	logger.Info(fmt.Sprintf("Server running at %s:%s", host, port))

	http.HandleFunc("/api/v1/categories", getCategoriesHandler)
	http.HandleFunc("/api/v1/categories/", getSingleCategoryHandler)
	http.HandleFunc("/api/v1/expenses", getExpensesHandler)
	http.HandleFunc("/api/v1/health", healthCheckHandler)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logger.Info(fmt.Sprintf("Request received: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr))
		fmt.Fprintln(w, "API is running. Use /api/v1/ for versioned endpoints.")
	})

	logger.Info("Server started successfully")
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
		logger.Error(fmt.Sprintf("Server failed to start: %v", err))
		os.Exit(1)
	}
}
