package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
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
