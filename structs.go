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

type chirpError struct {
	Error string `json:"error"`
}

// JSON PACKETS SENT BY SERVER

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserId    uuid.UUID `json:"user_id"`
}

// JSON PACKETS SENT BY CLIENT

type validateChirpParameters struct {
	Body string `json:"body"`
}

type createUserParameters struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

type createChirpParameters struct {
	Body   string    `json:"body"`
	UserId uuid.UUID `json:"user_id"`
}
