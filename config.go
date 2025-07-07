package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/jacosy/go-web-server/internal/auth"
	"github.com/jacosy/go-web-server/internal/database"
	"github.com/jacosy/go-web-server/internal/utils"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	env            string
	secretKey      string
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

	if createUserRequest.Password == "" || createUserRequest.Email == "" {
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

	utils.ResponseWithJSON(w, http.StatusCreated, UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Time,
		UpdatedAt: user.UpdatedAt.Time,
	})
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

	defaultExpiresIn := 1 * time.Hour
	reqExpiresIn := time.Duration(loginRequest.ExpiresInSeconds) * time.Second
	if reqExpiresIn <= 0 || reqExpiresIn > defaultExpiresIn {
		reqExpiresIn = defaultExpiresIn
	}

	jwtToken, err := auth.MakeJWT(user.ID, c.secretKey, reqExpiresIn)
	if err != nil {
		http.Error(w, "Failed to create JWT token", http.StatusInternalServerError)
	}

	utils.ResponseWithJSON(w, http.StatusOK, UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Time,
		UpdatedAt: user.UpdatedAt.Time,
		Token:     jwtToken,
	})
}
