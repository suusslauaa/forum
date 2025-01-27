package main

import (
	"forum/project"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3" // Импорт SQLite драйвера
)

func main() {
	db, err := project.InitDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// clearTable(db)

	project.CreateCategory(db, "General")
	project.CreateCategory(db, "Technology")

	err = project.CreatePost(db, "First Post", "This is the content of the first post.", 1)
	if err != nil {
		log.Fatal(err)
	}

	err = project.CreatePost(db, "Second Post", "This is the content of the second post.", 2)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/register", project.RegisterHandler)
	http.HandleFunc("/login", project.LoginHandler)
	http.HandleFunc("/account", project.ProfileHandler)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	http.HandleFunc("/", project.HomeHandler)
	http.ListenAndServe(":8080", nil)
}
