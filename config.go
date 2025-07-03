package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/jacosy/go-web-server/internal/auth"
	"github.com/jacosy/go-web-server/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	env            string
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
	if c.env != "dev" {
		http.Error(w, "This endpoint is only available in development mode", http.StatusForbidden)
		return
	}

	err := c.db.Reset(r.Context())
	if err != nil {
		http.Error(w, "Failed to reset user data", http.StatusInternalServerError)
		return
	}

	c.fileserverHits.Store(0)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Users and Metrics are reset successfully!"))
}

func (c *apiConfig) CreateUser(w http.ResponseWriter, r *http.Request) {
	var body map[string]string
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if body["username"] == "" || body["email"] == "" || body["password"] == "" {
		http.Error(w, "Invalid request body: username, email, and password are required", http.StatusBadRequest)
		return
	}

	hashedPwd, err := auth.HashPassword(body["password"])
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	user, err := c.db.CreateUser(r.Context(), database.CreateUserParams{
		Username:       body["username"],
		Email:          body["email"],
		HashedPassword: hashedPwd,
	})
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	respUser := User{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Time,
		UpdatedAt: user.UpdatedAt.Time,
	}
	data, err := json.Marshal(respUser)
	if err != nil {
		http.Error(w, "Failed to return user data", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Write(data)
}
