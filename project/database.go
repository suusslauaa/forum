package project

import (
	"database/sql"
	"log"
)

func InitDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./forum.db")
	if err != nil {
		return nil, err
	}

	createTable := `
    CREATE TABLE IF NOT EXISTS users (
        id TEXT PRIMARY KEY,
        username TEXT NOT NULL,
        email TEXT NOT NULL UNIQUE,
        password TEXT NOT NULL
    );
    `
	_, err = db.Exec(createTable)
	if err != nil {
		log.Fatal(err)
	}

	// Создание таблицы категорий
	createCategoryTableSQL := `CREATE TABLE IF NOT EXISTS categories (
		"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"name" TEXT NOT NULL UNIQUE
	);`
	_, err = db.Exec(createCategoryTableSQL)
	if err != nil {
		return nil, err
	}

	// Создание новой таблицы постов с category_id
	createPostTableSQL := `CREATE TABLE IF NOT EXISTS posts_new (
		"id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"title" TEXT,
		"content" TEXT,
		"category_id" INTEGER,
		FOREIGN KEY (category_id) REFERENCES categories(id)
	);`
	_, err = db.Exec(createPostTableSQL)
	if err != nil {
		return nil, err
	}

	// Перенос данных из старой таблицы в новую (если таблица posts уже существует)
	_, err = db.Exec(`
		INSERT INTO posts_new (id, title, content, category_id)
		SELECT id, title, content, category_id FROM posts
	`)
	if err != nil {
		return nil, err
	}

	// Удаление старой таблицы
	_, err = db.Exec(`DROP TABLE IF EXISTS posts`)
	if err != nil {
		return nil, err
	}

	// Переименование новой таблицы в posts
	_, err = db.Exec(`ALTER TABLE posts_new RENAME TO posts`)
	if err != nil {
		return nil, err
	}

	return db, nil
}
