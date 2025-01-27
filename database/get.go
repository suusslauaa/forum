package database

import "database/sql"

type Post struct {
	ID       int
	Title    string
	Content  string
	Category string
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

type Category struct {
	ID   int
	Name string
}
