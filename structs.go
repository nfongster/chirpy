package main

import (
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/nfongster/chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

// JSON PACKETS SENT BY CLIENT

type validateChirpParameters struct {
	Body string `json:"body"`
}

type createUserParameters struct {
	Email string `json:"email"`
}
