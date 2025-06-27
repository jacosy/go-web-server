package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jacosy/go-web-server/internal/database"
)

type Chirp struct {
	db *database.Queries
}

func NewChirpHandler(db *database.Queries) *Chirp {
	return &Chirp{db: db}
}

var profaneWords = map[string]struct{}{
	"kerfuffle": {},
	"sharbert":  {},
	"fornax":    {},
}

func (c *Chirp) CreateChirp(w http.ResponseWriter, r *http.Request) {
	req := &ChirptRequestModel{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Body) > 140 {
		http.Error(w, "Chirp body exceeds 140 characters", http.StatusBadRequest)
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	cleanedBody := getCleanedBody(req.Body)
	chirp, dbErr := c.db.CreateChirp(r.Context(), database.CreateChirpParams{
		UserID: userID,
		Body:   cleanedBody,
	})
	if dbErr != nil {
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
