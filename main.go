package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq" 
)

func main() {
	connStr := os.Getenv("DB_CONN_STRING")
	if connStr == "" {
		log.Println("DB_CONN_STRING environment variable not set. Using default.")
		connStr = "postgres://postgres:thok@localhost:5433/postgres?sslmode=disable"
	}
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to open database connection:", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}
	log.Println("Successfully connected to PostgreSQL database.")
	storage := NewDBStore(db)
	if err := storage.InitSchema(); err != nil {
		log.Fatal("Failed to initialize database schema:", err)
	}
	log.Println("Database schema initialized.")
	cache := NewKVCache(100)
	log.Println("In-memory LRU cache initialized.")
	server := NewServer(storage, cache)
	log.Println("HTTP server initialized.")
	http.HandleFunc("/kv/", server.kvHandler)
	port := ":8080"
	log.Printf("Server starting on port %s...", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
