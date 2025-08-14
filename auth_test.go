package main

import (
	"testing"

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
