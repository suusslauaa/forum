package database

import (
	"database/sql"
	"errors"
	"log"
)

type Comment struct {
	ID           int
	PostID       int
	UserID       int
	Author       string
	Content      string
	CreatedAt    string
	LikeCount    int
	DislikeCount int
}

type Post struct {
	ID           int
	Title        string
	Content      string
	AuthorID     int
	Author       string
	CategoryID   int
	Category     string
	LikeCount    int
	DislikeCount int
	ImagePath    string
	Comments     []Comment // Добавляем список комментариев
	CreatedAt    string
}

func GetPostByID(db *sql.DB, postID int) (Post, error) {
	var post Post
	// Получаем пост
	query := `
		SELECT 
			p.id, 
			p.title, 
			p.content, 
			u.username AS author, 
			p.category_id, 
			IFNULL(c.name, '') AS category, 
			p.liked AS like_count, 
			p.disliked AS dislike_count, 
			IFNULL(p.image_path, '') AS image,
			p.author_id 
		FROM posts p
		LEFT JOIN categories c ON p.category_id = c.id
		LEFT JOIN users u ON p.author_id = u.id 
		WHERE p.id = ?`

	// Выполняем запрос
	row := db.QueryRow(query, postID)

	// Проверяем ошибку при сканировании
	err := row.Scan(&post.ID, &post.Title, &post.Content, &post.Author, &post.CategoryID, &post.Category, &post.LikeCount, &post.DislikeCount, &post.ImagePath, &post.AuthorID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return post, errors.New("post not found")
		}
		log.Printf("Error scanning row: %v", err)
		return post, err
	}

	// Получаем комментарии для поста
	commentsQuery := `
		SELECT 
			c.id, 
			c.post_id, 
			c.user_id, 
			c.content, 
			c.created_at, 
			c.liked AS like_count, 
			c.disliked AS dislike_count, 
			u.username AS author
		FROM comments c
		LEFT JOIN users u ON c.user_id = u.id
		WHERE c.post_id = ?`

	rows, err := db.Query(commentsQuery, postID)
	if err != nil {
		log.Printf("Error querying comments: %v", err)
		return post, err
	}
	defer rows.Close()

	// Заполняем список комментариев
	for rows.Next() {
		var comment Comment
		err := rows.Scan(&comment.ID, &comment.PostID, &comment.UserID, &comment.Content, &comment.CreatedAt, &comment.LikeCount, &comment.DislikeCount, &comment.Author)
		if err != nil {
			log.Printf("Error scanning comment: %v", err)
			return post, err
		}
		post.Comments = append(post.Comments, comment)
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

func GetLikedPostsByUserID(db *sql.DB, userID int) ([]Post, error) {
	var posts []Post

	query := `
		SELECT 
			p.id, p.title, p.content, IFNULL(c.name, '') AS category, 
			u.username AS author
		FROM posts p
		LEFT JOIN categories c ON p.category_id = c.id
		LEFT JOIN users u ON p.author_id = u.id
		INNER JOIN post_reactions pr ON p.id = pr.post_id
		WHERE pr.user_id = ? AND pr.reaction_type = 'like'
	`

	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var post Post
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.Category, &post.Author)
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

func GetPostIDByCommentID(db *sql.DB, commentID int) (postID int, err error) {
	query := `
		SELECT 
			p.post_id
		FROM comments p
		WHERE p.id = ?`

	// Выполняем запрос
	row := db.QueryRow(query, commentID)

	// Проверяем ошибку при сканировании
	err = row.Scan(&postID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return postID, errors.New("post not found")
		}
		log.Printf("Error scanning row: %v", err)
		return postID, err
	}

	return
}

func GetUserIDByCommentID(db *sql.DB, commentID int) (userID int, err error) {
	query := `
	SELECT 
		c.user_id
	FROM comments c
	WHERE c.id = ?`

	// Выполняем запрос
	row := db.QueryRow(query, commentID)

	// Проверяем ошибку при сканировании
	err = row.Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return userID, errors.New("post not found")
		}
		log.Printf("Error scanning row: %v", err)
		return userID, err
	}

	return
}
