package handler

import (
	"encoding/json"
	"net/http"
	"strings"
)

type Chirp struct{}

type request struct {
	Body string `json:"body"`
}

type response struct {
	CleanedBody string `json:"cleaned_body"`
	Valid       bool   `json:"valid"`
}

var profaneWords = map[string]struct{}{
	"kerfuffle": {},
	"sharbert":  {},
	"fornax":    {},
}

func (c *Chirp) Validate(w http.ResponseWriter, r *http.Request) {
	req := &request{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Body) > 140 {
		http.Error(w, "Chirp body exceeds 140 characters", http.StatusBadRequest)
		return
	}

	data, err := json.Marshal(response{Valid: true, CleanedBody: getCleanedBody(req.Body)})
	if err != nil {
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
