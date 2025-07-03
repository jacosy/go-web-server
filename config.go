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
	var createUserRequest UserRequest
	if err := json.NewDecoder(r.Body).Decode(&createUserRequest); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if createUserRequest.Username == "" || createUserRequest.Password == "" || createUserRequest.Email == "" {
		http.Error(w, "Invalid request body: username, email, and password are required", http.StatusBadRequest)
		return
	}

	hashedPwd, err := auth.HashPassword(createUserRequest.Password)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	user, err := c.db.CreateUser(r.Context(), database.CreateUserParams{
		Username:       createUserRequest.Username,
		Email:          createUserRequest.Email,
		HashedPassword: hashedPwd,
	})
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	respUser := UserResponse{
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

func (c *apiConfig) LoginUser(w http.ResponseWriter, r *http.Request) {
	var loginRequest LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginRequest); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if loginRequest.Email == "" || loginRequest.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	user, err := c.db.GetUserByEmail(r.Context(), loginRequest.Email)
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	if err = auth.CheckPasswordHash(loginRequest.Password, user.HashedPassword); err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	res, err := json.Marshal(UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Time,
		UpdatedAt: user.UpdatedAt.Time,
	})
	if err != nil {
		http.Error(w, "Failed to return user data", http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", "application/json")
	w.Write(res)
}
