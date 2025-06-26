package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/jacosy/go-web-server/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
}

func (c *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (c *apiConfig) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, c.fileserverHits.Load())
}

func (c *apiConfig) ResetMetricsHandler(w http.ResponseWriter, r *http.Request) {
	c.fileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Metrics reset"))
}

func (c *apiConfig) CreateUser(w http.ResponseWriter, r *http.Request) {
	var body map[string]string
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := c.db.CreateUser(r.Context(), database.CreateUserParams{
		Username: body["username"],
		Email:    body["email"],
	})
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	data, err := json.Marshal(user)
	if err != nil {
		http.Error(w, "Failed to return user data", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write(data)
}
