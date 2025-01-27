package project

import "database/sql"

type Post struct {
	ID       int
	Title    string
	Content  string
	Category string
}

type Category struct {
	ID   int
	Name string
}

func ClearTable(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM posts")
	return err
}

func CreateCategory(db *sql.DB, name string) error {
	_, err := db.Exec(`INSERT INTO categories (name) VALUES (?)`, name)
	return err
}

func DeleteCategory(db *sql.DB, id int) error {
	_, err := db.Exec(`DELETE FROM categories WHERE id = ?`, id)
	return err
}

func CreatePost(db *sql.DB, title string, content string, categoryId int) error {
	// Проверка на существование поста с таким же заголовком и содержанием
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM posts WHERE title = ? AND content = ?)`
	err := db.QueryRow(query, title, content).Scan(&exists)
	if err != nil {
		return err
	}

	if exists {
		return nil // Если пост уже существует, ничего не делаем
	}

	// Вставка нового поста
	insertPostSQL := `INSERT INTO posts (title, content, category_id) VALUES (?, ?, ?)`
	_, err = db.Exec(insertPostSQL, title, content, categoryId)
	return err
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
