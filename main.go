package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

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
	if port == "" {
		port = "8082"
	}

	logger.Info("Starting expense tracker API server")

	if err := initDB(); err != nil {
		logger.Error(fmt.Sprintf("Failed to initialize database: %v", err))
		os.Exit(1)
	}

	logger.Info(fmt.Sprintf("Server running at %s:%s", host, port))

	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/", func(w http.ResponseWriter, r *http.Request) {
		logger.Info(fmt.Sprintf("API route called: %s %s", r.Method, r.URL.Path))

		if strings.HasPrefix(r.URL.Path, "/api/v1/categories") {
			if r.URL.Path == "/api/v1/categories" {
				if r.Method == http.MethodGet {
					logger.Info("Calling getCategoriesHandler")
					getCategoriesHandler(w, r)
				} else if r.Method == http.MethodPost {
					logger.Info("Calling createCategoryHandler")
					createCategoryHandler(w, r)
				} else {
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				}
			} else if strings.HasPrefix(r.URL.Path, "/api/v1/categories/") {
				logger.Info("Calling single category handler")
				path := strings.TrimPrefix(r.URL.Path, "/api/v1/categories/")
				if path != "" {
					if r.Method == http.MethodPut || r.Method == http.MethodPatch {
						updateCategoryHandler(w, r)
					} else if r.Method == http.MethodGet {
						getSingleCategoryHandler(w, r)
					} else {
						http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
					}
				} else {
					getCategoriesHandler(w, r)
				}
			}
		} else if strings.HasPrefix(r.URL.Path, "/api/v1/subcategories") {
			if r.URL.Path == "/api/v1/subcategories" {
				if r.Method == http.MethodPost {
					logger.Info("Calling createSubcategoryHandler")
					createSubcategoryHandler(w, r)
				} else {
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				}
			} else if strings.HasPrefix(r.URL.Path, "/api/v1/subcategories/") {
				if r.Method == http.MethodGet {
					getSingleSubcategoryHandler(w, r)
				} else if r.Method == http.MethodPut || r.Method == http.MethodPatch {
					updateSubcategoryHandler(w, r)
				} else {
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				}
			}
		} else if strings.HasPrefix(r.URL.Path, "/api/v1/expenses/") {
			path := strings.TrimPrefix(r.URL.Path, "/api/v1/expenses/")
			if path != "" {
				if r.Method == http.MethodDelete {
					deleteExpenseHandler(w, r)
				} else {
					http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				}
			} else {
				http.Error(w, "Expense ID is required", http.StatusBadRequest)
			}
		} else if r.URL.Path == "/api/v1/expenses" {
			if r.Method == http.MethodGet {
				getExpensesHandler(w, r)
			} else if r.Method == http.MethodPost {
				createExpenseHandler(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else if r.URL.Path == "/api/v1/subcategories-by-expense-count" {
			debugSubcategoriesByExpenseCountHandler(w, r)
		} else if r.URL.Path == "/api/v1/health" {
			healthCheckHandler(w, r)
		} else {
			logger.Info("No matching route found")
			http.Error(w, "Not found", http.StatusNotFound)
		}
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		logger.Info(fmt.Sprintf("Request received: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr))
		fmt.Fprintln(w, "API is running. Use /api/v1/ for versioned endpoints.")
	})

	logger.Info("Server started successfully")
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), mux); err != nil {
		logger.Error(fmt.Sprintf("Server failed to start: %v", err))
		os.Exit(1)
	}
}
