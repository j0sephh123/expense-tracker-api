package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
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
	if categories == nil {
		categories = []Category{}
	}
	json.NewEncoder(w).Encode(categories)
}

func getSingleCategoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/categories/")

	if path == "" {
		http.Error(w, "Category ID is required", http.StatusBadRequest)
		return
	}

	categoryID, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	rows, err := db.Query(`
		SELECT c.id, c.name, s.id, s.name 
		FROM categories c 
		LEFT JOIN subcategories s ON c.id = s.category_id 
		WHERE c.id = ?
		ORDER BY s.id
	`, categoryID)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to query category: %v", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var category *Category

	for rows.Next() {
		var catID int
		var catName string
		var subcategoryID sql.NullInt64
		var subcategoryName sql.NullString

		if err := rows.Scan(&catID, &catName, &subcategoryID, &subcategoryName); err != nil {
			logger.Error(fmt.Sprintf("Failed to scan category row: %v", err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if category == nil {
			category = &Category{
				ID:            catID,
				Name:          catName,
				Subcategories: []Subcategory{},
			}
		}

		if subcategoryID.Valid && subcategoryName.Valid {
			subcategory := Subcategory{
				ID:   int(subcategoryID.Int64),
				Name: subcategoryName.String,
			}
			category.Subcategories = append(category.Subcategories, subcategory)
		}
	}

	if err = rows.Err(); err != nil {
		logger.Error(fmt.Sprintf("Error iterating over rows: %v", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if category == nil {
		http.Error(w, "Category not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(category)
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

func parseCommaSeparatedInts(s string) ([]int, error) {
	if s == "" {
		return []int{}, nil
	}

	parts := strings.Split(s, ",")
	result := make([]int, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		val, err := strconv.Atoi(part)
		if err != nil {
			return nil, fmt.Errorf("invalid integer: %s", part)
		}
		result = append(result, val)
	}

	return result, nil
}

func getExpensesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userIDStr := r.URL.Query().Get("user_id")
	categoryIDStr := r.URL.Query().Get("category_id")
	subcategoryIDStr := r.URL.Query().Get("subcategory_id")
	dateFromStr := r.URL.Query().Get("date_from")
	dateToStr := r.URL.Query().Get("date_to")
	groupByStr := r.URL.Query().Get("group_by")
	orderByStr := r.URL.Query().Get("order_by")
	orderDirStr := r.URL.Query().Get("order_dir")
	aggregatesOnlyStr := r.URL.Query().Get("aggregates_only")

	logger.Info(fmt.Sprintf("Received user_id: '%s', category_id: '%s', subcategory_id: '%s', date_from: '%s', date_to: '%s', group_by: '%s', order_by: '%s', order_dir: '%s', aggregates_only: '%s'", userIDStr, categoryIDStr, subcategoryIDStr, dateFromStr, dateToStr, groupByStr, orderByStr, orderDirStr, aggregatesOnlyStr))

	if categoryIDStr != "" && subcategoryIDStr != "" {
		http.Error(w, "Cannot use both category_id and subcategory_id in the same query", http.StatusBadRequest)
		return
	}

	orderBy := "created_at"
	if orderByStr != "" {
		switch orderByStr {
		case "amount":
			orderBy = "e.amount"
		case "date":
			orderBy = "e.created_at"
		default:
			http.Error(w, "Invalid order_by parameter. Must be 'amount' or 'date'", http.StatusBadRequest)
			return
		}
	}

	orderDir := "DESC"
	if orderDirStr != "" {
		switch orderDirStr {
		case "asc":
			orderDir = "ASC"
		case "desc":
			orderDir = "DESC"
		default:
			http.Error(w, "Invalid order_dir parameter. Must be 'asc' or 'desc'", http.StatusBadRequest)
			return
		}
	}

	if dateFromStr != "" {
		if _, err := time.Parse("2006-01-02", dateFromStr); err != nil {
			http.Error(w, "Invalid date_from parameter. Must be in YYYY-MM-DD format", http.StatusBadRequest)
			return
		}
	}

	if dateToStr != "" {
		if _, err := time.Parse("2006-01-02", dateToStr); err != nil {
			http.Error(w, "Invalid date_to parameter. Must be in YYYY-MM-DD format", http.StatusBadRequest)
			return
		}
	}

	if groupByStr != "" {
		switch groupByStr {
		case "category", "subcategory", "user":
			// Valid group_by values
		default:
			http.Error(w, "Invalid group_by parameter. Must be 'category', 'subcategory', or 'user'", http.StatusBadRequest)
			return
		}
	}

	aggregatesOnly := false
	if aggregatesOnlyStr == "true" {
		aggregatesOnly = true
	}

	if groupByStr != "" {
		handleGroupedExpenses(w, userIDStr, categoryIDStr, subcategoryIDStr, dateFromStr, dateToStr, groupByStr, orderBy, orderDir)
		return
	}

	if aggregatesOnly {
		query := "SELECT SUM(e.amount) as total_amount FROM expenses e"
		var args []interface{}
		var conditions []string

		if userIDStr != "" {
			userID, err := strconv.Atoi(userIDStr)
			if err != nil {
				http.Error(w, "Invalid user_id parameter", http.StatusBadRequest)
				return
			}

			if userID == 0 {
				conditions = append(conditions, "e.user_id IS NULL")
			} else {
				conditions = append(conditions, "e.user_id = ?")
				args = append(args, userID)
			}
		}

		if categoryIDStr != "" {
			logger.Info(fmt.Sprintf("Processing category_id: '%s'", categoryIDStr))
			categoryIDs, err := parseCommaSeparatedInts(categoryIDStr)
			if err != nil {
				logger.Error(fmt.Sprintf("Failed to parse category_id: %v", err))
				http.Error(w, "Invalid category_id parameter", http.StatusBadRequest)
				return
			}

			logger.Info(fmt.Sprintf("Parsed category IDs: %v", categoryIDs))

			if len(categoryIDs) > 0 {
				query += " JOIN subcategories sc ON e.subcategory_id = sc.id"
				placeholders := make([]string, len(categoryIDs))
				for i := range categoryIDs {
					placeholders[i] = "?"
					args = append(args, categoryIDs[i])
				}
				condition := fmt.Sprintf("sc.category_id IN (%s)", strings.Join(placeholders, ","))
				conditions = append(conditions, condition)
				logger.Info(fmt.Sprintf("Added category condition: %s", condition))
			}
		}

		if subcategoryIDStr != "" {
			logger.Info(fmt.Sprintf("Processing subcategory_id: '%s'", subcategoryIDStr))
			subcategoryIDs, err := parseCommaSeparatedInts(subcategoryIDStr)
			if err != nil {
				logger.Error(fmt.Sprintf("Failed to parse subcategory_id: %v", err))
				http.Error(w, "Invalid subcategory_id parameter", http.StatusBadRequest)
				return
			}

			logger.Info(fmt.Sprintf("Parsed subcategory IDs: %v", subcategoryIDs))

			if len(subcategoryIDs) > 0 {
				placeholders := make([]string, len(subcategoryIDs))
				for i := range subcategoryIDs {
					placeholders[i] = "?"
					args = append(args, subcategoryIDs[i])
				}
				condition := fmt.Sprintf("e.subcategory_id IN (%s)", strings.Join(placeholders, ","))
				conditions = append(conditions, condition)
				logger.Info(fmt.Sprintf("Added subcategory condition: %s", condition))
			}
		}

		if dateFromStr != "" {
			conditions = append(conditions, "DATE(e.created_at) >= ?")
			args = append(args, dateFromStr)
		}

		if dateToStr != "" {
			conditions = append(conditions, "DATE(e.created_at) <= ?")
			args = append(args, dateToStr)
		}

		if len(conditions) > 0 {
			query += " WHERE " + strings.Join(conditions, " AND ")
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

	query := `
		SELECT 
			e.id, 
			e.amount, 
			e.subcategory_id, 
			e.user_id, 
			e.note, 
			e.created_at,
			u.email as user_email,
			s.name as subcategory_name,
			c.id as category_id,
			c.name as category_name
		FROM expenses e
		LEFT JOIN users u ON e.user_id = u.id
		LEFT JOIN subcategories s ON e.subcategory_id = s.id
		LEFT JOIN categories c ON s.category_id = c.id
	`
	var args []interface{}
	var conditions []string

	if userIDStr != "" {
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			http.Error(w, "Invalid user_id parameter", http.StatusBadRequest)
			return
		}

		if userID == 0 {
			conditions = append(conditions, "e.user_id IS NULL")
		} else {
			conditions = append(conditions, "e.user_id = ?")
			args = append(args, userID)
		}
	}

	if categoryIDStr != "" {
		categoryIDs, err := parseCommaSeparatedInts(categoryIDStr)
		if err != nil {
			http.Error(w, "Invalid category_id parameter", http.StatusBadRequest)
			return
		}

		if len(categoryIDs) > 0 {
			query += " JOIN subcategories sc ON e.subcategory_id = sc.id"
			placeholders := make([]string, len(categoryIDs))
			for i := range categoryIDs {
				placeholders[i] = "?"
				args = append(args, categoryIDs[i])
			}
			conditions = append(conditions, fmt.Sprintf("sc.category_id IN (%s)", strings.Join(placeholders, ",")))
		}
	}

	if subcategoryIDStr != "" {
		subcategoryIDs, err := parseCommaSeparatedInts(subcategoryIDStr)
		if err != nil {
			http.Error(w, "Invalid subcategory_id parameter", http.StatusBadRequest)
			return
		}

		if len(subcategoryIDs) > 0 {
			placeholders := make([]string, len(subcategoryIDs))
			for i := range subcategoryIDs {
				placeholders[i] = "?"
				args = append(args, subcategoryIDs[i])
			}
			conditions = append(conditions, fmt.Sprintf("e.subcategory_id IN (%s)", strings.Join(placeholders, ",")))
		}
	}

	if dateFromStr != "" {
		conditions = append(conditions, "DATE(e.created_at) >= ?")
		args = append(args, dateFromStr)
	}

	if dateToStr != "" {
		conditions = append(conditions, "DATE(e.created_at) <= ?")
		args = append(args, dateToStr)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += fmt.Sprintf(" ORDER BY %s %s", orderBy, orderDir)

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
		var userEmail sql.NullString
		var subcategoryName sql.NullString
		var categoryID sql.NullInt64
		var categoryName sql.NullString

		if err := rows.Scan(
			&expense.ID,
			&expense.Amount,
			&expense.SubcategoryID,
			&userID,
			&note,
			&expense.CreatedAt,
			&userEmail,
			&subcategoryName,
			&categoryID,
			&categoryName,
		); err != nil {
			logger.Error(fmt.Sprintf("Failed to scan expense row: %v", err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if userID.Valid {
			userIDValue := int(userID.Int64)
			expense.UserID = &userIDValue
		} else {
			expense.UserID = nil
		}

		if note.Valid {
			expense.Note = &note.String
		} else {
			expense.Note = nil
		}

		if userEmail.Valid {
			expense.UserEmail = &userEmail.String
		}

		if subcategoryName.Valid {
			expense.SubcategoryName = &subcategoryName.String
		}

		if categoryID.Valid {
			categoryIDValue := int(categoryID.Int64)
			expense.CategoryID = &categoryIDValue
		}

		if categoryName.Valid {
			expense.CategoryName = &categoryName.String
		}

		expenses = append(expenses, expense)
	}

	if err = rows.Err(); err != nil {
		logger.Error(fmt.Sprintf("Error iterating over rows: %v", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if expenses == nil {
		expenses = []Expense{}
	}
	json.NewEncoder(w).Encode(expenses)
}

func handleGroupedExpenses(w http.ResponseWriter, userIDStr, categoryIDStr, subcategoryIDStr, dateFromStr, dateToStr, groupByStr, orderBy, orderDir string) {
	var query string
	var args []interface{}
	var conditions []string

	switch groupByStr {
	case "category":
		query = `
			SELECT 
				c.id as category_id,
				c.name as category_name,
				SUM(e.amount) as total_amount,
				COUNT(*) as expense_count
			FROM expenses e
			JOIN subcategories s ON e.subcategory_id = s.id
			JOIN categories c ON s.category_id = c.id
		`
	case "subcategory":
		query = `
			SELECT 
				s.id as subcategory_id,
				s.name as subcategory_name,
				SUM(e.amount) as total_amount,
				COUNT(*) as expense_count
			FROM expenses e
			JOIN subcategories s ON e.subcategory_id = s.id
		`
	case "user":
		query = `
			SELECT 
				COALESCE(e.user_id, 0) as user_id,
				COALESCE(u.display_name, 'Unknown User') as user_name,
				SUM(e.amount) as total_amount,
				COUNT(*) as expense_count
			FROM expenses e
			LEFT JOIN users u ON e.user_id = u.id
		`
	}

	if userIDStr != "" {
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			http.Error(w, "Invalid user_id parameter", http.StatusBadRequest)
			return
		}

		if userID == 0 {
			conditions = append(conditions, "e.user_id IS NULL")
		} else {
			conditions = append(conditions, "e.user_id = ?")
			args = append(args, userID)
		}
	}

	if categoryIDStr != "" {
		categoryIDs, err := parseCommaSeparatedInts(categoryIDStr)
		if err != nil {
			http.Error(w, "Invalid category_id parameter", http.StatusBadRequest)
			return
		}

		if len(categoryIDs) > 0 {
			placeholders := make([]string, len(categoryIDs))
			for i := range categoryIDs {
				placeholders[i] = "?"
				args = append(args, categoryIDs[i])
			}
			conditions = append(conditions, fmt.Sprintf("s.category_id IN (%s)", strings.Join(placeholders, ",")))
		}
	}

	if subcategoryIDStr != "" {
		subcategoryIDs, err := parseCommaSeparatedInts(subcategoryIDStr)
		if err != nil {
			http.Error(w, "Invalid subcategory_id parameter", http.StatusBadRequest)
			return
		}

		if len(subcategoryIDs) > 0 {
			placeholders := make([]string, len(subcategoryIDs))
			for i := range subcategoryIDs {
				placeholders[i] = "?"
				args = append(args, subcategoryIDs[i])
			}
			conditions = append(conditions, fmt.Sprintf("e.subcategory_id IN (%s)", strings.Join(placeholders, ",")))
		}
	}

	if dateFromStr != "" {
		conditions = append(conditions, "DATE(e.created_at) >= ?")
		args = append(args, dateFromStr)
	}

	if dateToStr != "" {
		conditions = append(conditions, "DATE(e.created_at) <= ?")
		args = append(args, dateToStr)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " GROUP BY "
	switch groupByStr {
	case "category":
		query += "c.id, c.name"
	case "subcategory":
		query += "s.id, s.name"
	case "user":
		query += "e.user_id, u.display_name"
	}

	query += fmt.Sprintf(" ORDER BY total_amount %s", orderDir)

	rows, err := db.Query(query, args...)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to query grouped expenses: %v", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var groupedExpenses []map[string]interface{}

	for rows.Next() {
		var groupName string
		var totalAmount float64
		var expenseCount int
		var groupID int

		if err := rows.Scan(&groupID, &groupName, &totalAmount, &expenseCount); err != nil {
			logger.Error(fmt.Sprintf("Failed to scan grouped expense row: %v", err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		group := map[string]interface{}{
			"group_name": groupName,
			"total":      totalAmount,
			"count":      expenseCount,
		}

		groupedExpenses = append(groupedExpenses, group)
	}

	if err = rows.Err(); err != nil {
		logger.Error(fmt.Sprintf("Error iterating over grouped rows: %v", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if groupedExpenses == nil {
		groupedExpenses = []map[string]interface{}{}
	}
	json.NewEncoder(w).Encode(groupedExpenses)
}

func updateCategoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/categories/")
	if path == "" {
		http.Error(w, "Category ID is required", http.StatusBadRequest)
		return
	}

	categoryID, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	var requestBody struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if requestBody.Name == "" {
		http.Error(w, "Category name is required", http.StatusBadRequest)
		return
	}

	var existingID int
	err = db.QueryRow("SELECT id FROM categories WHERE name = ? AND id != ?", requestBody.Name, categoryID).Scan(&existingID)
	if err == nil {
		http.Error(w, "Category name already exists", http.StatusConflict)
		return
	} else if err != sql.ErrNoRows {
		logger.Error(fmt.Sprintf("Failed to check category name uniqueness: %v", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	result, err := db.Exec("UPDATE categories SET name = ? WHERE id = ?", requestBody.Name, categoryID)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to update category: %v", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get rows affected: %v", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Category not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Category updated successfully",
		"id":      categoryID,
		"name":    requestBody.Name,
	})
}

func getSingleSubcategoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/subcategories/")

	if path == "" {
		http.Error(w, "Subcategory ID is required", http.StatusBadRequest)
		return
	}

	subcategoryID, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid subcategory ID", http.StatusBadRequest)
		return
	}

	var subcategory Subcategory
	var categoryID int
	var categoryName string

	err = db.QueryRow(`
		SELECT s.id, s.name, s.category_id, c.name as category_name
		FROM subcategories s
		JOIN categories c ON s.category_id = c.id
		WHERE s.id = ?
	`, subcategoryID).Scan(&subcategory.ID, &subcategory.Name, &categoryID, &categoryName)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Subcategory not found", http.StatusNotFound)
			return
		}
		logger.Error(fmt.Sprintf("Failed to query subcategory: %v", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"id":            subcategory.ID,
		"name":          subcategory.Name,
		"category_id":   categoryID,
		"category_name": categoryName,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func updateSubcategoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/subcategories/")
	if path == "" {
		http.Error(w, "Subcategory ID is required", http.StatusBadRequest)
		return
	}

	subcategoryID, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid subcategory ID", http.StatusBadRequest)
		return
	}

	var requestBody struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if requestBody.Name == "" {
		http.Error(w, "Subcategory name is required", http.StatusBadRequest)
		return
	}

	var existingID int
	err = db.QueryRow("SELECT id FROM subcategories WHERE name = ? AND category_id = (SELECT category_id FROM subcategories WHERE id = ?) AND id != ?", requestBody.Name, subcategoryID, subcategoryID).Scan(&existingID)
	if err == nil {
		http.Error(w, "Subcategory name already exists in this category", http.StatusConflict)
		return
	} else if err != sql.ErrNoRows {
		logger.Error(fmt.Sprintf("Failed to check subcategory name uniqueness: %v", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	result, err := db.Exec("UPDATE subcategories SET name = ? WHERE id = ?", requestBody.Name, subcategoryID)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to update subcategory: %v", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get rows affected: %v", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Subcategory not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Subcategory updated successfully",
		"id":      subcategoryID,
		"name":    requestBody.Name,
	})
}

func createCategoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var requestBody struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if len(requestBody.Name) < 3 {
		http.Error(w, "Category name must be at least 3 characters long", http.StatusBadRequest)
		return
	}

	var existingID int
	err := db.QueryRow("SELECT id FROM categories WHERE name = ?", requestBody.Name).Scan(&existingID)
	if err == nil {
		http.Error(w, "Category name already exists", http.StatusConflict)
		return
	} else if err != sql.ErrNoRows {
		logger.Error(fmt.Sprintf("Failed to check category name uniqueness: %v", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	result, err := db.Exec("INSERT INTO categories (name) VALUES (?)", requestBody.Name)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to create category: %v", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	categoryID, err := result.LastInsertId()
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get last insert ID: %v", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"id":      categoryID,
		"name":    requestBody.Name,
		"message": "Category created successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func createSubcategoryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var requestBody struct {
		Name       string `json:"name"`
		CategoryID int    `json:"category_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if len(requestBody.Name) < 3 {
		http.Error(w, "Subcategory name must be at least 3 characters long", http.StatusBadRequest)
		return
	}

	if requestBody.CategoryID <= 0 {
		http.Error(w, "Valid category_id is required", http.StatusBadRequest)
		return
	}

	var categoryExists int
	err := db.QueryRow("SELECT id FROM categories WHERE id = ?", requestBody.CategoryID).Scan(&categoryExists)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Category not found", http.StatusNotFound)
			return
		}
		logger.Error(fmt.Sprintf("Failed to check category existence: %v", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var existingID int
	err = db.QueryRow("SELECT id FROM subcategories WHERE name = ? AND category_id = ?", requestBody.Name, requestBody.CategoryID).Scan(&existingID)
	if err == nil {
		http.Error(w, "Subcategory name already exists in this category", http.StatusConflict)
		return
	} else if err != sql.ErrNoRows {
		logger.Error(fmt.Sprintf("Failed to check subcategory name uniqueness: %v", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	result, err := db.Exec("INSERT INTO subcategories (name, category_id) VALUES (?, ?)", requestBody.Name, requestBody.CategoryID)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to create subcategory: %v", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	subcategoryID, err := result.LastInsertId()
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get last insert ID: %v", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"id":          subcategoryID,
		"name":        requestBody.Name,
		"category_id": requestBody.CategoryID,
		"message":     "Subcategory created successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

type CreateExpenseRequest struct {
	Amount        float64 `json:"amount"`
	SubcategoryID *int    `json:"subcategory_id,omitempty"`
	UserID        *int    `json:"user_id,omitempty"`
	Note          *string `json:"note,omitempty"`
}

func createExpenseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error(fmt.Sprintf("Failed to decode request body: %v", err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Amount <= 0 {
		http.Error(w, "Amount must be greater than 0", http.StatusBadRequest)
		return
	}

	var subcategoryID sql.NullInt64
	if req.SubcategoryID != nil {
		subcategoryID.Int64 = int64(*req.SubcategoryID)
		subcategoryID.Valid = true
	}

	var userID sql.NullInt64
	if req.UserID != nil {
		userID.Int64 = int64(*req.UserID)
		userID.Valid = true
	}

	var note sql.NullString
	if req.Note != nil {
		note.String = *req.Note
		note.Valid = true
	}

	result, err := db.Exec(`
		INSERT INTO expenses (amount, subcategory_id, user_id, note, created_at)
		VALUES (?, ?, ?, ?, NOW())
	`, req.Amount, subcategoryID, userID, note)

	if err != nil {
		logger.Error(fmt.Sprintf("Failed to create expense: %v", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	expenseID, err := result.LastInsertId()
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get last insert ID: %v", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"id":      expenseID,
		"amount":  req.Amount,
		"message": "Expense created successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

type SubcategoryExpenseCount struct {
	SubcategoryID   int    `json:"subcategory_id"`
	SubcategoryName string `json:"subcategory_name"`
	CategoryID      int    `json:"category_id"`
	CategoryName    string `json:"category_name"`
	ExpenseCount    int    `json:"expense_count"`
}

func debugSubcategoriesByExpenseCountHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	rows, err := db.Query(`
		SELECT 
			s.id as subcategory_id,
			s.name as subcategory_name,
			c.id as category_id,
			c.name as category_name,
			COUNT(e.id) as expense_count
		FROM subcategories s
		LEFT JOIN categories c ON s.category_id = c.id
		LEFT JOIN expenses e ON s.id = e.subcategory_id
		GROUP BY s.id, s.name, c.id, c.name
		ORDER BY expense_count DESC
	`)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to query subcategories by expense count: %v", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var subcategories []SubcategoryExpenseCount

	for rows.Next() {
		var subcategory SubcategoryExpenseCount
		if err := rows.Scan(
			&subcategory.SubcategoryID,
			&subcategory.SubcategoryName,
			&subcategory.CategoryID,
			&subcategory.CategoryName,
			&subcategory.ExpenseCount,
		); err != nil {
			logger.Error(fmt.Sprintf("Failed to scan subcategory row: %v", err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		subcategories = append(subcategories, subcategory)
	}

	if err = rows.Err(); err != nil {
		logger.Error(fmt.Sprintf("Error iterating over rows: %v", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if subcategories == nil {
		subcategories = []SubcategoryExpenseCount{}
	}
	json.NewEncoder(w).Encode(subcategories)
}

func deleteExpenseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/expenses/")
	if path == "" {
		http.Error(w, "Expense ID is required", http.StatusBadRequest)
		return
	}

	expenseID, err := strconv.Atoi(path)
	if err != nil {
		http.Error(w, "Invalid expense ID", http.StatusBadRequest)
		return
	}

	var existingID int
	err = db.QueryRow("SELECT id FROM expenses WHERE id = ?", expenseID).Scan(&existingID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Expense not found", http.StatusNotFound)
			return
		}
		logger.Error(fmt.Sprintf("Failed to check expense existence: %v", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	result, err := db.Exec("DELETE FROM expenses WHERE id = ?", expenseID)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to delete expense: %v", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to get rows affected: %v", err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Expense not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Expense deleted successfully",
		"id":      expenseID,
	})
}
