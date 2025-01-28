package database

import (
	"database/sql"
	"errors"
	"log"
)

type Post struct {
	ID           int
	Title        string
	Content      string
	AuthorID     int    // ID автора
	Author       string // Имя автора
	CategoryID   int
	Category     string
	LikeCount    int
	DislikeCount int
}

func GetPostByID(db *sql.DB, postID int) (Post, error) {
	var post Post
	query := `
		SELECT 
			p.id, 
			p.title, 
			p.content, 
			u.username AS author, 
			p.category_id,                -- Добавлено для получения ID категории
			IFNULL(c.name, '') AS category, 
			p.liked AS like_count, 
			p.disliked AS dislike_count, 
			p.author_id                   -- Получаем ID автора
		FROM posts p
		LEFT JOIN categories c ON p.category_id = c.id
		LEFT JOIN users u ON p.author_id = u.id 
		WHERE p.id = ?`

	// Выполняем запрос
	row := db.QueryRow(query, postID)

	// Проверяем ошибку при сканировании результатов
	err := row.Scan(
		&post.ID, &post.Title, &post.Content, &post.Author,
		&post.CategoryID, &post.Category,
		&post.LikeCount, &post.DislikeCount, &post.AuthorID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return post, errors.New("post not found")
		}
		log.Printf("Error scanning row: %v", err)
		return post, err
	}

	return post, nil
}

func GetPosts(db *sql.DB, categoryID int) ([]Post, error) {
	var rows *sql.Rows
	var err error

	if categoryID > 0 {
		rows, err = db.Query(`
			SELECT p.id, p.title, p.content, c.name 
			FROM posts p
			LEFT JOIN categories c ON p.category_id = c.id
			WHERE p.category_id = ?`, categoryID)
	} else {
		rows, err = db.Query(`
			SELECT p.id, p.title, p.content, c.name 
			FROM posts p
			LEFT JOIN categories c ON p.category_id = c.id`)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		var categoryName string
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &categoryName)
		if err != nil {
			return nil, err
		}
		post.Category = categoryName
		posts = append(posts, post)
	}

	return posts, nil
}

func GetPostsByUserID(db *sql.DB, userID int) ([]Post, error) {
	var posts []Post
	query := `
        SELECT p.id, p.title, p.content, IFNULL(c.name, '') AS category, p.liked AS like_count
        FROM posts p
        LEFT JOIN categories c ON p.category_id = c.id
        WHERE p.author_id = ?`

	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var post Post
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.Category, &post.LikeCount)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}
func GetPostsByUserIDAndCategory(db *sql.DB, userID int, categoryID string) ([]Post, error) {
	query := `
		SELECT p.id, p.title, p.content, c.name AS category 
		FROM posts p
		JOIN categories c ON p.category_id = c.id 
		WHERE p.author_id = ? AND p.category_id = ?`
	rows, err := db.Query(query, userID, categoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.Category); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	return posts, nil
}

func GetCategories(db *sql.DB) ([]Category, error) {
	rows, err := db.Query(`SELECT id, name FROM categories`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var category Category
		err := rows.Scan(&category.ID, &category.Name)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, nil
}

type Category struct {
	ID   int
	Name string
}
