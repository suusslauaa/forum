package database

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"

	"golang.org/x/crypto/bcrypt"
)

// hashPassword is a function that would hash the password before saving it
func hashPassword(password string) (string, error) {
	// For example, using bcrypt:
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("error hashing password: %w", err)
	}
	return string(hashedPassword), nil
}

// CreateCategory создает новую категорию
func CreateCategory(db *sql.DB, name string) error {
	_, err := db.Exec(`INSERT INTO categories (name) VALUES (?)`, name)
	return err
}

// CreatePost создает новый пост, привязанный к категории
func CreatePost(db *sql.DB, title, content string, authorID int, categoryID *int, createdAt, imagePath string) error {
	query := `INSERT INTO posts (title, content, author_id, category_id, created_at, image_path) 
				VALUES ($1, $2, $3, $4, $5, $6) RETURNING id;`

	// Используем подготовленный запрос для безопасности
	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	var postID int
	var imgPath interface{}
	if imagePath == "" {
		imgPath = nil
	} else {
		imgPath = imagePath
	}
	var catID interface{}
	if categoryID == nil {
		catID = nil
	} else {
		catID = *categoryID
	}

	// Выполняем запрос и получаем ID вставленного поста
	err = db.QueryRow(query, title, content, authorID, catID, createdAt, imgPath).Scan(&postID)
	if err != nil {
		return err
	}
	query = `INSERT INTO activities (user_id, activity_type, post_id, created_at) 
	           VALUES ($1, 'create_post', $2, CURRENT_TIMESTAMP)`
	_, err = db.Exec(query, authorID, postID)
	return nil
}

func CreateUser(db *sql.DB, email, username, password, role string) error {
	// Validate the role before inserting
	validRoles := map[string]bool{
		"guest":     true,
		"user":      true,
		"moderator": true,
		"admin":     true,
	}

	if !validRoles[role] {
		return fmt.Errorf("invalid role: %s", role)
	}

	// Hash the password (consider using bcrypt or Argon2)
	hashedPassword, err := hashPassword(password)
	if err != nil {
		return fmt.Errorf("error hashing password: %w", err)
	}

	query := `INSERT INTO users (email, username, password, role) VALUES (?, ?, ?, ?);`

	// Prepare the query
	stmt, err := db.Prepare(query)
	if err != nil {
		return fmt.Errorf("error preparing statement: %w", err)
	}
	defer stmt.Close()

	// Execute the query
	_, err = stmt.Exec(email, username, hashedPassword, role)
	if err != nil {
		return fmt.Errorf("error creating user: %w", err)
	}

	return nil
}

func ReadNotification(db *sql.DB, ID int) {
	fmt.Println(ID)
	_, err := db.Exec("UPDATE notifications SET is_read = 1 WHERE id = ?", ID)
	if err != nil {
		log.Println(err)
	}
}
