package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jacosy/go-web-server/internal/auth"
	"github.com/jacosy/go-web-server/internal/database"
)

type Chirp struct {
	db        *database.Queries
	secretKey string
}

func NewChirpHandler(db *database.Queries, secretKey string) *Chirp {
	return &Chirp{db: db, secretKey: secretKey}
}

var profaneWords = map[string]struct{}{
	"kerfuffle": {},
	"sharbert":  {},
	"fornax":    {},
}

func (c *Chirp) CreateChirp(w http.ResponseWriter, r *http.Request) {
	jwtToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	userID, err := auth.ValidateJWT(jwtToken, c.secretKey)
	if err != nil {
		http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	req := &ChirptRequestModel{}
	decoder := json.NewDecoder(r.Body)
	if err = decoder.Decode(req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Body) > 140 {
		http.Error(w, "Chirp body exceeds 140 characters", http.StatusBadRequest)
		return
	}

	cleanedBody := getCleanedBody(req.Body)
	chirp, dbErr := c.db.CreateChirp(r.Context(), database.CreateChirpParams{
		UserID: userID,
		Body:   cleanedBody,
	})
	if dbErr != nil {
		log.Println("Error creating chirp:", dbErr)
		http.Error(w, "Failed to create chirp", http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(convertChirpToResponseModel(chirp))
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(data)
}

func (c *Chirp) GetChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := c.db.GetAllChirps(r.Context())
	if err != nil {
		http.Error(w, "Failed to retrieve chirps", http.StatusInternalServerError)
		return
	}

	var chirpResponses []ChirpResponseModel
	for _, chirp := range chirps {
		chirpResponses = append(chirpResponses, convertChirpToResponseModel(chirp))
	}

	data, err := json.Marshal(chirpResponses)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (c *Chirp) GetChirpByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id") // get the string value of the path parameter
	if id == "" {
		http.Error(w, "Chirp ID is required", http.StatusBadRequest)
		return
	}

	chirpID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "Invalid chirp ID", http.StatusBadRequest)
		return
	}

	chirp, err := c.db.GetChirpByID(r.Context(), chirpID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Chirp not found", http.StatusNotFound)
			return
		}

		log.Println("Error retrieving chirp:", err)
		http.Error(w, "Failed to retrieve chirp", http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(convertChirpToResponseModel(chirp))
	if err != nil {
		log.Println("failed to JSON.Marshal a chirp instance:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func getCleanedBody(body string) string {
	words := strings.Split(body, " ")
	for i, str := range words {
		if _, exists := profaneWords[strings.ToLower(str)]; exists {
			words[i] = "****"
		}
	}
	return strings.Join(words, " ")
}

func convertChirpToResponseModel(chirp database.Chirp) ChirpResponseModel {
	return ChirpResponseModel{
		ID:        chirp.ID,
		UserID:    chirp.UserID,
		Body:      chirp.Body,
		CreatedAt: chirp.CreatedAt.Time,
		UpdatedAt: chirp.UpdatedAt.Time,
	}
}
