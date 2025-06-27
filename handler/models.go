package handler

import (
	"time"

	"github.com/google/uuid"
)

type ChirptRequestModel struct {
	UserID string `json:"user_id"`
	Body   string `json:"body"`
}

type ChirpResponseModel struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
