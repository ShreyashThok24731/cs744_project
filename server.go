package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

type Server struct {
	store *DBStore
	cache *KVCache
}

func NewServer(store *DBStore, cache *KVCache) *Server {
	return &Server{
		store: store,
		cache: cache,
	}
}

func (s *Server) kvHandler(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimPrefix(r.URL.Path, "/kv/")
	if key == "" && r.Method != "POST" {
		http.Error(w, "Key is required", http.StatusBadRequest)
		return
	}
	switch r.Method {
	case "GET":
		s.handleRead(w, r, key)
	case "POST":
		s.handleCreate(w, r)
	case "DELETE":
		s.handleDelete(w, r, key)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

type CreateRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (s *Server) handleCreate(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Key == "" || req.Value == "" {
		http.Error(w, "Key and value must not be empty", http.StatusBadRequest)
		return
	}
	if err := s.store.Put(req.Key, req.Value); err != nil {
		log.Printf("Failed to write to DB: %v\n", err)
		http.Error(w, "Failed to store key-value pair", http.StatusInternalServerError)
		return
	}
	s.cache.Put(req.Key, req.Value)
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Successfully created key: %s\n", req.Key)
	log.Printf("HANDLED CREATE: Key=%s\n", req.Key)
}

func (s *Server) handleRead(w http.ResponseWriter, r *http.Request, key string) {
	value, found := s.cache.Get(key)
	if found {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, value)
		log.Printf("HANDLED READ (Cache Hit): Key=%s\n", key)
		return
	}
	value, err := s.store.Get(key)
	if err != nil {
		log.Printf("Failed to get from DB: %v\n", err)
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	}
	s.cache.Put(key, value)
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, value)
	log.Printf("HANDLED READ (Cache Miss): Key=%s\n", key)
}

func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request, key string) {
	if err := s.store.Delete(key); err != nil {
		log.Printf("Failed to delete from DB: %v\n", err)
		http.Error(w, "Failed to delete key", http.StatusInternalServerError)
		return
	}
	s.cache.Delete(key)
	w.WriteHeader(http.StatusNoContent)
	log.Printf("HANDLED DELETE: Key=%s\n", key)
}
