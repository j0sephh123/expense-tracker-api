package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

func getCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rows, err := db.Query(`
		SELECT c.id, c.name, s.id, s.name 
		FROM categories c 
		LEFT JOIN subcategories s ON c.id = s.category_id 
		ORDER BY c.id, s.id
	`)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to query categories: %v", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var categories []Category
	var currentCategory *Category

	for rows.Next() {
		var categoryID int
		var categoryName string
		var subcategoryID sql.NullInt64
		var subcategoryName sql.NullString

		if err := rows.Scan(&categoryID, &categoryName, &subcategoryID, &subcategoryName); err != nil {
			logger.Error(fmt.Sprintf("Failed to scan category row: %v", err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if currentCategory == nil || currentCategory.ID != categoryID {
			if currentCategory != nil {
				categories = append(categories, *currentCategory)
			}
			currentCategory = &Category{
				ID:            categoryID,
				Name:          categoryName,
				Subcategories: []Subcategory{},
			}
		}

		if subcategoryID.Valid && subcategoryName.Valid {
			subcategory := Subcategory{
				ID:   int(subcategoryID.Int64),
				Name: subcategoryName.String,
			}
			currentCategory.Subcategories = append(currentCategory.Subcategories, subcategory)
		}
	}

	if currentCategory != nil {
		categories = append(categories, *currentCategory)
	}

	if err = rows.Err(); err != nil {
		logger.Error(fmt.Sprintf("Error iterating over rows: %v", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"status":    "healthy",
		"message":   "API is running",
		"timestamp": "2024-01-01T00:00:00Z",
	}
	json.NewEncoder(w).Encode(response)
}

func getExpensesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userIDStr := r.URL.Query().Get("user_id")
	aggregatesOnlyStr := r.URL.Query().Get("aggregates_only")

	logger.Info(fmt.Sprintf("Received user_id: '%s', aggregates_only: '%s'", userIDStr, aggregatesOnlyStr))

	aggregatesOnly := false
	if aggregatesOnlyStr == "true" {
		aggregatesOnly = true
	}

	if aggregatesOnly {
		query := "SELECT SUM(amount) as total_amount FROM expenses"
		var args []interface{}

		if userIDStr != "" {
			userID, err := strconv.Atoi(userIDStr)
			if err != nil {
				http.Error(w, "Invalid user_id parameter", http.StatusBadRequest)
				return
			}

			if userID == 0 {
				query += " WHERE user_id IS NULL"
			} else {
				query += " WHERE user_id = ?"
				args = append(args, userID)
			}
		}

		var totalAmount sql.NullFloat64
		err := db.QueryRow(query, args...).Scan(&totalAmount)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to query expense total: %v", err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"total_amount": totalAmount.Float64,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	query := "SELECT id, amount, subcategory_id, user_id, note, created_at FROM expenses"
	var args []interface{}

	if userIDStr != "" {
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			http.Error(w, "Invalid user_id parameter", http.StatusBadRequest)
			return
		}

		if userID == 0 {
			query += " WHERE user_id IS NULL"
		} else {
			query += " WHERE user_id = ?"
			args = append(args, userID)
		}
	}

	query += " ORDER BY created_at DESC"

	rows, err := db.Query(query, args...)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to query expenses: %v", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var expenses []Expense

	for rows.Next() {
		var expense Expense
		var userID sql.NullInt64
		var note sql.NullString

		if err := rows.Scan(&expense.ID, &expense.Amount, &expense.SubcategoryID, &userID, &note, &expense.CreatedAt); err != nil {
			logger.Error(fmt.Sprintf("Failed to scan expense row: %v", err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		expense.UserID = userID
		expense.Note = note
		expenses = append(expenses, expense)
	}

	if err = rows.Err(); err != nil {
		logger.Error(fmt.Sprintf("Error iterating over rows: %v", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(expenses)
}
