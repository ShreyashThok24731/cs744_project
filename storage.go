package main

import (
	"database/sql"
)

type DBStore struct {
	db *sql.DB
}

func NewDBStore(db *sql.DB) *DBStore {
	return &DBStore{db: db}
}

func (s *DBStore) InitSchema() error {
	query := `
    CREATE TABLE IF NOT EXISTS kv (
        key TEXT PRIMARY KEY,
        value TEXT
    );
    `
	_, err := s.db.Exec(query)
	return err
}

func (s *DBStore) Get(key string) (string, error) {
	var value string
	err := s.db.QueryRow("SELECT value FROM kv WHERE key = $1", key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", err 
	} else if err != nil {
		return "", err 
	}
	return value, nil
}

func (s *DBStore) Put(key, value string) error {
	query := `
    INSERT INTO kv (key, value) VALUES ($1, $2)
    ON CONFLICT (key) DO UPDATE SET value = $2;
    `
	_, err := s.db.Exec(query, key, value)
	return err
}

func (s *DBStore) Delete(key string) error {
	_, err := s.db.Exec("DELETE FROM kv WHERE key = $1", key)
	return err
}
