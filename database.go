package main

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

type Subcategory struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Category struct {
	ID            int           `json:"id"`
	Name          string        `json:"name"`
	Subcategories []Subcategory `json:"subcategories"`
}

type Expense struct {
	ID              int     `json:"id"`
	Amount          float64 `json:"amount"`
	SubcategoryID   int     `json:"subcategory_id"`
	UserID          *int    `json:"user_id"`
	Note            *string `json:"note"`
	CreatedAt       string  `json:"created_at"`
	UserEmail       *string `json:"user_email"`
	SubcategoryName *string `json:"subcategory_name"`
	CategoryID      *int    `json:"category_id"`
	CategoryName    *string `json:"category_name"`
}

type GroupedExpense struct {
	GroupName string    `json:"group_name"`
	Total     float64   `json:"total"`
	Count     int       `json:"count"`
	Expenses  []Expense `json:"expenses"`
}

var db *sql.DB

func initDB() error {
	var err error
	dsn := "root:your_password@tcp(localhost:3306)/expenses?parseTime=true"
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	if err = db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	logger.Info("Database connection established successfully")
	return nil
}
