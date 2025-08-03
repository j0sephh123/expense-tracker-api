package main

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
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

type User struct {
	ID          int     `json:"id"`
	UID         *string `json:"uid"`
	Email       string  `json:"email"`
	DisplayName *string `json:"display_name"`
	CreatedAt   string  `json:"created_at"`
	Password    *string `json:"-"` // Exclude from JSON responses
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	User  User   `json:"user"`
	Token string `json:"token"`
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

func getUserByEmail(email string) (*User, error) {
	var user User
	var password sql.NullString
	var uid sql.NullString
	var displayName sql.NullString

	query := `SELECT id, uid, email, display_name, created_at, password FROM users WHERE email = ?`
	err := db.QueryRow(query, email).Scan(&user.ID, &uid, &user.Email, &displayName, &user.CreatedAt, &password)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query user: %v", err)
	}

	if uid.Valid {
		user.UID = &uid.String
	}
	if displayName.Valid {
		user.DisplayName = &displayName.String
	}
	if password.Valid {
		user.Password = &password.String
	}

	return &user, nil
}

func verifyPassword(plainPassword, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
}
