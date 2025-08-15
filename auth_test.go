package main

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nfongster/chirpy/internal/auth"
)

func TestHashPassword(t *testing.T) {
	password := "password"
	hashedPassword, err := auth.HashPassword(password)
	if hashedPassword == "" || err != nil {
		t.Errorf("Failed to hash password \"%s\"", password)
	}
}

func TestCheckPasswordHash(t *testing.T) {
	password := "password"
	hashedPassword, _ := auth.HashPassword(password)

	if err := auth.CheckPasswordHash(password, hashedPassword); err != nil {
		t.Errorf("Failed to check password hash.  Error: %v", err)
	}
}

func TestCheckPasswordHashWrongPassword(t *testing.T) {
	password := "password"
	hashedPassword, _ := auth.HashPassword(password)

	if err := auth.CheckPasswordHash("password2", hashedPassword); err == nil {
		t.Errorf("Expected an error when checking a password hash against an incorrect source password.")
	}
}

func TestMakeAndValidateJWT(t *testing.T) {
	id := uuid.New()
	secret := "my_secret"
	ss, err := auth.MakeJWT(id, secret, 10*time.Second)

	if err != nil {
		t.Errorf("making JWT returned err: %v", err)
	}

	validatedId, err := auth.ValidateJWT(ss, secret)

	if err != nil {
		t.Errorf("validating JWT returned err: %v", err)
	}
	if validatedId != id {
		t.Errorf("JWT was validated, but id did not match original id")
	}
}
