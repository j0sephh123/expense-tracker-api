package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt/v5"
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
	Role        string  `json:"role"`
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
	user := os.Getenv("MYSQL_USER")
	password := os.Getenv("MYSQL_PASSWORD")
	host := os.Getenv("MYSQL_HOST")
	port := os.Getenv("MYSQL_PORT")
	database := os.Getenv("MYSQL_DATABASE")
	
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", user, password, host, port, database)
	
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

	query := `SELECT id, uid, email, display_name, created_at, password, role FROM users WHERE email = ?`
	err := db.QueryRow(query, email).Scan(&user.ID, &uid, &user.Email, &displayName, &user.CreatedAt, &password, &user.Role)

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

func generateToken(userID int, email string, role string) (string, error) {
	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		secretKey = "your-secret-key-change-in-production"
	}

	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"role":    role,
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(), // 7 days
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

func validateToken(tokenString string) (jwt.MapClaims, error) {
	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		secretKey = "your-secret-key-change-in-production"
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}
