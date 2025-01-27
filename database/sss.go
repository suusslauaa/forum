package database

import (
	"database/sql"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// User структура для пользователя
type User struct {
	ID           int
	Email        string
	Username     string
	PasswordHash string
}

// GetUserByEmail получает пользователя из базы данных по email
func GetUserByEmail(db *sql.DB, email string) (*User, error) {
	row := db.QueryRow("SELECT id, email, username, password FROM users WHERE email = ?", email)
	user := &User{}
	err := row.Scan(&user.ID, &user.Email, &user.Username, &user.PasswordHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Пользователь не найден
		}
		return nil, fmt.Errorf("ошибка при получении пользователя: %w", err)
	}
	return user, nil
}

// ComparePassword сравнивает введенный пароль с хешем пароля из базы данных
func ComparePassword(hashedPassword string, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
